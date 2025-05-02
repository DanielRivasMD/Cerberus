////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strings"

	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateHeader creates the Markdown table header.
func generateHeader() string {
	var builder strings.Builder
	builder.WriteString("| Repo                     | Commit | Age    | Language   | Lines  | Size    | Mean | Q1  | Q2  | Q3  | Q4  |\n")
	builder.WriteString("|--------------------------|--------|--------|------------|--------|---------|------|-----|-----|-----|-----|\n")
	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateBody creates the Markdown table body for the provided repository statistics.
func generateBody(stats RepoStats, repoName string, year int) string {
	var builder strings.Builder

	// Calculate average commits per month.
	repoAgeMonths := calculateRepoAgeInMonths(stats.Age)
	averageCommits := 0
	if repoAgeMonths > 0 {
		averageCommits = stats.Commits / repoAgeMonths
	}

	// Aggregate commits by quarter for the specified year.
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

	// Notice we now use %s for language since getColoredLanguage returns a padded string.
	builder.WriteString(fmt.Sprintf(
		"| %-24s | %-6d | %-6s | %s | %-6d | %-7s | %-4d | %-3d | %-3d | %-3d | %-3d |\n",
		repoName,
		stats.Commits,
		stats.Age,
		getColoredLanguage(stats.Language, 10),
		stats.Lines,
		stats.Size,
		averageCommits,         // Average commits per month.
		quarterlyCommits["Q1"], // Q1 commits.
		quarterlyCommits["Q2"], // Q2 commits.
		quarterlyCommits["Q3"], // Q3 commits.
		quarterlyCommits["Q4"], // Q4 commits.
	))
	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateMD creates the Markdown table for multiple repositories.
func generateMD(repoNames []string, year int) string {
	var builder strings.Builder
	builder.WriteString(generateHeader()) // Add the header once.

	// Record original directory.
	originalDir := recallDir()

	// Iterate over the directories and generate the body for each repo.
	for _, repoName := range repoNames {
		// Change directory.
		if len(repoNames) > 1 {
			changeDir(repoName)
		}

		// Collect repo data.
		statsItem, err := populateRepoStats(year)
		checkErr(err)

		// Format output.
		builder.WriteString(generateBody(statsItem, repoName, year))

		// Return directory.
		changeDir(originalDir)
	}

	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// getColoredLanguage pads the input language string to the provided width,
// then applies color based on its base language (first token).
func getColoredLanguage(language string, width int) string {
	// First, pad the entire language string to the required width.
	padded := fmt.Sprintf("%-"+fmt.Sprintf("%d", width)+"s", language)
	// Split the language string by spaces to determine the base language.
	parts := strings.Split(language, " ")
	baseLanguage := strings.ToLower(parts[0])
	switch baseLanguage {
	case "go":
		return chalk.Blue.Color(padded)
	case "julia":
		return chalk.Magenta.Color(padded)
	case "python":
		return chalk.Green.Color(padded)
	case "r":
		return chalk.Cyan.Color(padded)
	case "rust":
		return chalk.Yellow.Color(padded)
	case "shell":
		return chalk.Red.Color(padded)
	default:
		return padded // Return uncolored if no match.
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// calculateRepoAgeInMonths calculates repository age in months given an "Xy Ym" format.
func calculateRepoAgeInMonths(age string) int {
	years, months := 0, 0
	fmt.Sscanf(age, "%dy %dm", &years, &months) // Parse "Xy Ym" format.
	return (years * 12) + months
}

////////////////////////////////////////////////////////////////////////////////////////////////////
