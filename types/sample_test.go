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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSample(t *testing.T) {
	Convey("Key combines sample Name and RunID", t, func() {
		s := &Sample{
			SampleID: "sample",
			RunID:    "run",
		}
		So(s.Key(), ShouldEqual, "sample.run")
	})

	Convey("SampleName() combines selection and experiment replicate", t, func() {
		s := &Sample{
			Selection:           SelectionInput,
			ExperimentReplicate: 1,
		}
		So(s.SampleName(), ShouldEqual, "input1")

		s = &Sample{
			Selection:           SelectionOutput,
			ExperimentReplicate: 2,
		}
		So(s.SampleName(), ShouldEqual, "output2")
	})

	Convey("SelectionID() returns 0 for input and 1 for output", t, func() {
		s := &Sample{
			Selection: SelectionInput,
		}
		So(s.SelectionID(), ShouldEqual, 0)

		s = &Sample{
			Selection: SelectionOutput,
		}
		So(s.SelectionID(), ShouldEqual, 1)

		s = &Sample{}
		So(s.SelectionID(), ShouldEqual, 0)
	})

	Convey("SelectionReplicate() converts the Selection to a replicate number", t, func() {
		s := &Sample{
			Selection: SelectionInput,
		}
		So(s.SelectionReplicate(), ShouldEqual, "")

		s = &Sample{
			Selection: SelectionOutput,
		}
		So(s.SelectionReplicate(), ShouldEqual, "1")

		s = &Sample{}
		So(s.SelectionReplicate(), ShouldEqual, "")
	})

	Convey("You can convert strings to Selections", t, func() {
		s, err := StringToSelection("input")
		So(err, ShouldBeNil)
		So(s, ShouldEqual, SelectionInput)

		s, err = StringToSelection("output")
		So(err, ShouldBeNil)
		So(s, ShouldEqual, SelectionOutput)

		s, err = StringToSelection("foo")
		So(err, ShouldEqual, ErrInvalidSelection)
		So(s, ShouldEqual, Selection(""))
	})

	// TODO: Generations() testable here?
}
