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
	"github.com/wtsi-hgi/dimsum-automation/types"
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
	// DimSumMetaData reads sheets "Libraries", "Experiments" and "Samples" from
	// the sheet with the given id and extracts metadata for columns relevant to
	// DimSum, returning a slice of Library that each contain a slice of their
	// Experiments, that each contain a slice of their Samples.
	DimSumMetaData(sheetID string) (types.Libraries, error)
}

type cache struct {
	libs       map[string]types.Libraries
	lastUpdate time.Time
	lifetime   time.Duration
	mu         sync.RWMutex
}

func newCache(lifetime time.Duration) *cache {
	return &cache{
		libs:     make(map[string]types.Libraries),
		lifetime: lifetime,
	}
}

func (c *cache) getData(sponsor string) (bool, types.Libraries) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached := c.lastUpdate.Add(c.lifetime).After(time.Now())
	data := c.libs[sponsor]

	return cached, data
}

func (c *cache) storeData(sponsor string, data types.Libraries) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.libs[sponsor] = data
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

// ForSponsor returns all libraries for the given sponsor that have experiements
// that have samples where manual_qc is 1 and where there is corresponding
// metadata in our google sheet. It caches database queries, so results can be
// up to CacheLifetime old.
//
// If you have prefetching enabled, this always returns immediately with the
// result of the last successful prefetch, which might have been longer than
// CacheLifetime ago, if the last actual prefetch failed (see Err()).
func (c *Client) ForSponsor(sponsor string) (types.Libraries, error) {
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

func (c *Client) freshForSponsorQuery(sponsor string) (types.Libraries, error) {
	samples, err := c.mc.SamplesForSponsor(sponsor)
	if err != nil {
		return nil, err
	}

	libs, err := c.sc.DimSumMetaData(c.sheetID)
	if err != nil {
		return nil, err
	}

	mlwhSampleLookup := make(map[string]int, len(samples))

	for i, s := range samples {
		mlwhSampleLookup[s.SampleName] = i
	}

	// TODO: apply mlwh sample and study metadata to libs, remove any libs and
	// experiments that don't have mlwh samples

	for _, lib := range libs {
		// goodExps := make([]*sheets.Experiment, 0, len(lib.Experiments))

		for _, exp := range lib.Experiments {
			// goodSamples := make([]*sheets.Sample, 0, len(exp.Samples))

			for _, sample := range exp.Samples {
				_, ok := mlwhSampleLookup[sample.SampleName]
				if !ok {
					continue
				}

				// mlwhSample := samples[i]
			}
		}
	}

	return libs, nil
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
