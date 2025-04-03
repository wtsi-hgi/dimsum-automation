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

package dimsum

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/samples"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

func TestDimsum(t *testing.T) {
	Convey("Given library and sample info", t, func() {
		exp := "exp"
		sample1 := "sample1"
		sample2 := "sample2"
		run := "run"

		libMeta := sheets.LibraryMetaData{
			LibraryID:        "lib1",
			ExperimentID:     exp,
			Wt:               "wt",
			Cutadapt5First:   "ACTG",
			Cutadapt5Second:  "TCGA",
			MaxSubstitutions: 3,
		}

		testSamples := []samples.Sample{
			{
				Sample: mlwh.Sample{
					SampleName: sample1,
					SampleID:   sample1 + "_id",
					RunID:      run,
				},
				MetaData: sheets.MetaData{
					Selection:       0,
					Replicate:       1,
					Time:            0.5,
					OD:              0.1,
					LibraryMetaData: libMeta,
				},
			},
			{
				Sample: mlwh.Sample{
					SampleName: sample2,
					SampleID:   sample2 + "_id",
					RunID:      run,
				},
				MetaData: sheets.MetaData{
					Selection:       1,
					Replicate:       2,
					Time:            0.6,
					OD:              0.2,
					LibraryMetaData: libMeta,
				},
			},
		}

		Convey("You can generate an experiment design file", func() {
			dir := t.TempDir()

			design, err := NewExperimentDesign(testSamples)
			So(err, ShouldBeNil)
			So(design, ShouldResemble, ExperimentDesign{
				{
					ID:              exp,
					SampleName:      sample1,
					Replicate:       1,
					Selection:       0,
					Pair1:           sample1 + "_id." + run + pair1FastqSuffix,
					Pair2:           sample1 + "_id." + run + pair2FastqSuffix,
					CellDensity:     0.1,
					Generations:     float32(1),
					SelectionTime:   0.5,
					LibraryMetaData: libMeta,
				},
				{
					ID:              exp,
					SampleName:      sample2,
					Replicate:       2,
					Selection:       1,
					Pair1:           sample2 + "_id." + run + pair1FastqSuffix,
					Pair2:           sample2 + "_id." + run + pair2FastqSuffix,
					CellDensity:     0.2,
					Generations:     float32(2),
					SelectionTime:   0.6,
					LibraryMetaData: libMeta,
				},
			})
			So(design.ID(), ShouldEqual, exp)

			designPath, err := design.Write(dir)
			So(err, ShouldBeNil)
			So(designPath, ShouldEqual,
				filepath.Join(dir, experimentDesignPrefix+exp+experimentDesignSuffix))

			ts0 := testSamples[0]
			ts0m := ts0.MetaData
			ts1 := testSamples[1]
			ts1m := ts1.MetaData

			d, err := os.ReadFile(designPath)
			So(err, ShouldBeNil)
			So(string(d), ShouldEqual, fmt.Sprintf(
				"sample_name\texperiment_replicate\tselection_id\tselection_replicate\ttechnical_replicate\t"+
					"pair1\tpair2\tgenerations\tcell_density\tselection_time\n"+
					"%s\t%d\t%d\t%s\t%d\t%s_id.run_1.fastq.gz\t%s_id.run_2.fastq.gz\t%d\t%.3f\t%.1f\n"+
					"%s\t%d\t%d\t%s\t%d\t%s_id.run_1.fastq.gz\t%s_id.run_2.fastq.gz\t%d\t%.3f\t%.1f\n",
				sample1, ts0m.Replicate, ts0m.Selection, "", 1, sample1, sample1, 1, ts0m.OD, ts0m.Time,
				sample2, ts1m.Replicate, ts1m.Selection, "1", 1, sample2, sample2, 2, ts1m.OD, ts1m.Time,
			))

			Convey("Then you can generate a dimsum command line", func() {
				exe := "/path/to/DiMSum"
				fastqDir := "/path/to/fastqs"
				vsearchMinQual := 20
				startStage := 0
				fitnessMinInputCountAny := 10
				fitnessMinInputCountAll := 0
				barcodeIdentityPath := "barcode_identity.txt"

				So(design.LibraryMetaData(), ShouldResemble, libMeta)

				dimsum := New(exe, fastqDir, barcodeIdentityPath, vsearchMinQual, startStage,
					fitnessMinInputCountAny, fitnessMinInputCountAll, libMeta)
				So(dimsum, ShouldNotBeNil)

				So(dimsum.Key(testSamples), ShouldEqual, "exp/sample1.run,sample2.run/69b24c9009b4933a204a8d2aace78d566eb8b31b")

				cmd := dimsum.Command(dir, libMeta)
				So(cmd, ShouldEqual, fmt.Sprintf(
					"%s -i %s -l %s -g %s -e %s --cutadapt5First %s --cutadapt5Second %s "+
						"-n %d -a %.2f -q %d -o %s -p %s -s %d -w %s -c %d "+
						"--fitnessMinInputCountAny %d --fitnessMinInputCountAll %d "+
						"--maxSubstitutions %d --mutagenesisType %s --retainIntermediateFiles %s "+
						"--mixedSubstitutions %s --experimentDesignPairDuplicates %s "+
						"--barcodeIdentityPath %s",
					exe, fastqDir, DefaultFastqExtension, "TRUE", designPath,
					libMeta.Cutadapt5First+cutAdaptRequired+"TCGA"+cutAdaptOptional,
					libMeta.Cutadapt5Second+cutAdaptRequired+"CAGT"+cutAdaptOptional,
					DefaultCutAdaptMinLength, DefaultCutAdaptErrorRate,
					vsearchMinQual, filepath.Join(dir, outputSubdir), dimsumProjectPrefix+exp,
					startStage, libMeta.Wt, DefaultCores, fitnessMinInputCountAny,
					fitnessMinInputCountAll, 3,
					DefaultMutagenesisType, "T", "F", "F", barcodeIdentityPath,
				))

				dimsum = New(exe, fastqDir, "", vsearchMinQual, startStage,
					fitnessMinInputCountAny, fitnessMinInputCountAll, libMeta)
				So(dimsum, ShouldNotBeNil)

				cmd = dimsum.Command(dir, libMeta)
				So(cmd, ShouldNotContainSubstring, "--barcodeIdentityPath")
				So(dimsum.Key(testSamples), ShouldEqual, "exp/sample1.run,sample2.run/631c90f196443c203f4eeea856da242fafcc1793")
			})
		})
	})
}
