/*
Copyright © 2024 Daniel Rivas <danielrivasmd@gmail.com>

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
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:     "cerberus",
	Long:    helpRoot,
	Example: exampleRoot,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func Execute() {
	horus.CheckErr(rootCmd.Execute())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var RootFlags rootFlags

type rootFlags struct {
	verbose bool
	output  string
}

type defaults struct {
	repoLen     int
	commitLen   int
	ageLen      int
	languageLen int
	linesLen    int
	sizeLen     int
	meanLen     int
	qLen        int
	overviewLen int
	licenseLen  int
	remoteLen   int
}

var Defaults = defaults{
	repoLen:     25,
	commitLen:   6,
	ageLen:      6,
	languageLen: 15,
	linesLen:    6,
	sizeLen:     7,
	meanLen:     4,
	qLen:        3,
	overviewLen: 92,
	licenseLen:  7,
	remoteLen:   95,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&RootFlags.verbose, "verbose", "v", false, "Verbose")
	rootCmd.PersistentFlags().StringVarP(&RootFlags.output, "output", "o", "", "File output")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func generateRememberCSV(repoNames []string) error {
	var outFile *os.File
	var err error

	if strings.TrimSpace(RootFlags.output) != "" {
		// create or truncate the file
		outFile, err = os.Create(RootFlags.output)
		if err != nil {
			return horus.Wrap(err, "generateRememberCSV", "failed to create output file: "+RootFlags.output)
		}
		defer outFile.Close()
	} else {
		outFile = os.Stdout
	}

	w := csv.NewWriter(outFile)
	defer w.Flush()

	// optional header
	if err := w.Write([]string{"repoName", "repoURL"}); err != nil {
		return err
	}

	for _, repo := range repoNames {
		remoteURL, err := resolveRemoteURL(repo)
		if err != nil {
			remoteURL = ""
		}
		if err := w.Write([]string{repo, remoteURL}); err != nil {
			return err
		}
	}
	return nil
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
//   - sample: a pointer to a sample instance (used for header generation)
//   - repoNames: slice of repository names to process
//   - populateFunc: a function that, given a repo name, returns a pointer to a populated instance of T
//   - fieldSizes: a slice of column widths (in characters) used for Markdown formatting
//   - skip: a map of field names to skip
//   - extra: an extra parameter (for example, the year for stats formatting)
//   - aligners: a map whose keys are header names and values are Alignment settings (left, right, center)
//   - outputFile: if non-empty, the generated CSV table will be written to this file.
//     The Markdown output is always returned.
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

	// Generate Markdown header.
	markdownBuilder.WriteString(generateMarkdownHeader(sample, fieldSizes, skip, aligners))
	// For CSV output, get the raw (unformatted) headers.
	if outputFile != "" {
		csvHeaders := getHeaders(sample, skip)
		csvRows = append(csvRows, csvHeaders)
	}

	originalDir, err := domovoi.RecallDir()
	horus.CheckErr(err)

	// Process each repository.
	for _, repoName := range repoNames {
		// Change directory if processing multiple repositories.
		if len(repoNames) > 1 {
			err := domovoi.ChangeDir(repoName)
			horus.CheckErr(err)
		}

		instance, err := populateFunc(repoName)
		if err != nil {
			panic(horus.NewHerror("generateGenericMD", "populateFunc failed for repository", err, map[string]any{"repoName": repoName}))
		}

		// Append Markdown row.
		markdownBuilder.WriteString(generateMarkdownRow(instance, fieldSizes, skip, extra, aligners))
		// Append CSV row with raw values.
		if outputFile != "" {
			values := getValues(instance, skip)
			csvRows = append(csvRows, values)
		}

		err = domovoi.ChangeDir(originalDir)
		horus.CheckErr(err)
	}

	markdownOutput := markdownBuilder.String()

	// Write CSV output to file if required.
	if outputFile != "" {
		var csvBuf bytes.Buffer
		csvWriter := csv.NewWriter(&csvBuf)
		for _, row := range csvRows {
			if err := csvWriter.Write(row); err != nil {
				panic(horus.Wrap(err, "generateGenericMD", "failed to write CSV row"))
			}
		}
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			panic(horus.Wrap(err, "generateGenericMD", "failed to flush CSV writer"))
		}

		err := os.WriteFile(outputFile, csvBuf.Bytes(), 0644)
		if err != nil {
			panic(horus.Wrap(err, "generateGenericMD", "failed to write CSV output to file"))
		}
	}

	return markdownOutput
}

////////////////////////////////////////////////////////////////////////////////////////////////////

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

func resolveRemoteURL(repo string) (string, error) {
	repoPath := filepath.Join(repository, repo)
	out, _, err := domovoi.CaptureExecCmd("git", "-C", repoPath, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func repoAge() (string, error) {
	// get first commit
	out, _, ε := domovoi.CaptureExecCmd("git", "log", "--reverse", "--format=%ci")
	horus.CheckErr(ε)

	// split output into individual lines
	commitDates := strings.Split(string(out), "\n")
	if len(commitDates) == 0 || strings.TrimSpace(commitDates[0]) == "" {
		return "", fmt.Errorf("no commit dates found in the repository")
	}

	// use oldest commit date
	firstCommitDateStr := strings.TrimSpace(commitDates[0])

	// parse first commit date
	layout := "2006-01-02 15:04:05 -0700" // git commit date format
	firstCommitDate, ε := time.Parse(layout, firstCommitDateStr)
	horus.CheckErr(ε)

	// calculate difference
	currentDate := time.Now()
	years := currentDate.Year() - firstCommitDate.Year()
	months := int(currentDate.Month()) - int(firstCommitDate.Month())

	// Adjust for negative months (e.g., December - February)
	if months < 0 {
		years--
		months += 12
	}

	// Format the result
	formattedAge := fmt.Sprintf("%dy %dm", years, months)

	return formattedAge, nil
}

// cloneRepository clones a Git repository from the specified URL into the target directory.
// It wraps any errors using horus.Wrap to provide detailed context.
func cloneRepository(repoURL, targetDir string) error {
	out, _, err := domovoi.CaptureExecCmd("git", "clone", repoURL, targetDir)
	if err != nil {
		return horus.Wrap(err, "cloneRepository", "failed to clone repository: "+repoURL)
	}

	// Optionally process the output if needed.
	_ = strings.TrimSpace(string(out))
	return nil
}

// cloneRepositoriesFromCSV reads a CSV file whose rows contain Git repository details,
// where the first column is the repository name and the second column is the repository URL.
// It then clones each repository into a subdirectory under targetDir using the provided repository name.
// If targetDir is an empty string, it defaults to "." (the current directory).
// Before opening the CSV file, it verifies its existence using domovoi.FileExist with an anonymous NotFoundAction.
func cloneRepositoriesFromCSV(csvFile, targetDir string) error {
	// Set default targetDir if not provided.
	if strings.TrimSpace(targetDir) == "" {
		targetDir = "."
	}

	// Verify that the CSV file exists.
	_, err := domovoi.FileExist(csvFile, func(filePath string) (bool, error) {
		panic(horus.NewHerror("cloneRepositoriesFromCSV", "CSV file does not exist", nil,
			map[string]any{"csvFile": filePath}))
	}, true)
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to check existence of CSV file: "+csvFile)
	}

	// Open the CSV file.
	file, err := os.Open(csvFile)
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to open CSV file: "+csvFile)
	}
	defer file.Close()

	// Read all records from CSV.
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return horus.Wrap(err, "cloneRepositoriesFromCSV", "failed to parse CSV file: "+csvFile)
	}

	// Skip header if present.
	start := 0
	if len(records) > 0 && strings.EqualFold(records[0][0], "repoName") {
		start = 1
	}

	// Process each row. Expecting two columns: [repoName, repoURL].
	for i := start; i < len(records); i++ {
		record := records[i]
		if len(record) < 2 {
			continue
		}

		repoName := strings.TrimSpace(record[0])
		repoURL := strings.TrimSpace(record[1])

		// Skip if either field is empty.
		if repoName == "" || repoURL == "" {
			continue
		}

		finalTargetDir := filepath.Join(targetDir, repoName)

		// Clone the repository.
		if err := cloneRepository(repoURL, finalTargetDir); err != nil {
			return horus.Wrap(err, "cloneRepositoriesFromCSV",
				"failed to clone repository at row "+strconv.Itoa(i+1))
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// parse repo size
func repoSize() (string, error) {
	var size int
	err := filepath.Walk(".", func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += int(info.Size())
		}
		return err
	})
	return formatRepoSize(size), err
}

func formatRepoSize(size int) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%d GB", size/GB)
	case size >= MB:
		return fmt.Sprintf("%d MB", size/MB)
	case size >= KB:
		return fmt.Sprintf("%d KB", size/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func commitFrequency(year int) (map[string]int, error) {
	// initialize map with all months set 0
	commitFrequency := map[string]int{
		fmt.Sprintf("%d-01", year): 0,
		fmt.Sprintf("%d-02", year): 0,
		fmt.Sprintf("%d-03", year): 0,
		fmt.Sprintf("%d-04", year): 0,
		fmt.Sprintf("%d-05", year): 0,
		fmt.Sprintf("%d-06", year): 0,
		fmt.Sprintf("%d-07", year): 0,
		fmt.Sprintf("%d-08", year): 0,
		fmt.Sprintf("%d-09", year): 0,
		fmt.Sprintf("%d-10", year): 0,
		fmt.Sprintf("%d-11", year): 0,
		fmt.Sprintf("%d-12", year): 0,
	}

	// get commit dates within specified year
	out, _, ε := domovoi.CaptureExecCmd("git", "log", "--since", fmt.Sprintf("%d-01-01", year), "--until", fmt.Sprintf("%d-12-31", year), "--format=%ci")
	horus.CheckErr(ε)

	// process output & group by month
	commitDates := strings.Split(string(out), "\n")
	layout := "2006-01-02 15:04:05 -0700" // git date format

	for _, dateStr := range commitDates {
		if strings.TrimSpace(dateStr) == "" {
			continue // skip empty lines
		}

		commitTime, err := time.Parse(layout, dateStr)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}

		// use "YYYY-MM" format for grouping by month
		month := commitTime.Format("2006-01")
		commitFrequency[month]++
	}

	return commitFrequency, nil
}

// detectLicense identifies license type
func detectLicense(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	horus.CheckErr(err)

	content := strings.ToLower(string(data))

	// mapping common license identifiers
	licenseKeywords := map[string]string{
		"mit license":                "MIT",
		"apache license":             "Apache-2.0",
		"gnu general public license": "GPL",
		"bsd license":                "BSD",
		"mozilla public license":     "MPL",
		"creative commons":           "CC",
		"eclipse public license":     "EPL",
	}

	for keyword, licenseType := range licenseKeywords {
		if strings.Contains(content, keyword) {
			return licenseType, nil
		}
	}

	return "Unknown License", nil
}

// trimmer returns the substring of desc up to the first period or newline.
func trimmer(desc string) string {
	if idx := strings.IndexAny(desc, ".\n"); idx >= 0 {
		return strings.TrimSpace(desc[:idx+1])
	}
	return strings.TrimSpace(desc)
}

// parseReadme extracts the content under "## Overview" from the given file,
// joins the lines with a space (thus removing newlines),
// and then limits the returned string to at most maxChars characters.
func parseReadme(filename string, maxChars int) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", horus.NewCategorizedHerror(
			"parse readme",
			"file_open_error",
			"failed to open file",
			err,
			map[string]any{"filename": filename},
		)
	}
	defer file.Close()

	var descriptionLines []string
	scanner := bufio.NewScanner(file)
	inDescription := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Begin capturing after the "## Overview" heading.
		if strings.HasPrefix(line, "## Overview") {
			inDescription = true
			continue
		}

		// Stop capturing when a new main heading is encountered.
		if inDescription && strings.HasPrefix(line, "## ") {
			break
		}

		if inDescription {
			descriptionLines = append(descriptionLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", horus.NewCategorizedHerror(
			"parse readme",
			"scanner_error",
			"scanner encountered an error",
			err,
			nil,
		)
	}

	// Join the lines with a newline
	result := strings.Join(descriptionLines, "\n")
	result = trimmer(result)

	// Truncate the result if it exceeds maxChars.
	if len(result) > maxChars {
		result = result[:maxChars]
	}
	return result, nil
}

// counts total commits
func countCommits() (int, error) {
	out, _, ε := domovoi.CaptureExecCmd("git", "rev-list", "--count", "HEAD")
	horus.CheckErr(ε)

	commits, err := strconv.Atoi(strings.TrimSpace(string(out)))
	return commits, err
}

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
		// If the value is "0", return the padded string in a dim style
		return chalk.Dim.TextStyle(padded)
	}
	return padded
}

// calculateRepoAgeInMonths calculates repository age in months given an "Xy Ym" format
func calculateRepoAgeInMonths(age string) int {
	years, months := 0, 0
	_, err := fmt.Sscanf(age, "%dy %dm", &years, &months) // Parse "Xy Ym" format
	horus.CheckErr(err)
	return (years * 12) + months
}

// TrimGitHubRemote removes the "https://github.com/" prefix from the provided remote string,
// if it exists. Otherwise, it returns the string unchanged
func TrimGitHubRemote(remote string) string {
	const prefix = "https://github.com/"
	return strings.TrimPrefix(remote, prefix)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type Pair struct {
	Number     int
	Percentage int
}

type Tokei struct {
	Files    Pair
	Lines    Pair
	Code     Pair
	Comments Pair
	Blanks   Pair
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func populateTokei(tokeiOutput string) (Tokei, string, error) {
	lines := strings.Split(tokeiOutput, "\n")
	var totalFiles, totalLines, totalCode, totalComments, totalBlanks int
	var dominantFiles, dominantLines, dominantCode, dominantComments, dominantBlanks int
	var dominantLanguage string

	for _, line := range lines {
		// skip separator lines
		if strings.HasPrefix(line, "=") || strings.TrimSpace(line) == "" {
			continue
		}

		// split line into parts
		parts := strings.Fields(line)
		if len(parts) < 6 { // ensure enough columns
			continue
		}

		// parse total row
		if strings.ToLower(parts[0]) == "total" {
			totalFiles, _ = strconv.Atoi(parts[1])
			totalLines, _ = strconv.Atoi(parts[2])
			totalCode, _ = strconv.Atoi(parts[3])
			totalComments, _ = strconv.Atoi(parts[4])
			totalBlanks, _ = strconv.Atoi(parts[5])
			continue
		}

		// parse data each language
		language := parts[0]
		files, _ := strconv.Atoi(parts[1])
		lines, _ := strconv.Atoi(parts[2])
		code, _ := strconv.Atoi(parts[3])
		comments, _ := strconv.Atoi(parts[4])
		blanks, _ := strconv.Atoi(parts[5])

		// update dominant language
		if lines > dominantLines {
			dominantLanguage = language
			dominantFiles = files
			dominantLines = lines
			dominantCode = code
			dominantComments = comments
			dominantBlanks = blanks
		}
	}

	// error if total zero (invalid data)
	if totalFiles == 0 || totalLines == 0 {
		return Tokei{}, "", fmt.Errorf("total counts are zero, invalid data")
	}

	// Tokei struct
	result := Tokei{
		Files: Pair{
			Number:     dominantFiles,
			Percentage: calculatePercentage(dominantFiles, totalFiles),
		},
		Lines: Pair{
			Number:     dominantLines,
			Percentage: calculatePercentage(dominantLines, totalLines),
		},
		Code: Pair{
			Number:     dominantCode,
			Percentage: calculatePercentage(dominantCode, totalCode),
		},
		Comments: Pair{
			Number:     dominantComments,
			Percentage: calculatePercentage(dominantComments, totalComments),
		},
		Blanks: Pair{
			Number:     dominantBlanks,
			Percentage: calculatePercentage(dominantBlanks, totalBlanks),
		},
	}

	return result, dominantLanguage, nil
}


// calculate percentages dominant language
func calculatePercentage(dominant, total int) int {
	if total == 0 {
		return 0
	}
	return (dominant * 100) / total
}

////////////////////////////////////////////////////////////////////////////////////////////////////
