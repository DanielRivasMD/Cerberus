////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// getHeaders extracts the field names of a struct,
// omitting any fields present in skipFields.
func getHeaders(v interface{}, skipFields map[string]bool) []string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if skipFields[fieldName] {
			continue
		}
		headers = append(headers, fieldName)
	}
	return headers
}

// getValues returns raw string representations of field values from a struct, skipping fields in skipFields.
// No padding or alignment is applied here.
func getValues(v interface{}, skipFields map[string]bool) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	var values []string
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		if skipFields[fieldName] {
			continue
		}
		// Return the raw value as string.
		formatted := fmt.Sprintf("%v", val.Field(i).Interface())
		values = append(values, formatted)
	}
	return values
}

// generateMarkdownHeader creates the Markdown header row (with bold text)
// and a separator row. It uses runewidth to calculate visible widths.
// For terminal rendering, ANSI Bold is applied directly.
func generateMarkdownHeader(v interface{}, fieldSizes []int, skipFields map[string]bool) string {
	headers := getHeaders(v, skipFields)
	var builder strings.Builder

	// Header row
	builder.WriteString("| ")
	for i, header := range headers {
		// Compute needed padding using visible width.
		visibleWidth := runewidth.StringWidth(header)
		padLength := fieldSizes[i] - visibleWidth
		if padLength < 0 {
			padLength = 0
		}
		var cell string
		// Left-align first column ("Repo"), right-align others.
		if i == 0 {
			cell = header + strings.Repeat(" ", padLength)
		} else {
			cell = strings.Repeat(" ", padLength) + header
		}
		// Wrap with ANSI Bold codes.
		boldCell := "\033[1m" + cell + "\033[0m"
		builder.WriteString(boldCell + " | ")
	}
	builder.WriteString("\n")

	// Separator row (simple dashes)
	builder.WriteString("|")
	for i := range headers {
		builder.WriteString(strings.Repeat("-", fieldSizes[i]+2) + "|")
	}
	builder.WriteString("\n")
	return builder.String()
}

// generateMarkdownRow creates a Markdown table row for a single struct instance.
// It updates computed fields (Mean, Q1–Q4) when v is a *RepoStats and applies alignment:
// the first column ("Repo") is left aligned, while all others are right aligned.
// Also, for the "Language" column, getColoredLanguage is applied.
func generateMarkdownRow(v interface{}, fieldSizes []int, skipFields map[string]bool, year int) string {
	// If v is a *RepoStats, update computed fields.
	if repoStats, ok := v.(*RepoStats); ok {
		repoAgeMonths := calculateRepoAgeInMonths(repoStats.Age)
		averageCommits := 0
		if repoAgeMonths > 0 {
			averageCommits = repoStats.Commit / repoAgeMonths
		}

		quarterlyCommits := map[string]int{
			"Q1": repoStats.Frequency[fmt.Sprintf("%d-01", year)] +
				repoStats.Frequency[fmt.Sprintf("%d-02", year)] +
				repoStats.Frequency[fmt.Sprintf("%d-03", year)],
			"Q2": repoStats.Frequency[fmt.Sprintf("%d-04", year)] +
				repoStats.Frequency[fmt.Sprintf("%d-05", year)] +
				repoStats.Frequency[fmt.Sprintf("%d-06", year)],
			"Q3": repoStats.Frequency[fmt.Sprintf("%d-07", year)] +
				repoStats.Frequency[fmt.Sprintf("%d-08", year)] +
				repoStats.Frequency[fmt.Sprintf("%d-09", year)],
			"Q4": repoStats.Frequency[fmt.Sprintf("%d-10", year)] +
				repoStats.Frequency[fmt.Sprintf("%d-11", year)] +
				repoStats.Frequency[fmt.Sprintf("%d-12", year)],
		}

		repoStats.Mean = averageCommits
		repoStats.Q1 = quarterlyCommits["Q1"]
		repoStats.Q2 = quarterlyCommits["Q2"]
		repoStats.Q3 = quarterlyCommits["Q3"]
		repoStats.Q4 = quarterlyCommits["Q4"]
	}

	// Get headers and raw values.
	headers := getHeaders(v, skipFields)
	values := getValues(v, skipFields)
	var builder strings.Builder

	builder.WriteString("| ")
	for i, value := range values {
		// For the "Language" column, apply color.
		if i < len(headers) && headers[i] == "Language" {
			value = getColoredLanguage(value, fieldSizes[i])
		}
		// Compute visible width and calculate the necessary padding.
		visibleWidth := runewidth.StringWidth(value)
		padLength := fieldSizes[i] - visibleWidth
		if padLength < 0 {
			padLength = 0
		}
		var cell string
		// Left align only the first column ("Repo"); right align others.
		if i == 0 {
			cell = value + strings.Repeat(" ", padLength)
		} else {
			cell = strings.Repeat(" ", padLength) + value
		}
		builder.WriteString(cell + " | ")
	}
	builder.WriteString("\n")
	return builder.String()
}

// generateMD creates the Markdown table for one or more repositories.
// It uses our header and row generators (which update computed fields and right align all but the first column).
func generateMD(repoNames []string, year int) string {
	var builder strings.Builder

	// Create a sample instance of RepoStats for header generation.
	var statsSample RepoStats

	// Define fields to skip.
	skip := map[string]bool{
		"Remote":    true,
		"Files":     true,
		"Frequency": true,
	}

	// Field sizes (in characters) for the displayed fields, in order:
	// Repo, Language, Age, Commit, Lines, Size, Mean, Q1, Q2, Q3, Q4.
	fieldSizes := []int{25, 6, 6, 15, 6, 7, 4, 3, 3, 3, 3}

	// Generate the header row.
	builder.WriteString(generateMarkdownHeader(statsSample, fieldSizes, skip))

	// Record original directory (stub).
	originalDir := recallDir()

	// Process each repository.
	for _, repoName := range repoNames {
		// Change directory if handling multiple repositories.
		if len(repoNames) > 1 {
			changeDir(repoName)
		}

		// Collect repository stats.
		stats, err := populateRepoStats(year)
		if err != nil {
			panic(err)
		}

		// Assign repo name.
		stats.Repo = repoName

		// Append a Markdown row for this repo.
		builder.WriteString(generateMarkdownRow(&stats, fieldSizes, skip, year))

		// Return to the original directory.
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
