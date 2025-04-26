////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strconv"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func parseTokei(tokeiOutput string) (string, error) {
	lines := strings.Split(tokeiOutput, "\n")
	var totalFiles, totalLines, totalCode, totalComments, totalBlanks int
	var dominantLanguage string
	var dominantFiles, dominantLines, dominantCode, dominantComments, dominantBlanks int

	for _, line := range lines {
		// skip separator lines
		if strings.HasPrefix(line, "=") || strings.TrimSpace(line) == "" {
			continue
		}

		// split the line into parts
		parts := strings.Fields(line)
		if len(parts) < 6 { // ensure the line has enough columns
			continue
		}

		if strings.ToLower(parts[0]) == "total" { // parse the totals row
			totalFiles, _ = strconv.Atoi(parts[1])
			totalLines, _ = strconv.Atoi(parts[2])
			totalCode, _ = strconv.Atoi(parts[3])
			totalComments, _ = strconv.Atoi(parts[4])
			totalBlanks, _ = strconv.Atoi(parts[5])
			continue
		}

		// parse the data for each language
		language := parts[0]
		files, _ := strconv.Atoi(parts[1])
		lines, _ := strconv.Atoi(parts[2])
		code, _ := strconv.Atoi(parts[3])
		comments, _ := strconv.Atoi(parts[4])
		blanks, _ := strconv.Atoi(parts[5])

		// update the dominant language if this one has more files
		if files > dominantFiles {
			dominantLanguage = language
			dominantFiles = files
			dominantLines = lines
			dominantCode = code
			dominantComments = comments
			dominantBlanks = blanks
		}
	}

	// calculate percentages for the dominant language
	if totalFiles == 0 || totalLines == 0 || totalCode == 0 || totalComments == 0 || totalBlanks == 0 {
		return "", fmt.Errorf("total counts are zero, invalid data")
	}
	filesPercentage := (float64(dominantFiles) / float64(totalFiles)) * 100
	linesPercentage := (float64(dominantLines) / float64(totalLines)) * 100
	codePercentage := (float64(dominantCode) / float64(totalCode)) * 100
	commentsPercentage := (float64(dominantComments) / float64(totalComments)) * 100
	blanksPercentage := (float64(dominantBlanks) / float64(totalBlanks)) * 100

	// format the result
	result := fmt.Sprintf(
		"%s %d (%.2f%%) %d (%.2f%%) %d (%.2f%%) %d (%.2f%%) %d (%.2f%%)",
		dominantLanguage,
		dominantFiles, filesPercentage,
		dominantLines, linesPercentage,
		dominantCode, codePercentage,
		dominantComments, commentsPercentage,
		dominantBlanks, blanksPercentage,
	)
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
