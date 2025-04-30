////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"os"

	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// listDirs lists all subdirectories within the given directory path and reports errors using horus.
func listDirs(dirPath string) ([]string, error) {
	var directories []string

	// Open the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// Create a categorized error using horus
		herr := horus.NewCategorizedHerror(
			"list directories",
			"directory_error",
			fmt.Sprintf("failed to read directory '%s'", dirPath),
			err,
			map[string]any{"target_directory": dirPath},
		)

		// Optionally, log the error for debugging
		fmt.Println(horus.FormatError(herr, horus.JSONFormatter))
		return nil, herr
	}

	// Iterate through directory entries
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	return directories, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
