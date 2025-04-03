/*******************************************************************************
 * Copyright (c) 2025 Genome Research Ltd.
 *
 * Authors:
 *	- Sendu Bala <sb10@sanger.ac.uk>
 *
 * Permission is hereby granted, free of charge, to any person obtaining
 * a copy of this software and associated documentation files (the
 * "Software"), to deal in the Software without restriction, including
 * without limitation the rights to use, copy, modify, merge, publish,
 * distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to
 * the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
 * CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
 * TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/

package itl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	FastqPair1Suffix       = "_1.fastq.gz"
	FastqPair2Suffix       = "_2.fastq.gz"
	ErrFastqExistsDiffSize = Error("fastq file already exists with a different size")

	fastqOutputPathSuffix = ".output"
	fastqOutputSubDir     = "fastq"
	dirPerm               = 0755
)

// FastqCreator holds the information needed to create fastq files for a sample.
type FastqCreator struct {
	sampleRun sampleRun
	tsvPath   string
	finalDir  string
}

// Command returns a command line for irods_to_lustre that will use our TSV file
// to download crams and convert them to FASTQs.
func (fc *FastqCreator) Command() string {
	outputPath := fc.outputPathPrefix()

	return fmt.Sprintf(
		"irods_to_lustre --run_mode csv_samples_id --input_samples_csv %s "+
			"--samples_to_process -1 --run_imeta_study false --run_iget_study_cram true "+
			"--run_merge_crams true --run_crams_to_fastq true --filter_manual_qc true "+
			"--outdir %s%s -w %s.work",
		fc.tsvPath, outputPath, fastqOutputPathSuffix, outputPath,
	)
}

func (fc *FastqCreator) outputPathPrefix() string {
	return filepath.Join(".", fc.sampleRun.Key())
}

// CopyFastqFiles moves the pair 1 and 2 fastq files created by irods_to_lustre
// to our final fastq directory, renaming them to be based on sampleRun instead
// of just sampleID.
//
// If the destination files already exist and have the same size, nothing is
// done. If they have different sizes, an error is returned.
func (fc *FastqCreator) CopyFastqFiles() error {
	sourceDir := filepath.Join(fc.outputPathPrefix()+fastqOutputPathSuffix, fastqOutputSubDir)

	for _, suffix := range []string{FastqPair1Suffix, FastqPair2Suffix} {
		sourceFile := filepath.Join(sourceDir, fc.sampleRun.sampleID+suffix)
		destFile := fc.sampleRun.FastqPath(fc.finalDir, suffix)

		if err := moveFile(sourceFile, destFile); err != nil {
			return err
		}
	}

	return nil
}

// FastqBasenamePrefix returns the prefix for the fastq files based on the
// sample ID and run ID. Appending the suffixes FastqPair1Suffix and
// FastqPair2Suffix will give the full names of the fastq files.
func FastqBasenamePrefix(sampleID, runID string) string {
	return sampleRun{sampleID: sampleID, runID: runID}.Key()
}

// moveFile moves a file from src to dst. If the destination file already exists
// and has the same size, nothing is done. If it exists with a different size,
// an error is returned. If it doesn't exist, a rename is attempted. If that
// fails, a copy is attempted. If that fails, an error is returned.
func moveFile(src, dst string) error {
	if err := checkExistingFile(src, dst); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), dirPerm); err != nil {
		return err
	}

	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	return copyAndRemove(src, dst)
}

// checkExistingFile checks if destination file exists and compares sizes with
// source.
func checkExistingFile(src, dst string) error {
	dstInfo, err := os.Stat(dst)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.Size() == dstInfo.Size() {
		return nil
	}

	return ErrFastqExistsDiffSize
}

// copyAndRemove copies src to dst and removes src if successful.
func copyAndRemove(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if err = dstFile.Close(); err != nil {
		return err
	}

	return os.Remove(src)
}
