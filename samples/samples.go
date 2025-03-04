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
	"time"

	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

type Error string

func (e Error) Error() string { return string(e) }

type MLWHClient interface {
	// SamplesForSponsor returns all samples for the given sponsor, including
	// study and run information.
	SamplesForSponsor(sponsor string) ([]mlwh.Sample, error)
}

type SheetsClient interface {
	// DimSumMetaData reads sheets "Libraries" and "Samples" from the sheet with
	// the given id and merges the results for columns relevant to DimSum,
	// returning a map where keys are sample_id.
	DimSumMetaData(sheetID string) (map[string]sheets.MetaData, error)
}

type cache struct {
	samples    map[string][]mlwh.Sample
	metadata   map[string]sheets.MetaData
	lastUpdate time.Time
	lifetime   time.Duration
	mu         sync.RWMutex
}

func newCache(lifetime time.Duration) *cache {
	return &cache{
		samples:  make(map[string][]mlwh.Sample),
		metadata: make(map[string]sheets.MetaData),
		lifetime: lifetime,
	}
}

func (c *cache) getData(sponsor string) (bool, []mlwh.Sample, map[string]sheets.MetaData) {
	c.mu.RLock()
	cached := c.lastUpdate.Add(c.lifetime).After(time.Now())
	samples := c.samples[sponsor]
	metadata := c.metadata
	c.mu.RUnlock()

	return cached, samples, metadata
}

func (c *cache) storeData(sponsor string, samples []mlwh.Sample, metadata map[string]sheets.MetaData) {
	c.mu.Lock()
	c.samples[sponsor] = samples
	c.metadata = metadata
	c.lastUpdate = time.Now()
	c.mu.Unlock()
}

// Client can connect to MLWH and Google Sheets to get sample information.
type Client struct {
	mc      MLWHClient
	sc      SheetsClient
	sheetID string
	cache   *cache
}

// ClientOptions are options for creating a new Client.
type ClientOptions struct {
	// SheetID is the id of the google sheet to get metadata from.
	SheetID string

	// CacheLifetime is the maximum age of cached results.
	CacheLifetime time.Duration
}

// New returns a new Client that can connect to MLWH and the google sheet with
// the given id to retrieve sample information.
func New(mc MLWHClient, sc SheetsClient, opts ClientOptions) *Client {
	c := &Client{
		mc:      mc,
		sc:      sc,
		sheetID: opts.SheetID,
		cache:   newCache(opts.CacheLifetime),
	}

	return c
}

// Sample represents a sample in the MLWH combined with metadata taken from
// Google Sheets.
type Sample struct {
	mlwh.Sample
	sheets.MetaData
}

// ForSponsor returns all samples for the given sponsor where manual_qc is 1 and
// where there is corresponding metadata in our google sheet. It caches database
// queries, so results can be up to CacheLifetime old.
func (c *Client) ForSponsor(sponsor string) ([]Sample, error) {
	cached, samples, metadata := c.cache.getData(sponsor)

	if !cached {
		var err error

		samples, metadata, err = c.doLiveQueries(sponsor)
		if err != nil {
			return nil, err
		}

		c.cache.storeData(sponsor, samples, metadata)
	}

	result := make([]Sample, 0, len(metadata))

	for _, s := range samples {
		meta, ok := metadata[s.SampleName]
		if !ok {
			continue
		}

		result = append(result, newSample(s, meta))
	}

	return result, nil
}

func (c *Client) doLiveQueries(sponsor string) ([]mlwh.Sample, map[string]sheets.MetaData, error) {
	samples, err := c.mc.SamplesForSponsor(sponsor)
	if err != nil {
		return nil, nil, err
	}

	metadata, err := c.sc.DimSumMetaData(c.sheetID)
	if err != nil {
		return nil, nil, err
	}

	return samples, metadata, nil
}

func newSample(s mlwh.Sample, meta sheets.MetaData) Sample {
	return Sample{
		Sample: mlwh.Sample{
			SampleID:   s.SampleID,
			SampleName: s.SampleName,
			RunID:      s.RunID,
			StudyID:    s.StudyID,
			StudyName:  s.StudyName,
		},
		MetaData: sheets.MetaData{
			Selection: meta.Selection,
			Replicate: meta.Replicate,
			Time:      meta.Time,
			OD:        meta.OD,
			LibraryMetaData: sheets.LibraryMetaData{
				LibraryID:       meta.LibraryID,
				Wt:              meta.Wt,
				Cutadapt5First:  meta.Cutadapt5First,
				Cutadapt5Second: meta.Cutadapt5Second,
			},
		},
	}
}
