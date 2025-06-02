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

// Alignment represents a simple alignment setting.
type Alignment struct {
	Dir string
}

var (
	// Predefined alignment values.
	AlignLeft   = Alignment{"left"}
	AlignRight  = Alignment{"right"}
	AlignCenter = Alignment{"center"}
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

// centerAligned centers the content within width characters.
// It splits the extra space into left and right portions.
func centerAligned(content string, width int) string {
	visible := runewidth.StringWidth(content)
	padTotal := width - visible
	if padTotal <= 0 {
		return content
	}
	padLeft := padTotal / 2
	padRight := padTotal - padLeft
	return strings.Repeat(" ", padLeft) + content + strings.Repeat(" ", padRight)
}

// formatCell applies any header‑specific processing (coloring, trimming, etc.)
// and then aligns the text using the alignment specified in the aligners map.
// If no alignment is provided for a header, it defaults to left alignment for the first column
// and right alignment for subsequent columns.
func formatCell(header, value string, fieldSize, idx int, aligners map[string]Alignment) string {
	// Header-specific processing.
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
	// Only apply dimming when the value is "0". Otherwise, leave the value intact.
	if idx > 0 && value == "0" {
		value = getDimIfZero(value, fieldSize)
	}

	// Lookup alignment for this header.
	var alignment Alignment
	var ok bool
	if aligners != nil {
		alignment, ok = aligners[header]
	}
	// Default: if no alignment is set, use left for the first column,
	// otherwise default to right alignment.
	if !ok {
		if idx == 0 {
			alignment = AlignLeft
		} else {
			alignment = AlignRight
		}
	}

	// Now select the proper helper based on the alignment value.
	var cell string
	switch alignment.Dir {
	case "left":
		cell = leftAligned(value, fieldSize)
	case "center":
		cell = centerAligned(value, fieldSize)
	case "right":
		cell = rightAligned(value, fieldSize)
	default:
		panic(horus.NewHerror("formatCell", "invalid alignment value", fmt.Errorf("invalid alignment: %s", alignment.Dir), nil))
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
// ANSI Bold is applied via chalk.
// It accepts an aligners map whose keys are header names and values indicate alignment.
func generateMarkdownHeader(v interface{}, fieldSizes []int, skipFields map[string]bool, aligners map[string]Alignment) string {
	headers := getHeaders(v, skipFields)
	var builder strings.Builder

	// Header row.
	builder.WriteString("| ")
	for i, header := range headers {
		// Look up the desired alignment for this header.
		alignment, ok := aligners[header]
		if !ok {
			// Default to right alignment if none is provided.
			alignment = AlignRight
		}

		var cell string
		// Use a switch statement to select the proper helper.
		switch alignment.Dir {
		case "left":
			cell = leftAligned(header, fieldSizes[i])
		case "center":
			cell = centerAligned(header, fieldSizes[i])
		case "right":
			cell = rightAligned(header, fieldSizes[i])
		default:
			panic(horus.NewHerror("generateMarkdownHeader", "invalid alignment value", fmt.Errorf("invalid alignment: %s", alignment.Dir), nil))
		}

		// Wrap with Bold ANSI codes.
		boldCell := chalk.Bold.TextStyle(cell)
		builder.WriteString(boldCell + " | ")
	}
	builder.WriteString("\n")

	// Separator row: Add two extra characters (one each for the leading and trailing spaces).
	builder.WriteString("|")
	for i := range headers {
		builder.WriteString(strings.Repeat("-", fieldSizes[i]+2) + "|")
	}
	builder.WriteString("\n")
	return builder.String()
}

// generateMarkdownRow creates a Markdown table row for a single struct instance.
// It updates computed fields (Mean, Q1–Q4) when v is a *RepoStats and applies alignment
// based on the provided aligners map.
func generateMarkdownRow(v interface{}, fieldSizes []int, skipFields map[string]bool, year int, aligners map[string]Alignment) string {
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

	headers := getHeaders(v, skipFields)
	values := getValues(v, skipFields)
	var builder strings.Builder

	builder.WriteString("| ")
	for i, value := range values {
		// Process each cell using our generic helper.
		cell := formatCell(headers[i], value, fieldSizes[i], i, aligners)
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
// - aligners: a map whose keys are header names and values are Alignment settings (left, right, center)
func generateGenericMD[T any](
	sample *T,
	repoNames []string,
	populateFunc func(repoName string) (*T, error),
	fieldSizes []int,
	skip map[string]bool,
	extra int,
	aligners map[string]Alignment,
) string {
	var builder strings.Builder

	// Generate header row.
	builder.WriteString(generateMarkdownHeader(sample, fieldSizes, skip, aligners))

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

		builder.WriteString(generateMarkdownRow(instance, fieldSizes, skip, extra, aligners))
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
