////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func generateMD(stats RepoStats, repoName string, year int) string {
	var builder strings.Builder

	// header
	builder.WriteString("| Repo        | Commit | Age    | Language   | Lines  | Size    | Mean | Q1  | Q2  | Q3  | Q4  |\n")
	builder.WriteString("|-------------|--------|--------|------------|--------|---------|------|-----|-----|-----|-----|\n")

	// builder.WriteString("| Repo        | Remote                                        | Commit | Age    | Language   | Lines  | Size    | Mean | Q1  | Q2  | Q3  | Q4  |\n")
	// builder.WriteString("|-------------|-----------------------------------------------|--------|--------|------------|--------|---------|------|-----|-----|-----|-----|\n")

	// calculate average commits per month
	repoAgeMonths := calculateRepoAgeInMonths(stats.Age)
	averageCommits := 0
	if repoAgeMonths > 0 {
		averageCommits = stats.Commits / repoAgeMonths
	}

	// aggregate commits by quarter for the specified year
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

	// data row
	builder.WriteString(fmt.Sprintf(
		"| %-11s | %-6d | %-6s | %-10s | %-6d | %-7s | %-4d | %-3d | %-3d | %-3d | %-3d |\n",
		repoName,
		stats.Commits,
		stats.Age,
		stats.Language,
		stats.Lines,
		stats.Size,
		averageCommits,         // average commits per month
		quarterlyCommits["Q1"], // Q1 commits
		quarterlyCommits["Q2"], // Q2 commits
		quarterlyCommits["Q3"], // Q3 commits
		quarterlyCommits["Q4"], // Q4 commits
	))

	return builder.String()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func calculateRepoAgeInMonths(age string) int {
	years, months := 0, 0
	fmt.Sscanf(age, "%dy %dm", &years, &months) // parse "Xy Ym" format
	return (years * 12) + months
}

///////////////////////////////////////////////////////////////////////////////////////////////////
