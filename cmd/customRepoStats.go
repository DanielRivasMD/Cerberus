////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"strconv"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoStats represents the repository statistics.
type RepoStats struct {
	Language  string // Main programming language of the repository
	Age       string // Age of the repository
	Commits   int    // Total number of commits
	Remote    string
	Lines     int // Total lines of code
	Files     int
	Size      string         // Size of the repository
	Frequency map[string]int // Commit frequency by month (e.g., "2025-01": 10)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func populateRepoStats(year int) (RepoStats, error) {
	// initialize RepoStats
	stats := RepoStats{}

	// // list files
	// files, ε := listFiles(repo)
	// if ε != nil {
	// 	return stats, ε
	// }

	// // declare switches
	// readmeFound := false
	// licenseFound := false

	// // iterate on files
	// for _, file := range files {
	// 	if file == "README.md" {
	// 		readmeFound = true
	// 		description, err := parseReadme(repo + "/" + file)
	// 		if err != nil {
	// 			return stats, err
	// 		}
	// 		fmt.Println("Extracted Description:\n", description)
	// 	}

	// 	if file == "LICENSE" {
	// 		licenseFound = true
	// 		licenseType, err := detectLicense(repo + "/" + file)
	// 		if err != nil {
	// 			return stats, err
	// 		}
	// 		fmt.Println("License is: ", licenseType)
	// 	}
	// }

	// if !readmeFound {
	// 	fmt.Println("README.md not found in the directory.")
	// }
	// if !licenseFound {
	// 	fmt.Println("LICENSE not found in the directory.")
	// }

	// fetch repository metrics
	tokeiOut, _, ε := captureExecCmd("tokei", "-C")
	if ε != nil {
		return stats, ε
	}

	// define language & lines
	tokeiStats, language, ε := popualteTokei(tokeiOut)
	if ε != nil {
		return stats, ε
	}
	stats.Language = language + " " + strconv.Itoa(tokeiStats.Lines.Percentage) + "%"
	stats.Lines = tokeiStats.Lines.Number

	// define age
	age, ε := repoAge()
	if ε != nil {
		return stats, ε
	}
	stats.Age = age

	// define number commits
	commitCount, ε := countCommits()
	if ε != nil {
		return stats, ε
	}
	stats.Commits = commitCount

	// define remote
	remoteURL, ε := getRemote()
	if ε != nil {
		return stats, ε
	}
	stats.Remote = remoteURL

	// define repo size
	size, ε := repoSize()
	if ε != nil {
		return stats, ε
	}
	stats.Size = size

	// define commit frecuency
	commitFrequency, err := commitFrequency(year)
	if err != nil {
		return stats, err
	}
	stats.Frequency = commitFrequency

	return stats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
