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
		"projectName",
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
		"experimentDesignPairDuplicates",
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

		ws := row[25]
		if ws == "" {
			ws = lib.WildtypeSequence
		}

		ms := lib.MaxSubstitutions
		if row[30] != "" {
			ms = c.ToInt(row[30])
		}

		exps[i] = &types.Experiment{
			ExperimentID:                   row[1],
			Assay:                          row[2],
			ProjectName:                    row[3],
			StartStage:                     c.ToInt(row[4]),
			StopStage:                      c.ToInt(row[5]),
			BarcodeDesignPath:              row[6],
			BarcodeErrorRate:               c.ToFloatString(row[7]),
			ExperimentDesignPairDuplicates: c.ToBool(row[8]),
			CountPath:                      row[9],
			BarcodeIdentityPath:            row[10],
			Cutadapt5First:                 row[11],
			Cutadapt5Second:                row[12],
			CutadaptMinLength:              c.ToInt(row[13]),
			CutadaptErrorRate:              c.ToFloatString(row[14]),
			CutadaptOverlap:                c.ToInt(row[15]),
			CutadaptCut5First:              row[16],
			CutadaptCut5Second:             row[17],
			CutadaptCut3First:              row[18],
			CutadaptCut3Second:             row[19],
			VsearchMinQual:                 c.ToInt(row[20]),
			VsearchMaxQual:                 c.ToInt(row[21]),
			VsearchMaxee:                   c.ToInt(row[22]),
			VsearchMinovlen:                c.ToInt(row[23]),
			ReverseComplement:              c.ToBool(row[24]),
			WildtypeSequence:               ws,
			PermittedSequences:             row[26],
			SequenceType:                   c.ToSequenceType(row[27]),
			MutagenesisType:                c.ToMutagenesisType(row[28]),
			Indels:                         row[29],
			MaxSubstitutions:               ms,
			MixedSubstitutions:             c.ToBool(row[31]),
			FitnessMinInputCountAll:        c.ToInt(row[32]),
			FitnessMinInputCountAny:        c.ToInt(row[33]),
			FitnessMinOutputCountAll:       c.ToInt(row[34]),
			FitnessMinOutputCountAny:       c.ToInt(row[35]),
			FitnessNormalise:               c.ToBool(row[36]),
			FitnessErrorModel:              c.ToBool(row[37]),
			FitnessDropoutPseudocount:      c.ToInt(row[38]),
			RetainedReplicates:             row[39],
			Stranded:                       c.ToBool(row[40]),
			Paired:                         c.ToBool(row[41]),
			SynonymSequencePath:            row[42],
			TransLibrary:                   c.ToBool(row[43]),
			TransLibraryReverseComplement:  c.ToBool(row[44]),
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
		"sample_id",
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
			SampleID:            row[1],
			Selection:           c.ToSelection(row[2]),
			ExperimentReplicate: c.ToInt(row[3]),
			SelectionTime:       c.ToFloatString(row[4]),
			CellDensity:         c.ToFloatString(row[5]),
			CellDensityFloat:    c.ToFloat(row[5]),
		}

		exp := exps[expI]
		exp.Samples = append(exp.Samples, samples[i])
	}

	return c.Err
}
