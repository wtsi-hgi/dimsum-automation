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
	"path/filepath"
)

// FastqCreator is a struct that holds the information needed to create fastq
// files for a sample.
type FastqCreator struct {
	sampleRun sampleRun
	tsvPath   string
	outputDir string
}

// Command returns a command line for irods_to_lustre that will use our TSV file
// to download crams and convert them to FASTQs.
func (fc *FastqCreator) Command() string {
	outputPath := filepath.Join(fc.outputDir, fc.sampleRun.Key())

	return fmt.Sprintf(
		"irods_to_lustre --run_mode csv_samples_id --input_samples_csv %s "+
			"--samples_to_process -1 --run_imeta_study false --run_iget_study_cram true "+
			"--run_merge_crams true --run_crams_to_fastq true --filter_manual_qc true "+
			"--outdir %s.output -w %s.work",
		fc.tsvPath, outputPath, outputPath,
	)
}
