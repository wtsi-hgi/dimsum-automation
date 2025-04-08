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

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/wtsi-hgi/dimsum-automation/config"
	"github.com/wtsi-hgi/dimsum-automation/types"
)

func TestSheets(t *testing.T) {
	c, err := config.FromEnv("..")
	if err != nil {
		SkipConvey("skipping sheet tests without DIMSUM_AUTOMATION_* set", t, func() {})

		return
	}

	sc, err := ServiceCredentialsFromFile(c.CredentialsPath)
	if err != nil {
		SkipConvey("skipping sheet tests without valid credentials.json", t, func() {})

		return
	}

	Convey("Given real service credentials, you can make a Sheets", t, func() {
		sheets, err := New(sc)
		So(err, ShouldBeNil)
		So(sheets, ShouldNotBeNil)

		Convey("Which you can use to Read the contents of named sheets", func() {
			sheetL, err := sheets.Read(c.SheetID, "libraries")
			So(err, ShouldBeNil)
			So(sheetL, ShouldNotBeNil)
			So(sheetL.ColumnHeaders, ShouldContain, "library_id")
			So(sheetL.ColumnHeaders, ShouldContain, "wildtypeSequence")
			So(sheetL.ColumnHeaders, ShouldContain, "maxSubstitutions")
			So(len(sheetL.Rows), ShouldBeGreaterThan, 0)
			So(sheetL.Rows[0][0], ShouldNotBeBlank)

			sheetE, err := sheets.Read(c.SheetID, "experiments")
			So(err, ShouldBeNil)
			So(sheetE, ShouldNotBeNil)
			So(sheetE.ColumnHeaders, ShouldContain, "library_id")
			So(sheetE.ColumnHeaders, ShouldContain, "experiment_id")
			So(sheetE.ColumnHeaders, ShouldContain, "projectName")
			So(sheetE.ColumnHeaders, ShouldContain, "startStage")
			So(sheetE.ColumnHeaders, ShouldContain, "stopStage")
			So(sheetE.ColumnHeaders, ShouldContain, "barcodeDesignPath")
			So(sheetE.ColumnHeaders, ShouldContain, "barcodeErrorRate")
			So(sheetE.ColumnHeaders, ShouldContain, "experimentDesignPairDuplicates")
			So(sheetE.ColumnHeaders, ShouldContain, "countPath")
			So(sheetE.ColumnHeaders, ShouldContain, "barcodeIdentityPath")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadapt5First")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadapt5Second")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadaptMinLength")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadaptErrorRate")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadaptOverlap")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadaptCut5First")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadaptCut5Second")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadaptCut3First")
			So(sheetE.ColumnHeaders, ShouldContain, "cutadaptCut3Second")
			So(sheetE.ColumnHeaders, ShouldContain, "vsearchMinQual")
			So(sheetE.ColumnHeaders, ShouldContain, "vsearchMaxQual")
			So(sheetE.ColumnHeaders, ShouldContain, "vsearchMaxee")
			So(sheetE.ColumnHeaders, ShouldContain, "vsearchMinovlen")
			So(sheetE.ColumnHeaders, ShouldContain, "reverseComplement")
			So(sheetE.ColumnHeaders, ShouldContain, "wildtypeSequence")
			So(sheetE.ColumnHeaders, ShouldContain, "permittedSequences")
			So(sheetE.ColumnHeaders, ShouldContain, "sequenceType")
			So(sheetE.ColumnHeaders, ShouldContain, "mutagenesisType")
			So(sheetE.ColumnHeaders, ShouldContain, "indels")
			So(sheetE.ColumnHeaders, ShouldContain, "maxSubstitutions")
			So(sheetE.ColumnHeaders, ShouldContain, "mixedSubstitutions")
			So(sheetE.ColumnHeaders, ShouldContain, "fitnessMinInputCountAll")
			So(sheetE.ColumnHeaders, ShouldContain, "fitnessMinInputCountAny")
			So(sheetE.ColumnHeaders, ShouldContain, "fitnessMinOutputCountAll")
			So(sheetE.ColumnHeaders, ShouldContain, "fitnessMinOutputCountAny")
			So(sheetE.ColumnHeaders, ShouldContain, "fitnessNormalise")
			So(sheetE.ColumnHeaders, ShouldContain, "fitnessErrorModel")
			So(sheetE.ColumnHeaders, ShouldContain, "fitnessDropoutPseudocount")
			So(sheetE.ColumnHeaders, ShouldContain, "retainedReplicates")
			So(sheetE.ColumnHeaders, ShouldContain, "stranded")
			So(sheetE.ColumnHeaders, ShouldContain, "paired")
			So(sheetE.ColumnHeaders, ShouldContain, "experimentDesignPairDuplicates")
			So(sheetE.ColumnHeaders, ShouldContain, "synonymSequencePath")
			So(sheetE.ColumnHeaders, ShouldContain, "transLibrary")
			So(sheetE.ColumnHeaders, ShouldContain, "transLibraryReverseComplement")
			So(len(sheetE.Rows), ShouldBeGreaterThan, 0)
			So(sheetE.Rows[0][0], ShouldNotBeBlank)

			sheetS, err := sheets.Read(c.SheetID, "samples")
			So(err, ShouldBeNil)
			So(sheetS, ShouldNotBeNil)
			So(sheetS.ColumnHeaders, ShouldContain, "experiment_id")
			So(sheetS.ColumnHeaders, ShouldContain, "sample_id")
			So(sheetS.ColumnHeaders, ShouldContain, "selection")
			So(sheetS.ColumnHeaders, ShouldContain, "experiment_replicate")
			So(sheetS.ColumnHeaders, ShouldContain, "selection_time")
			So(sheetS.ColumnHeaders, ShouldContain, "cell_density")
			So(len(sheetS.Rows), ShouldBeGreaterThan, 0)
			So(sheetS.Rows[0][0], ShouldNotBeBlank)

			_, err = sheets.Read(c.SheetID, "~invalid")
			So(err, ShouldNotBeNil)

			_, err = sheets.Read("invalid", "Libraries")
			So(err, ShouldNotBeNil)

			Convey("Then get specific columns of information", func() {
				cols, err := sheetL.Columns("library_id", "maxSubstitutions")
				So(err, ShouldBeNil)
				So(len(cols), ShouldBeGreaterThan, 0)
				So(len(cols[0]), ShouldEqual, 2)

				_, err = sheetL.Columns("library_id", "foo")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Which you can use to retrieve the data needed for DimSum", func() {
			libs, err := sheets.DimSumMetaData(c.SheetID)
			So(err, ShouldBeNil)
			So(len(libs), ShouldBeGreaterThan, 0)

			var lib762 *types.Library

			for _, lib := range libs {
				if lib.LibraryID == "762" {
					lib762 = lib

					break
				}
			}

			So(lib762, ShouldNotBeNil)
			So(lib762.MaxSubstitutions, ShouldEqual, 2)

			So(len(libs[0].Experiments), ShouldBeGreaterThan, 0)
			So(len(libs[0].Experiments[0].Samples), ShouldBeGreaterThan, 0)

			lib, err := libs.Subset([]*types.Sample{
				{SampleID: "AM762abstart1", RunID: "?"},
				{SampleID: "AM762abstart5", RunID: "?"},
			})
			So(err, ShouldBeNil)
			So(lib.LibraryID, ShouldEqual, "762")
			So(lib.WildtypeSequence, ShouldEqual, "AAGGTCATGGAAATAAAGCTGATCAAGGGCCCAAAAGGACTTGGGTTCTCTATCGCAGGCGGAGTTGGCAACCAGCATATCCCCGGGGATAACTCAATCTACGTAACCAAAATTATCGAAGGCGGGGCAGCTCATAAGGATGGTCGACTT") //nolint:lll
			So(lib.MaxSubstitutions, ShouldEqual, 2)
			So(len(lib.Experiments), ShouldEqual, 1)

			exp := lib.Experiments[0]
			So(exp.ExperimentID, ShouldEqual, "762_abundance")
			So(exp.Assay, ShouldEqual, "AbundancePCA")
			So(exp.StopStage, ShouldEqual, 5)
			So(exp.ExperimentDesignPairDuplicates, ShouldEqual, false)
			So(exp.Cutadapt5First, ShouldEqual, "GGGAGGTGGAGCTAGC;required...CAAATTGGAGATAAAATCTTGGC;optional")
			So(exp.Cutadapt5Second, ShouldEqual, "GCCAAGATTTTATCTCCAATTTG;required...GCTAGCTCCACCTCCC;optional")
			So(exp.CutadaptMinLength, ShouldEqual, 100)
			So(exp.CutadaptErrorRate, ShouldEqual, "0.2")
			So(exp.VsearchMinQual, ShouldEqual, 20)
			So(exp.MutagenesisType, ShouldEqual, types.MutagenesisTypeRandom)
			So(exp.MixedSubstitutions, ShouldEqual, false)
			So(exp.FitnessMinInputCountAll, ShouldEqual, 10)
			So(len(exp.Samples), ShouldEqual, 2)

			s1 := exp.Samples[0]
			So(s1.SampleID, ShouldEqual, "AM762abstart1")
			So(s1.Selection, ShouldEqual, types.SelectionInput)
			So(s1.ExperimentReplicate, ShouldEqual, 1)
			So(s1.SelectionTime, ShouldEqual, "")
			So(s1.CellDensity, ShouldEqual, "0.05")

			s2 := exp.Samples[1]
			So(s2.SampleID, ShouldEqual, "AM762abstart5")
			So(s2.Selection, ShouldEqual, types.SelectionOutput)
			So(s2.ExperimentReplicate, ShouldEqual, 2)
			So(s2.SelectionTime, ShouldEqual, "34.5")
			So(s2.CellDensity, ShouldEqual, "1.27")

			_, err = libs.Subset([]*types.Sample{
				{SampleID: "AM762abstart1", RunID: "?"},
				{SampleID: "AM762808start2", RunID: "?"},
			})
			So(err, ShouldNotBeNil)

			_, err = libs.Subset([]*types.Sample{{SampleID: "AM762abstart1"}})
			So(err, ShouldNotBeNil)

			_, err = libs.Subset([]*types.Sample{{SampleID: "foo", RunID: "bar"}})
			So(err, ShouldNotBeNil)
		})
	})
}
