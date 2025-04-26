////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// create Markdown table
func generateMD(stats RepoStats, repoName string) string {
	var builder strings.Builder

	// header
	builder.WriteString("| Repo        | Language | Age    | Commits | Remote                                   | Lines  | Size    | Frequency |")
	for i := 1; i <= 12; i++ {
		builder.WriteString(fmt.Sprintf("  %02d |", i))
	}
	builder.WriteString("\n")

	// separator
	builder.WriteString("|-------------|----------|--------|---------|------------------------------------------|--------|---------|-----------|")
	for i := 1; i <= 12; i++ {
		builder.WriteString("-----|")
	}
	builder.WriteString("\n")

	// data
	// frequency := "Monthly Breakdown"
	builder.WriteString(fmt.Sprintf(
		"| %-11s | %-8s | %-6s | %-7d | %-40s | %-6d | %-7d | %-9s |",
		repoName,
		stats.Language,
		stats.Age,
		stats.Count,
		stats.Remote,
		stats.Lines,
		stats.Files,
		formatRepoSize(stats.Size),
	))
	for i := 1; i <= 12; i++ {
		month := fmt.Sprintf("%d-%02d", time.Now().Year(), i)
		builder.WriteString(fmt.Sprintf(" %-3d |", stats.Frequency[month]))
	}
	builder.WriteString("\n")

	return builder.String()
}

///////////////////////////////////////////////////////////////////////////////////////////////////
