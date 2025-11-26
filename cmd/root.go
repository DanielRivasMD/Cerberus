/*
Copyright © 2024 Daniel Rivas <danielrivasmd@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:     "cerberus",
	Long:    helpRoot,
	Example: exampleRoot,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func Execute() {
	horus.CheckErr(rootCmd.Execute())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var RootFlags rootFlags

type rootFlags struct {
	verbose bool
	output  string
}

type defaults struct {
	repoLen     int
	commitLen   int
	ageLen      int
	languageLen int
	linesLen    int
	sizeLen     int
	meanLen     int
	qLen        int
	overviewLen int
	licenseLen  int
	remoteLen   int
}

var Defaults = defaults{
	repoLen:     25,
	commitLen:   6,
	ageLen:      6,
	languageLen: 15,
	linesLen:    6,
	sizeLen:     7,
	meanLen:     4,
	qLen:        3,
	overviewLen: 92,
	licenseLen:  7,
	remoteLen:   95,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&RootFlags.verbose, "verbose", "v", false, "Verbose")
	rootCmd.PersistentFlags().StringVarP(&RootFlags.output, "output", "o", "", "File output")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateRememberCSV writes repoName,repoURL rows either to stdout or to the file specified by RootFlags.output.
func generateRememberCSV(repoNames []string) error {
	var outFile *os.File
	var err error

	if strings.TrimSpace(RootFlags.output) != "" {
		// create or truncate the file
		outFile, err = os.Create(RootFlags.output)
		if err != nil {
			return horus.Wrap(err, "generateRememberCSV", "failed to create output file: "+RootFlags.output)
		}
		defer outFile.Close()
	} else {
		outFile = os.Stdout
	}

	w := csv.NewWriter(outFile)
	defer w.Flush()

	// optional header
	if err := w.Write([]string{"repoName", "repoURL"}); err != nil {
		return err
	}

	for _, repo := range repoNames {
		remoteURL, err := resolveRemoteURL(repo)
		if err != nil {
			remoteURL = ""
		}
		if err := w.Write([]string{repo, remoteURL}); err != nil {
			return err
		}
	}
	return nil
}

func resolveRemoteURL(repo string) (string, error) {
	repoPath := filepath.Join(repository, repo)
	out, _, err := domovoi.CaptureExecCmd("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func handleGit(reportType string, verbose bool) error {
	ok, err := domovoi.DirExist(".git", horus.NullAction(false), verbose)
	if err != nil {
		return err
	}

	var repoNames []string
	if !ok {
		repoNames, err = domovoi.ListDirs(repository)
		if err != nil {
			return err
		}
	} else {
		if repository == "." {
			repoName, err := domovoi.CurrentDir()
			if err != nil {
				return err
			}
			repository = repoName
		}
		repoNames = append(repoNames, repository)
	}

	switch reportType {
	case "stats":
		fmt.Println(generateStatsMD(repoNames, year))
	case "describe":
		fmt.Println(generateDescribeMD(repoNames))
	case "remember":
		return generateRememberCSV(repoNames)
	default:
		return fmt.Errorf("unknown report type: %s", reportType)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
