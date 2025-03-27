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

package dimsum

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wtsi-hgi/dimsum-automation/itl"
	"github.com/wtsi-hgi/dimsum-automation/samples"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrMultipleExperiments = Error("multiple experiments in samples")

	DefaultFastqExtension          = ".fastq"
	DefaultGzipped                 = true
	DefaultCutAdaptMinLength       = 100
	DefaultCutAdaptErrorRate       = 0.2
	DefaultCores                   = 4
	DefaultMaxSubstitutions        = 2
	DefaultMixedSubstitutions      = false
	DefaultMutagenesisType         = "random"
	DefaultRetainIntermediateFiles = true
	DefaultDesignPairDuplicates    = false

	pair1FastqSuffix       = "_1.fastq.gz"
	pair2FastqSuffix       = "_2.fastq.gz"
	experimentDesignPrefix = "dimsumDesign_"
	experimentDesignSuffix = ".txt"
	experimentDesignHeader = "sample_name\texperiment_replicate\tselection_id\tselection_replicate\t" +
		"technical_replicate\tpair1\tpair2\tgenerations\tcell_density\tselection_time\n"
	outputSubdir        = "outputs"
	cutAdaptRequired    = ":required..."
	cutAdaptOptional    = ":optional"
	dimsumProjectPrefix = "dimsumRun_"
)

// Experiment represents a single experiment's metadata.
type Experiment struct {
	ID          string  // ID is the unique identifier for the experiment.
	SampleName  string  // SampleName is the unique name for the sample.
	Replicate   int     // Replicate is the replicate number of the experiment.
	Selection   int     // Selection is the selection number of the experiment.
	Pair1       string  // Pair1 is the filename of the first pair of FASTQ files.
	Pair2       string  // Pair2 is the filename of the second pair of FASTQ files.
	CellDensity float32 // CellDensity is the cell density at the time of sampling.
	// Generations is the amount of times the cells have divided between 0.05 and the final cell density.
	Generations   float32
	SelectionTime float32 // SelectionTime is the selection time.
}

// SelectionReplicate converts the selection number to a replicate number.
func (e Experiment) SelectionReplicate() string {
	if e.Selection == 1 {
		return "1"
	}

	return ""
}

type ExperimentDesign []Experiment

// ID returns the ID of the first experiment in the design.
// It is assumed that all experiments in the design have the same ID.
// This is a precondition for the NewExperimentDesign function.
func (ed ExperimentDesign) ID() string {
	return ed[0].ID
}

// NewExperimentDesign creates an experiment design from the given samples.
// It returns an error if there are multiple experiments in the samples.
func NewExperimentDesign(samples []samples.Sample) (ExperimentDesign, error) {
	design := make(ExperimentDesign, 0, len(samples))
	experiments := make(map[string]int)

	for _, sample := range samples {
		fastqBasenamePrefix := itl.FastqBasenamePrefix(sample.SampleID, sample.RunID)

		exp := Experiment{
			ID:            sample.ExperimentID,
			SampleName:    sample.SampleName,
			Replicate:     sample.Replicate,
			Selection:     sample.Selection,
			Pair1:         fastqBasenamePrefix + itl.FastqPair1Suffix,
			Pair2:         fastqBasenamePrefix + itl.FastqPair2Suffix,
			CellDensity:   sample.OD,
			Generations:   sample.Generations(),
			SelectionTime: sample.Time,
		}

		design = append(design, exp)
		experiments[exp.ID]++
	}

	if len(experiments) > 1 {
		return nil, ErrMultipleExperiments
	}

	return design, nil
}

// Write writes an experiment design to a file that includes our ID in the
// basename in the given directory and returns the path to the file.
func (ed ExperimentDesign) Write(dir string) (string, error) {
	designPath := experimentDesignPath(dir, ed[0].ID)

	file, err := os.Create(designPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err = file.WriteString(experimentDesignHeader); err != nil {
		return "", err
	}

	for _, exp := range ed {
		line := fmt.Sprintf("%s\t%d\t%d\t%s\t%d\t%s\t%s\t%.0f\t%.3f\t%.1f\n",
			exp.SampleName, exp.Replicate, exp.Selection, exp.SelectionReplicate(), 1,
			exp.Pair1, exp.Pair2, exp.Generations, exp.CellDensity, exp.SelectionTime)

		if _, err = file.WriteString(line); err != nil {
			return "", err
		}
	}

	return designPath, nil
}

func experimentDesignPath(dir, experiment string) string {
	return filepath.Join(dir, experimentDesignPrefix+experiment+experimentDesignSuffix)
}

// DimSum represents the parameters for running DiMSum. All parameters are
type DimSum struct {
	// Required parameters
	Exe                     string // Path to the DiMSum executable
	FastqDir                string // Directory containing FASTQ files
	BarcodeIdentityPath     string // Path to the barcode identity file; can be blank
	Experiment              string // Name of the experiment
	VSearchMinQual          int    // Minimum quality score for VSearch
	StartStage              int    // Stage to start the analysis from
	FitnessMinInputCountAny int    // Minimum input count for any fitness calculation
	FitnessMinInputCountAll int    // Minimum input count for all fitness calculations

	// Optional parameters
	FastqExtension          string  // Extension of FASTQ files
	Gzipped                 bool    // Whether the FASTQ files are gzipped
	CutAdaptMinLength       int     // Minimum length for CutAdapt
	CutAdaptErrorRate       float32 // Error rate for CutAdapt
	Cores                   int     // Number of cores to use
	MaxSubstitutions        int     // Maximum number of substitutions allowed
	MixedSubstitutions      bool    // Whether mixed substitutions are allowed
	MutagenesisType         string  // Type of mutagenesis
	RetainIntermediateFiles bool    // Whether to retain intermediate files
	DesignPairDuplicates    bool    // Whether to design pair duplicates
}

// New creates a new DimSum instance with default values for the properties not
// supplied.
//
// Parameters:
//   - exe: Path to the DiMSum executable.
//   - fastqDir: Directory containing FASTQ files.
//   - barcodeIdentityPath: Path to the barcode identity file. This can be blank.
//   - experiment: Name of the experiment.
//   - vsearchMinQual: Minimum quality score for VSearch.
//   - startStage: Stage to start the analysis from.
//   - fitnessMinInputCountAny: Minimum input count for any fitness calculation.
//   - fitnessMinInputCountAll: Minimum input count for all fitness calculations.
func New(exe, fastqDir, barcodeIdentityPath, experiment string,
	vsearchMinQual, startStage, fitnessMinInputCountAny, fitnessMinInputCountAll int) DimSum {
	return DimSum{
		Exe:                     exe,
		FastqDir:                fastqDir,
		BarcodeIdentityPath:     barcodeIdentityPath,
		Experiment:              experiment,
		VSearchMinQual:          vsearchMinQual,
		StartStage:              startStage,
		FitnessMinInputCountAny: fitnessMinInputCountAny,
		FitnessMinInputCountAll: fitnessMinInputCountAll,

		FastqExtension:          DefaultFastqExtension,
		Gzipped:                 DefaultGzipped,
		CutAdaptMinLength:       DefaultCutAdaptMinLength,
		CutAdaptErrorRate:       DefaultCutAdaptErrorRate,
		Cores:                   DefaultCores,
		MaxSubstitutions:        DefaultMaxSubstitutions,
		MixedSubstitutions:      DefaultMixedSubstitutions,
		MutagenesisType:         DefaultMutagenesisType,
		RetainIntermediateFiles: DefaultRetainIntermediateFiles,
		DesignPairDuplicates:    DefaultDesignPairDuplicates,
	}
}

// Command generates the DiMSum command to execute.
//
// Parameters:
//   - dir: The directory where the output files will be stored.
//   - libMeta: Metadata for the library used in the experiment.
func (d *DimSum) Command(dir string, libMeta sheets.LibraryMetaData) string {
	cmd := fmt.Sprintf("%s -i %s -l %s -g %s -e %s --cutadapt5First %s --cutadapt5Second %s "+
		"-n %d -a %.2f -q %d -o %s -p %s -s %d -w %s -c %d "+
		"--fitnessMinInputCountAny %d --fitnessMinInputCountAll %d "+
		"--maxSubstitutions %d --mutagenesisType %s --retainIntermediateFiles %s "+
		"--mixedSubstitutions %s --experimentDesignPairDuplicates %s",
		d.Exe, d.FastqDir, d.FastqExtension, d.gzippedStr(), experimentDesignPath(dir, d.Experiment),
		libMeta.Cutadapt5First+cutAdaptRequired+
			reverseComplement(libMeta.Cutadapt5Second)+cutAdaptOptional,
		libMeta.Cutadapt5Second+cutAdaptRequired+
			reverseComplement(libMeta.Cutadapt5First)+cutAdaptOptional,
		d.CutAdaptMinLength, d.CutAdaptErrorRate,
		d.VSearchMinQual, filepath.Join(dir, outputSubdir), dimsumProjectPrefix+d.Experiment,
		d.StartStage, libMeta.Wt, d.Cores, d.FitnessMinInputCountAny,
		d.FitnessMinInputCountAll, d.MaxSubstitutions,
		d.MutagenesisType, d.retainIntermediateFilesStr(), d.mixedSubstitutionsStr(),
		d.designPairDuplicatesStr(),
	)

	if d.BarcodeIdentityPath != "" {
		cmd += " --barcodeIdentityPath " + d.BarcodeIdentityPath
	}

	return cmd
}

func (d *DimSum) gzippedStr() string {
	return boolToStr(d.Gzipped)
}

func boolToStr(b bool) string {
	if b {
		return "TRUE"
	}

	return "FALSE"
}

func (d *DimSum) retainIntermediateFilesStr() string {
	return boolToLetter(d.RetainIntermediateFiles)
}

func boolToLetter(b bool) string {
	return boolToStr(b)[0:1]
}

func (d *DimSum) mixedSubstitutionsStr() string {
	return boolToLetter(d.MixedSubstitutions)
}

func (d *DimSum) designPairDuplicatesStr() string {
	return boolToLetter(d.DesignPairDuplicates)
}

func reverseComplement(seq string) string {
	seq = strings.ToUpper(seq)
	result := make([]byte, len(seq))

	for i, j := 0, len(seq)-1; j >= 0; i, j = i+1, j-1 {
		switch seq[j] {
		case 'A':
			result[i] = 'T'
		case 'T':
			result[i] = 'A'
		case 'G':
			result[i] = 'C'
		case 'C':
			result[i] = 'G'
		default:
			result[i] = seq[j]
		}
	}

	return string(result)
}
