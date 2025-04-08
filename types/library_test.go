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

func TestLibrary(t *testing.T) {
	Convey("Clone lets you copy a Library with a new Experiment and Samples", t, func() {
		exps := []*Experiment{
			{
				ExperimentID: "exp1",
				Samples:      []*Sample{{SampleName: "sample1"}},
			},
			{
				ExperimentID: "exp2",
				Samples:      []*Sample{{SampleName: "sample2"}},
			},
		}

		orig := &Library{
			LibraryID:   "lib1",
			Experiments: exps,
		}
		cloned := orig.Clone(&Experiment{ExperimentID: "exp3"}, []*Sample{{SampleName: "sample3"}})

		So(cloned.LibraryID, ShouldEqual, "lib1")
		So(cloned.Experiments, ShouldHaveLength, 1)
		So(cloned.Experiments[0].ExperimentID, ShouldEqual, "exp3")
		So(cloned.Experiments[0].Samples, ShouldHaveLength, 1)
		So(cloned.Experiments[0].Samples[0].SampleName, ShouldEqual, "sample3")

		cloned.LibraryID = "lib2"

		So(orig.LibraryID, ShouldEqual, "lib1")
	})

	Convey("Given some Libraries", t, func() {
		lib1 := &Library{
			LibraryID: "lib1",
			StudyID:   "study1",
			Experiments: []*Experiment{
				{
					ExperimentID: "exp1",
					Samples: []*Sample{
						{SampleName: "sample1", RunID: "run1"},
						{SampleName: "sample2", RunID: "run1"},
					},
				},
			},
		}

		lib2 := &Library{
			LibraryID: "lib2",
			StudyID:   "study2",
			Experiments: []*Experiment{
				{
					ExperimentID: "exp2",
					Samples: []*Sample{
						{SampleName: "sample3", RunID: "run2"},
						{SampleName: "sample4", RunID: "run2"},
					},
				},
				{
					ExperimentID: "exp3",
					Samples: []*Sample{
						{SampleName: "sample5", RunID: "run3"},
						{SampleName: "sample6", RunID: "run3"},
					},
				},
			},
		}

		libraries := Libraries{lib1, lib2}

		Convey("Subset returns an error if no samples are requested", func() {
			_, err := libraries.Subset([]*Sample{})
			So(err, ShouldEqual, ErrNoSamplesRequested)
		})

		Convey("Subset returns an error if samples lack SampleName or RunID", func() {
			_, err := libraries.Subset([]*Sample{{SampleName: "sample1"}})
			So(err, ShouldEqual, ErrNoSamplesRequested)

			_, err = libraries.Subset([]*Sample{{RunID: "run1"}})
			So(err, ShouldEqual, ErrNoSamplesRequested)
		})

		Convey("Subset returns an error if samples are not found", func() {
			_, err := libraries.Subset([]*Sample{{SampleName: "nonexistent", RunID: "run1"}})
			So(err, ShouldEqual, ErrSamplesNotFound)
		})

		Convey("Subset returns an error if samples are in different experiments", func() {
			_, err := libraries.Subset([]*Sample{
				{SampleName: "sample1", RunID: "run1"},
				{SampleName: "sample3", RunID: "run2"},
			})
			So(err, ShouldEqual, ErrNotAllSamplesInSameExperiment)
		})

		Convey("Subset successfully returns a subset library with found samples", func() {
			result, err := libraries.Subset([]*Sample{
				{SampleName: "sample1", RunID: "run1"},
				{SampleName: "sample2", RunID: "run1"},
			})
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.LibraryID, ShouldEqual, "lib1")
			So(result.StudyID, ShouldEqual, "study1")
			So(result.Experiments, ShouldHaveLength, 1)
			So(result.Experiments[0].ExperimentID, ShouldEqual, "exp1")
			So(result.Experiments[0].Samples, ShouldHaveLength, 2)
		})

		Convey("Subset works with a subset of samples in an experiment", func() {
			result, err := libraries.Subset([]*Sample{
				{SampleName: "sample1", RunID: "run1"},
			})
			So(err, ShouldBeNil)
			So(result.Experiments[0].Samples, ShouldHaveLength, 1)
			So(result.Experiments[0].Samples[0].SampleName, ShouldEqual, "sample1")
		})

		Convey("Subset correctly handles the Sample.Key method for matching", func() {
			samples := []*Sample{
				{SampleName: "sample5", RunID: "run3"},
			}
			result, err := libraries.Subset(samples)
			So(err, ShouldBeNil)
			So(result.LibraryID, ShouldEqual, "lib2")
			So(result.Experiments[0].ExperimentID, ShouldEqual, "exp3")
		})
	})
}
