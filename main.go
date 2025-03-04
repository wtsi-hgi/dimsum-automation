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

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wtsi-hgi/dimsum-automation/config"
	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/samples"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

const (
	sponsor       = "Ben Lehner"
	cacheLifetime = 1 * time.Minute
)

func main() {
	c, err := config.FromEnv()
	if err != nil {
		log.Fatal(err)
	}

	sc, err := sheets.ServiceCredentialsFromConfig(c)
	if err != nil {
		log.Fatalf("unable to load credentials: %v", err)
	}

	sheets, err := sheets.New(sc)
	if err != nil {
		log.Fatalf("unable to retrieve Sheets client: %v", err)
	}

	metadata, err := sheets.DimSumMetaData(c.SheetID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("All samples from google sheet (sample name, replicate, library id, cutadapt5first):\n")

	for sample, meta := range metadata {
		fmt.Printf("%s, %d, %s, %s\n", sample, meta.Replicate, meta.LibraryID, meta.Cutadapt5First)
	}

	db, err := mlwh.New(mlwh.MySQLConfigFromConfig(c))
	if err != nil {
		log.Fatalf("unable to connect to MLWH: %v", err)
	}

	mlwhSamples, err := db.SamplesForSponsor(sponsor)
	if err != nil {
		log.Fatalf("unable to get samples: %v", err)
	}

	fmt.Printf("\nExample samples found in MLWH (sample name, id, study name):\n")

	for _, sample := range mlwhSamples[0:5] {
		fmt.Printf("%s, %s, %s\n", sample.SampleName, sample.SampleID, sample.StudyName)
	}

	fmt.Printf("\nMerged sample info:\n")

	client := samples.New(db, sheets, samples.ClientOptions{
		SheetID: c.SheetID, CacheLifetime: cacheLifetime,
	})

	clientSamples, err := client.ForSponsor(sponsor)
	if err != nil {
		log.Fatalf("unable to get samples: %v", err)
	}

	for _, sample := range clientSamples {
		fmt.Printf("%s, %s, %s, %d, %s, %s\n",
			sample.SampleName, sample.SampleID, sample.StudyName,
			sample.Replicate, sample.LibraryID, sample.Cutadapt5First)
	}
}
