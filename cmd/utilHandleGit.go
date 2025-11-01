////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

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
