////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateHeader creates the Markdown table header.
func generateHeader() string {
	var builder strings.Builder
	builder.WriteString("| Repo                     | Commit | Age    | Language   | Lines  | Size    | Mean | Q1  | Q2  | Q3  | Q4  |\n")
	builder.WriteString("|--------------------------|--------|--------|------------|--------|---------|------|-----|-----|-----|-----|\n")
	return builder.String()
}

// generateBody creates the Markdown table body for the provided repository statistics.
func generateBody(stats RepoStats, repoName string, year int) string {
	var builder strings.Builder

	// Calculate average commits per month
	repoAgeMonths := calculateRepoAgeInMonths(stats.Age)
	averageCommits := 0
	if repoAgeMonths > 0 {
		averageCommits = stats.Commits / repoAgeMonths
	}

	// Aggregate commits by quarter for the specified year
	quarterlyCommits := map[string]int{
		"Q1": stats.Frequency[fmt.Sprintf("%d-01", year)] +
			stats.Frequency[fmt.Sprintf("%d-02", year)] +
			stats.Frequency[fmt.Sprintf("%d-03", year)],
		"Q2": stats.Frequency[fmt.Sprintf("%d-04", year)] +
			stats.Frequency[fmt.Sprintf("%d-05", year)] +
			stats.Frequency[fmt.Sprintf("%d-06", year)],
		"Q3": stats.Frequency[fmt.Sprintf("%d-07", year)] +
			stats.Frequency[fmt.Sprintf("%d-08", year)] +
			stats.Frequency[fmt.Sprintf("%d-09", year)],
		"Q4": stats.Frequency[fmt.Sprintf("%d-10", year)] +
			stats.Frequency[fmt.Sprintf("%d-11", year)] +
			stats.Frequency[fmt.Sprintf("%d-12", year)],
	}

	// Add data row
	builder.WriteString(fmt.Sprintf(
		"| %-24s | %-6d | %-6s | %-10s | %-6d | %-7s | %-4d | %-3d | %-3d | %-3d | %-3d |\n",
		repoName,
		stats.Commits,
		stats.Age,
		stats.Language,
		stats.Lines,
		stats.Size,
		averageCommits,         // Average commits per month
		quarterlyCommits["Q1"], // Q1 commits
		quarterlyCommits["Q2"], // Q2 commits
		quarterlyCommits["Q3"], // Q3 commits
		quarterlyCommits["Q4"], // Q4 commits
	))
	return builder.String()
}

// generateMD creates the Markdown table for multiple repositories.
func generateMD(repoNames []string, year int) string {
	var builder strings.Builder
	builder.WriteString(generateHeader()) // Add the header once

	// record original directory
	originalDir := recallDir()

	// Iterate over the directories and generate the body for each repo
	for _, repoName := range repoNames {

		// change directory
		changeDir(repoName)

		// Collect repo data
		statsItem, err := populateRepoStats(year)
		checkErr(err)

		// format output
		builder.WriteString(generateBody(statsItem, repoName, year))

		// return directory
		changeDir(originalDir)
	}

	return builder.String()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

// calculateRepoAgeInMonths calculate repository age in months.
func calculateRepoAgeInMonths(age string) int {
	years, months := 0, 0
	fmt.Sscanf(age, "%dy %dm", &years, &months) // parse "Xy Ym" format
	return (years * 12) + months
}

///////////////////////////////////////////////////////////////////////////////////////////////////
