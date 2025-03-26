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
		sampleName1 := "sample1"
		sampleName2 := "sample2"

		testSamples := []samples.Sample{
			{
				Sample: mlwh.Sample{
					StudyID:    studyID,
					RunID:      runID1,
					SampleName: sampleName1,
					SampleID:   sampleName1 + "_id",
				},
			},
			{
				Sample: mlwh.Sample{
					StudyID:    studyID,
					RunID:      runID2,
					SampleName: sampleName1,
					SampleID:   sampleName1 + "_id",
				},
			},
			{
				Sample: mlwh.Sample{
					StudyID:    studyID,
					RunID:      runID1,
					SampleName: sampleName2,
					SampleID:   sampleName2 + "_id",
				},
			},
		}

		Convey("You can generate irods_to_lustre command lines, filter the initial tsv, and copy the fastqs", func() {
			testSamplesTSVPath, err := filepath.Abs(filepath.Join("testdata", "samples.tsv"))
			So(err, ShouldBeNil)

			dir := t.TempDir()
			t.Chdir(dir)

			itl, err := New(testSamples)
			So(err, ShouldBeNil)
			So(itl, ShouldNotBeNil)
			So(itl.studyID, ShouldEqual, studyID)
			So(itl.sampleRuns, ShouldResemble, []sampleRun{
				{sampleID: "sample1_id", runID: "run1"},
				{sampleID: "sample1_id", runID: "run2"},
				{sampleID: "sample2_id", runID: "run1"},
			})

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

			fcs, err := itl.FilterSamplesTSV(testSamplesTSVPath, ".")
			So(err, ShouldBeNil)
			So(fcs, ShouldHaveLength, len(testSamples))

			for i, sampleRun := range []string{
				sampleName1 + "_id." + runID1,
				sampleName1 + "_id." + runID2,
				sampleName2 + "_id." + runID1,
			} {
				cmd := fcs[i].Command()
				So(cmd, ShouldEqual,
					fmt.Sprintf(
						"irods_to_lustre --run_mode csv_samples_id --input_samples_csv %[1]s.tsv "+
							"--samples_to_process -1 --run_imeta_study false --run_iget_study_cram true "+
							"--run_merge_crams true --run_crams_to_fastq true --filter_manual_qc true "+
							"--outdir %[1]s.output -w %[1]s.work",
						sampleRun,
					),
				)

				So(fileContents(
					filepath.Join(".", sampleRun+".tsv")),
					ShouldEqual,
					fileContents(testSamplesTSVPath+"."+sampleRun))

				err = createTestFastqFiles(sampleRun)
				So(err, ShouldBeNil)

				finalDir := filepath.Join(dir, "final")

				err = os.MkdirAll(finalDir, userPerm)
				So(err, ShouldBeNil)

				err = fcs[i].CopyFastqFiles(finalDir)
				So(err, ShouldBeNil)

				for _, suffix := range []string{"_1.fastq.gz", "_2.fastq.gz", ".fastq.gz"} {
					expectedBasename := sampleRun[:7] + "_id" + suffix

					So(fileContents(
						filepath.Join(finalDir, sampleRun+suffix)),
						ShouldEqual,
						fileContents(filepath.Join(sampleRun+".output", fastqOutputSubDir, expectedBasename)))
				}
			}
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

func fileContents(path string) string {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err.Error()
	}

	return string(contents)
}

func createTestFastqFiles(sampleRun string) error {
	dir := filepath.Join(".", sampleRun+".output", fastqOutputSubDir)

	err := os.MkdirAll(dir, userPerm)
	if err != nil {
		return err
	}

	sampleID := sampleRun[:7] + "_id"

	for _, suffix := range []string{"_1.fastq.gz", "_2.fastq.gz", ".fastq.gz"} {
		path := filepath.Join(dir, sampleID+suffix)

		err := os.WriteFile(path, []byte(sampleRun+" "+suffix), userPerm)
		if err != nil {
			return err
		}
	}

	return nil
}
