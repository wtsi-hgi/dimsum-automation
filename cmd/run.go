/*******************************************************************************
 * Copyright (c) 2025 Genome Research Ltd.
 *
 * Author: Sendu Bala <sb10@sanger.ac.uk>
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

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wtsi-hgi/dimsum-automation/config"
	"github.com/wtsi-hgi/dimsum-automation/dimsum"
	"github.com/wtsi-hgi/dimsum-automation/itl"
	"github.com/wtsi-hgi/dimsum-automation/types"
)

const (
	ErrBadOutputDir    = Error("output directory must not be a sub-directory of the current working directory")
	ErrSamplesRequired = Error("at least one sampleName:runID pair is required")

	dirPerm    = 0755
	outputFlag = "output"
)

// options for this cmd.
var (
	itlOutput                     string
	dimsumOutput                  string
	dimsumFastqDir                string
	dimsumBarcodeIdentityPath     string
	dimsumVsearchMinQual          int
	dimsumStartStage              int
	dimsumFitnessMinInputCountAny int
	dimsumFitnessMinInputCountAll int
	dimsumCutAdaptMinLength       int
	dimsumCutAdaptErrorRate       float32
	dimsumMixedSubstitutions      bool
	dimsumMutagenesisType         string
	dimsumDesignPairDuplicates    bool
)

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run commands.",
	Long: `Run commands.

The various run sub-commands run the various steps of the workflow on the local
machine in the current working directory.

It is intended that these will be automatically run in a workflow using wr.
`,
}

// irodsToLustreCmd represents the irodsToLustre command.
var irodsToLustreCmd = &cobra.Command{
	Use:   "irods-to-lustre",
	Short: "Run irods_to_lustre to get sample FASTQ files.",
	Long: `Run irods_to_lustre to get sample FASTQ files.

iRODS commands and the irods_to_lustre pipeline must be in your PATH before
calling this command.

Given desired samples, crams will be downloaded from iRODS, merged as necessary
and FASTQ files created. The samples must be from the same study, otherwise an
error will be raised. You must also specify an output directory with the -o
option, which will be created if it doesn't exist.

Samples should be supplied as a series of sampleName:runID pairs. An example
command line could look like this:
$ dimsum-automation run irods_to_lustre -o /output/dir AMA1:1234 AMA2:5678

Note that the current working directory will be used for various working files
and it is expected that you delete this directory afterwards, ie. that you run
this command via wr without --cwd_matters. -o must therefore not be a sub
directory of the current working directory, or the working directory itself.

If output files already exist in the output directory for a sample, the process
will be skipped for that sample.
`,
	Run: func(_ *cobra.Command, nameRunStrs []string) {
		desired := subsetDesiredSamples(nameRunStrs)

		err := validateOutputDir(itlOutput)
		if err != nil {
			die(err)
		}

		itl, err := itl.New(desired, itlOutput)
		if err != nil {
			die(err)
		}

		if len(itl.SampleNameRuns()) == 0 {
			info("fastqs for these samples already exist in the output directory")

			return
		}

		cmd, tsvPath := itl.GenerateSamplesTSVCommand()

		infof("running command to generate samples TSV file:\n%s", cmd)

		err = executeCmd(cmd)
		if err != nil {
			die(err)
		}

		fcs, err := itl.FilterSamplesTSV(tsvPath)
		if err != nil {
			die(err)
		}

		for _, fc := range fcs {
			cmd = fc.Command()

			infof("running command to get fastq file for %s:\n%s", fc.IDRun(), cmd)

			err = executeCmd(cmd)
			if err != nil {
				die(err)
			}

			err = fc.MoveFastqFiles()
			if err != nil {
				die(err)
			}
		}

		infof("fastq files for %d samples downloaded to %s", len(fcs), itlOutput)
	},
}

func validateOutputDir(outputDir string) error {
	absOut, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if absOut == wd || strings.HasPrefix(absOut, wd) {
		return ErrBadOutputDir
	}

	if _, err := os.Stat(outputDir); err != nil {
		err = createDirIfNotExist(outputDir, err)
		if err != nil {
			return err
		}
	}

	return nil
}

func createDirIfNotExist(dir string, statErr error) error {
	if !os.IsNotExist(statErr) {
		return statErr
	}

	return os.MkdirAll(dir, dirPerm)
}

func subsetDesiredSamples(nameRunStrs []string) *types.Library {
	nameRuns := nameRunStrsToNameRuns(nameRunStrs)

	c, err := config.FromEnv()
	if err != nil {
		die(err)
	}

	db, s, err := getDBAndSheets(c)
	if err != nil {
		die(err)
	}

	libs, err := sponsorLibs(c, db, s)
	if err != nil {
		die(err)
	}

	filtered, err := libs.Subset(nameRuns)
	if err != nil {
		die(err)
	}

	return filtered
}

func nameRunStrsToNameRuns(nameRunStrs []string) []*types.Sample {
	result := make([]*types.Sample, 0, len(nameRunStrs))
	done := make(map[string]bool)

	for _, nameRunStr := range nameRunStrs {
		if done[nameRunStr] {
			continue
		}

		parts := strings.Split(nameRunStr, ":")
		if len(parts) != 2 {
			dief("invalid sampleName:runID pair: %s", nameRunStr)
		}

		result = append(result, &types.Sample{
			SampleID: parts[0],
			RunID:    parts[1],
		})

		done[nameRunStr] = true
	}

	if len(result) == 0 {
		die(ErrSamplesRequired)
	}

	return result
}

func executeCmd(cmd string) error {
	execCmd := exec.Command("bash", "-c", "set -o pipefail; "+cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// dimsumCmd represents the dimsum command.
var dimsumCmd = &cobra.Command{
	Use:   "dimsum",
	Short: "Run dimsum.",
	Long: `Run dimsum.

DiMSum must be in your PATH.

The fastqs for your samples should already have been generated using the
irods-to-lustre sub-command. You must supply the -o directory of that command
as the -f option to this command.

Given desired samples, this command will run DiMSum on the appropriate FASTQ
files, generating the needed DiMSum experiment design file.

The samples must be from the same study, and share the same dimsum-related
library metadata, otherwise an error will be raised.

You must also specify an output directory with the -o option, which will be
created if it doesn't exist. In this output directory, a unique sub-directory
will be created corresponding to your choice of samples and dimsum options. If
that unique sub-directory already exists and has files in it, an error will be
raised.

Samples should be supplied as a series of sampleName:runID pairs. All other
options should be supplied before these. An example command line could look like
this:
$ dimsum-automation run dimsum -o /output/dir -f /fastqs/dir \
    --barcodeIdentityPath /path/to/barcode AMA1:1234 AMA2:5678

Note that the current working directory will be used for various working files
and it is expected that you delete this directory afterwards, ie. that you run
this command via wr without --cwd_matters. -o must therefore not be a sub
directory of the current working directory, or the working directory itself.
`,
	Run: func(_ *cobra.Command, nameRunStrs []string) {
		lib := subsetDesiredSamples(nameRunStrs)

		design, err := dimsum.NewExperimentDesign(lib.Experiments[0])
		if err != nil {
			die(err)
		}

		d := dimsum.New(dimsumFastqDir, design)

		d.VSearchMinQual = dimsumVsearchMinQual
		d.StartStage = dimsumStartStage
		d.FitnessMinInputCountAny = dimsumFitnessMinInputCountAny
		d.FitnessMinInputCountAll = dimsumFitnessMinInputCountAll
		d.CutAdaptMinLength = dimsumCutAdaptMinLength
		d.CutAdaptErrorRate = dimsumCutAdaptErrorRate
		d.MixedSubstitutions = dimsumMixedSubstitutions
		d.MutagenesisType = dimsumMutagenesisType
		d.DesignPairDuplicates = dimsumDesignPairDuplicates

		err = validateOutputDir(dimsumOutput)
		if err != nil {
			die(err)
		}

		uniqueDimsumOutputDir := dimsumUniqueOutputDir(d, dimsumOutput, lib.Experiments[0].Samples)

		dir := "."

		experimentPath, err := design.Write(dir)
		if err != nil {
			die(err)
		}

		infof("created experiment design file: %s", experimentPath)

		cmd, err := d.Command(design)
		if err != nil {
			die(err)
		}

		infof("will run dimsum:\n%s", cmd)

		err = executeCmd(cmd)
		if err != nil {
			die(err)
		}

		infof("then would move output files to %s", uniqueDimsumOutputDir)
	},
}

func dimsumUniqueOutputDir(d dimsum.DimSum, outputDir string, desired []*types.Sample) string {
	uniqueDimsumOutputDir := filepath.Join(outputDir, d.Key(desired))

	if _, err := os.Stat(uniqueDimsumOutputDir); err == nil {
		entries, readErr := os.ReadDir(uniqueDimsumOutputDir)
		if readErr != nil {
			die(readErr)
		}

		if len(entries) > 0 {
			dief("unique dimsum output directory %s already exists and is not empty", uniqueDimsumOutputDir)
		}
	} else if !os.IsNotExist(err) {
		die(err)
	}

	err := os.MkdirAll(uniqueDimsumOutputDir, dirPerm)
	if err != nil {
		die(err)
	}

	return uniqueDimsumOutputDir
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.AddCommand(irodsToLustreCmd)
	runCmd.AddCommand(dimsumCmd)

	// flags specific to these sub-commands
	irodsToLustreCmd.Flags().StringVarP(&itlOutput, outputFlag, "o", "",
		"output directory for FASTQ files")
	markFlagRequired(irodsToLustreCmd, outputFlag)

	dimsumCmd.Flags().StringVarP(&dimsumOutput, outputFlag, "o", "",
		"output directory")
	markFlagRequired(dimsumCmd, outputFlag)
	dimsumCmd.Flags().StringVarP(&dimsumFastqDir, "fastqs", "f", "",
		"directory containing FASTQ files")
	markFlagRequired(dimsumCmd, "fastqs")

	dimsumCmd.Flags().StringVar(&dimsumBarcodeIdentityPath, "barcodeIdentityPath", "",
		"path to your barcode identity file")
	dimsumCmd.Flags().IntVar(&dimsumVsearchMinQual, "vsearchMinQual", dimsum.DefaultVsearchMinQual,
		"passed through to dimsum")
	dimsumCmd.Flags().IntVar(&dimsumStartStage, "startStage", dimsum.DefaultStartStage,
		"passed through to dimsum")
	dimsumCmd.Flags().IntVar(&dimsumFitnessMinInputCountAny, "fitnessMinInputCountAny", dimsum.DefaultFitnessMinInputCountAny,
		"passed through to dimsum")
	dimsumCmd.Flags().IntVar(&dimsumFitnessMinInputCountAll, "fitnessMinInputCountAll", dimsum.DefaultFitnessMinInputCountAll,
		"passed through to dimsum")
	dimsumCmd.Flags().IntVar(&dimsumCutAdaptMinLength, "cutAdaptMinLength", dimsum.DefaultCutAdaptMinLength,
		"passed through to dimsum")
	dimsumCmd.Flags().Float32Var(&dimsumCutAdaptErrorRate, "cutAdaptErrorRate", dimsum.DefaultCutAdaptErrorRate,
		"passed through to dimsum")
	dimsumCmd.Flags().BoolVar(&dimsumMixedSubstitutions, "mixedSubstitutions", dimsum.DefaultMixedSubstitutions,
		"passed through to dimsum")
	dimsumCmd.Flags().StringVar(&dimsumMutagenesisType, "mutagenesisType", dimsum.DefaultMutagenesisType,
		"passed through to dimsum")
	dimsumCmd.Flags().BoolVar(&dimsumDesignPairDuplicates, "designPairDuplicates", dimsum.DefaultDesignPairDuplicates,
		"passed through to dimsum")
}

func markFlagRequired(cmd *cobra.Command, flagName string) {
	err := cmd.MarkFlagRequired(flagName)
	if err != nil {
		die(err)
	}
}
