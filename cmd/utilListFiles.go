////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"os"

	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// listFiles returns a slice of filenames from the given directory.
// If an error occurs while reading the directory, the error is wrapped
// using horus.NewCategorizedHerror.
func listFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, horus.NewCategorizedHerror(
			"list files",
			"io",
			"failed to list files in directory",
			err,
			map[string]any{"directory": directory},
		)
	}

	var files []string
	for _, entry := range entries {
		// ensures only regular files are considered.
		if entry.Type().IsRegular() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
