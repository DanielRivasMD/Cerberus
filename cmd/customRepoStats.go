////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strconv"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

type RepoStats struct {
	Language  string
	Age       string
	Count     int
	Remote    string
	Lines     int
	Files     int
	Size      int
	Frequency map[string]int
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func populateRepoStats(repo string, year int) (RepoStats, error) {
	// initialize RepoStats
	stats := RepoStats{}

	// list files
	files, ε := listFiles(repo)
	if ε != nil {
		return stats, ε
	}

	// declare switches
	readmeFound := false
	licenseFound := false

	// iterate on files
	for _, file := range files {
		if file == "README.md" {
			readmeFound = true
			description, err := parseReadme(repo + "/" + file)
			if err != nil {
				return stats, err
			}
			fmt.Println("Extracted Description:\n", description)
		}

		if file == "LICENSE" {
			licenseFound = true
			licenseType, err := detectLicense(repo + "/" + file)
			if err != nil {
				return stats, err
			}
			fmt.Println("License is: ", licenseType)
		}
	}

	if !readmeFound {
		fmt.Println("README.md not found in the directory.")
	}
	if !licenseFound {
		fmt.Println("LICENSE not found in the directory.")
	}

	// fetch repository metrics
	tokeiOut, _, ε := execCmdCapture("tokei", "-C")
	if ε != nil {
		return stats, ε
	}

	language, ε := parseTokei(tokeiOut) // dominant language
	if ε != nil {
		return stats, ε
	}
	stats.Language = language

	repoAge, ε := repoAge(repo)
	if ε != nil {
		return stats, ε
	}
	stats.Age = repoAge

	commitCount, ε := countCommits(repo)
	if ε != nil {
		return stats, ε
	}
	stats.Count = commitCount

	remoteURL, ε := getRemote(repo)
	if ε != nil {
		return stats, ε
	}
	stats.Remote = remoteURL

	linesOfCode, ε := parseTokei(tokeiOut)
	if ε != nil {
		return stats, ε
	}
	stats.Lines, _ = strconv.Atoi(linesOfCode)

	size, ε := repoSize(repo)
	if ε != nil {
		return stats, ε
	}
	stats.Size, _ = strconv.Atoi(size)

	commitFrequency, err := commitFrequency(repo, year)
	if err != nil {
		return stats, err
	}
	stats.Frequency = commitFrequency

	return stats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
