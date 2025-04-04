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
			sheetL, err := sheets.Read(c.SheetID, "Libraries")
			So(err, ShouldBeNil)
			So(sheetL, ShouldNotBeNil)
			So(sheetL.ColumnHeaders[0:9], ShouldResemble, []string{
				"library_id", "experiment_id", "dimsum_wt", "dimsum_cutadapt5First", "dimsum_cutadapt5Second",
				"dimsum_maxSubstitutions", "Assay", "uniprot_WT", "binder_uniprot_id",
			})

			So(len(sheetL.Rows), ShouldBeGreaterThan, 0)
			So(sheetL.Rows[0][0], ShouldNotBeBlank)

			sheetS, err := sheets.Read(c.SheetID, "Samples")
			So(err, ShouldBeNil)
			So(sheetS, ShouldNotBeNil)
			So(sheetS.ColumnHeaders[0:6], ShouldResemble, []string{
				"experiment_id", "sample_id", "selection", "replicate", "time", "OD",
			})

			So(len(sheetS.Rows), ShouldBeGreaterThan, 0)
			So(sheetS.Rows[0][0], ShouldNotBeBlank)

			_, err = sheets.Read(c.SheetID, "~invalid")
			So(err, ShouldNotBeNil)

			_, err = sheets.Read("invalid", "Libraries")
			So(err, ShouldNotBeNil)

			Convey("Then get specific columns of information", func() {
				cols, err := sheetL.Columns("library_id", "experiment_id",
					"dimsum_wt", "dimsum_cutadapt5First", "dimsum_cutadapt5Second")
				So(err, ShouldBeNil)
				So(len(cols), ShouldBeGreaterThan, 0)
				So(len(cols[0]), ShouldEqual, 5)

				_, err = sheetL.Columns("library_id", "foo")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Which you can use to retrieve the merged data needed for DimSum", func() {
			samples, err := sheets.DimSumMetaData(c.SheetID)
			So(err, ShouldBeNil)
			So(len(samples), ShouldBeGreaterThan, 0)
			So(samples["AM762abstart5"], ShouldResemble, MetaData{
				Selection: 1,
				Replicate: 2,
				Time:      34.5,
				OD:        1.27,
				LibraryMetaData: LibraryMetaData{
					LibraryID:        "762",
					ExperimentID:     "762_abundance",
					Wt:               "AAGGTCATGGAAATAAAGCTGATCAAGGGCCCAAAAGGACTTGGGTTCTCTATCGCAGGCGGAGTTGGCAACCAGCATATCCCCGGGGATAACTCAATCTACGTAACCAAAATTATCGAAGGCGGGGCAGCTCATAAGGATGGTCGACTT",
					Cutadapt5First:   "GGGAGGTGGAGCTAGC",
					Cutadapt5Second:  "GCCAAGATTTTATCTCCAATTTG",
					MaxSubstitutions: 2,
				},
			})
		})
	})
}
