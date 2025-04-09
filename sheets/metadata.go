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

import "github.com/wtsi-hgi/dimsum-automation/types"

const (
	ErrNoData            = Error("no data found in sheet")
	ErrMissingLibrary    = Error("experiment's library not found in libraries sheet")
	ErrMissingExperiment = Error("sample's experiment not found in experiments sheet")
)

// DimSumMetaData reads sheets "Libraries", "Experiments" and "Samples" from the
// sheet with the given id and extracts metadata for columns relevant to DimSum,
// returning a slice of Library that each contain a slice of their Experiments,
// that each contain a slice of their Samples.
func (s *Sheets) DimSumMetaData(sheetID string) (types.Libraries, error) {
	libs, libLookup, err := s.getLibraryMetaData(sheetID)
	if err != nil {
		return nil, err
	}

	exps, expLookup, err := s.getExperimentMetaData(sheetID, libs, libLookup)
	if err != nil {
		return nil, err
	}

	err = s.getSampleMetaData(sheetID, exps, expLookup)
	if err != nil {
		return nil, err
	}

	return libs, nil
}

func (s *Sheets) getLibraryMetaData(sheetID string) (types.Libraries, map[string]int, error) { //nolint:funlen
	sheet, err := s.Read(sheetID, "libraries")
	if err != nil {
		return nil, nil, err
	}

	if len(sheet.Rows) == 0 {
		return nil, nil, ErrNoData
	}

	libRows, err := sheet.Columns(
		"library_id",
		"wildtypeSequence",
		"maxSubstitutions",
	)
	if err != nil {
		return nil, nil, err
	}

	libs := make(types.Libraries, len(libRows))
	lookup := make(map[string]int, len(libRows))

	c := converter{}

	for i, row := range libRows {
		libs[i] = &types.Library{
			LibraryID:        row[0],
			WildtypeSequence: row[1],
			MaxSubstitutions: c.ToInt(row[2]),
		}

		lookup[row[0]] = i
	}

	return libs, lookup, c.Err
}

func (s *Sheets) getExperimentMetaData( //nolint:gocognit,gocyclo,funlen
	sheetID string, libs types.Libraries, libLookup map[string]int,
) ([]*types.Experiment, map[string]int, error) {
	sheet, err := s.Read(sheetID, "experiments")
	if err != nil {
		return nil, nil, err
	}

	if len(sheet.Rows) == 0 {
		return nil, nil, ErrNoData
	}

	expRows, err := sheet.Columns(
		"library_id",
		"experiment_id",
		"Assay",
		"startStage",
		"stopStage",
		"barcodeDesignPath",
		"barcodeErrorRate",
		"experimentDesignPairDuplicates",
		"countPath",
		"barcodeIdentityPath",
		"cutadapt5First",
		"cutadapt5Second",
		"cutadaptMinLength",
		"cutadaptErrorRate",
		"cutadaptOverlap",
		"cutadaptCut5First",
		"cutadaptCut5Second",
		"cutadaptCut3First",
		"cutadaptCut3Second",
		"vsearchMinQual",
		"vsearchMaxQual",
		"vsearchMaxee",
		"vsearchMinovlen",
		"reverseComplement",
		"wildtypeSequence",
		"permittedSequences",
		"sequenceType",
		"mutagenesisType",
		"indels",
		"maxSubstitutions",
		"mixedSubstitutions",
		"fitnessMinInputCountAll",
		"fitnessMinInputCountAny",
		"fitnessMinOutputCountAll",
		"fitnessMinOutputCountAny",
		"fitnessNormalise",
		"fitnessErrorModel",
		"fitnessDropoutPseudocount",
		"retainedReplicates",
		"stranded",
		"paired",
		"synonymSequencePath",
		"transLibrary",
		"transLibraryReverseComplement",
	)
	if err != nil {
		return nil, nil, err
	}

	exps := make([]*types.Experiment, len(expRows))
	lookup := make(map[string]int, len(expRows))

	c := converter{}

	for i, row := range expRows {
		libI, ok := libLookup[row[0]]
		if !ok {
			return nil, nil, ErrMissingLibrary
		}

		lib := libs[libI]

		ws := row[24]
		if ws == "" {
			ws = lib.WildtypeSequence
		}

		ms := lib.MaxSubstitutions
		if row[29] != "" {
			ms = c.ToInt(row[29])
		}

		exps[i] = &types.Experiment{
			ExperimentID:                   row[1],
			Assay:                          row[2],
			StartStage:                     c.ToInt(row[3]),
			StopStage:                      c.ToInt(row[4]),
			BarcodeDesignPath:              row[5],
			BarcodeErrorRate:               c.ToFloatString(row[6]),
			ExperimentDesignPairDuplicates: c.ToBool(row[7]),
			CountPath:                      row[8],
			BarcodeIdentityPath:            row[9],
			Cutadapt5First:                 row[10],
			Cutadapt5Second:                row[11],
			CutadaptMinLength:              c.ToInt(row[12]),
			CutadaptErrorRate:              c.ToFloatString(row[13]),
			CutadaptOverlap:                c.ToInt(row[14]),
			CutadaptCut5First:              row[15],
			CutadaptCut5Second:             row[16],
			CutadaptCut3First:              row[17],
			CutadaptCut3Second:             row[18],
			VsearchMinQual:                 c.ToInt(row[19]),
			VsearchMaxQual:                 c.ToInt(row[20]),
			VsearchMaxee:                   c.ToInt(row[21]),
			VsearchMinovlen:                c.ToInt(row[22]),
			ReverseComplement:              c.ToBool(row[23]),
			WildtypeSequence:               ws,
			PermittedSequences:             row[25],
			SequenceType:                   c.ToSequenceType(row[26]),
			MutagenesisType:                c.ToMutagenesisType(row[27]),
			Indels:                         row[28],
			MaxSubstitutions:               ms,
			MixedSubstitutions:             c.ToBool(row[30]),
			FitnessMinInputCountAll:        c.ToInt(row[31]),
			FitnessMinInputCountAny:        c.ToInt(row[32]),
			FitnessMinOutputCountAll:       c.ToInt(row[33]),
			FitnessMinOutputCountAny:       c.ToInt(row[34]),
			FitnessNormalise:               c.ToBool(row[35]),
			FitnessErrorModel:              c.ToBool(row[36]),
			FitnessDropoutPseudocount:      c.ToInt(row[37]),
			RetainedReplicates:             row[38],
			Stranded:                       c.ToBool(row[39]),
			Paired:                         c.ToBool(row[40]),
			SynonymSequencePath:            row[41],
			TransLibrary:                   c.ToBool(row[42]),
			TransLibraryReverseComplement:  c.ToBool(row[43]),
		}

		lib.Experiments = append(lib.Experiments, exps[i])

		lookup[row[1]] = i
	}

	return exps, lookup, c.Err
}

func (s *Sheets) getSampleMetaData( //nolint:funlen
	sheetID string, exps []*types.Experiment, expLookup map[string]int) error {
	sheet, err := s.Read(sheetID, "samples")
	if err != nil {
		return err
	}

	if len(sheet.Rows) == 0 {
		return ErrNoData
	}

	sampleRows, err := sheet.Columns(
		"experiment_id",
		"mlwh_sample_name",
		"selection",
		"experiment_replicate",
		"selection_time",
		"cell_density",
	)
	if err != nil {
		return err
	}

	samples := make([]*types.Sample, len(sampleRows))

	c := converter{}

	for i, row := range sampleRows {
		expI, ok := expLookup[row[0]]
		if !ok {
			return ErrMissingExperiment
		}

		samples[i] = &types.Sample{
			SampleName:          row[1],
			Selection:           c.ToSelection(row[2]),
			ExperimentReplicate: c.ToInt(row[3]),
			SelectionTime:       c.ToFloatString(row[4]),
			CellDensity:         c.ToFloatString(row[5]),
		}

		exp := exps[expI]
		exp.Samples = append(exp.Samples, samples[i])
	}

	return c.Err
}
