////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"bufio"
	"os"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// parseReadme extracts the content under "### Description"
func parseReadme(filename string) (string, error) {
	file, err := os.Open(filename)
	checkErr(err)
	defer file.Close()

	var descriptionLines []string
	scanner := bufio.NewScanner(file)
	inDescription := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// check "## Overview"
		if strings.HasPrefix(line, "## Overview") {
			inDescription = true
			continue
		}

		// stop reading when another heading (##)
		if inDescription && strings.HasPrefix(line, "## ") {
			break
		}

		if inDescription {
			descriptionLines = append(descriptionLines, line)
		}
	}

	checkErr(scanner.Err())

	return strings.Join(descriptionLines, "\n"), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
