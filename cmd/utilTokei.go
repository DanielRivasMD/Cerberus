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
	"fmt"
	"strconv"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// Pair holds a number and a percentage.
type Pair struct {
	Number     int
	Percentage int
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Tokei holds dominant language stats.
type Tokei struct {
	Files    Pair
	Lines    Pair
	Code     Pair
	Comments Pair
	Blanks   Pair
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// populateTokei parses tokei -C output and returns dominant language stats.
func populateTokei(tokeiOutput string) (Tokei, string, error) {
	lines := strings.Split(tokeiOutput, "\n")
	var totalFiles, totalLines, totalCode, totalComments, totalBlanks int
	var dominantFiles, dominantLines, dominantCode, dominantComments, dominantBlanks int
	var dominantLanguage string

	for _, line := range lines {
		if strings.HasPrefix(line, "=") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 6 {
			continue
		}
		if strings.ToLower(parts[0]) == "total" {
			totalFiles, _ = strconv.Atoi(parts[1])
			totalLines, _ = strconv.Atoi(parts[2])
			totalCode, _ = strconv.Atoi(parts[3])
			totalComments, _ = strconv.Atoi(parts[4])
			totalBlanks, _ = strconv.Atoi(parts[5])
			continue
		}
		lang := parts[0]
		files, _ := strconv.Atoi(parts[1])
		linesCount, _ := strconv.Atoi(parts[2])
		code, _ := strconv.Atoi(parts[3])
		comments, _ := strconv.Atoi(parts[4])
		blanks, _ := strconv.Atoi(parts[5])

		if linesCount > dominantLines {
			dominantLanguage = lang
			dominantFiles = files
			dominantLines = linesCount
			dominantCode = code
			dominantComments = comments
			dominantBlanks = blanks
		}
	}

	if totalFiles == 0 || totalLines == 0 {
		return Tokei{}, "", fmt.Errorf("total counts are zero, invalid tokei data")
	}

	result := Tokei{
		Files:    Pair{Number: dominantFiles, Percentage: calculatePercentage(dominantFiles, totalFiles)},
		Lines:    Pair{Number: dominantLines, Percentage: calculatePercentage(dominantLines, totalLines)},
		Code:     Pair{Number: dominantCode, Percentage: calculatePercentage(dominantCode, totalCode)},
		Comments: Pair{Number: dominantComments, Percentage: calculatePercentage(dominantComments, totalComments)},
		Blanks:   Pair{Number: dominantBlanks, Percentage: calculatePercentage(dominantBlanks, totalBlanks)},
	}
	return result, dominantLanguage, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
