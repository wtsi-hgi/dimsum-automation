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
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/wtsi-hgi/dimsum-automation/config"
	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

const (
	sponsor = "Ben Lehner"
	errMock = Error("mock error")
)

type mockMLWH struct {
	msamples  []mlwh.Sample
	queryTime time.Duration
	err       error
	mu        sync.RWMutex
}

func (m *mockMLWH) SamplesForSponsor(sponsor string) ([]mlwh.Sample, error) {
	time.Sleep(m.queryTime)

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.msamples, m.err
}

func (m *mockMLWH) setSamples(samples []mlwh.Sample) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.msamples = samples
}

func (m *mockMLWH) setError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.err = err
}

func (m *mockMLWH) Close() error {
	return nil
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
				ManualQC:   true,
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
		exp := "exp"
		libMeta := sheets.LibraryMetaData{ExperimentID: exp}

		smeta := map[string]sheets.MetaData{
			"sample1": {Replicate: 1, LibraryMetaData: libMeta},
			"sample3": {Replicate: 2, LibraryMetaData: libMeta},
			"sample4": {Replicate: 3, LibraryMetaData: libMeta},
			"sample5": {Replicate: 4, LibraryMetaData: libMeta},
		}
		sclient := &mockSheets{smeta: smeta}

		allowedAge := 2 * mlwhQueryTime
		c := New(mclient, sclient, ClientOptions{
			SheetID:       "sheetID",
			CacheLifetime: allowedAge,
			Prefetch:      []string{sponsor},
		})
		createTime := time.Now()

		defer c.Close()

		So(c.LastPrefetchSuccess(), ShouldHappenBefore, createTime)

		Convey("You can get info about samples belonging to a given sponsor", func() {
			start := time.Now()
			samples, err := c.ForSponsor(sponsor)
			So(err, ShouldBeNil)
			So(len(samples), ShouldEqual, 3)
			So(samples, ShouldResemble, Samples{
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

			So(time.Since(start), ShouldBeLessThan, mlwhQueryTime)
			So(c.LastPrefetchSuccess(), ShouldHappenBefore, createTime)

			Convey("Queries to mlwh and sheets are cached", func() {
				mclient.setSamples(msamples[0:1])

				time.Sleep(mlwhQueryTime / 2)

				start = time.Now()
				cachedSamples, err := c.ForSponsor(sponsor)
				So(err, ShouldBeNil)
				So(cachedSamples, ShouldResemble, samples)

				So(time.Since(start), ShouldBeLessThan, mlwhQueryTime)
				So(time.Since(createTime), ShouldBeLessThan, mlwhQueryTime)
				So(c.LastPrefetchSuccess(), ShouldHappenBefore, createTime)

				Convey("And the cache expires and auto-renews", func() {
					time.Sleep(allowedAge * 2)

					start = time.Now()
					freshSamples, err := c.ForSponsor(sponsor)
					So(err, ShouldBeNil)
					So(len(freshSamples), ShouldEqual, 1)
					So(freshSamples, ShouldResemble, Samples{
						{
							Sample:   msamples[0],
							MetaData: smeta[msamples[0].SampleName],
						},
					})

					So(time.Since(start), ShouldBeLessThan, mlwhQueryTime)
					So(c.LastPrefetchSuccess(), ShouldHappenAfter, createTime)
				})

				Convey("Prefetch errors are captured", func() {
					mclient.setError(errMock)
					So(c.Err(), ShouldBeNil)

					time.Sleep(allowedAge * 2)

					So(c.Err(), ShouldEqual, errMock)

					freshSamples, err := c.ForSponsor(sponsor)
					So(err, ShouldBeNil)
					So(len(freshSamples), ShouldEqual, 3)
					So(c.Err(), ShouldEqual, errMock)
					So(c.LastPrefetchSuccess(), ShouldHappenBefore, createTime)
				})
			})

			Convey("You can filter those for desired samples", func() {
				subset, err := samples.Filter([]NameRun{
					{Name: msamples[0].SampleName, Run: msamples[0].RunID},
					{Name: msamples[2].SampleName, Run: msamples[2].RunID},
				})
				So(err, ShouldBeNil)
				So(len(subset), ShouldEqual, 2)
				So(subset[0].SampleName, ShouldEqual, msamples[0].SampleName)
				So(subset[0].RunID, ShouldEqual, msamples[0].RunID)
				So(subset[1].SampleName, ShouldEqual, msamples[2].SampleName)
				So(subset[1].RunID, ShouldEqual, msamples[2].RunID)

				_, err = samples.Filter(nil)
				So(err, ShouldEqual, ErrNoNameRun)

				_, err = samples.Filter([]NameRun{{Name: "", Run: ""}})
				So(err, ShouldEqual, ErrInvalidNameRun)

				_, err = samples.Filter([]NameRun{
					{Name: "1", Run: "1"},
					{Name: "2", Run: "1"},
					{Name: "3", Run: "1"},
					{Name: "4", Run: "1"},
				})
				So(err, ShouldEqual, ErrNameRunsNotFound)

				_, err = samples.Filter([]NameRun{
					{Name: msamples[0].SampleName, Run: msamples[0].RunID},
					{Name: "foo", Run: "bar"},
				})
				So(err, ShouldEqual, ErrNameRunsNotFound)
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

		c := New(mlwh, sheets, ClientOptions{
			SheetID:       c.SheetID,
			CacheLifetime: 1 * time.Minute,
		})

		Convey("You can get un-cached, un-prefetched info about samples belonging to a given sponsor", func() {
			start := time.Now()
			samples, err := c.ForSponsor(sponsor)
			So(err, ShouldBeNil)
			So(len(samples), ShouldBeGreaterThan, 0)
			So(samples[0].SampleID, ShouldNotBeEmpty)
			So(samples[0].SampleName, ShouldNotBeEmpty)
			So(samples[0].RunID, ShouldNotBeEmpty)
			So(samples[0].StudyID, ShouldNotBeEmpty)
			So(samples[0].StudyName, ShouldNotBeEmpty)
			So(samples[0].ManualQC, ShouldBeTrue)
			So(samples[0].Replicate, ShouldBeGreaterThan, 0)
			So(samples[0].OD, ShouldBeGreaterThan, 0)
			So(samples[0].LibraryID, ShouldNotBeEmpty)
			So(samples[0].ExperimentID, ShouldNotBeEmpty)
			So(samples[0].Wt, ShouldNotBeEmpty)
			So(samples[0].Cutadapt5First, ShouldNotBeEmpty)
			So(samples[0].Cutadapt5Second, ShouldNotBeEmpty)
			So(time.Since(start), ShouldBeGreaterThan, 100*time.Millisecond)

			Convey("Which is then cached and filterable", func() {
				start = time.Now()
				cachedSamples, err := c.ForSponsor(sponsor)
				So(err, ShouldBeNil)
				So(cachedSamples, ShouldResemble, samples)
				So(time.Since(start), ShouldBeLessThan, 100*time.Millisecond)

				first := samples[0]
				last := samples[len(samples)-1]

				subset, err := cachedSamples.Filter([]NameRun{
					{Name: first.SampleName, Run: first.RunID},
					{Name: last.SampleName, Run: last.RunID},
				})
				So(err, ShouldBeNil)
				So(len(subset), ShouldEqual, len(samples))
				So(subset[0].SampleName, ShouldEqual, first.SampleName)
				So(subset[0].RunID, ShouldEqual, first.RunID)

				if len(subset) > 1 {
					So(subset[1].SampleName, ShouldEqual, last.SampleName)
					So(subset[1].RunID, ShouldEqual, last.RunID)
				}
			})
		})
	})
}
