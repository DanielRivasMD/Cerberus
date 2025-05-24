////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"os"

	"github.com/DanielRivasMD/domovoi"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoDescribe represents the repository features.
type RepoDescribe struct {
	Repo     string
	Remote   string
	Overview string
	License  string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func populateRepoDescribe() (RepoDescribe, error) {
	// initialize RepoDescribe
	describe := RepoDescribe{}

	// list files
	pwd, err := os.Getwd()
	if err != nil {
		return describe, err
	}

	files, err := domovoi.ListFiles(pwd)
	if err != nil {
		return describe, err
	}

	// iterate on files
	for _, file := range files {
		if file == "README.md" {
			describe.Overview, err = parseReadme(file, overviewLen)
			if err != nil {
				return describe, err
			}
		}

		if file == "LICENSE" {
			describe.License, err = detectLicense(file)
			if err != nil {
				return describe, err
			}
		}
	}

	// define remote
	remoteURL, err := getRemote()
	if err != nil {
		return describe, err
	}
	describe.Remote = remoteURL

	return describe, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateDescribeMD generates the Markdown table for the describe command.
func generateDescribeMD(repoNames []string) string {
	// Define column widths for: Repo, Remote, Overview, License.
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

	// Extra parameter is not used for describe (pass 0).
	return generateGenericMD(&sample, repoNames, populateFunc, fieldSizes, skip, 0)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
