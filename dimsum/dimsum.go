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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wtsi-hgi/dimsum-automation/types"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrMultipleExperiments = Error("multiple experiments in samples")

	DefaultVsearchMinQual          = 20
	DefaultStartStage              = 0
	DefaultFitnessMinInputCountAny = 10
	DefaultFitnessMinInputCountAll = 0

	DimSumExe                      = "DiMSum"
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

type Row struct {
	types.Sample
}

type Rows []Row

// ExperimentDesign represents a single experiment's metadata.
type ExperimentDesign struct {
	*types.Experiment
	Rows
}

// NewExperimentDesign creates an experiment design from the Experiment.
func NewExperimentDesign(exp *types.Experiment) (ExperimentDesign, error) {
	// fastqBasenamePrefix := itl.FastqBasenamePrefix(sample.SampleName, sample.RunID)
	// Pair1:           fastqBasenamePrefix + itl.FastqPair1Suffix,
	// Pair2:           fastqBasenamePrefix + itl.FastqPair2Suffix,
	// Generations:     sample.Generations(),

	// TODO: form rows from the samples in exp

	return ExperimentDesign{
		Experiment: exp,
	}, nil
}

// Write writes an experiment design to a file that includes our ID in the
// basename in the given directory and returns the path to the file.
func (ed ExperimentDesign) Write(dir string) (string, error) {
	designPath := experimentDesignPath(dir, ed.ExperimentID)

	file, err := os.Create(designPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err = file.WriteString(experimentDesignHeader); err != nil {
		return "", err
	}

	for _, row := range ed.Rows {
		line := fmt.Sprintf("%s\t%d\t%d\t%s\t%d\t%s\t%s\t%.0f\t%s\t%s\n",
			row.DimsumSampleName(), row.ExperimentReplicate, row.SelectionID(), row.SelectionReplicate(), 1, //TODO: technical replicate
			"TODO: Pair1", "TODO: Pair2", row.Generations(), row.CellDensity, row.SelectionTime)

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
	FastqDir                string // Directory containing FASTQ files
	BarcodeIdentityPath     string // Path to the barcode identity file; can be blank
	Experiment              string // Name of the experiment
	VSearchMinQual          int    // Minimum quality score for VSearch
	StartStage              int    // Stage to start the analysis from
	FitnessMinInputCountAny int    // Minimum input count for any fitness calculation
	FitnessMinInputCountAll int    // Minimum input count for all fitness calculations

	// Optional parameters
	FastqExtension          string  // Extension of FASTQ files
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
// defined in the Experiment.
//
// Parameters:
//   - fastqDir: Directory containing FASTQ files.
//   - ed: ExperimentDesign with all experiment details.
func New(fastqDir string, ed ExperimentDesign) DimSum {
	maxSubs := ed.Experiment.MaxSubstitutions
	if maxSubs == 0 {
		maxSubs = DefaultMaxSubstitutions
	}

	return DimSum{
		FastqDir:            fastqDir,
		BarcodeIdentityPath: ed.Experiment.BarcodeIdentityPath,
		Experiment:          ed.Experiment.ExperimentID,
		MaxSubstitutions:    maxSubs,

		// TODO: use exp values, and default to these Defaults if not set, or
		// perhaps easier, just set the default values in sheets pkg.
		VSearchMinQual:          DefaultVsearchMinQual,
		StartStage:              DefaultStartStage,
		FitnessMinInputCountAny: DefaultFitnessMinInputCountAny,
		FitnessMinInputCountAll: DefaultFitnessMinInputCountAll,
		FastqExtension:          DefaultFastqExtension,
		CutAdaptMinLength:       DefaultCutAdaptMinLength,
		CutAdaptErrorRate:       DefaultCutAdaptErrorRate,
		Cores:                   DefaultCores,
		MixedSubstitutions:      DefaultMixedSubstitutions,
		MutagenesisType:         DefaultMutagenesisType,
		RetainIntermediateFiles: DefaultRetainIntermediateFiles,
		DesignPairDuplicates:    DefaultDesignPairDuplicates,
	}
}

// Key generates a unique key that includes our Experiment, the given sample
// names and runIDs (sorted), and a condensed encoded representation of all our
// other properties.
func (d *DimSum) Key(samples []*types.Sample) string {
	sampleInfo := make([]string, len(samples))

	for i, sample := range samples {
		sampleInfo[i] = fmt.Sprintf("%s.%s", sample.SampleName, sample.RunID)
	}

	sort.Strings(sampleInfo)

	// TODO: include more/all of the properties in the key
	combinedProps := fmt.Sprintf("%s_%d_%d_%d_%d_%d_%.2f_%d_%t_%s_%t",
		d.BarcodeIdentityPath, d.VSearchMinQual, d.StartStage,
		d.FitnessMinInputCountAny, d.FitnessMinInputCountAll,
		d.CutAdaptMinLength, d.CutAdaptErrorRate, d.MaxSubstitutions,
		d.MixedSubstitutions, d.MutagenesisType, d.DesignPairDuplicates)

	hasher := sha1.New()
	hasher.Write([]byte(combinedProps))
	encodedProps := hex.EncodeToString(hasher.Sum(nil))

	return filepath.Join(d.Experiment, strings.Join(sampleInfo, ","), encodedProps)
}

// TODO: make Key be an explicit initial "temp" output path method that returns
// experiment ID/samplesnameIDs, then a final output path that would be the
// hash in a subdir of that.

// Command generates the DiMSum command to execute. It assumes you will run the
// command in the current working directory, and output files will be set to be
// written to a subdirectory called "outputs", which will be created if it
// doesn't exist.
//
// Parameters:
//   - ed: ExperimentDesign with all experiment details.
func (d *DimSum) Command(ed ExperimentDesign) (string, error) {
	if err := os.MkdirAll(outputSubdir, 0755); err != nil {
		return "", err
	}

	libMeta := ed.Experiment

	cmd := fmt.Sprintf("%s -i %s -l %s -g %s -e %s --cutadapt5First %s --cutadapt5Second %s "+
		"-n %d -a %.2f -q %d -o %s -p %s -s %d -w %s -c %d "+
		"--fitnessMinInputCountAny %d --fitnessMinInputCountAll %d "+
		"--maxSubstitutions %d --mutagenesisType %s --retainIntermediateFiles %s "+
		"--mixedSubstitutions %s --experimentDesignPairDuplicates %s",
		DimSumExe, d.FastqDir, d.FastqExtension, "T", experimentDesignPath(".", d.Experiment),
		libMeta.Cutadapt5First,
		libMeta.Cutadapt5Second,
		d.CutAdaptMinLength, d.CutAdaptErrorRate,
		d.VSearchMinQual, outputSubdir, dimsumProjectPrefix+d.Experiment,
		d.StartStage, libMeta.WildtypeSequence, d.Cores, d.FitnessMinInputCountAny,
		d.FitnessMinInputCountAll, d.MaxSubstitutions,
		d.MutagenesisType, "T", "T", "T",
		//TODO: use libMeta values for these, converting the bools to "T" or "F"
		// libMeta.MixedSubstitutions,
		// libMeta.ExperimentDesignPairDuplicates,
	)

	if d.BarcodeIdentityPath != "" {
		cmd += " --barcodeIdentityPath " + d.BarcodeIdentityPath
	}

	return cmd, nil
}

// TODO: maybe DimSum struct replaces ExperimentDesign struct
