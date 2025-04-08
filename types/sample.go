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

import (
	"fmt"
	"math"
)

const (
	ErrInvalidSelection = Error("invalid selection")

	generationsMin = 0.05
)

type Selection string

const (
	SelectionInput  Selection = "input"
	SelectionOutput Selection = "output"
)

// StringToSelection converts a string to a Selection type.
func StringToSelection(s string) (Selection, error) {
	switch Selection(s) {
	case SelectionInput:
		return SelectionInput, nil
	case SelectionOutput:
		return SelectionOutput, nil
	default:
		return "", ErrInvalidSelection
	}
}

type Sample struct {
	MLWHSampleID        string
	RunID               string
	ManualQC            string
	SampleID            string
	Selection           Selection
	ExperimentReplicate int
	SelectionTime       string
	CellDensity         string
	CellDensityFloat    float32
}

// Key returns a unique key for this sample, which is the SampleID and RunID
// concatenated with a period.
func (s *Sample) Key() string {
	return s.SampleID + "." + s.RunID
}

// SampleName is the selection and replicate number, eg. "input1" or "output2".
func (s *Sample) SampleName() string {
	return fmt.Sprintf("%s%d", s.Selection, s.ExperimentReplicate)
}

// SelectionID returns 0 for input and 1 for output.
func (s *Sample) SelectionID() int {
	switch s.Selection {
	case SelectionInput:
		return 0
	case SelectionOutput:
		return 1
	default:
		return 0
	}
}

// SelectionReplicate converts the Selection to a replicate number.
func (s *Sample) SelectionReplicate() string {
	if s.Selection == SelectionOutput {
		return "1"
	}

	return ""
}

// TODO: Pair1, Pair2, proper Generations() calc; probably these are dimsum pkg
// methods during experiment file creation when looking over a slice of samples

// Generations is the amount of times the cells have divided between input and
// output, ie. log2(output cell density / input cell density).
func (s *Sample) Generations() float32 {
	if s.CellDensityFloat == 0 || s.Selection == SelectionInput {
		return 0
	}

	// TODO: This is a bit of a hack, we should be using the input cell density
	// from the corresponding input sample, not generationsMin

	return float32(math.Log2(float64(s.CellDensityFloat / generationsMin)))
}
