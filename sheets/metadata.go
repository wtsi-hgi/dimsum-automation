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
	"math"
	"strconv"
)

const (
	ErrNoData         = Error("no data found in sheet")
	ErrMissingLibrary = Error("sample library not found in Libraries sheet")

	generationsMin = 0.05
)

// LibraryMetaData holds library metadata needed by DimSum.
type LibraryMetaData struct {
	LibraryID       string
	ExperimentID    string
	Wt              string
	Cutadapt5First  string
	Cutadapt5Second string
}

// MetaData holds information needed by DimSum for a sample.
type MetaData struct {
	Selection int
	Replicate int
	Time      float32
	OD        float32
	LibraryMetaData
}

// Generations is the amount of times the cells have divided between 0.05 and
// the OD, ie. log2(OD/0.05).
func (m MetaData) Generations() float32 {
	if m.OD == 0 {
		return 0
	}

	return float32(math.Log2(float64(m.OD / generationsMin)))
}

// DimSumMetaData reads sheets "Libraries" and "Samples" from the sheet with the
// given id and merges the results for columns relevant to DimSum, returning a
// map where keys are sample_id.
func (s *Sheets) DimSumMetaData(sheetID string) (map[string]MetaData, error) {
	libMeta, err := s.getLibraryMetaData(sheetID)
	if err != nil {
		return nil, err
	}

	sheet, err := s.Read(sheetID, "Samples")
	if err != nil {
		return nil, err
	}

	if len(sheet.Rows) == 0 {
		return nil, ErrNoData
	}

	sampleRows, err := sheet.Columns(
		"sample_id",
		"selection",
		"replicate",
		"time",
		"OD",
		"library_id",
	)
	if err != nil {
		return nil, err
	}

	m := make(map[string]MetaData, len(sampleRows))

	for _, row := range sampleRows {
		lib, ok := libMeta[row[5]]
		if !ok {
			return nil, ErrMissingLibrary
		}

		selection, replicate, dimsumTime, od, err := stringsToNumbers(row[1:5])
		if err != nil {
			return nil, err
		}

		m[row[0]] = MetaData{
			Selection:       selection,
			Replicate:       replicate,
			Time:            dimsumTime,
			OD:              od,
			LibraryMetaData: lib,
		}
	}

	return m, nil
}

func (s *Sheets) getLibraryMetaData(sheetID string) (map[string]LibraryMetaData, error) {
	sheet, err := s.Read(sheetID, "Libraries")
	if err != nil {
		return nil, err
	}

	if len(sheet.Rows) == 0 {
		return nil, ErrNoData
	}

	libRows, err := sheet.Columns(
		"library_id",
		"experiment_id",
		"dimsum_wt",
		"dimsum_cutadapt5First",
		"dimsum_cutadapt5Second",
	)
	if err != nil {
		return nil, err
	}

	m := make(map[string]LibraryMetaData, len(libRows))

	for _, row := range libRows {
		m[row[1]] = LibraryMetaData{
			LibraryID:       row[0],
			ExperimentID:    row[1],
			Wt:              row[2],
			Cutadapt5First:  row[3],
			Cutadapt5Second: row[4],
		}
	}

	return m, nil
}

func stringsToNumbers(numStrs []string) (selection int, replicate int, dimsumTime float32, od float32, err error) {
	selection, err = strconv.Atoi(numStrs[0])
	if err != nil {
		return
	}

	replicate, err = strconv.Atoi(numStrs[1])
	if err != nil {
		return
	}

	var val64 float64

	if numStrs[2] != "" {
		val64, err = strconv.ParseFloat(numStrs[2], 32)
		if err != nil {
			return
		}

		dimsumTime = float32(val64)
	}

	val64, err = strconv.ParseFloat(numStrs[3], 32)
	if err != nil {
		return
	}

	od = float32(val64)

	return
}
