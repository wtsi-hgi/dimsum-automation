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

package mlwh

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/wtsi-hgi/dimsum-automation/config"
)

const sponsor = "Ben Lehner"

func TestMLWH(t *testing.T) {
	c, err := config.FromEnv("..")
	if err != nil {
		SkipConvey("skipping mlwh tests without DIMSUM_AUTOMATION_* set", t, func() {})

		return
	}

	Convey("Given a working New MLWH", t, func() {
		mlwh, err := New(MySQLConfigFromConfig(c))
		So(err, ShouldBeNil)
		So(mlwh, ShouldNotBeNil)

		Convey("You can get info about samples belonging to a given sponsor", func() {
			samples, err := mlwh.SamplesForSponsor(sponsor)
			So(err, ShouldBeNil)
			So(len(samples), ShouldBeGreaterThan, 10)
			So(samples[0].SampleID, ShouldNotBeEmpty)
			So(samples[0].SampleName, ShouldNotBeEmpty)
			So(samples[0].RunID, ShouldNotBeEmpty)
			So(samples[0].StudyID, ShouldNotBeEmpty)
			So(samples[0].StudyName, ShouldNotBeEmpty)

			passed := 0
			failed := 0

			for _, sample := range samples {
				if sample.ManualQC == "1" {
					passed++
				} else {
					failed++
				}
			}

			So(passed, ShouldBeGreaterThan, 0)
			So(failed, ShouldBeGreaterThan, 0)

			samples, err = mlwh.SamplesForSponsor("invalid sponsor")
			So(err, ShouldBeNil)
			So(len(samples), ShouldEqual, 0)
		})
	})
}
