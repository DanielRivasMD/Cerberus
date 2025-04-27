////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func generateMD(stats RepoStats, repoName string) string {
	var builder strings.Builder

	// header
	builder.WriteString("| Repo        | Remote                                        | Commit | Age    | Language | Lines  | Size    | Mean | Q1  | Q2  | Q3  | Q4  |\n")
	builder.WriteString("|-------------|-----------------------------------------------|--------|--------|----------|--------|---------|------|-----|-----|-----|-----|\n")

	// calculate average commits per month
	repoAgeMonths := calculateRepoAgeInMonths(stats.Age)
	averageCommits := 0
	if repoAgeMonths > 0 {
		averageCommits = stats.Commits / repoAgeMonths
	}

	// aggregate commits by quarter
	quarterlyCommits := map[string]int{
		"Q1": stats.Frequency[fmt.Sprintf("%d-01", time.Now().Year())] +
			stats.Frequency[fmt.Sprintf("%d-02", time.Now().Year())] +
			stats.Frequency[fmt.Sprintf("%d-03", time.Now().Year())],

		"Q2": stats.Frequency[fmt.Sprintf("%d-04", time.Now().Year())] +
			stats.Frequency[fmt.Sprintf("%d-05", time.Now().Year())] +
			stats.Frequency[fmt.Sprintf("%d-06", time.Now().Year())],

		"Q3": stats.Frequency[fmt.Sprintf("%d-07", time.Now().Year())] +
			stats.Frequency[fmt.Sprintf("%d-08", time.Now().Year())] +
			stats.Frequency[fmt.Sprintf("%d-09", time.Now().Year())],

		"Q4": stats.Frequency[fmt.Sprintf("%d-10", time.Now().Year())] +
			stats.Frequency[fmt.Sprintf("%d-11", time.Now().Year())] +
			stats.Frequency[fmt.Sprintf("%d-12", time.Now().Year())],
	}

	// data row
	builder.WriteString(fmt.Sprintf(
		"| %-11s | %-45s | %-6d | %-6s | %-8s | %-6d | %-7s | %-4d | %-3d | %-3d | %-3d | %-3d |\n",
		repoName,
		stats.Remote,
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
