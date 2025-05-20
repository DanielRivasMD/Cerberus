////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"reflect"
	"strings"

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
			continue // skip the field if it's in the skipFields map
		}
		headers = append(headers, fieldName)
	}
	return headers
}

// getValues returns the string representations of the struct’s field values,
// formatted with fixed widths taken from the fieldSizes slice,
// and omits any fields that are in skipFields.
func getValues(v interface{}, fieldSizes []int, skipFields map[string]bool) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var values []string
	col := 0
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		if skipFields[fieldName] {
			continue // skip the field if it's in the skipFields map
		}
		width := 0
		if col < len(fieldSizes) {
			width = fieldSizes[col]
		}
		formatted := fmt.Sprintf("%-*v", width, val.Field(i).Interface())
		values = append(values, formatted)
		col++
	}
	return values
}

// generateMarkdownHeader creates a Markdown header row including a separator.
// It uses the provided fieldSizes slice to pad each header and accepts
// a skipFields map so that specified fields are not printed.
// Instead of using chalk (which introduces hidden ANSI codes that confuse width calculations),
// we wrap the padded header value in Markdown bold markers.
func generateMarkdownHeader(v interface{}, fieldSizes []int, skipFields map[string]bool) string {
	headers := getHeaders(v, skipFields)
	var builder strings.Builder

	// Header row
	builder.WriteString("| ")
	for i, header := range headers {
		width := 0
		if i < len(fieldSizes) {
			width = fieldSizes[i]
		}
		// First pad the header with raw text.
		paddedHeader := fmt.Sprintf("%-*s", width, header)
		// Then wrap it in markdown bold markers.
		boldHeader := chalk.Bold.TextStyle(paddedHeader)
		builder.WriteString(boldHeader + " | ")
	}
	builder.WriteString("\n")

	// Separator row (use the original unformatted header lengths)
	builder.WriteString("|")
	for i, header := range headers {
		width := fieldSizes[i]
		if width <= 0 {
			width = len(header)
		}
		// We account for the extra padding spaces.
		builder.WriteString(strings.Repeat("-", width+2) + "|")
	}
	builder.WriteString("\n")
	return builder.String()
}

// generateMarkdownRow creates a Markdown table row for a single struct instance.
// It accepts a skipFields map so that fields (e.g., "Remote", "Files", "Frequency")
// can be omitted from the output.
// Additionally, if the provided instance is of type *RepoStats, it calculates and
// populates the computed fields (Mean, Q1, Q2, Q3, Q4) for the specified year.
// Finally, for the "Language" column only, it uses getColoredLanguage to generate
// a colored & padded string.
func generateMarkdownRow(v interface{}, fieldSizes []int, skipFields map[string]bool, year int) string {
	// If v is a pointer to RepoStats, update its computed fields.
	if repoStats, ok := v.(*RepoStats); ok {
		// Calculate average commits per month.
		repoAgeMonths := calculateRepoAgeInMonths(repoStats.Age)
		averageCommits := 0
		if repoAgeMonths > 0 {
			averageCommits = repoStats.Commit / repoAgeMonths
		}

		// Aggregate commits by quarter for the specified year.
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

		// Update the computed fields.
		repoStats.Mean = averageCommits
		repoStats.Q1 = quarterlyCommits["Q1"]
		repoStats.Q2 = quarterlyCommits["Q2"]
		repoStats.Q3 = quarterlyCommits["Q3"]
		repoStats.Q4 = quarterlyCommits["Q4"]
	}

	// Retrieve headers and values from the instance, omitting the skipped fields.
	headers := getHeaders(v, skipFields)
	values := getValues(v, fieldSizes, skipFields)
	var builder strings.Builder

	builder.WriteString("| ")
	for i, value := range values {
		// For the "Language" column, update the value using getColoredLanguage.
		if i < len(headers) && headers[i] == "Language" {
			value = getColoredLanguage(value, fieldSizes[i])
		}
		// Format the cell according to the given width.
		builder.WriteString(fmt.Sprintf("%-*s | ", fieldSizes[i], value))
	}
	builder.WriteString("\n")
	return builder.String()
}

// generateMD creates the Markdown table for one or more repositories.
// It uses our latest header & row generators that include computed fields
// and special handling for the Language column via getColoredLanguage.
func generateMD(repoNames []string, year int) string {
	var builder strings.Builder

	// Create a sample instance of RepoStats for header generation.
	// (Only its fields are used, not its values.)
	var statsSample RepoStats

	// Define the skip fields map: these fields will be omitted from the final table.
	skip := map[string]bool{
		"Remote":    true,
		"Files":     true,
		"Frequency": true,
	}

	// Define field sizes for the displayed (remaining) fields.
	// In our RepoStats, after skipping, the order is assumed to be:
	// Repo, Language, Age, Commit, Lines, Size, Mean, Q1, Q2, Q3, Q4.
	// Here we use:
	// - 25 for Repo,
	// - 6 for Language,
	// - 6 for Age,
	// - 15 for Commit,
	// - 6 for Lines,
	// - 7 for Size,
	// - 4 for Mean, and
	// - 3 for each quarter.
	fieldSizes := []int{25, 6, 6, 15, 6, 7, 4, 3, 3, 3, 3}

	// Generate the header row (with bold headers) and the separator line.
	builder.WriteString(generateMarkdownHeader(statsSample, fieldSizes, skip))

	// Record the original directory.
	originalDir := recallDir()

	// Iterate over all the repository names.
	for _, repoName := range repoNames {
		// If processing multiple repositories,
		// change directory to the current repository.
		if len(repoNames) > 1 {
			changeDir(repoName)
		}

		// Collect repository statistics (populates computed fields via populateRepoStats).
		stats, err := populateRepoStats(year)
		checkErr(err)

		// Assign repo name (new field "Repo") to include the repository column.
		stats.Repo = repoName

		// Generate and append a Markdown row for the repository.
		// The row generator will update computed fields and, for the "Language" column,
		// it will call getColoredLanguage.
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
