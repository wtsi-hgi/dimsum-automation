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
	"github.com/wtsi-hgi/dimsum-automation/types"
)

func TestDimsum(t *testing.T) {
	Convey("Given library, experiement and sample info", t, func() {
		sample1 := "sample1"
		sample2 := "sample2"
		run := "run"

		testSamples := []*types.Sample{
			{
				SampleName:          sample1,
				SampleID:            sample1 + "_id",
				RunID:               run,
				Selection:           types.SelectionInput,
				ExperimentReplicate: 1,
				TechnicalReplicate:  1,
				SelectionTime:       "0.5",
				CellDensity:         "0.1",
			},
			{
				SampleName:          sample2,
				SampleID:            sample2 + "_id",
				RunID:               run,
				Selection:           types.SelectionOutput,
				ExperimentReplicate: 2,
				TechnicalReplicate:  1,
				SelectionTime:       "0.6",
				CellDensity:         "0.2",
			},
		}

		barcodeIdentityPath := "barcode_identity.txt"
		exp := &types.Experiment{
			ExperimentID:        "exp",
			BarcodeIdentityPath: barcodeIdentityPath,
			WildtypeSequence:    "wt",
			MaxSubstitutions:    3,
			Cutadapt5First:      "ACTG",
			Cutadapt5Second:     "TCGA",
			Samples:             testSamples,
		}

		Convey("You can generate an experiment design file", func() {
			dir := t.TempDir()

			design, err := NewExperimentDesign(exp)
			So(err, ShouldBeNil)
			So(design, ShouldResemble, ExperimentDesign{
				Experiment: exp,
				Samples: []*types.Sample{
					{
						SampleName:          sample1,
						SampleID:            sample1 + "_id",
						RunID:               run,
						Selection:           types.SelectionInput,
						ExperimentReplicate: 1,
						TechnicalReplicate:  1,
						SelectionTime:       "0.5",
						CellDensity:         "0.1",
						Pair1:               sample1 + "_id." + run + pair1FastqSuffix,
						Pair2:               sample1 + "_id." + run + pair2FastqSuffix,
					},
					{
						SampleName:          sample2,
						SampleID:            sample2 + "_id",
						RunID:               run,
						Selection:           types.SelectionOutput,
						ExperimentReplicate: 2,
						TechnicalReplicate:  1,
						SelectionTime:       "0.6",
						CellDensity:         "0.2",
						Pair1:               sample2 + "_id." + run + pair1FastqSuffix,
						Pair2:               sample2 + "_id." + run + pair2FastqSuffix,
					},
				},
			})
			So(design.ExperimentID, ShouldEqual, exp.ExperimentID)

			designPath, err := design.Write(dir)
			So(err, ShouldBeNil)
			So(designPath, ShouldEqual,
				filepath.Join(dir, experimentDesignPrefix+exp.ExperimentID+experimentDesignSuffix))

			ts0 := testSamples[0]
			ts1 := testSamples[1]

			d, err := os.ReadFile(designPath)
			So(err, ShouldBeNil)
			So(string(d), ShouldEqual, fmt.Sprintf(
				"sample_name\texperiment_replicate\tselection_id\tselection_replicate\ttechnical_replicate\t"+
					"pair1\tpair2\tgenerations\tcell_density\tselection_time\n"+
					"%s\t%d\t%d\t%s\t%d\t%s_id.run_1.fastq.gz\t%s_id.run_2.fastq.gz\t%d\t%s\t%s\n"+
					"%s\t%d\t%d\t%s\t%d\t%s_id.run_1.fastq.gz\t%s_id.run_2.fastq.gz\t%d\t%s\t%s\n",
				"input1", ts0.ExperimentReplicate, ts0.SelectionID(), ts0.SelectionReplicate(),
				1, sample1, sample1, 1, ts0.CellDensity, ts0.SelectionTime,
				"output2", ts1.ExperimentReplicate, ts1.SelectionID(), ts1.SelectionReplicate(),
				1, sample2, sample2, 2, ts1.CellDensity, ts1.SelectionTime,
			))

			//TODO: proper test for generations value being correct for an
			// output with a corresponding input of cell density other than 0.05

			Convey("Then you can generate a dimsum command line", func() {
				fastqDir := "/path/to/fastqs"

				dimsum := New(fastqDir, design)
				So(dimsum, ShouldNotBeNil)

				So(dimsum.Key(testSamples), ShouldEqual, "exp/sample1.run,sample2.run/69b24c9009b4933a204a8d2aace78d566eb8b31b")

				cmd, err := dimsum.Command()
				So(err, ShouldBeNil)

				So(cmd, ShouldEqual, fmt.Sprintf(
					"%s -i %s -l %s -g %s -e %s --cutadapt5First %s --cutadapt5Second %s "+
						"-n %d -a %.2f -q %d -o %s -p %s -s %d -w %s -c %d "+
						"--fitnessMinInputCountAny %d --fitnessMinInputCountAll %d "+
						"--maxSubstitutions %d --mutagenesisType %s --retainIntermediateFiles %s "+
						"--mixedSubstitutions %s --experimentDesignPairDuplicates %s "+
						"--barcodeIdentityPath %s",
					DimSumExe, fastqDir, DefaultFastqExtension, "T", filepath.Base(designPath),
					exp.Cutadapt5First, exp.Cutadapt5Second,
					DefaultCutAdaptMinLength, DefaultCutAdaptErrorRate,
					DefaultVsearchMinQual, outputSubdir, dimsumProjectPrefix+exp.ExperimentID,
					DefaultStartStage, exp.WildtypeSequence, DefaultCores, DefaultFitnessMinInputCountAny,
					DefaultFitnessMinInputCountAll, 3,
					DefaultMutagenesisType, "T", "F", "F", barcodeIdentityPath,
				))

				_, err = os.Stat(outputSubdir)
				So(err, ShouldBeNil)

				dimsum = New(fastqDir, design)
				So(dimsum, ShouldNotBeNil)

				exp.BarcodeIdentityPath = ""
				cmd, err = dimsum.Command()
				So(err, ShouldBeNil)
				So(cmd, ShouldNotContainSubstring, "--barcodeIdentityPath")
				So(dimsum.Key(testSamples), ShouldEqual, "exp/sample1.run,sample2.run/631c90f196443c203f4eeea856da242fafcc1793")
			})
		})
	})
}
