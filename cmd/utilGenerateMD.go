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
	builder.WriteString("| Repo        | Remote                                        | Commit | Age    | Language | Lines  | Size    | Mean |")
	for i := 1; i <= 12; i++ {
		builder.WriteString(fmt.Sprintf("  %02d |", i))
	}
	builder.WriteString("\n")

	// separator
	builder.WriteString("|-------------|-----------------------------------------------|--------|--------|----------|--------|---------|------|")
	for i := 1; i <= 12; i++ {
		builder.WriteString("-----|")
	}
	builder.WriteString("\n")

	// calculate average commits per month
	repoAgeMonths := calculateRepoAgeInMonths(stats.Age)
	averageCommits := 0
	if repoAgeMonths > 0 {
		averageCommits = stats.Commits / repoAgeMonths
	}

	// data row
	builder.WriteString(fmt.Sprintf(
		"| %-11s | %-45s | %-6d | %-6s | %-8s | %-6d | %-7s | %-4d |",
		repoName,
		stats.Remote,
		stats.Commits,
		stats.Age,
		stats.Language,
		stats.Lines,
		stats.Size,
		averageCommits, // average commits per month
	))
	for i := 1; i <= 12; i++ {
		month := fmt.Sprintf("%d-%02d", time.Now().Year(), i) // format: YYYY-MM
		monthCommits := stats.Frequency[month]                // month frequency
		builder.WriteString(fmt.Sprintf(" %-3d |", monthCommits))
	}
	builder.WriteString("\n")

	return builder.String()
}

///////////////////////////////////////////////////////////////////////////////////////////////////

func calculateRepoAgeInMonths(age string) int {
	years, months := 0, 0
	fmt.Sscanf(age, "%dy %dm", &years, &months) // parse "Xy Ym" format
	return (years * 12) + months
}

///////////////////////////////////////////////////////////////////////////////////////////////////
