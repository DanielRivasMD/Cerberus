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
	repository string
	year       int
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

	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// inGit function to be executed if '.git' is found
func inGit() {

	// Vectors to hold stats and repo names
	var repoNames []string

	// Change repo name report if the repo is "."
	if repository == "." {
		repository = currentDir()
	}

	// collect repo
	repoNames = append(repoNames, repository)

	// Generate and print the final report
	table := generateMD(repoNames, year)
	fmt.Println(table)
}

// outGit function to be executed if '.git' is not found
func outGit() {
	// fmt.Println("'.git' directory not found! Executing outGit...")

	repoNames, _ := listDirs(repository)

	// Generate and print the final report
	table := generateMD(repoNames, year)
	fmt.Println(table)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(guardCmd)

	// flags
	guardCmd.Flags().StringVarP(&repository, "repo", "r", ".", "Repository")
	guardCmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "Year for commit frequency calculation (default: current year)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
