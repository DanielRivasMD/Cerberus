////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"os"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoDescribe represents the repository features.
type RepoDescribe struct {
	Repo     string
	Overview string
	License  string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// populateRepoDescribe gathers information about a repository,
// wrapping any errors using the Horus library for additional context.
func populateRepoDescribe() (RepoDescribe, error) {
	// initialize RepoDescribe
	describe := RepoDescribe{}

	// list files
	pwd, err := os.Getwd()
	if err != nil {
		return describe, horus.Wrap(err, "populateRepoDescribe", "failed to get current working directory")
	}

	files, err := domovoi.ListFiles(pwd)
	if err != nil {
		return describe, horus.Wrap(err, "populateRepoDescribe", "failed to list files in directory")
	}

	// iterate on files
	for _, file := range files {
		if file == "README.md" {
			describe.Overview, err = parseReadme(file, overviewLen)
			if err != nil {
				return describe, horus.Wrap(err, "populateRepoDescribe", "failed to parse README.md file")
			}
		}

		if file == "LICENSE" {
			describe.License, err = detectLicense(file)
			if err != nil {
				return describe, horus.Wrap(err, "populateRepoDescribe", "failed to detect LICENSE file")
			}
		}
	}

	return describe, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateDescribeMD generates the Markdown table for the describe command.
func generateDescribeMD(repoNames []string) string {
	// Define column widths for: Repo, Remote, Overview, License.
	// Note: "Remote" is being skipped, so only the remaining fields will appear.
	fieldSizes := []int{repoLen, overviewLen, licenseLen}
	skip := map[string]bool{
		"Remote": true,
	}

	var sample RepoDescribe // used solely for header generation

	// populateFunc for describe command.
	populateFunc := func(repoName string) (*RepoDescribe, error) {
		d, err := populateRepoDescribe()
		if err != nil {
			return nil, err
		}
		d.Repo = repoName
		return &d, nil
	}

	// Create an aligners map that forces all fields to be left aligned.
	aligners := map[string]Alignment{
		"Repo":     AlignLeft,
		"Overview": AlignLeft,
		"License":  AlignLeft,
	}

	// Extra parameter is not used for describe (pass 0).
	return generateGenericMD(&sample, repoNames, populateFunc, fieldSizes, skip, 0, aligners)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
