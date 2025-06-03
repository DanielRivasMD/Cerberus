////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"strings"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// extracts repository remote URL
func getRemote() (string, error) {
	out, _, err := domovoi.CaptureExecCmd("git", "remote", "-v")
	if err != nil {
		return "", horus.Wrap(err, "getRemote", "failed to capture git remote output")
	}

	return parseRemoteURL(string(out)), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func parseRemoteURL(remoteOutput string) string {
	// split by lines
	lines := strings.Split(remoteOutput, "\n")
	for _, line := range lines {
		// check for 'origin' remote and extract URL
		if strings.HasPrefix(line, "origin") {
			parts := strings.Fields(line) // split line into components
			if len(parts) >= 2 {
				// remote URL typically second field
				return parts[1]
			}
		}
	}
	// return empty string if no URL is found
	return ""
}

////////////////////////////////////////////////////////////////////////////////////////////////////
