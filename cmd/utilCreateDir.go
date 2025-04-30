////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"os"

	"github.com/DanielRivasMD/horus"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// Example of a NotFoundAction to create a directory if it doesn't exist
func createDir(dirPath string) horus.NotFoundAction {
	return func(address string) error {
		fmt.Printf("Attempting to create directory: %s\n", address)
		err := os.Mkdir(address, 0755)
		if err != nil {
			return horus.NewCategorizedHerror(
				"create directory",
				"directory_creation_error",
				"failed to create directory",
				err,
				map[string]any{"path": address},
			)
		}
		fmt.Printf("Directory successfully created: %s\n", address)
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
