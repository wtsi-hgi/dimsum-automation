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
}

// Read retrieves the contents of a given document and sheet within that
// document. The id of a Google sheet is the long string of characters in the
// URL when viewing that document.
func (s *Sheets) Read(docID, sheetName string) (*Sheet, error) {
	valRange, err := s.srv.Spreadsheets.Values.Get(docID, sheetName).Do()
	if err != nil {
		return nil, err
	}

	if len(valRange.Values) == 0 {
		return nil, nil
	}

	var header []string

	rows := make([][]string, len(valRange.Values)-1)

	for i, row := range valRange.Values {
		if i == 0 {
			header = rowToStringSlice(row)
		} else {
			rows[i-1] = rowToStringSlice(row)
		}
	}

	return &Sheet{
		ColumnHeaders: header,
		Rows:          rows,
	}, nil
}

func rowToStringSlice(in []any) []string {
	out := make([]string, len(in))

	for i, cols := range in {
		out[i] = fmt.Sprint(cols)
	}

	return out
}
