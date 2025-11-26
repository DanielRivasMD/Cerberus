////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"strconv"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoStats represents the repository statistics.
// Empty fields will use their zero values initially,
// ensuring that the automatic header generator includes these columns.
type RepoStats struct {
	Repo      string         // Repo name
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

// populateRepoStats gathers statistics for a repository for the given year,
// wrapping any errors using the Horus library for additional context.
func populateRepoStats(year int) (RepoStats, error) {
	// initialize RepoStats
	stats := RepoStats{}

	// fetch repository metrics
	tokeiOut, _, err := domovoi.CaptureExecCmd("tokei", "-C")
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to capture tokei output")
	}

	// define language & lines
	tokeiStats, language, err := populateTokei(tokeiOut)
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to parse tokei output")
	}
	stats.Language = language + " " + strconv.Itoa(tokeiStats.Lines.Percentage) + "%"
	stats.Lines = tokeiStats.Lines.Number

	// define age
	age, err := repoAge()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to determine repository age")
	}
	stats.Age = age

	// define number of commits
	commitCount, err := countCommits()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to count commits")
	}
	stats.Commit = commitCount

	// define repo size
	size, err := repoSize()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to determine repository size")
	}
	stats.Size = size

	// define commit frequency
	commitFrequency, err := commitFrequency(year)
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to fetch commit frequency")
	}
	stats.Frequency = commitFrequency

	return stats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// populateHistoricalRepoStats collects historical data by invoking populateRepoStats for each year
// in the range from startYear to endYear (inclusive). It returns a map where the keys are years and
// the values are the corresponding RepoStats structs.
// If startYear is greater than endYear, it returns an error.
func populateHistoricalRepoStats(startYear, endYear int) (map[int]RepoStats, error) {
	// Validate the year range.
	if startYear > endYear {
		return nil, horus.NewHerror(
			"populateHistoricalRepoStats",
			"invalid year range: startYear must not be greater than endYear",
			nil,
			map[string]any{"startYear": startYear, "endYear": endYear},
		)
	}

	historicalStats := make(map[int]RepoStats)
	for year := startYear; year <= endYear; year++ {
		stats, err := populateRepoStats(year)
		if err != nil {
			return nil, horus.Wrap(
				err,
				"populateHistoricalRepoStats",
				"failed to collect stats for year "+strconv.Itoa(year),
			)
		}
		historicalStats[year] = stats
	}

	return historicalStats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateStatsMD generates the Markdown table for the stats command.
func generateStatsMD(repoNames []string, year int) string {
	// Define field sizes for: Repo, Language, Age, Commit, Lines, Size, Mean, Q1, Q2, Q3, Q4.
	// (Note: Ensure that the order of fieldSizes matches the header order in your RepoStats struct.)
	fieldSizes := []int{Defaults.repoLen, Defaults.commitLen, Defaults.ageLen, Defaults.languageLen, Defaults.linesLen, Defaults.sizeLen, Defaults.meanLen, Defaults.qLen, Defaults.qLen, Defaults.qLen, Defaults.qLen}
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
		"Repo":     AlignLeft,
		"Language": AlignLeft,
	}

	// Pass the year as the extra parameter.
	return generateGenericMD(&sample, repoNames, populateFunc, fieldSizes, skip, year, aligners, output)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
