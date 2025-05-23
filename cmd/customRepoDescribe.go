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

	// declare switches
	readmeFound := false
	licenseFound := false

	// iterate on files
	for _, file := range files {
		if file == "README.md" {
			readmeFound = true
			describe.Overview, err = parseReadme(file)
			if err != nil {
				return describe, err
			}
		}

		if file == "LICENSE" {
			licenseFound = true
			describe.License, err = detectLicense(file)
			if err != nil {
				return describe, err
			}
			// fmt.Println("License is: ", describe.License)
		}
	}

	if !readmeFound {
		// fmt.Println("README.md not found in the directory.")
	}

	if !licenseFound {
		// fmt.Println("LICENSE not found in the directory.")
	}

	// define remote
	remoteURL, err := getRemote()
	if err != nil {
		return describe, err
	}
	describe.Remote = remoteURL

	// fmt.Println(describe.Remote)

	return describe, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateDescribeMD generates the Markdown table for the describe command.
func generateDescribeMD(repoNames []string) string {
	// Define column widths for: Repo, Remote, Overview, License.
	fieldSizes := []int{17, 99, 7}
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
