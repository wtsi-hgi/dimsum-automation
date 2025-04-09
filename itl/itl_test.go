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
	"github.com/wtsi-hgi/dimsum-automation/types"
)

func TestITL(t *testing.T) {
	testSamplesTSVPath, err := filepath.Abs(filepath.Join("testdata", "samples.tsv"))
	if err != nil {
		t.Fatalf("Failed to get absolute path of testdata file: %v", err)
	}

	Convey("Given desired samples in an experiment in a library", t, func() {
		studyID := "study1"
		runID1 := "run1"
		runID2 := "run2"
		sampleName1 := "sample1"
		sampleName2 := "sample2"

		testSamples := []*types.Sample{
			{
				RunID:      runID1,
				SampleName: sampleName1,
				SampleID:   sampleName1 + "_id",
			},
			{
				RunID:      runID2,
				SampleName: sampleName1,
				SampleID:   sampleName1 + "_id",
			},
			{
				RunID:      runID1,
				SampleName: sampleName2,
				SampleID:   sampleName2 + "_id",
			},
		}

		testLib := &types.Library{
			LibraryID: "lib",
			StudyID:   studyID,
			Experiments: []*types.Experiment{
				{
					ExperimentID: "exp",
					Samples:      testSamples,
				},
			},
		}

		Convey("You can generate irods_to_lustre command lines, filter the initial tsv, and move the fastqs", func() {
			dir := t.TempDir()
			t.Chdir(dir)

			finalDir := t.TempDir()

			itl, err := New(testLib, finalDir)
			So(err, ShouldBeNil)
			So(itl, ShouldNotBeNil)
			So(itl.studyID, ShouldEqual, studyID)
			So(itl.Samples(), ShouldResemble, []*Sample{
				{Sample: types.Sample{SampleID: "sample1_id", RunID: "run1"}},
				{Sample: types.Sample{SampleID: "sample1_id", RunID: "run2"}},
				{Sample: types.Sample{SampleID: "sample2_id", RunID: "run1"}},
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

			fcs, err := itl.FilterSamplesTSV(testSamplesTSVPath)
			So(err, ShouldBeNil)
			So(fcs, ShouldHaveLength, len(testSamples))

			for i, sr := range []string{
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
						sr,
					),
				)

				So(fcs[i].IDRun(), ShouldEqual, sr)

				So(fileContents(filepath.Join(".", sr+".tsv")),
					ShouldEqual,
					fileContents(testSamplesTSVPath+"."+sr))

				err = createTestFastqFiles(sr)
				So(err, ShouldBeNil)

				suffixes := []string{FastqPair1Suffix, FastqPair2Suffix}

				expectedContents := make([]string, len(suffixes))
				sourcePaths := make([]string, len(suffixes))

				for i, suffix := range suffixes {
					expectedBasename := sr[:7] + "_id" + suffix
					path := filepath.Join(sr+".output", fastqOutputSubDir, expectedBasename)
					expectedContents[i] = fileContents(path)
					sourcePaths[i] = path
				}

				err = fcs[i].MoveFastqFiles()
				So(err, ShouldBeNil)

				for i, suffix := range suffixes {
					So(fileContents(filepath.Join(finalDir, sr+suffix)),
						ShouldEqual,
						expectedContents[i])

					_, err := os.Stat(sourcePaths[i])
					So(err, ShouldNotBeNil)
					So(os.IsNotExist(err), ShouldBeTrue)
				}
			}
		})

		Convey("itl ignores samples where the fastqs already exist", func() {
			dir := t.TempDir()
			t.Chdir(dir)

			finalDir := t.TempDir()

			doneSR := sampleName1 + "_id." + runID2
			fastq1 := filepath.Join(finalDir, doneSR+FastqPair1Suffix)
			err := os.WriteFile(fastq1, []byte("done"), userPerm)
			So(err, ShouldBeNil)

			_, err = New(testLib, finalDir)
			So(err, ShouldNotBeNil)

			fastq2 := filepath.Join(finalDir, doneSR+FastqPair2Suffix)
			err = os.WriteFile(fastq2, []byte("done"), userPerm)
			So(err, ShouldBeNil)

			itl, err := New(testLib, finalDir)
			So(err, ShouldBeNil)
			So(itl, ShouldNotBeNil)
			So(itl.studyID, ShouldEqual, studyID)
			So(itl.Samples(), ShouldResemble, []*Sample{
				{Sample: types.Sample{SampleID: "sample1_id", RunID: "run1"}},
				{Sample: types.Sample{SampleID: "sample2_id", RunID: "run1"}},
			})

			fcs, err := itl.FilterSamplesTSV(testSamplesTSVPath)
			So(err, ShouldBeNil)
			So(fcs, ShouldHaveLength, len(testSamples)-1)
		})

		Convey("You can't make a new ITL with multiple or no experiments", func() {
			dir := t.TempDir()

			testLib.Experiments = append(testLib.Experiments, &types.Experiment{
				ExperimentID: "exp2",
			})

			itl, err := New(testLib, dir)
			So(err, ShouldNotBeNil)
			So(itl, ShouldBeNil)

			testLib.Experiments = nil

			_, err = New(testLib, dir)
			So(err, ShouldNotBeNil)

			_, err = New(nil, dir)
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
