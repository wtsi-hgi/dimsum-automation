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
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/samples"
)

func TestITL(t *testing.T) {
	Convey("Given desired samples", t, func() {
		studyID := "study1"
		runID1 := "run1"
		runID2 := "run2"
		sampleID1 := "sample1"
		sampleID2 := "sample2"

		testSamples := []samples.Sample{
			{
				Sample: mlwh.Sample{
					StudyID:  studyID,
					RunID:    runID1,
					SampleID: sampleID1,
				},
			},
			{
				Sample: mlwh.Sample{
					StudyID:  studyID,
					RunID:    runID2,
					SampleID: sampleID1,
				},
			},
			{
				Sample: mlwh.Sample{
					StudyID:  studyID,
					RunID:    runID1,
					SampleID: sampleID2,
				},
			},
		}

		Convey("You can generate irods_to_lustre command lines", func() {
			itl, err := New(testSamples)
			So(err, ShouldBeNil)
			So(itl, ShouldNotBeNil)
			So(itl.studyID, ShouldEqual, studyID)
			So(itl.sampleIDs, ShouldHaveLength, 2)

			sort.Strings(itl.sampleIDs)
			So(itl.sampleIDs, ShouldResemble, []string{sampleID1, sampleID2})

			cmd, tsvPath := itl.GenerateSamplesTSVCommand()
			So(cmd, ShouldEqual,
				fmt.Sprintf(
					"irods_to_lustre --run_mode study_id --input_studies %s "+
						"--samples_to_process -1 --run_imeta_study true --run_iget_study_cram false "+
						"--run_merge_crams false --run_crams_to_fastq false --filter_manual_qc true "+
						"--outdir %s -w %s",
					studyID, tsvOutputDir, tsvWorkDir,
				),
			)
			So(tsvPath, ShouldEqual, tsvOutputPath)

			cmd, outputDir := itl.CreateFastqsCommand(tsvPath)
			So(cmd, ShouldEqual,
				fmt.Sprintf(
					"irods_to_lustre --run_mode csv_samples_id --input_samples_csv %s "+
						"--samples_to_process -1 --run_imeta_study false --run_iget_study_cram true "+
						"--run_merge_crams true --run_crams_to_fastq true --filter_manual_qc true "+
						"--outdir %s -w %s",
					tsvPath, fastqOutputDir, fastqWorkDir,
				),
			)
			So(outputDir, ShouldEqual, fastqFinalDir)
		})

		Convey("You can't make a new ITL with samples from multiple studies, or no studies", func() {
			testSamples[0].Sample.StudyID = "study2"

			itl, err := New(testSamples)
			So(err, ShouldNotBeNil)
			So(itl, ShouldBeNil)

			testSamples[0].Sample.StudyID = ""

			_, err = New([]samples.Sample{testSamples[0]})
			So(err, ShouldNotBeNil)

			_, err = New(nil)
			So(err, ShouldNotBeNil)
		})
	})
}

// TODO: test and implement a tsv filterer to pick out the samples we want
