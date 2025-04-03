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

package itl

import (
	"os"
	"strings"
)

const (
	ErrNoSamplesFound = Error("no matching samples found in TSV file")
)

func createPerSampleRunTSV(inputTSVPath string, sr sampleRun) (string, error) {
	data, err := os.ReadFile(inputTSVPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return "", ErrNoSamplesFound
	}

	filteredLines, err := filterLinesForSampleRun(lines, sr)
	if err != nil {
		return "", err
	}

	outPath := sr.TSVPath()

	return outPath, writeFilteredTSV(outPath, filteredLines)
}

func filterLinesForSampleRun(lines []string, sr sampleRun) ([]string, error) {
	if len(lines) == 0 {
		return nil, ErrNoSamplesFound
	}

	header := lines[0]
	dataLines := lines[1:]

	matchingLines := filterMatchingSampleRuns(dataLines, sr)

	if len(matchingLines) == 0 {
		return nil, ErrNoSamplesFound
	}

	result := append([]string{header}, matchingLines...)

	return result, nil
}

func filterMatchingSampleRuns(lines []string, sr sampleRun) []string {
	var matchingLines []string

	for _, line := range lines {
		if isMatchingSampleRun(line, sr) {
			matchingLines = append(matchingLines, line)
		}
	}

	return matchingLines
}

func isMatchingSampleRun(line string, sr sampleRun) bool {
	if line == "" {
		return false
	}

	fields := strings.Split(line, "\t")
	if len(fields) < minTSVColumns {
		return false
	}

	return fields[1] == sr.sampleID && fields[3] == sr.runID
}

func writeFilteredTSV(outPath string, filteredLines []string) error {
	output := strings.Join(filteredLines, "\n") + "\n"

	return os.WriteFile(outPath, []byte(output), userPerm)
}
