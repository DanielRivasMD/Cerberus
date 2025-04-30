/*
Copyright © 2025 Daniel Rivas <danielrivasmd@gmail.com>

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

import (
	"fmt"
	"os"
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// declarations
var (
	repo string
	year int
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// guardCmd
var guardCmd = &cobra.Command{
	Use:   "guard",
	Short: "" + chalk.Yellow.Color("") + ".",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) + chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `
`,

	Example: `
` + chalk.Cyan.Color("") + ` help ` + chalk.Yellow.Color("") + chalk.Yellow.Color("guard"),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {

		// Path to check for the '.git' directory
		dirPath := ".git"

		// Using horus library's CheckDirExist to check for '.git'
		err := horus.CheckDirExist(dirPath, horus.LogNotFound("The '.git' directory is missing."))

		// Decide what to do based on the result
		if err == nil {
			inGit() // '.git' is found
		} else if os.IsNotExist(err) || horus.IsHerror(err) {
			outGit() // '.git' is not found or logged as missing
		} else {
			fmt.Println(horus.FormatError(err, horus.JSONFormatter)) // Handle unexpected errors
		}

		// // Path to check for the '.git' directory
		// dirPath := ".git"

		// // Using horus library's CheckDirExist to check for '.git'
		// err := horus.CheckDirExist(
		// 	dirPath,
		// 	horus.LogNotFound("The '.git' directory is missing."), // Log warning if not found
		// )

		// // Decide what to do based on the result
		// if err != nil {
		// 	// Handle the error if needed
		// 	if os.IsNotExist(err) {
		// 		outGit() // '.git' is not found
		// 	} else {
		// 		fmt.Println(horus.FormatError(err, horus.JSONFormatter)) // Other errors
		// 	}
		// } else {
		// 	inGit() // '.git' is found
		// }

		// // // Example 1: Checking for a directory with a custom NotFoundAction
		// dirPath := "test"
		// err := horus.CheckDirExist(dirPath, createDir(dirPath))
		// if err != nil {
		// 	// Print error in JSON format
		// 	fmt.Println(horus.FormatError(err, horus.JSONFormatter))
		// }

		// // Example 2: Logging a not-found resource
		// logAction := horus.LogNotFound("This is a test resource that is missing.")
		// err = logAction("./nonexistent_resource")
		// if err != nil {
		// 	fmt.Println(horus.FormatError(err, horus.JSONFormatter))
		// }

		// // directoryToCheck := ".git"

		// // Define the action to take if the directory does NOT exist
		// notFoundAction := horus.NotFoundAction(notGit) // Type conversion to match NotFoundAction

		// err := horus.CheckDirExist(directoryToCheck, notFoundAction)
		// if err != nil {
		// 	fmt.Println("Error during directory check:", err)
		// 	if herr, ok := horus.AsHerror(err); ok {
		// 		fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		// 	}
		// 	fmt.Println("Failed to determine Git status due to an error.")
		// 	os.Exit(1)
		// 	return
		// }

		// If CheckDirExist returns nil, it means the initial check was successful.
		// The directory either existed, or the notFoundAction ran without error.
		// In this specific scenario where notGit doesn't create the directory,
		// if we reach here, the directory likely existed.

		// // Execute logic for when the directory exists.
		// // We can assume it exists if CheckDirExist didn't return an error.
		// inGit()

		// err := horus.CheckDirExist("nonexistent_dir", createDirectory)
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// 	if herr, ok := horus.AsHerror(err); ok {
		// 		fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		// 		if herr.Err != nil {
		// 			fmt.Printf("  Underlying Error: %v\n", herr.Err)
		// 		}
		// 	}
		// }

		// err = horus.CheckDirExist("another_nonexistent_dir", logNotFound)
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// 	if herr, ok := horus.AsHerror(err); ok {
		// 		fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		// 	}
		// }

		// err = horus.CheckDirExist("yet_another_nonexistent_dir", nil)
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// 	if herr, ok := horus.AsHerror(err); ok {
		// 		fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		// 	}
		// }

		// err = horus.CheckDirExist("existing_dir", createDirectory)
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// 	if herr, ok := horus.AsHerror(err); ok {
		// 		fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		// 	}
		// }

		// err := horus.CheckDirExist(".git", executeLogic)
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// 	if herr, ok := horus.AsHerror(err); ok {
		// 		fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		// 	}
		// }

		// gitPresence := dirExist(".git")
		// if err != nil {
		// 	fmt.Println("Error processing config:", err)
		// 	// if msg := horus.UserMessage(err); msg != "" {
		// 	// 	fmt.Printf("  User Message: %s\n", msg)
		// 	// }
		// 	// if step, ok := horus.Detail(err, "step"); ok {
		// 	// 	fmt.Printf("  Step: %v\n", step)
		// 	// }
		// }

		// if gitPresence {

		// 	// collect repo data
		// 	stats, ε := populateRepoStats(repo, year)
		// 	checkErr(ε)

		// 	// change repo name report
		// 	if repo == "." {
		// 		repo = currentDir()
		// 	}

		// 	// generate report
		// 	table := generateMD(stats, repo, year)
		// 	fmt.Println(table)
		// } else {

		// 	originalDir := recallDir()
		// 	println(originalDir)
		// 	dirs, _ := listFiles(originalDir)
		// 	println(dirs)

		// }
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// inGit function to be executed if '.git' is found
func inGit() {
	fmt.Println("'.git' directory found! Executing inGit...")
	// Add your specific logic here
}

// outGit function to be executed if '.git' is not found
func outGit() {
	fmt.Println("'.git' directory not found! Executing outGit...")
	// Add your specific logic here
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(guardCmd)

	// flags
	guardCmd.Flags().StringVarP(&repo, "repo", "r", ".", "Repository")
	guardCmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "Year for commit frequency calculation (default: current year)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
