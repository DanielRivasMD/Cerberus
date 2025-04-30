////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"os"

	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// createFile attempts to create a file at the specified path and reports errors using horus.
func createFile(path string) error {
	fmt.Printf("Attempting to create file '%s'.\n", path)

	file, err := os.Create(path)
	if err != nil {
		// Report the error using horus
		herr := horus.NewCategorizedHerror(
			"create file",
			"file_creation_error",
			fmt.Sprintf("failed to create file '%s'", path),
			err,
			map[string]any{"file_path": path},
		)

		// Log the error for debugging
		fmt.Println(horus.FormatError(herr, horus.JSONFormatter))
		return herr
	}
	defer file.Close()

	fmt.Printf("File '%s' created successfully.\n", path)
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
