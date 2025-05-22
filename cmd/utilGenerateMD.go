////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"reflect"
	"strconv"
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
		// Wrap with Bold codes.
		boldCell := chalk.Bold.TextStyle(cell)
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
		// For the "Age" column, apply color.
		if i < len(headers) && headers[i] == "Age" {
			value = getColoredAge(value, fieldSizes[i])
		}

		// For the "Language" column, apply color.
		if i < len(headers) && headers[i] == "Language" {
			value = getColoredLanguage(value, fieldSizes[i])
		}

		// For the "Size" column, apply color.
		if i < len(headers) && headers[i] == "Size" {
			value = getColoredSize(value, fieldSizes[i])
		}

		// For other columns, dim if zero.
		if i > 0 {
			value = getDimIfZero(value, fieldSizes[i])
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

// generateDescribeMD creates the Markdown table for repositories described via RepoDescribe.
func generateDescribeMD(repoNames []string) string {
	var builder strings.Builder

	// Create a sample instance of RepoDescribe for header generation.
	var describeSample RepoDescribe

	// Our generic functions use a skipFields map.
	// For this struct, no field is skipped.
	skip := map[string]bool{}

	// Define the field sizes (in characters) for each column in order: Repo, Remote, Overview, License.
	fieldSizes := []int{20, 40, 50, 20}

	builder.WriteString(generateMarkdownHeader(describeSample, fieldSizes, skip))

	// Record the original directory.
	originalDir := recallDir()

	// Iterate over each repository name.
	for _, repoName := range repoNames {
		// Change directory if processing multiple repositories.
		if len(repoNames) > 1 {
			changeDir(repoName)
		}

		// Populate repository description.
		describe, err := populateRepoDescribe()
		if err != nil {
			panic(err)
		}

		// Set the Repo field from the external repository name.
		describe.Repo = repoName

		// Append a row for this repository.
		// The extra 'year' parameter is passed as 0 since it's not used here.
		builder.WriteString(generateMarkdownRow(&describe, fieldSizes, skip, 0))

		// Return to the original directory.
		changeDir(originalDir)
	}

	return builder.String()
}

// generateStatsMD creates the Markdown table for one or more repositories.
// It uses our header and row generators (which update computed fields and right align all but the first column).
func generateStatsMD(repoNames []string, year int) string {
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

// getColoredAge pads the input age string to the provided width,
// splitting it into two parts (e.g., "1y" and "3m" from "1y 3m") so that
// the year part is left aligned and the month part is right aligned, then
// applies chalk color: bold if the record does not contain "0y", otherwise dim.
func getColoredAge(age string, width int) string {
	// Split the age string on whitespace.
	parts := strings.Fields(age)
	var padded string
	if len(parts) < 2 {
		// If there are fewer than two parts, simply pad the entire string to the width.
		padded = fmt.Sprintf("%-"+strconv.Itoa(width)+"s", age)
	} else {
		yearPart := parts[0]
		monthPart := parts[1]
		// Calculate the visible widths.
		yearWidth := runewidth.StringWidth(yearPart)
		monthWidth := runewidth.StringWidth(monthPart)
		// Calculate filler spaces ensuring at least one space between tokens.
		fillerWidth := width - (yearWidth + monthWidth)
		if fillerWidth < 1 {
			fillerWidth = 1
		}
		// Construct the padded age string.
		padded = yearPart + strings.Repeat(" ", fillerWidth) + monthPart
	}
	// If the age record does not contain "0y", render in bold; otherwise, use dim.
	if !strings.Contains(age, "0y") {
		return chalk.Bold.TextStyle(padded)
	}
	return chalk.Dim.TextStyle(padded)
}

// splitting by whitespace so that the second token is right aligned.
// Color is applied based on the base language (the first token) using chalk.
func getColoredLanguage(language string, width int) string {
	parts := strings.Fields(language)
	var padded string
	if len(parts) < 2 {
		// Fallback: simply pad the whole string to the required width.
		padded = fmt.Sprintf("%-"+strconv.Itoa(width)+"s", language)
	} else {
		first := parts[0]
		second := parts[1]
		firstWidth := runewidth.StringWidth(first)
		secondWidth := runewidth.StringWidth(second)
		// Ensure there is at least one space between tokens.
		fillerWidth := width - (firstWidth + secondWidth)
		if fillerWidth < 1 {
			fillerWidth = 1
		}
		padded = first + strings.Repeat(" ", fillerWidth) + second
	}
	baseLanguage := strings.ToLower(strings.Split(language, " ")[0])
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

// getColoredSize pads the input age string to the provided width,
// splitting it into two parts (e.g., "1y" and "3m" from "1y 3m") so that
// the year part is left aligned and the month part is right aligned, then
// applies chalk color: bold if the record contains "MB", otherwise dim.
func getColoredSize(age string, width int) string {
	// Split the age string on whitespace.
	parts := strings.Fields(age)
	var padded string
	if len(parts) < 2 {
		// If there are fewer than two parts, simply pad the entire string to the width.
		padded = fmt.Sprintf("%-"+strconv.Itoa(width)+"s", age)
	} else {
		yearPart := parts[0]
		monthPart := parts[1]
		// Calculate visible widths.
		yearWidth := runewidth.StringWidth(yearPart)
		monthWidth := runewidth.StringWidth(monthPart)
		// Calculate filler spaces ensuring at least one space between tokens.
		fillerWidth := width - (yearWidth + monthWidth)
		if fillerWidth < 1 {
			fillerWidth = 1
		}
		// Construct the padded age string.
		padded = yearPart + strings.Repeat(" ", fillerWidth) + monthPart
	}
	// If the original age string contains "MB", render in bold; otherwise, use dim.
	if strings.Contains(age, "MB") {
		return chalk.Bold.TextStyle(padded)
	}
	return chalk.Dim.TextStyle(padded)
}

func getDimIfZero(value string, width int) string {
	// Right-align the value within the given width.
	padded := fmt.Sprintf("%*s", width, value)
	if value == "0" {
		// If the value is "0", return the padded string in a dim style.
		return chalk.Dim.TextStyle(padded)
	}
	return padded
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// calculateRepoAgeInMonths calculates repository age in months given an "Xy Ym" format.
func calculateRepoAgeInMonths(age string) int {
	years, months := 0, 0
	fmt.Sscanf(age, "%dy %dm", &years, &months) // Parse "Xy Ym" format.
	return (years * 12) + months
}

////////////////////////////////////////////////////////////////////////////////////////////////////
