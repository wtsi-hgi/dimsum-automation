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

package types

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrNoSamplesRequested            = Error("no samples requested")
	ErrSamplesNotFound               = Error("samples not found")
	ErrNotAllSamplesInSameExperiment = Error("not all samples in the same experiment")
)

type Library struct {
	StudyID          string
	StudyName        string
	LibraryID        string
	WildtypeSequence string
	MaxSubstitutions int
	Experiments      []*Experiment
}

type Libraries []*Library

// Subset returns a new Library containing only the experiment with the desired
// samples inside it. If the given samples belong to more than one experiment,
// an error is returned. If the samples are not found, an error is returned. The
// given samples must have at least MLWHSampleName and RunID set, or they will
// be ignored.
func (l Libraries) Subset(desired []*Sample) (*Library, error) {
	valid, err := getValidSamples(desired)
	if err != nil {
		return nil, err
	}

	return l.findMatchingLibrary(valid)
}

// getValidSamples extracts valid samples from input and returns a map of their
// keys.
func getValidSamples(desired []*Sample) (map[string]bool, error) {
	valid := make(map[string]bool, len(desired))

	for _, s := range desired {
		if s.SampleID == "" || s.RunID == "" {
			continue
		}

		valid[s.Key()] = true
	}

	if len(valid) == 0 {
		return nil, ErrNoSamplesRequested
	}

	return valid, nil
}

// findMatchingLibrary searches for a library that contains all the desired
// samples in one experiment.
func (l Libraries) findMatchingLibrary(desired map[string]bool) (*Library, error) {
	for _, lib := range l {
		for _, exp := range lib.Experiments {
			samples := findDesiredSamplesInExperiment(exp, desired)

			if len(samples) == 0 {
				continue
			}

			if len(samples) != len(desired) {
				return nil, ErrNotAllSamplesInSameExperiment
			}

			return lib.Clone(exp, samples), nil
		}
	}

	return nil, ErrSamplesNotFound
}

// findDesiredSamplesInExperiment collects all samples in the experiment that
// match the desired keys.
func findDesiredSamplesInExperiment(exp *Experiment, desired map[string]bool) []*Sample {
	var samples []*Sample

	for _, sample := range exp.Samples {
		if desired[sample.Key()] {
			samples = append(samples, sample)
		}
	}

	return samples
}

// Clone returns a new Library with the given experiment and samples inside it.
// It otherwise has the same values as the original Library.
func (l *Library) Clone(exp *Experiment, samples []*Sample) *Library {
	newL := *l
	newL.Experiments = []*Experiment{exp.Clone(samples)}

	return &newL
}
