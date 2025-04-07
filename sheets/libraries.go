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

package sheets

const (
	ErrNoSamplesRequested            = Error("no samples requested")
	ErrSamplesNotFound               = Error("samples not found")
	ErrNotAllSamplesInSameExperiment = Error("not all samples in the same experiment")
)

type Library struct {
	LibraryID        string
	WildtypeSequence string
	MaxSubstitutions int
	Experiments      []*Experiment
}

type Libraries []*Library

// Subset returns a new Library containing only the experiment with the desired
// samples inside it. If the given sampleIDs belong to more than one experiment,
// an error is returned. If the sampleIDs are not found, an error is returned.
func (l Libraries) Subset(sampleIDs ...string) (*Library, error) { //nolint:gocognit,gocyclo,funlen
	if len(sampleIDs) == 0 {
		return nil, ErrNoSamplesRequested
	}

	samples := make([]*Sample, 0, len(sampleIDs))

	desired := make(map[string]bool, len(sampleIDs))
	for _, sampleID := range sampleIDs {
		desired[sampleID] = true
	}

	for _, lib := range l {
		for _, exp := range lib.Experiments {
			for _, sample := range exp.Samples {
				if desired[sample.SampleID] {
					samples = append(samples, sample)
				}
			}

			if len(samples) == 0 {
				continue
			}

			if len(samples) != len(sampleIDs) {
				return nil, ErrNotAllSamplesInSameExperiment
			}

			return lib.Clone(exp, samples), nil
		}
	}

	return nil, ErrSamplesNotFound
}

// Clone returns a new Library with the given experiment and samples inside it.
// It otherwise has the same values as the original Library.
func (l *Library) Clone(exp *Experiment, samples []*Sample) *Library {
	newL := *l
	newL.Experiments = []*Experiment{exp.Clone(samples)}

	return &newL
}
