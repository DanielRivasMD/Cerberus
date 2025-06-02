////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/mattn/go-runewidth"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// leftAligned formats the content so that it is padded on the right to meet the required width.
func leftAligned(content string, width int) string {
	visible := runewidth.StringWidth(content)
	pad := width - visible
	if pad < 0 {
		pad = 0
	}
	return content + strings.Repeat(" ", pad)
}

// rightAligned formats the content so that it is padded on the left to meet the required width.
func rightAligned(content string, width int) string {
	visible := runewidth.StringWidth(content)
	pad := width - visible
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + content
}

// formatCell applies header-specific formatting (coloring, trimming, dimming) then aligns the text.
func formatCell(header, value string, fieldSize, idx int) string {
	// Apply special processing for particular columns.
	switch header {
	case "Age":
		value = getColoredAge(value, fieldSize)
	case "Language":
		value = getColoredLanguage(value, fieldSize)
	case "Size":
		value = getColoredSize(value, fieldSize)
	case "Remote":
		value = TrimGitHubRemote(value)
	}
	// For non-first columns, dim zero-values.
	if idx > 0 {
		value = getDimIfZero(value, fieldSize)
	}

	// Choose alignment based on column index.
	var cell string
	if idx == 0 {
		cell = leftAligned(value, fieldSize)
	} else {
		cell = rightAligned(value, fieldSize)
	}
	return cell
}

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

// generateMarkdownHeader creates the Markdown header row and a separator row.
// For terminal rendering, ANSI Bold is applied (using chalk).
// In this version, we use separate helper functions for left- and right alignment.
func generateMarkdownHeader(v interface{}, fieldSizes []int, skipFields map[string]bool) string {
	headers := getHeaders(v, skipFields)
	var builder strings.Builder

	// Header row: For this example, let's say the first column ("Repo") is left aligned,
	// all others right aligned.
	builder.WriteString("| ")
	for i, header := range headers {
		var cell string
		if i == 0 {
			cell = leftAligned(header, fieldSizes[i])
		} else {
			cell = rightAligned(header, fieldSizes[i])
		}
		// Wrap with Bold codes.
		boldCell := chalk.Bold.TextStyle(cell)
		builder.WriteString(boldCell + " | ")
	}
	builder.WriteString("\n")

	// Separator row (simple dashes): add two extra characters for the leading and trailing spaces.
	builder.WriteString("|")
	for i := range headers {
		builder.WriteString(strings.Repeat("-", fieldSizes[i]+2) + "|")
	}
	builder.WriteString("\n")
	return builder.String()
}

// generateMarkdownRow creates a Markdown table row for a single struct instance.
// It updates computed fields (Mean, Q1–Q4) when v is a *RepoStats and applies alignment
// (first column left aligned, others right aligned).
func generateMarkdownRow(v interface{}, fieldSizes []int, skipFields map[string]bool, year int) string {
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

	headers := getHeaders(v, skipFields)
	values := getValues(v, skipFields)
	var builder strings.Builder

	builder.WriteString("| ")
	for i, value := range values {
		// Process each cell generically via our helper:
		cell := formatCell(headers[i], value, fieldSizes[i], i)
		builder.WriteString(cell + " | ")
	}
	builder.WriteString("\n")
	return builder.String()
}

// --------------------------------------------------------------------------------
// Generic Table Generator Using Go Generics (Go 1.18+)
// --------------------------------------------------------------------------------

// generateGenericMD is a generic function that builds a Markdown table for any type T.
// - sample: a pointer to a sample instance (used for header generation)
// - repoNames: slice of repository names to process
// - populateFunc: a function that, given a repo name, returns a pointer to a populated instance of T
// - fieldSizes: a slice of column widths (in characters)
// - skip: a map of field names to skip
// - extra: an extra parameter (for example, the year for stats formatting)
func generateGenericMD[T any](
	sample *T,
	repoNames []string,
	populateFunc func(repoName string) (*T, error),
	fieldSizes []int,
	skip map[string]bool,
	extra int,
) string {
	var builder strings.Builder

	// Generate header row.
	builder.WriteString(generateMarkdownHeader(sample, fieldSizes, skip))

	originalDir, err := domovoi.RecallDir()
	horus.CheckErr(err)

	for _, repoName := range repoNames {
		// Change directory if processing multiple repositories.
		if len(repoNames) > 1 {
			err := domovoi.ChangeDir(repoName)
			horus.CheckErr(err)
		}

		instance, err := populateFunc(repoName)
		if err != nil {
			// Wrap the error using Horus and then panic.
			panic(horus.NewHerror("generateGenericMD", "populateFunc failed for repository", err, map[string]any{"repoName": repoName}))
		}

		builder.WriteString(generateMarkdownRow(instance, fieldSizes, skip, extra))
		err = domovoi.ChangeDir(originalDir)
		horus.CheckErr(err)
	}

	return builder.String()
}

//////////////////////////////////////////////////////////////////////////////////////////////////

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

// TrimGitHubRemote removes the "https://github.com/" prefix from the provided remote string,
// if it exists. Otherwise, it returns the string unchanged.
func TrimGitHubRemote(remote string) string {
	const prefix = "https://github.com/"
	return strings.TrimPrefix(remote, prefix)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
