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
	"os"
	"strings"

	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("provide the path to credentials.json as an argument")
	}

	spreadsheetId := os.Getenv("DIMSUM_AUTOMATION_SPREADSHEETID")
	if spreadsheetId == "" {
		log.Fatal("set DIMSUM_AUTOMATION_SPREADSHEETID to sheet ID")
	}

	sc, err := sheets.ServiceCredentialsFromFile(os.Args[1])
	if err != nil {
		log.Fatalf("unable to load credentials: %v", err)
	}

	sheets, err := sheets.New(sc)
	if err != nil {
		log.Fatalf("unable to retrieve Sheets client: %v", err)
	}

	sheet, err := sheets.Read(spreadsheetId, "Libraries")
	if err != nil {
		log.Fatalf("unable to retrieve data from sheet: %v", err)
	}

	if len(sheet.Rows) == 0 {
		fmt.Println("no data found")
	} else {
		rows, err := sheet.Columns(
			"library_id",
			"dimsum_wt",
			"dimsum_cutadapt5First",
			"dimsum_cutadapt5Second",
		)
		if err != nil {
			log.Fatal(err)
		}

		for _, row := range rows {
			fmt.Printf("%s\n", strings.Join(row, ", "))
		}
	}
}
