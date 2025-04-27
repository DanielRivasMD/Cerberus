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
	Size      int64
	Frequency map[string]int
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func populateRepoStats(repo string, year int) (RepoStats, error) {
	// initialize RepoStats
	stats := RepoStats{}

	// list files
	files, err := listFiles(repo)
	if err != nil {
		return stats, err
	}

	readmeFound := false
	licenseFound := false
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
	tokeiOut, _, err := execCmdCapture("tokei", "-C")
	if err != nil {
		return stats, err
	}

	language, err := parseTokei(tokeiOut) // dominant language
	if err != nil {
		return stats, err
	}
	stats.Language = language

	repoAge, err := repoAge(repo)
	if err != nil {
		return stats, err
	}
	stats.Age = repoAge

	commitCount, err := countCommits(repo)
	if err != nil {
		return stats, err
	}
	stats.Count = commitCount

	remoteURL, err := getRemote(repo)
	if err != nil {
		return stats, err
	}
	stats.Remote = remoteURL

	linesOfCode, err := parseTokei(tokeiOut)
	if err != nil {
		return stats, err
	}
	stats.Lines, _ = strconv.Atoi(linesOfCode)

	// size, err := repoSize(repo)
	// if err != nil {
	// 	return stats, err
	// }
	// stats.Size, _ = strconv.Atoi(size)

	commitFrequency, err := commitFrequency(repo, year)
	if err != nil {
		return stats, err
	}
	stats.Frequency = commitFrequency

	return stats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
