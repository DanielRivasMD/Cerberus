////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"strconv"

	"github.com/DanielRivasMD/domovoi"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoStats represents the repository statistics.
// Empty fields will use their zero values initially,
// ensuring that the automatic header generator includes these columns.
type RepoStats struct {
	Repo      string         // Repo name
	Remote    string         // Remote URL of the repository
	Commit    int            // Total number of commits
	Age       string         // Age of the repository (e.g., "3y 2m")
	Language  string         // Main programming language of the repository, along with details (e.g., "Go 80%")
	Lines     int            // Total lines of code
	Files     int            // Total number of files; may be populated later
	Size      string         // Size of the repository (e.g., "5MB")
	Frequency map[string]int // Commit frequency by month (e.g., "2025-01": 10)

	// The following are computed values, added to support additional table columns.
	Mean int // Average commits per month (to be computed via custom logic)
	Q1   int // Commits for Quarter 1 (computed)
	Q2   int // Commits for Quarter 2 (computed)
	Q3   int // Commits for Quarter 3 (computed)
	Q4   int // Commits for Quarter 4 (computed)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func populateRepoStats(year int) (RepoStats, error) {
	// initialize RepoStats
	stats := RepoStats{}

	// fetch repository metrics
	tokeiOut, _, ε := domovoi.CaptureExecCmd("tokei", "-C")
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
	stats.Commit = commitCount

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

// generateStatsMD generates the Markdown table for the stats command.
func generateStatsMD(repoNames []string, year int) string {
	// Define field sizes for: Repo, Language, Age, Commit, Lines, Size, Mean, Q1, Q2, Q3, Q4.
	// (Note: Ensure that the order of fieldSizes matches the header order in your RepoStats struct.)
	fieldSizes := []int{repoLen, commitLen, ageLen, languageLen, linesLen, sizeLen, meanLen, qLen, qLen, qLen, qLen}
	skip := map[string]bool{
		"Remote":    true,
		"Files":     true,
		"Frequency": true,
	}

	var sample RepoStats // Sample instance for header generation

	// populateFunc for stats command.
	populateFunc := func(repoName string) (*RepoStats, error) {
		s, err := populateRepoStats(year)
		if err != nil {
			return nil, err
		}
		s.Repo = repoName
		return &s, nil
	}

	// Create an aligners map such that only the "Repo" column is left aligned.
	// For any header not explicitly provided, the default behavior in formatCell
	// will be: if index==0 then left aligned, otherwise right aligned.
	aligners := map[string]Alignment{
		"Repo": AlignLeft,
		"Language": AlignLeft,
	}

	// Pass the year as the extra parameter.
	return generateGenericMD(&sample, repoNames, populateFunc, fieldSizes, skip, year, aligners)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
