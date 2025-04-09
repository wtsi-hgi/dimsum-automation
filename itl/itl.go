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
	"os"
	"path/filepath"

	"github.com/wtsi-hgi/dimsum-automation/types"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrNoStudy             = Error("study not specified")
	ErrMultipleExperiments = Error("samples from multiple experiments provided")
	ErrMissingFastqFile    = Error("one fastq file for sample run already exists, but not the other")

	tsvOutputDir   = "./tsv_output"
	tsvOutputPath  = tsvOutputDir + "/metadata/samples.tsv"
	tsvWorkDir     = "./tsv_work"
	fastqOutputDir = "./fastq_output"
	fastqWorkDir   = "./fastq_work"
	fastqFinalDir  = fastqOutputDir + "/" + fastqOutputSubDir
	minTSVColumns  = 4
	tsvExtension   = ".tsv"

	userPerm = 0700
)

// Sample is a types.Sample with extra methods.
type Sample struct {
	types.Sample
}

func (s *Sample) Key() string {
	return fmt.Sprintf("%s.%s", s.SampleID, s.RunID)
}

func (s *Sample) TSVPath() string {
	return filepath.Join(".", s.Key()+tsvExtension)
}

func (s *Sample) FastqPath(outputDir, pairSuffix string) string {
	return filepath.Join(outputDir, s.Key()+pairSuffix)
}

// ITL lets you use irods_to_lustre to get fastqs for certain samples.
type ITL struct {
	studyID  string
	samples  []*Sample
	fastqDir string
}

// New creates a new ITL for the samples within the given library.
//
// Supply the final output directory for the fastq files you'll create by
// running the GenerateSamplesTSVCommand() command, followed by
// FilterSamplesTSV(), followed by the FastqCreator.Command() commands.
//
// If the output directory already contains the fastq files for a sample that
// chain of operations would create, that input sample will be ignored. If it
// contains some but not all of the fastq files for a sample, an error will be
// returned.
//
// You can use Samples() to get the Samples of the unignored samples we will
// operate on. If none are returned, you won't need to do anything, as all your
// desired fastq files already exist.
func New(lib *types.Library, fastqDir string) (*ITL, error) {
	if lib == nil || lib.StudyID == "" {
		return nil, ErrNoStudy
	}

	samples, err := extractSamples(lib)
	if err != nil {
		return nil, err
	}

	todo, err := todoSamples(samples, fastqDir)
	if err != nil {
		return nil, err
	}

	return &ITL{
		studyID:  lib.StudyID,
		samples:  todo,
		fastqDir: fastqDir,
	}, nil
}

// extractSamples finds all the unique samples in the given Library, validating
// that there's only one experiement.
func extractSamples(lib *types.Library) ([]*Sample, error) {
	if len(lib.Experiments) != 1 {
		return nil, ErrMultipleExperiments
	}

	inputSamples := lib.Experiments[0].Samples

	sampleMap := make(map[string]*Sample, len(inputSamples))
	sampleOrder := make([]string, 0, len(inputSamples))

	for _, input := range inputSamples {
		key := input.Key()
		s := &Sample{
			Sample: types.Sample{
				SampleID: input.SampleID,
				RunID:    input.RunID,
			},
		}

		if _, exists := sampleMap[key]; !exists {
			sampleMap[key] = s

			sampleOrder = append(sampleOrder, key)
		}
	}

	samples := make([]*Sample, len(sampleMap))

	for i, key := range sampleOrder {
		samples[i] = sampleMap[key]
	}

	return samples, nil
}

// todoSamples checks if the fastq files for each sample already exist in the
// fastq directory. It returns a slice of samples that need to be processed, or
// an error if any of the samples have only one fastq file already present.
func todoSamples(inputs []*Sample, fastqDir string) ([]*Sample, error) {
	todo := make([]*Sample, 0, len(inputs))

	for _, input := range inputs {
		found, err := checkFastqFiles(input, fastqDir)
		if err != nil {
			return nil, err
		}

		if found {
			continue
		}

		todo = append(todo, input)
	}

	return todo, nil
}

// checkFastqFiles checks if the fastq files for a sample already exist in the
// fastq directory. If they both do, returns true, or if none do, returns false.
// If only one fastq file exists, it returns an error.
func checkFastqFiles(input *Sample, fastqDir string) (bool, error) {
	pair1 := input.FastqPath(fastqDir, FastqPair1Suffix)
	pair2 := input.FastqPath(fastqDir, FastqPair2Suffix)

	if fileExists(pair1) && fileExists(pair2) {
		return true, nil
	}

	if fileExists(pair1) || fileExists(pair2) {
		return true, ErrMissingFastqFile
	}

	return false, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

// Samples returns the samples run in the ITL.
//
// This is useful for checking which samples will be processed by the
// GenerateSamplesTSVCommand() command, and which ones already exist in the
// fastq directory.
func (i *ITL) Samples() []*Sample {
	return i.samples
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
// returns a slice of FastqCreator.
func (i *ITL) FilterSamplesTSV(inputTSVPath string) ([]FastqCreator, error) {
	fcs := make([]FastqCreator, 0, len(i.samples))

	for _, s := range i.samples {
		tsvPath, err := createPerSampleRunTSV(inputTSVPath, s)
		if err != nil {
			return nil, err
		}

		fcs = append(fcs, FastqCreator{
			sample:   s,
			tsvPath:  tsvPath,
			finalDir: i.fastqDir,
		})
	}

	return fcs, nil
}
