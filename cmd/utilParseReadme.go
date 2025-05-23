////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"bufio"
	"os"
	"strings"

	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

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

	// Join the lines with a space (removing any newline characters).
	result := strings.Join(descriptionLines, " ")
	result = strings.TrimSpace(result)

	// Truncate the result if it exceeds maxChars.
	if len(result) > maxChars {
		result = result[:maxChars]
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
