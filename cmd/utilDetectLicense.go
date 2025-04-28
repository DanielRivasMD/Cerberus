////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"os"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// detectLicense identifies license type
func detectLicense(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	checkErr(err)

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

///////////////////////////////////////////////////////////////////////////////////////////////////
