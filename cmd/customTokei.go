////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"strconv"
	"strings"
)

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

func popualteTokei(tokeiOutput string) (Tokei, string, error) {
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

////////////////////////////////////////////////////////////////////////////////////////////////////

// calculate percentages dominant language
func calculatePercentage(dominant, total int) int {
	if total == 0 {
		return 0
	}
	return (dominant * 100) / total
}

////////////////////////////////////////////////////////////////////////////////////////////////////
