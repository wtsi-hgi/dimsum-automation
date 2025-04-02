/*******************************************************************************
 * Copyright (c) 2025 Genome Research Ltd.
 *
 * Author: Sendu Bala <sb10@sanger.ac.uk>
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

package cmd

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wtsi-hgi/dimsum-automation/config"
	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/samples"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

const (
	sponsor       = "Ben Lehner"
	cacheLifetime = 10 * time.Minute
)

// infoCmd represents the info command.
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get sample info.",
	Long: `Get sample info from MLWH and the GenGen Google sheet.

This is a work in progress. It currently shows you the sample names in the
Google sheet, some of the sample names in MLWH, and all the merged details we
can extract for samples found in both.

You can then use the sample names and run IDs from merged details as input to
the run sub-commands.
`,
	Run: func(_ *cobra.Command, _ []string) {
		err := sampleInfo()
		if err != nil {
			die("%s", err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
}

func sampleInfo() error {
	c, err := config.FromEnv()
	if err != nil {
		return err
	}

	sc, err := sheets.ServiceCredentialsFromConfig(c)
	if err != nil {
		return err
	}

	sheets, err := sheets.New(sc)
	if err != nil {
		return err
	}

	db, err := mlwh.New(mlwh.MySQLConfigFromConfig(c))
	if err != nil {
		return err
	}

	metadata, err := sheets.DimSumMetaData(c.SheetID)
	if err != nil {
		return err
	}

	cliPrint("All samples names from google sheet:\n")

	sampleNames := make([]string, 0, len(metadata))

	// for sample, meta := range metadata {
	// 	fmt.Printf("%s, %d, %s, %s\n", sample, meta.Replicate, meta.LibraryID, meta.Cutadapt5First)
	// }

	for sample := range metadata {
		sampleNames = append(sampleNames, sample)
	}

	sort.Strings(sampleNames)
	cliPrint(strings.Join(sampleNames, ","))

	mlwhSamples, err := db.SamplesForSponsor(sponsor)
	if err != nil {
		return err
	}

	cliPrint("\n\nExample samples names found in MLWH:\n")

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
				cliPrintf("Sample %s appears %d times\n", key, count) // 2 of these
			}
		}

		for key, runMap := range diffRun {
			if len(runMap) > 1 {
				cliPrintf("Sample %s has %d different run IDs\n", key, len(runMap)) // many of these
			}
		}

		for key, studyMap := range diffStudy {
			if len(studyMap) > 1 {
				cliPrintf("Sample %s has %d different study IDs\n", key, len(studyMap)) // none of these
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

	cliPrint(strings.Join(subset, ","))

	cliPrint("\n\nMerged sample info:\n")

	client := samples.New(db, sheets, samples.ClientOptions{
		SheetID:       c.SheetID,
		CacheLifetime: cacheLifetime,
		Prefetch:      []string{sponsor},
	})

	defer client.Close()

	clientSamples, err := client.ForSponsor(sponsor)
	if err != nil {
		return err
	}

	for _, sample := range clientSamples {
		bytes, _ := json.MarshalIndent(sample, "", "  ") //nolint:errcheck,errchkjson
		cliPrint(string(bytes))
	}

	cliPrint("\n")

	// fmt.Printf("\nIf first sample above was selected, command lines would be:\n\n")

	// itl, err := itl.New(clientSamples[0:1])
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// cmd1, _ := itl.GenerateSamplesTSVCommand()
	// fmt.Printf(
	// 	"$ %s\n\n[\n and then the tsv output would be split in to per-desired sample-run files\n"+
	// 		" and then irods_to_lustre run on each to get the fastq files,\n"+
	// 		" which would be moved to a certain folder\n]\n\n",
	// 	cmd1,
	// )

	// design, err := dimsum.NewExperimentDesign(clientSamples[0:1])
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// dir := "./"

	// experimentPath, err := design.Write(dir)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fastqDir := "./fastq_dir"
	// exe := "/path/to/DiMSum"
	// vsearchMinQual := 20
	// startStage := 0
	// fitnessMinInputCountAny := 10
	// fitnessMinInputCountAll := 0
	// barcodeIdentityPath := "barcode_identity.txt"

	// d := dimsum.New(exe, fastqDir, barcodeIdentityPath, design.ID(), vsearchMinQual, startStage,
	// 	fitnessMinInputCountAny, fitnessMinInputCountAll)
	// cmd3 := d.Command(dir, clientSamples[0].LibraryMetaData)

	// fmt.Printf("$ %s\n\nNB: %s was created, but barcode_identity.txt is a placeholder\n", cmd3, experimentPath)

	return nil
}
