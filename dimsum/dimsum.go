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

	"github.com/wtsi-hgi/dimsum-automation/samples"
	"github.com/wtsi-hgi/dimsum-automation/sheets"
)

const (
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

	experiementDesignPrefix = "dimsumDesign_"
	experiementDesignSuffix = ".txt"
	outputSubdir            = "outputs"
	cutAdaptRequired        = ":required..."
	cutAdaptOptional        = ":optional"
	dimsumProjectPrefix     = "dimsumRun_"
)

type Error string

func (e Error) Error() string { return string(e) }

type Experiment struct {
	SampleID      string
	Replicate     int
	Selection     int
	Pair1         string
	Pair2         string
	CellDensity   float32
	SelectionTime float32
}

type ExperimentDesign []Experiment

// NewExperimentDesign creates an experiment design from the given samples.
func NewExperimentDesign(samples []samples.Sample) ExperimentDesign {
	design := make(ExperimentDesign, 0, len(samples))

	for _, sample := range samples {
		exp := Experiment{
			SampleID:      sample.Sample.SampleID,
			Replicate:     sample.MetaData.Replicate,
			Selection:     sample.MetaData.Selection,
			Pair1:         sample.Sample.SampleID + "_1.fastq.gz",
			Pair2:         sample.Sample.SampleID + "_2.fastq.gz",
			CellDensity:   sample.MetaData.OD,
			SelectionTime: sample.MetaData.Time,
		}

		design = append(design, exp)
	}

	return design
}

// Write writes an experiment design to a file that includes the given
// experiment in the basename in the given directory and returns the path to the
// file.
func (ed ExperimentDesign) Write(dir, experiment string) (string, error) {
	designPath := experimentDesignPath(dir, experiment)

	file, err := os.Create(designPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	const header = "sample_name\texperiment_replicate\tselection_id\tselection_replicate\t" +
		"technical_replicate\tpair1\tpair2\tgenerations\tcell_density\tselection_time\n"

	_, err = file.WriteString(header)
	if err != nil {
		return "", err
	}

	for _, exp := range ed {
		selectionReplicate := ""
		if exp.Selection == 1 {
			selectionReplicate = "1"
		}

		line := fmt.Sprintf("%s\t%d\t%d\t%s\t%d\t%s\t%s\t%s\t%.3f\t%.1f\n",
			exp.SampleID, exp.Replicate, exp.Selection, selectionReplicate, 1,
			exp.Pair1, exp.Pair2, "", exp.CellDensity, exp.SelectionTime)

		_, err = file.WriteString(line)
		if err != nil {
			return "", err
		}
	}

	return designPath, nil
}

func experimentDesignPath(dir, experiment string) string {
	return filepath.Join(dir, experiementDesignPrefix+experiment+experiementDesignSuffix)
}

// DimSum represents the parameters for running DiMSum. All parameters are
// required, but using New() will default many of them to usually fixed values.
type DimSum struct {
	// Required parameters
	Exe                     string
	FastqDir                string
	BarcodeIdentityPath     string
	Experiment              string
	VSearchMinQual          int
	StartStage              int
	FitnessMinInputCountAny int
	FitnessMinInputCountAll int

	// Optional parameters
	FastqExtension          string
	Gzipped                 bool
	CutAdaptMinLength       int
	CutAdaptErrorRate       float32
	Cores                   int
	MaxSubstitutions        int
	MixedSubstitutions      bool
	MutagenesisType         string
	RetainIntermediateFiles bool
	DesignPairDuplicates    bool
}

// New creates a new DimSum instance with default values for the properties not
// supplied.
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
func (d *DimSum) Command(dir string, libMeta sheets.LibraryMetaData) string {
	cmd := fmt.Sprintf("%s -i %s -l %s -g %s -e %s --cutadapt5First %s --cutadapt5Second %s "+
		"-n %d -a %.2f -q %d -o %s -p %s -s %d -w %s -c %d "+
		"--fitnessMinInputCountAny %d --fitnessMinInputCountAll %d "+
		"--maxSubstitutions %d --mutagenesisType %s --retainIntermediateFiles %s "+
		"--mixedSubstitutions %s --experimentDesignPairDuplicates %s "+
		"--barcodeIdentityPath %s",
		d.Exe, d.FastqDir, d.FastqExtension, d.gzippedStr(), experimentDesignPath(dir, d.Experiment),
		libMeta.Cutadapt5First+cutAdaptRequired+
			reverseCompliment(libMeta.Cutadapt5Second)+cutAdaptOptional,
		libMeta.Cutadapt5Second+cutAdaptRequired+
			reverseCompliment(libMeta.Cutadapt5First)+cutAdaptOptional,
		d.CutAdaptMinLength, d.CutAdaptErrorRate,
		d.VSearchMinQual, filepath.Join(dir, outputSubdir), dimsumProjectPrefix+d.Experiment,
		d.StartStage, libMeta.Wt, d.Cores, d.FitnessMinInputCountAny,
		d.FitnessMinInputCountAll, d.MaxSubstitutions,
		d.MutagenesisType, d.retainIntermediateFilesStr(), d.mixedSubstitutionsStr(),
		d.designPairDuplicatesStr(), d.BarcodeIdentityPath,
	)

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

func reverseCompliment(seq string) string {
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
