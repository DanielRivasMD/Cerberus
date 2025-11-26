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
	"fmt"

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

var (
	verbose bool
	output  string
)

type defaultVals struct {
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

var Defaults = defaultVals{
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "File output")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// handleGit is a generic function that performs the ".git" check and then
// dispatches to the appropriate report generation function.
//
// reportType should be "stats" or "describe".
// repository is the path (or name) for the repo to report on.
// year is used in stats reporting.
// verbose controls extra output in domovoi.DirExist.
func handleGit(reportType string, verbose bool) error {
	// Check if the .git directory exists.
	ok, err := domovoi.DirExist(".git", horus.NullAction(false), verbose)
	if err != nil {
		return err
	}

	var repoNames []string

	if !ok {
		// OUT branch: The ".git" directory does not exist.
		// List the subdirectories from the given repository.
		repoNames, err = domovoi.ListDirs(repository)
		if err != nil {
			return err
		}
		// Generate the report based on reportType.
		switch reportType {
		case "stats":
			fmt.Println(generateStatsMD(repoNames, year))
		case "describe":
			fmt.Println(generateDescribeMD(repoNames))
		case "remember":
			fmt.Println(generateRememberMD(repoNames))
		default:
			return fmt.Errorf("unknown report type: %s", reportType)
		}
	} else {
		// IN branch: The ".git" directory is present.
		// If the repository is ".", then update it from the current directory.
		if repository == "." {
			repoName, err := domovoi.CurrentDir()
			if err != nil {
				return err
			}
			repository = repoName
		}
		// Create a slice with just this repository.
		repoNames = append(repoNames, repository)
		switch reportType {
		case "stats":
			fmt.Println(generateStatsMD(repoNames, year))
		case "describe":
			fmt.Println(generateDescribeMD(repoNames))
		case "remember":
			fmt.Println(generateRememberMD(repoNames))
		default:
			return fmt.Errorf("unknown report type: %s", reportType)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
