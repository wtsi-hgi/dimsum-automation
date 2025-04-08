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

func TestExperiment(t *testing.T) {
	Convey("Clone lets you copy an Experiment with new Samples", t, func() {
		samples := []*Sample{
			{
				SampleID: "sample1",
			},
			{
				SampleID: "sample2",
			},
		}

		orig := &Experiment{
			ExperimentID: "exp1",
			Samples:      samples,
		}
		cloned := orig.Clone([]*Sample{
			{
				SampleID: "sample3",
			},
		})

		So(cloned.ExperimentID, ShouldEqual, "exp1")
		So(cloned.Samples, ShouldHaveLength, 1)
		So(cloned.Samples[0].SampleID, ShouldEqual, "sample3")

		cloned.ExperimentID = "exp2"

		So(orig.ExperimentID, ShouldEqual, "exp1")
	})

	Convey("You can convert a string to SequenceType", t, func() {
		st, err := StringToSequenceType("noncoding")
		So(err, ShouldBeNil)
		So(st, ShouldEqual, SequenceTypeNC)

		st, err = StringToSequenceType("coding")
		So(err, ShouldBeNil)
		So(st, ShouldEqual, SequenceTypeC)

		st, err = StringToSequenceType("auto")
		So(err, ShouldBeNil)
		So(st, ShouldEqual, SequenceTypeAuto)

		st, err = StringToSequenceType("")
		So(err, ShouldBeNil)
		So(st, ShouldEqual, SequenceTypeAuto)

		_, err = StringToSequenceType("foo")
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrInvalidSequenceType)
	})

	Convey("You can convert a string to MutagenesisType", t, func() {
		mt, err := StringToMutagenesisType("random")
		So(err, ShouldBeNil)
		So(mt, ShouldEqual, MutagenesisTypeRandom)

		mt, err = StringToMutagenesisType("codon")
		So(err, ShouldBeNil)
		So(mt, ShouldEqual, MutagenesisTypeCodon)

		mt, err = StringToMutagenesisType("")
		So(err, ShouldBeNil)
		So(mt, ShouldEqual, MutagenesisTypeRandom)

		_, err = StringToMutagenesisType("foo")
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrInvalidMutagenesisType)
	})
}
