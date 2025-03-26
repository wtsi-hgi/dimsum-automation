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

	"github.com/wtsi-hgi/dimsum-automation/samples"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrNoStudies       = Error("no samples with studies provided")
	ErrMultipleStudies = Error("samples from multiple studies")

	tsvOutputDir      = "./tsv_output"
	tsvOutputPath     = tsvOutputDir + "/metadata/samples.tsv"
	tsvWorkDir        = "./tsv_work"
	fastqOutputDir    = "./fastq_output"
	fastqWorkDir      = "./fastq_work"
	fastqOutputSubDir = "fastq"
	fastqFinalDir     = fastqOutputDir + "/" + fastqOutputSubDir
	minTSVColumns     = 4
	tsvExtension      = ".tsv"

	userPerm = 0700
)

type sampleRun struct {
	sampleID string
	runID    string
}

func (s sampleRun) Key() string {
	return fmt.Sprintf("%s.%s", s.sampleID, s.runID)
}

func (s sampleRun) TSVPath(outputDir string) string {
	return filepath.Join(outputDir, s.Key()+tsvExtension)
}

// ITL lets you use irods_to_lustre to get fastqs for certain samples.
type ITL struct {
	studyID    string
	sampleRuns []sampleRun
}

// New creates a new ITL for the given samples, checking that all samples are
// from the same study.
func New(inputSamples []samples.Sample) (*ITL, error) {
	if len(inputSamples) == 0 {
		return nil, ErrNoStudies
	}

	studyID := inputSamples[0].Sample.StudyID
	if studyID == "" {
		return nil, ErrNoStudies
	}

	sampleRunMap := make(map[string]sampleRun, len(inputSamples))
	sampleRunOrder := make([]string, 0, len(inputSamples))

	for _, sample := range inputSamples {
		if sample.Sample.StudyID != studyID {
			return nil, ErrMultipleStudies
		}

		key := sample.Sample.SampleID + "." + sample.Sample.RunID
		sr := sampleRun{
			sampleID: sample.Sample.SampleID,
			runID:    sample.Sample.RunID,
		}

		if _, exists := sampleRunMap[key]; !exists {
			sampleRunMap[key] = sr

			sampleRunOrder = append(sampleRunOrder, key)
		}
	}

	sampleRuns := make([]sampleRun, 0, len(sampleRunMap))

	for _, key := range sampleRunOrder {
		sampleRuns = append(sampleRuns, sampleRunMap[key])
	}

	return &ITL{
		studyID:    studyID,
		sampleRuns: sampleRuns,
	}, nil
}

// GenerateSamplesTSVCommand returns a command line for irods_to_lustre that
// will generate a TSV file of the sample metadata for our study. It also
// returns the path to that TSV file.
func (i *ITL) GenerateSamplesTSVCommand() (string, string) {
	return fmt.Sprintf(
		"irods_to_lustre --run_mode study_id --input_studies %s "+
			"--samples_to_process -1 --run_imeta_study true --run_iget_study_cram false "+
			"--run_merge_crams false --run_crams_to_fastq false --filter_manual_qc true "+
			"--outdir %s -w %s",
		i.studyID, tsvOutputDir, tsvWorkDir,
	), tsvOutputPath
}

// FilterSamplesTSV creates a TSV file for each sample run in the ITL and
// returns a slice of FastqCreator structs, each containing the path to the TSV
// file and the output directory for that sample run.
func (i *ITL) FilterSamplesTSV(inputTSVPath, outputDir string) ([]FastqCreator, error) {
	fcs := make([]FastqCreator, 0, len(i.sampleRuns))

	for _, sr := range i.sampleRuns {
		tsvPath, err := createPerSampleRunTSV(inputTSVPath, outputDir, sr)
		if err != nil {
			return nil, err
		}

		fcs = append(fcs, FastqCreator{
			sampleRun: sr,
			tsvPath:   tsvPath,
			outputDir: outputDir,
		})
	}

	return fcs, nil
}
