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

// parseReadme extracts the content under "### Description"
func parseReadme(filename string) (string, error) {
	file, err := os.Open(filename)
	horus.CheckErr(err)
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

	horus.CheckErr(scanner.Err())

	return strings.Join(descriptionLines, "\n"), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
