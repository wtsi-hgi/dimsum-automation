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

package sheets

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	googleSheets "google.golang.org/api/sheets/v4"
)

type Error string

func (e Error) Error() string { return string(e) }

const ErrColumnNotFound = Error("column not found in sheet")

// Sheets allows the retrival of sheets from Google docs.
type Sheets struct {
	srv *googleSheets.Service
}

// New returns a Sheets that you can Get() sheets from Google docs with.
func New(sc *ServiceCredentials) (*Sheets, error) {
	ctx := context.Background()
	client := sc.toJWTConfig().Client(ctx)

	srv, err := googleSheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &Sheets{srv: srv}, nil
}

// Sheet contains the retrieved cells in a Google sheet.
type Sheet struct {
	ColumnHeaders []string
	Rows          [][]string

	headerLookup map[string]int
}

// Read retrieves the contents of a given document and sheet within that
// document. The id of a Google sheet is the long string of characters in the
// URL when viewing that document.
func (s *Sheets) Read(sheetID, sheetName string) (*Sheet, error) {
	valRange, err := s.srv.Spreadsheets.Values.Get(sheetID, sheetName).Do()
	if err != nil {
		return nil, err
	}

	if len(valRange.Values) == 0 {
		return nil, nil
	}

	var header []string

	headerLookup := make(map[string]int, len(valRange.Values[0]))
	rows := make([][]string, len(valRange.Values)-1)

	for i, row := range valRange.Values {
		if i == 0 {
			header = rowToStringSlice(row)
			for i, head := range header {
				headerLookup[head] = i
			}
		} else {
			rows[i-1] = rowToStringSlice(row)
		}
	}

	return &Sheet{
		ColumnHeaders: header,
		Rows:          rows,
		headerLookup:  headerLookup,
	}, nil
}

func rowToStringSlice(in []any) []string {
	out := make([]string, len(in))

	for i, cols := range in {
		out[i] = fmt.Sprint(cols)
	}

	return out
}

// Columns returns a slice for each row in the sheet (like Rows property), but
// each slice only has values from the columns with the given column headers.
//
// Will return an error if given cols are not amongst ColumnHeaders.
func (s *Sheet) Columns(cols ...string) ([][]string, error) {
	colIndexes := make([]int, len(cols))

	for i, col := range cols {
		colIndex, ok := s.headerLookup[col]
		if !ok {
			return nil, ErrColumnNotFound
		}

		colIndexes[i] = colIndex
	}

	rows := make([][]string, len(s.Rows))

	for i, wholeRow := range s.Rows {
		row := make([]string, len(colIndexes))

		for j, colIndex := range colIndexes {
			row[j] = wholeRow[colIndex]
		}

		rows[i] = row
	}

	return rows, nil
}
