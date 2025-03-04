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

package samples

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/wtsi-hgi/dimsum-automation/config"
	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

const sponsor = "Ben Lehner"

type mockMLWH struct {
	msamples  []mlwh.Sample
	queryTime time.Duration
}

func (m *mockMLWH) SamplesForSponsor(sponsor string) ([]mlwh.Sample, error) {
	time.Sleep(m.queryTime)
	return m.msamples, nil
}

type mockSheets struct{ smeta map[string]sheets.MetaData }

func (m *mockSheets) DimSumMetaData(sheetID string) (map[string]sheets.MetaData, error) {
	return m.smeta, nil
}

func TestSamplesMock(t *testing.T) {
	Convey("Given mock mlwh and sheets connections", t, func() {
		msamples := []mlwh.Sample{
			{
				SampleID:   "sampleID1",
				SampleName: "sample1",
				RunID:      "run1",
				StudyID:    "studyID1",
				StudyName:  "study1",
			},
			{
				SampleID:   "sampleID2",
				SampleName: "sample2",
				RunID:      "run2",
				StudyID:    "studyID1",
				StudyName:  "study1",
			},
			{
				SampleID:   "sampleID3",
				SampleName: "sample3",
				RunID:      "run3",
				StudyID:    "studyID2",
				StudyName:  "study2",
			},
			{
				SampleID:   "sampleID4",
				SampleName: "sample4",
				RunID:      "run4",
				StudyID:    "studyID3",
				StudyName:  "study3",
			},
		}
		mlwhQueryTime := 100 * time.Millisecond
		mclient := &mockMLWH{msamples: msamples, queryTime: mlwhQueryTime}

		smeta := map[string]sheets.MetaData{
			"sample1": {Replicate: 1},
			"sample3": {Replicate: 2},
			"sample4": {Replicate: 3},
			"sample5": {Replicate: 4},
		}
		sclient := &mockSheets{smeta: smeta}

		allowedAge := 2 * mlwhQueryTime
		s := New(mclient, sclient, ClientOptions{SheetID: "sheetID", CacheLifetime: allowedAge})

		Convey("You can get info about samples belonging to a given sponsor", func() {
			start := time.Now()
			samples, err := s.ForSponsor(sponsor)
			So(err, ShouldBeNil)
			So(len(samples), ShouldEqual, 3)
			So(samples, ShouldResemble, []Sample{
				{
					Sample:   msamples[0],
					MetaData: smeta[msamples[0].SampleName],
				},
				{
					Sample:   msamples[2],
					MetaData: smeta[msamples[2].SampleName],
				},
				{
					Sample:   msamples[3],
					MetaData: smeta[msamples[3].SampleName],
				},
			})

			uncachedTime := time.Since(start)
			So(uncachedTime, ShouldBeGreaterThan, mlwhQueryTime)

			Convey("Queries to mlwh and sheets are cached", func() {
				mclient.msamples = msamples[0:1]

				start = time.Now()
				cachedSamples, err := s.ForSponsor(sponsor)
				So(err, ShouldBeNil)
				So(cachedSamples, ShouldResemble, samples)

				cachedTime := time.Since(start)
				So(cachedTime, ShouldBeLessThan, mlwhQueryTime)

				Convey("But the cache expires", func() {
					time.Sleep(allowedAge)

					start = time.Now()
					freshSamples, err := s.ForSponsor(sponsor)
					So(err, ShouldBeNil)
					So(len(freshSamples), ShouldEqual, 1)
					So(freshSamples, ShouldResemble, []Sample{
						{
							Sample:   msamples[0],
							MetaData: smeta[msamples[0].SampleName],
						},
					})

					cachedTime := time.Since(start)
					So(cachedTime, ShouldBeGreaterThan, mlwhQueryTime)
				})
			})
		})
	})
}

func TestSamplesReal(t *testing.T) {
	c, err := config.FromEnv("..")
	if err != nil {
		SkipConvey("skipping real samples tests without DIMSUM_AUTOMATION_* set", t, func() {})

		return
	}

	Convey("Given mlwh and sheets connections", t, func() {
		mlwh, err := mlwh.New(mlwh.MySQLConfigFromConfig(c))
		So(err, ShouldBeNil)

		sc, err := sheets.ServiceCredentialsFromConfig(c)
		So(err, ShouldBeNil)

		sheets, err := sheets.New(sc)
		So(err, ShouldBeNil)

		s := New(mlwh, sheets, ClientOptions{SheetID: c.SheetID})

		Convey("You can get info about samples belonging to a given sponsor", func() {
			samples, err := s.ForSponsor(sponsor)
			So(err, ShouldBeNil)
			So(len(samples), ShouldBeGreaterThan, 0)
			So(samples[0].SampleID, ShouldNotBeEmpty)
			So(samples[0].SampleName, ShouldNotBeEmpty)
			So(samples[0].RunID, ShouldNotBeEmpty)
			So(samples[0].StudyID, ShouldNotBeEmpty)
			So(samples[0].StudyName, ShouldNotBeEmpty)
			So(samples[0].Replicate, ShouldBeGreaterThan, 0)
			So(samples[0].OD, ShouldBeGreaterThan, 0)
			So(samples[0].LibraryID, ShouldNotBeEmpty)
			So(samples[0].Wt, ShouldNotBeEmpty)
			So(samples[0].Cutadapt5First, ShouldNotBeEmpty)
			So(samples[0].Cutadapt5Second, ShouldNotBeEmpty)
		})
	})
}
