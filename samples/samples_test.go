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

type mockSheets struct{ smeta sheets.Libraries }

func (m *mockSheets) DimSumMetaData(sheetID string) (sheets.Libraries, error) {
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
		lib := &sheets.Library{
			LibraryID: "lib",
			Experiments: []*sheets.Experiment{
				{
					ExperimentID: exp,
					Samples: []*sheets.Sample{
						{
							SampleID:            "sample1",
							ExperimentReplicate: 1,
						},
						{
							SampleID:            "sample3",
							ExperimentReplicate: 2,
						},
						{
							SampleID:            "sample4",
							ExperimentReplicate: 3,
						},
						{
							SampleID:            "sample5",
							ExperimentReplicate: 4,
						},
					},
				},
			},
		}

		sclient := &mockSheets{smeta: []*sheets.Library{lib}}

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
			mergedLibs, err := c.ForSponsor(sponsor)
			So(err, ShouldBeNil)
			So(len(mergedLibs), ShouldEqual, 0)
			mergedLib := mergedLibs[0]
			So(mergedLib, ShouldResemble, &sheets.Library{
				LibraryID: "lib",
				Experiments: []*sheets.Experiment{
					{
						ExperimentID: exp,
						Samples: []*sheets.Sample{
							{
								SampleID:            "sample1",
								ExperimentReplicate: 1,
							},
							{
								SampleID:            "sample3",
								ExperimentReplicate: 2,
							},
							{
								SampleID:            "sample4",
								ExperimentReplicate: 3,
							},
							{
								SampleID:            "sample5",
								ExperimentReplicate: 4,
							},
						},
					},
				},
			})

			So(time.Since(start), ShouldBeLessThan, mlwhQueryTime)
			So(c.LastPrefetchSuccess(), ShouldHappenBefore, createTime)

			Convey("Queries to mlwh and sheets are cached", func() {
				mclient.setSamples(msamples[0:1])

				time.Sleep(mlwhQueryTime / 2)

				start = time.Now()
				cachedLibs, err := c.ForSponsor(sponsor)
				So(err, ShouldBeNil)
				So(cachedLibs, ShouldResemble, mergedLibs)

				So(time.Since(start), ShouldBeLessThan, mlwhQueryTime)
				So(time.Since(createTime), ShouldBeLessThan, mlwhQueryTime)
				So(c.LastPrefetchSuccess(), ShouldHappenBefore, createTime)

				Convey("And the cache expires and auto-renews", func() {
					time.Sleep(allowedAge * 2)

					start = time.Now()
					freshLibs, err := c.ForSponsor(sponsor)
					So(err, ShouldBeNil)
					So(len(freshLibs), ShouldEqual, 1)
					So(freshLibs[0], ShouldResemble, &sheets.Library{
						LibraryID: "lib",
						Experiments: []*sheets.Experiment{
							{
								ExperimentID: exp,
								Samples: []*sheets.Sample{
									{
										SampleID:            "sample1",
										ExperimentReplicate: 1,
									},
								},
							},
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

					freshLibs, err := c.ForSponsor(sponsor)
					So(err, ShouldBeNil)
					So(len(freshLibs), ShouldEqual, 3)
					So(c.Err(), ShouldEqual, errMock)
					So(c.LastPrefetchSuccess(), ShouldHappenBefore, createTime)
				})
			})

			Convey("You can filter those for desired samples", func() {
				subset, err := mergedLibs.Subset(
					sheets.NameRun{Name: msamples[0].SampleName, Run: msamples[0].RunID},
					sheets.NameRun{Name: msamples[2].SampleName, Run: msamples[2].RunID},
				)
				So(err, ShouldBeNil)

				samples := subset.Experiments[0].Samples
				So(len(samples), ShouldEqual, 2)
				So(samples[0].SampleName, ShouldEqual, msamples[0].SampleName)
				So(samples[0].RunID, ShouldEqual, msamples[0].RunID)
				So(samples[1].SampleName, ShouldEqual, msamples[2].SampleName)
				So(samples[1].RunID, ShouldEqual, msamples[2].RunID)
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

		s, err := sheets.New(sc)
		So(err, ShouldBeNil)

		c := New(mlwh, s, ClientOptions{
			SheetID:       c.SheetID,
			CacheLifetime: 1 * time.Minute,
		})

		Convey("You can get un-cached, un-prefetched info about samples belonging to a given sponsor", func() {
			start := time.Now()
			libs, err := c.ForSponsor(sponsor)
			So(err, ShouldBeNil)
			So(len(libs), ShouldBeGreaterThan, 0)

			lib := libs[0]
			So(lib.LibraryID, ShouldNotBeBlank)
			So(lib.WildtypeSequence, ShouldNotBeBlank)
			So(lib.MaxSubstitutions, ShouldBeGreaterThan, 0)
			So(lib.StudyID, ShouldNotBeBlank)
			So(lib.StudyName, ShouldNotBeBlank)
			So(len(lib.Experiments), ShouldBeGreaterThan, 0)

			exp := lib.Experiments[0]
			So(exp.ExperimentID, ShouldNotBeBlank)
			So(exp.WildtypeSequence, ShouldNotBeBlank)
			So(exp.MaxSubstitutions, ShouldBeGreaterThan, 0)
			So(exp.BarcodeIdentityPath, ShouldNotBeBlank)
			So(exp.Cutadapt5First, ShouldNotBeBlank)
			So(exp.Cutadapt5Second, ShouldNotBeBlank)
			So(len(exp.Samples), ShouldBeGreaterThan, 0)

			sample := exp.Samples[0]
			So(sample.SampleID, ShouldNotBeBlank)
			So(sample.SampleName, ShouldNotBeBlank)
			So(sample.RunID, ShouldNotBeBlank)
			So(sample.ManualQC, ShouldBeTrue)
			So(sample.ExperimentReplicate, ShouldBeGreaterThan, 0)
			So(sample.CellDensity, ShouldNotBeBlank)

			So(time.Since(start), ShouldBeGreaterThan, 100*time.Millisecond)

			Convey("Which is then cached and filterable", func() {
				start = time.Now()
				cachedLibs, err := c.ForSponsor(sponsor)
				So(err, ShouldBeNil)
				So(cachedLibs, ShouldResemble, libs)
				So(time.Since(start), ShouldBeLessThan, 100*time.Millisecond)

				first := exp.Samples[0]
				last := exp.Samples[len(exp.Samples)-1]

				subset, err := cachedLibs.Subset(
					sheets.NameRun{Name: first.SampleID, Run: first.RunID},
					sheets.NameRun{Name: last.SampleID, Run: last.RunID},
				)
				So(err, ShouldBeNil)
				So(len(subset.Experiments), ShouldEqual, 1)
				So(len(subset.Experiments[0].Samples), ShouldBeGreaterThan, 0)
				So(subset.Experiments[0].Samples[0].SampleID, ShouldEqual, first.SampleID)

				if len(subset.Experiments[0].Samples) > 1 {
					So(subset.Experiments[0].Samples[1].SampleID, ShouldEqual, last.SampleID)
				}
			})
		})
	})
}
