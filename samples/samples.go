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

const (
	ErrInvalidNameRun   = Error("both name and run must be set")
	ErrNoNameRun        = Error("no name and run provided")
	ErrNameRunsNotFound = Error("no samples found for given names and runs")
)

type MLWHClient interface {
	// SamplesForSponsor returns all samples for the given sponsor, including
	// study and run information.
	SamplesForSponsor(sponsor string) ([]mlwh.Sample, error)

	// Close closes the connection to the MLWH database.
	Close() error
}

type SheetsClient interface {
	// DimSumMetaData reads sheets "Libraries" and "Samples" from the sheet with
	// the given id and merges the results for columns relevant to DimSum,
	// returning a map where keys are sample_id.
	DimSumMetaData(sheetID string) (map[string]sheets.MetaData, error)
}

type cache struct {
	samples    map[string][]Sample
	lastUpdate time.Time
	lifetime   time.Duration
	mu         sync.RWMutex
}

func newCache(lifetime time.Duration) *cache {
	return &cache{
		samples:  make(map[string][]Sample),
		lifetime: lifetime,
	}
}

func (c *cache) getData(sponsor string) (bool, []Sample) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached := c.lastUpdate.Add(c.lifetime).After(time.Now())
	data := c.samples[sponsor]

	return cached, data
}

func (c *cache) storeData(sponsor string, data []Sample) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.samples[sponsor] = data
	c.lastUpdate = time.Now()
}

func (c *cache) lastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lastUpdate
}

// Client can connect to MLWH and Google Sheets to get sample information.
type Client struct {
	mc      MLWHClient
	sc      SheetsClient
	sheetID string
	cache   *cache

	stopCh chan struct{}
	stopMu sync.RWMutex

	err   error
	errMu sync.RWMutex
}

// ClientOptions are options for creating a new Client.
type ClientOptions struct {
	// SheetID is the id of the google sheet to get metadata from.
	SheetID string

	// CacheLifetime is the maximum age of cached results.
	CacheLifetime time.Duration

	// Prefetch fetches ForSponsor() results for the given sponsors every
	// CacheLifetime so that you never have to wait for a query and they're as
	// fresh as possible. Errors are not returned, but can be checked with
	// Err().
	Prefetch []string
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

	if len(opts.Prefetch) > 0 && opts.CacheLifetime > 0 {
		c.asyncForSponsors(opts.Prefetch)
		go c.prefetch(opts.CacheLifetime, opts.Prefetch)
	}

	return c
}

func (c *Client) asyncForSponsors(sponsors []string) {
	for _, sponsor := range sponsors {
		result, err := c.freshForSponsorQuery(sponsor)

		c.errMu.Lock()
		c.err = err
		c.errMu.Unlock()

		if err != nil {
			return
		}

		c.cache.storeData(sponsor, result)
	}
}

func (c *Client) prefetch(sleepTime time.Duration, sponsors []string) {
	c.stopMu.Lock()
	stopCh := make(chan struct{})
	c.stopCh = stopCh
	c.stopMu.Unlock()

	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.asyncForSponsors(sponsors)
		case <-stopCh:
			return
		}
	}
}

// Err returns the last error that occurred during prefetching (ie. errors from
// calling ForSponsor() in the background). Successful prefetches clear the
// error.
func (c *Client) Err() error {
	c.errMu.RLock()
	defer c.errMu.RUnlock()

	return c.err
}

// LastPrefetchSuccess returns the time of the last successful prefetch. If no
// prefetch has succeeded yet, the zero time is returned.
func (c *Client) LastPrefetchSuccess() time.Time {
	return c.cache.lastUpdated()
}

// Sample represents a sample in the MLWH combined with metadata taken from
// Google Sheets.
type Sample struct {
	mlwh.Sample
	sheets.MetaData
}

// NameRun lets you specify a sample name and run id, for filtering Samples.
type NameRun struct {
	Name string
	Run  string
}

// Samples is a slice of Sample, from which you can get a subset based on
// NameRuns.
type Samples []Sample

// Filter returns a subset of the samples that match the given names and runs.
// Returns an error if not all NameRuns are found in the samples, or no valid
// NameRuns are provided.
func (s Samples) Filter(nameRuns []NameRun) (Samples, error) {
	nrMap := make(map[string]bool, len(nameRuns))

	for _, nr := range nameRuns {
		if nr.Name == "" || nr.Run == "" {
			return nil, ErrInvalidNameRun
		}

		nrMap[nr.Name+"."+nr.Run] = true
	}

	if len(nrMap) == 0 {
		return nil, ErrNoNameRun
	}

	if len(nrMap) > len(s) {
		return nil, ErrNameRunsNotFound
	}

	result := make(Samples, 0, len(nrMap))

	for _, sample := range s {
		key := sample.SampleName + "." + sample.RunID
		if nrMap[key] {
			result = append(result, sample)
			delete(nrMap, key)
		}
	}

	if len(nrMap) != 0 {
		return nil, ErrNameRunsNotFound
	}

	return result, nil
}

// ForSponsor returns all samples for the given sponsor where manual_qc is 1 and
// where there is corresponding metadata in our google sheet. It caches database
// queries, so results can be up to CacheLifetime old.
//
// If you have prefetching enabled, this always returns immediately with the
// result of the last successful prefetch, which might have been longer than
// CacheLifetime ago, if the last actual prefetch failed (see Err()).
func (c *Client) ForSponsor(sponsor string) (Samples, error) {
	cached, result := c.cache.getData(sponsor)

	c.stopMu.RLock()
	stopCh := c.stopCh
	c.stopMu.RUnlock()

	if !cached && stopCh == nil {
		var err error

		result, err = c.freshForSponsorQuery(sponsor)
		if err != nil {
			return nil, err
		}

		c.cache.storeData(sponsor, result)
	}

	return result, nil
}

func (c *Client) freshForSponsorQuery(sponsor string) ([]Sample, error) {
	samples, err := c.mc.SamplesForSponsor(sponsor)
	if err != nil {
		return nil, err
	}

	metadata, err := c.sc.DimSumMetaData(c.sheetID)
	if err != nil {
		return nil, err
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

func newSample(s mlwh.Sample, meta sheets.MetaData) Sample {
	return Sample{
		Sample: mlwh.Sample{
			SampleID:   s.SampleID,
			SampleName: s.SampleName,
			RunID:      s.RunID,
			StudyID:    s.StudyID,
			StudyName:  s.StudyName,
			ManualQC:   s.ManualQC,
		},
		MetaData: sheets.MetaData{
			Selection: meta.Selection,
			Replicate: meta.Replicate,
			Time:      meta.Time,
			OD:        meta.OD,
			LibraryMetaData: sheets.LibraryMetaData{
				LibraryID:        meta.LibraryID,
				ExperimentID:     meta.ExperimentID,
				Wt:               meta.Wt,
				Cutadapt5First:   meta.Cutadapt5First,
				Cutadapt5Second:  meta.Cutadapt5Second,
				MaxSubstitutions: meta.MaxSubstitutions,
			},
		},
	}
}

// Close closes database connections and stops prefetching.
func (c *Client) Close() error {
	err := c.mc.Close()

	c.stopMu.Lock()
	defer c.stopMu.Unlock()

	if c.stopCh != nil {
		close(c.stopCh)
		c.stopCh = nil
	}

	return err
}
