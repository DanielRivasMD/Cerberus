/*
Copyright © 2026 Daniel Rivas <danielrivasmd@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	AlignLeft   = Alignment{"left"}
	AlignRight  = Alignment{"right"}
	AlignCenter = Alignment{"center"}
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------------
// Git handling
// ----------------------------------------------------------------------

// handleGit processes a repository report (stats, describe, remember)
func handleGit(reportType string, verbose bool) error {
	ok, err := domovoi.DirExist(".git", horus.NullAction(false), verbose)
	if err != nil {
		return err
	}

	var repoNames []string
	if !ok {
		repoNames, err = domovoi.ListDirs(repository)
		if err != nil {
			return err
		}
	} else {
		if repository == "." {
			repoName, err := domovoi.CurrentDir()
			if err != nil {
				return err
			}
			repository = repoName
		}
		repoNames = append(repoNames, repository)
	}

	switch reportType {
	case "stats":
		fmt.Println(generateStatsMD(repoNames, year))
	case "describe":
		fmt.Println(generateDescribeMD(repoNames))
	case "remember":
		return generateRememberCSV(repoNames)
	default:
		return fmt.Errorf("unknown report type: %s", reportType)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// resolveRemoteURL returns the origin URL of a Git repository
func resolveRemoteURL(repo string) (string, error) {
	repoPath := filepath.Join(repository, repo)
	out, _, err := domovoi.CaptureExecCmd("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TrimGitHubRemote removes the "https://github.com/" prefix if present
func TrimGitHubRemote(remote string) string {
	const prefix = "https://github.com/"
	return strings.TrimPrefix(remote, prefix)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// calculateRepoAgeInMonths converts "Xy Ym" format to total months
func calculateRepoAgeInMonths(age string) int {
	years, months := 0, 0
	_, err := fmt.Sscanf(age, "%dy %dm", &years, &months)
	if err != nil {
		// fallback: assume 0
		return 0
	}
	return (years * 12) + months
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------------
// Markdown / CSV generation (generic)
// ----------------------------------------------------------------------

// generateGenericMD builds a Markdown table for any type T.
//   - sample: pointer to sample instance (for headers)
//   - repoNames: list of repository names to process
//   - populateFunc: function that returns a populated T given a repo name
//   - fieldSizes: slice of column widths (characters)
//   - skip: map of field names to exclude
//   - extra: extra integer parameter (e.g., year for stats)
//   - aligners: map of header -> Alignment
//   - outputFile: if non-empty, CSV is written to this file
func generateGenericMD[T any](
	sample *T,
	repoNames []string,
	populateFunc func(repoName string) (*T, error),
	fieldSizes []int,
	skip map[string]bool,
	extra int,
	aligners map[string]Alignment,
	outputFile string,
) string {
	var markdownBuilder strings.Builder
	var csvRows [][]string

	markdownBuilder.WriteString(generateMarkdownHeader(sample, fieldSizes, skip, aligners))
	if outputFile != "" {
		csvHeaders := getHeaders(sample, skip)
		csvRows = append(csvRows, csvHeaders)
	}

	originalDir, err := domovoi.RecallDir()
	horus.CheckErr(err)

	for _, repoName := range repoNames {
		if len(repoNames) > 1 {
			err := domovoi.ChangeDir(repoName)
			horus.CheckErr(err)
		}

		instance, err := populateFunc(repoName)
		if err != nil {
			panic(horus.NewHerror("generateGenericMD", "populateFunc failed for repository", err,
				map[string]any{"repoName": repoName}))
		}

		markdownBuilder.WriteString(generateMarkdownRow(instance, fieldSizes, skip, extra, aligners))
		if outputFile != "" {
			values := getValues(instance, skip)
			csvRows = append(csvRows, values)
		}

		err = domovoi.ChangeDir(originalDir)
		horus.CheckErr(err)
	}

	markdownOutput := markdownBuilder.String()

	if outputFile != "" {
		// Write CSV to file
		f, err := os.Create(outputFile)
		if err != nil {
			panic(horus.Wrap(err, "generateGenericMD", "failed to create CSV file"))
		}
		defer f.Close()
		w := csv.NewWriter(f)
		for _, row := range csvRows {
			if err := w.Write(row); err != nil {
				panic(horus.Wrap(err, "generateGenericMD", "failed to write CSV row"))
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			panic(horus.Wrap(err, "generateGenericMD", "CSV flush error"))
		}
	}

	return markdownOutput
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateMarkdownHeader creates the header and separator rows.
func generateMarkdownHeader(v interface{}, fieldSizes []int, skipFields map[string]bool, aligners map[string]Alignment) string {
	headers := getHeaders(v, skipFields)
	var builder strings.Builder

	// Header row
	builder.WriteString("| ")
	for i, header := range headers {
		alignment, ok := aligners[header]
		if !ok {
			alignment = AlignRight
		}
		var cell string
		switch alignment.Dir {
		case "left":
			cell = leftAligned(header, fieldSizes[i])
		case "center":
			cell = centerAligned(header, fieldSizes[i])
		default: // right
			cell = rightAligned(header, fieldSizes[i])
		}
		builder.WriteString(chalk.Bold.TextStyle(cell) + " | ")
	}
	builder.WriteString("\n")

	// Separator row
	builder.WriteString("|")
	for i := range headers {
		builder.WriteString(strings.Repeat("-", fieldSizes[i]+2) + "|")
	}
	builder.WriteString("\n")
	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateMarkdownRow creates a single table row.
func generateMarkdownRow(v interface{}, fieldSizes []int, skipFields map[string]bool, extra int, aligners map[string]Alignment) string {
	// If v is *RepoStats, update computed fields (Mean, Q1-Q4) based on extra (year)
	if repoStats, ok := v.(*RepoStats); ok {
		repoAgeMonths := calculateRepoAgeInMonths(repoStats.Age)
		averageCommits := 0
		if repoAgeMonths > 0 {
			averageCommits = repoStats.Commit / repoAgeMonths
		}
		quarterlyCommits := map[string]int{
			"Q1": repoStats.Frequency[fmt.Sprintf("%d-01", extra)] +
				repoStats.Frequency[fmt.Sprintf("%d-02", extra)] +
				repoStats.Frequency[fmt.Sprintf("%d-03", extra)],
			"Q2": repoStats.Frequency[fmt.Sprintf("%d-04", extra)] +
				repoStats.Frequency[fmt.Sprintf("%d-05", extra)] +
				repoStats.Frequency[fmt.Sprintf("%d-06", extra)],
			"Q3": repoStats.Frequency[fmt.Sprintf("%d-07", extra)] +
				repoStats.Frequency[fmt.Sprintf("%d-08", extra)] +
				repoStats.Frequency[fmt.Sprintf("%d-09", extra)],
			"Q4": repoStats.Frequency[fmt.Sprintf("%d-10", extra)] +
				repoStats.Frequency[fmt.Sprintf("%d-11", extra)] +
				repoStats.Frequency[fmt.Sprintf("%d-12", extra)],
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
	for i, val := range values {
		cell := formatCell(headers[i], val, fieldSizes[i], i, aligners)
		builder.WriteString(cell + " | ")
	}
	builder.WriteString("\n")
	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// formatCell applies header‑specific coloring and alignment.
func formatCell(header, value string, fieldSize, idx int, aligners map[string]Alignment) string {
	// Header-specific processing
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
	// Dim "0" values in numeric columns (idx > 0)
	if idx > 0 && value == "0" {
		value = getDimIfZero(value, fieldSize)
	}

	// Determine alignment
	alignment, ok := aligners[header]
	if !ok {
		if idx == 0 {
			alignment = AlignLeft
		} else {
			alignment = AlignRight
		}
	}

	switch alignment.Dir {
	case "left":
		return leftAligned(value, fieldSize)
	case "center":
		return centerAligned(value, fieldSize)
	default:
		return rightAligned(value, fieldSize)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------------
// Field reflection helpers
// ----------------------------------------------------------------------

func getHeaders(v interface{}, skipFields map[string]bool) []string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if !skipFields[name] {
			headers = append(headers, name)
		}
	}
	return headers
}

func getValues(v interface{}, skipFields map[string]bool) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	var values []string
	for i := 0; i < val.NumField(); i++ {
		name := typ.Field(i).Name
		if skipFields[name] {
			continue
		}
		values = append(values, fmt.Sprintf("%v", val.Field(i).Interface()))
	}
	return values
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------------
// Alignment helpers
// ----------------------------------------------------------------------

func leftAligned(content string, width int) string {
	vis := runewidth.StringWidth(content)
	pad := width - vis
	if pad < 0 {
		pad = 0
	}
	return content + strings.Repeat(" ", pad)
}

func rightAligned(content string, width int) string {
	vis := runewidth.StringWidth(content)
	pad := width - vis
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + content
}

func centerAligned(content string, width int) string {
	vis := runewidth.StringWidth(content)
	pad := width - vis
	if pad <= 0 {
		return content
	}
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + content + strings.Repeat(" ", right)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------------
// Colored cell formatters
// ----------------------------------------------------------------------

func getColoredAge(age string, width int) string {
	parts := strings.Fields(age)
	var padded string
	if len(parts) < 2 {
		padded = fmt.Sprintf("%-*s", width, age)
	} else {
		yearPart, monthPart := parts[0], parts[1]
		yearWidth := runewidth.StringWidth(yearPart)
		monthWidth := runewidth.StringWidth(monthPart)
		filler := width - (yearWidth + monthWidth)
		if filler < 1 {
			filler = 1
		}
		padded = yearPart + strings.Repeat(" ", filler) + monthPart
	}
	if !strings.Contains(age, "0y") {
		return chalk.Bold.TextStyle(padded)
	}
	return chalk.Dim.TextStyle(padded)
}

func getColoredLanguage(language string, width int) string {
	parts := strings.Fields(language)
	var padded string
	if len(parts) < 2 {
		padded = fmt.Sprintf("%-*s", width, language)
	} else {
		first, second := parts[0], parts[1]
		firstWidth := runewidth.StringWidth(first)
		secondWidth := runewidth.StringWidth(second)
		filler := width - (firstWidth + secondWidth)
		if filler < 1 {
			filler = 1
		}
		padded = first + strings.Repeat(" ", filler) + second
	}
	base := strings.ToLower(strings.Split(language, " ")[0])
	switch base {
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
		return padded
	}
}

func getColoredSize(size string, width int) string {
	parts := strings.Fields(size)
	var padded string
	if len(parts) < 2 {
		padded = fmt.Sprintf("%-*s", width, size)
	} else {
		numPart, unitPart := parts[0], parts[1]
		numWidth := runewidth.StringWidth(numPart)
		unitWidth := runewidth.StringWidth(unitPart)
		filler := width - (numWidth + unitWidth)
		if filler < 1 {
			filler = 1
		}
		padded = numPart + strings.Repeat(" ", filler) + unitPart
	}
	if strings.Contains(size, "MB") {
		return chalk.Bold.TextStyle(padded)
	}
	return chalk.Dim.TextStyle(padded)
}

func getDimIfZero(value string, width int) string {
	padded := fmt.Sprintf("%*s", width, value)
	if value == "0" {
		return chalk.Dim.TextStyle(padded)
	}
	return padded
}

////////////////////////////////////////////////////////////////////////////////////////////////////
