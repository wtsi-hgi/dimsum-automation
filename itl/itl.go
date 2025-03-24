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

	"github.com/wtsi-hgi/dimsum-automation/samples"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrNoStudies       = Error("no samples with studies provided")
	ErrMultipleStudies = Error("samples from multiple studies")

	tsvOutputDir   = "./tsv_output"
	tsvOutputPath  = tsvOutputDir + "/metadata/samples.tsv"
	tsvWorkDir     = "./tsv_work"
	fastqOutputDir = "./fastq_output"
	fastqWorkDir   = "./fastq_work"
	fastqFinalDir  = fastqOutputDir + "/fastq"
)

// ITL lets you use irods_to_lustre to get fastqs for certain samples.
type ITL struct {
	studyID   string
	sampleIDs []string
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

	sampleIDsMap := make(map[string]struct{}, len(inputSamples))

	for _, sample := range inputSamples {
		if sample.Sample.StudyID != studyID {
			return nil, ErrMultipleStudies
		}

		sampleIDsMap[sample.Sample.SampleID] = struct{}{}
	}

	sampleIDs := make([]string, 0, len(sampleIDsMap))

	for sampleID := range sampleIDsMap {
		sampleIDs = append(sampleIDs, sampleID)
	}

	return &ITL{
		studyID:   studyID,
		sampleIDs: sampleIDs,
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

// CreateFastqsCommand returns a command line for irods_to_lustre that will
// use the given TSV file to download crams and convert them to FASTQs, as
// well as the output directory where the FASTQ files will be placed.
func (i *ITL) CreateFastqsCommand(tsvPath string) (string, string) {
	cmd := fmt.Sprintf(
		"irods_to_lustre --run_mode csv_samples_id --input_samples_csv %s "+
			"--samples_to_process -1 --run_imeta_study false --run_iget_study_cram true "+
			"--run_merge_crams true --run_crams_to_fastq true --filter_manual_qc true "+
			"--outdir %s -w %s",
		tsvPath, fastqOutputDir, fastqWorkDir,
	)

	return cmd, fastqFinalDir
}
