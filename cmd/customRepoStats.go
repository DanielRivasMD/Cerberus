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
