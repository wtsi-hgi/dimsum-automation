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
	"time"

	"github.com/spf13/cobra"
	"github.com/wtsi-hgi/dimsum-automation/config"
	"github.com/wtsi-hgi/dimsum-automation/mlwh"
	"github.com/wtsi-hgi/dimsum-automation/samples"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
	"github.com/wtsi-hgi/dimsum-automation/types"
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
			die(err)
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

	db, sheets, err := getDBAndSheets(c)
	if err != nil {
		return err
	}

	cliPrint("Extracted library => experiement => sample info:\n")

	libs, err := sponsorLibs(c, db, sheets)
	if err != nil {
		return err
	}

	for _, lib := range libs {
		bytes, _ := json.MarshalIndent(lib, "", "  ") //nolint:errcheck,errchkjson
		cliPrint(string(bytes))
	}

	cliPrint("\n")

	return nil
}

func getDBAndSheets(c *config.Config) (*mlwh.MLWH, *sheets.Sheets, error) {
	db, err := mlwh.New(mlwh.MySQLConfigFromConfig(c))
	if err != nil {
		return nil, nil, err
	}

	sc, err := sheets.ServiceCredentialsFromConfig(c)
	if err != nil {
		return nil, nil, err
	}

	s, err := sheets.New(sc)
	if err != nil {
		return nil, nil, err
	}

	return db, s, err
}

func sponsorLibs(c *config.Config, db *mlwh.MLWH, s *sheets.Sheets) (types.Libraries, error) {
	client := samples.New(db, s, samples.ClientOptions{
		SheetID:       c.SheetID,
		CacheLifetime: cacheLifetime,
		Prefetch:      []string{sponsor},
	})

	defer client.Close()

	return client.ForSponsor(sponsor)
}
