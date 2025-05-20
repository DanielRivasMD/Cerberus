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

// statsCmd
var statsCmd = &cobra.Command{
	Use:     "stats",
	Aliases: []string{"s"},
	Short:   "Collect repository stats",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) + chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `
`,

	Example: `
` + chalk.Cyan.Color("") + ` help ` + chalk.Yellow.Color("") + chalk.Yellow.Color("guard"),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {

		// Path to check for the '.git' directory
		dirPath := ".git"

		// Check directory existence with a placeholder action in case of missing directory.
		ok, err := horus.CheckDirExist(dirPath, horus.NullAction(), verbose)
		if err != nil {
			// handle error: maybe log it, stop execution, etc.
		}

		if !ok {
			// Directory doesn't exist even after our neutral action.
			// Now you can list directories in the parent folder or take other actions.
			statsOutGit() // '.git' is not found or logged as missing
			if err != nil {
				// handle error listing directories
			}
		} else {
			statsInGit() // '.git' is found
		}

	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// statsOutGit function to be executed if '.git' is not found
func statsOutGit() {
	// Collect repositories
	repoNames, _ := listDirs(repository)

	// Generate and print the final report
	table := generateMD(repoNames, year)
	fmt.Println(table)
}

// statsInGit function to be executed if '.git' is found
func statsInGit() {
	// Vectors to hold stats and repo names
	var repoNames []string

	// Change repo name report if the repo is "."
	if repository == "." {
		repository = currentDir()
	}

	// collect repo
	repoNames = append(repoNames, repository)

	table := generateMD(repoNames, year)
	fmt.Println(table)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(statsCmd)

	// flags
	statsCmd.Flags().StringVarP(&repository, "repo", "r", ".", "Repository")
	statsCmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "Year for commit frequency calculation (default: current year)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
