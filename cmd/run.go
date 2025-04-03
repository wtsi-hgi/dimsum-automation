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
	"github.com/wtsi-hgi/dimsum-automation/itl"
	"github.com/wtsi-hgi/dimsum-automation/samples"
)

const (
	ErrBadOutputDir    = Error("output directory must not be a sub-directory of the current working directory")
	ErrSamplesRequired = Error("at least one sampleName:runID pair is required")

	dirPerm = 0755
)

// options for this cmd.
var itlOutput string

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
		nameRuns := nameRunStrsToNameRuns(nameRunStrs)

		err := validateOutputDir(itlOutput)
		if err != nil {
			die(err)
		}

		ss, err := getAllSponsorSamples()
		if err != nil {
			die(err)
		}

		filtered, err := ss.Filter(nameRuns)
		if err != nil {
			die(err)
		}

		itl, err := itl.New(filtered, itlOutput)
		if err != nil {
			die(err)
		}

		cmd, tsvPath := itl.GenerateSamplesTSVCommand()

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

			err = executeCmd(cmd)
			if err != nil {
				die(err)
			}

			err = fc.CopyFastqFiles()
			if err != nil {
				die(err)
			}
		}
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

func getAllSponsorSamples() (samples.Samples, error) {
	c, err := config.FromEnv()
	if err != nil {
		return nil, err
	}

	db, sheets, err := getDBAndSheets(c)
	if err != nil {
		return nil, err
	}

	return sponsorSamples(c, db, sheets)
}

func nameRunStrsToNameRuns(nameRunStrs []string) []samples.NameRun {
	result := make([]samples.NameRun, 0, len(nameRunStrs))
	done := make(map[string]bool)

	for _, nameRunStr := range nameRunStrs {
		if done[nameRunStr] {
			continue
		}

		parts := strings.Split(nameRunStr, ":")
		if len(parts) != 2 {
			dief("invalid sampleName:runID pair: %s", nameRunStr)
		}

		result = append(result, samples.NameRun{
			Name: parts[0],
			Run:  parts[1],
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

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.AddCommand(irodsToLustreCmd)

	// flags specific to these sub-commands
	irodsToLustreCmd.Flags().StringVarP(&itlOutput, "output", "o", "",
		"output directory for FASTQ files")

	err := irodsToLustreCmd.MarkFlagRequired("output")
	if err != nil {
		die(err)
	}
}
