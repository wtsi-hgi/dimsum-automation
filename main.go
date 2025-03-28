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
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/wtsi-hgi/dimsum-automation/config"
	"github.com/wtsi-hgi/dimsum-automation/dimsum"
	"github.com/wtsi-hgi/dimsum-automation/itl"
	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/samples"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

const (
	sponsor       = "Ben Lehner"
	cacheLifetime = 10 * time.Minute
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

	db, err := mlwh.New(mlwh.MySQLConfigFromConfig(c))
	if err != nil {
		log.Fatalf("unable to connect to MLWH: %v", err)
	}

	metadata, err := sheets.DimSumMetaData(c.SheetID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("All samples names from google sheet:\n")

	sampleNames := make([]string, 0, len(metadata))

	// for sample, meta := range metadata {
	// 	fmt.Printf("%s, %d, %s, %s\n", sample, meta.Replicate, meta.LibraryID, meta.Cutadapt5First)
	// }

	for sample := range metadata {
		sampleNames = append(sampleNames, sample)
	}

	sort.Strings(sampleNames)
	fmt.Println(strings.Join(sampleNames, ","))

	mlwhSamples, err := db.SamplesForSponsor(sponsor)
	if err != nil {
		log.Fatalf("unable to get samples: %v", err)
	}

	fmt.Printf("\nExample samples names found in MLWH:\n")

	sampleNames = make([]string, 0, len(mlwhSamples))

	// for _, sample := range mlwhSamples[0:5] {
	// 	fmt.Printf("%s, %s, %s\n", sample.SampleName, sample.SampleID, sample.StudyName)
	// }

	allSame := make(map[string]int)
	diffRun := make(map[string]map[string]bool)
	diffStudy := make(map[string]map[string]bool)

	for _, sample := range mlwhSamples {
		allSame[sample.SampleName+":"+sample.RunID+":"+sample.StudyID]++

		key := sample.SampleName + ":" + sample.StudyID
		if diffRun[key] == nil {
			diffRun[key] = make(map[string]bool)
		}

		diffRun[key][sample.RunID] = true

		key = sample.SampleName + ":" + sample.RunID
		if diffStudy[key] == nil {
			diffStudy[key] = make(map[string]bool)
		}

		diffStudy[key][sample.StudyID] = true

		sampleNames = append(sampleNames, sample.SampleName)
	}

	if false { //nolint:nestif
		for key, count := range allSame {
			if count > 1 {
				fmt.Printf("Sample %s appears %d times\n", key, count) // 2 of these
			}
		}

		for key, runMap := range diffRun {
			if len(runMap) > 1 {
				fmt.Printf("Sample %s has %d different run IDs\n", key, len(runMap)) // many of these
			}
		}

		for key, studyMap := range diffStudy {
			if len(studyMap) > 1 {
				fmt.Printf("Sample %s has %d different study IDs\n", key, len(studyMap)) // none of these
			}
		}
	}

	sort.Strings(sampleNames)

	subset := make([]string, 0, len(mlwhSamples))

	for i, sample := range sampleNames {
		if i%10 == 0 {
			subset = append(subset, sample)
		}
	}

	fmt.Println(strings.Join(subset, ","))

	fmt.Printf("\nMerged sample info:\n")

	client := samples.New(db, sheets, samples.ClientOptions{
		SheetID:       c.SheetID,
		CacheLifetime: cacheLifetime,
		Prefetch:      []string{sponsor},
	})

	defer client.Close()

	clientSamples, err := client.ForSponsor(sponsor)
	if err != nil {
		log.Fatalf("unable to get samples: %v", err)
	}

	for _, sample := range clientSamples {
		bytes, _ := json.MarshalIndent(sample, "", "  ") //nolint:errcheck,errchkjson
		fmt.Println(string(bytes))
	}

	fmt.Printf("\nIf first sample above was selected, command lines would be:\n\n")

	itl, err := itl.New(clientSamples[0:1])
	if err != nil {
		log.Fatal(err)
	}

	cmd1, _ := itl.GenerateSamplesTSVCommand()
	fmt.Printf(
		"$ %s\n\n[\n and then the tsv output would be split in to per-desired sample-run files\n"+
			" and then irods_to_lustre run on each to get the fastq files,\n"+
			" which would be moved to a certain folder\n]\n\n",
		cmd1,
	)

	design, err := dimsum.NewExperimentDesign(clientSamples[0:1])
	if err != nil {
		log.Fatal(err)
	}

	dir := "./"

	experimentPath, err := design.Write(dir)
	if err != nil {
		log.Fatal(err)
	}

	fastqDir := "./fastq_dir"
	exe := "/path/to/DiMSum"
	vsearchMinQual := 20
	startStage := 0
	fitnessMinInputCountAny := 10
	fitnessMinInputCountAll := 0
	barcodeIdentityPath := "barcode_identity.txt"

	d := dimsum.New(exe, fastqDir, barcodeIdentityPath, design.ID(), vsearchMinQual, startStage,
		fitnessMinInputCountAny, fitnessMinInputCountAll)
	cmd3 := d.Command(dir, clientSamples[0].LibraryMetaData)

	fmt.Printf("$ %s\n\nNB: %s was created, but barcode_identity.txt is a placeholder\n", cmd3, experimentPath)
}
