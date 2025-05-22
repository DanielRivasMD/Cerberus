////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"os"
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

// TODO: use horus to catch errors
// TODO: find why multiple repos error out
func populateRepoDescribe() (RepoDescribe, error) {
	// initialize RepoDescribe
	describe := RepoDescribe{}

	// list files
	pwd, err := os.Getwd()
	if err != nil {
		return describe, err
	}

	files, err := listFiles(pwd)
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
			fmt.Println("Extracted Description:\n", describe.Overview)
		}

		if file == "LICENSE" {
			licenseFound = true
			describe.License, err = detectLicense(file)
			if err != nil {
				return describe, err
			}
			fmt.Println("License is: ", describe.License)
		}
	}

	if !readmeFound {
		fmt.Println("README.md not found in the directory.")
	}
	if !licenseFound {
		fmt.Println("LICENSE not found in the directory.")
	}

	// define remote
	remoteURL, err := getRemote()
	if err != nil {
		return describe, err
	}
	describe.Remote = remoteURL

	fmt.Println(describe.Remote)

	return describe, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
