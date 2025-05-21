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

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// declarations
var ()

////////////////////////////////////////////////////////////////////////////////////////////////////

// describeCmd
var describeCmd = &cobra.Command{
	Use:     "describe",
	Aliases: []string{"d"},
	Short:   "Describe main repository features",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) + chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `
`,

	Example: `
` + chalk.Cyan.Color("") + ` help ` + chalk.Yellow.Color("") + chalk.Yellow.Color("describe"),

	//////////////////////////////////////////////////////////////////////////////////////////////////

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
			describeOutGit() // '.git' is not found or logged as missing
			if err != nil {
				// handle error listing directories
			}
		} else {
			describeInGit() // '.git' is found
		}

	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// describeOutGit function to be executed if '.git' is not found
func describeOutGit() {

	// repoNames, _ := listDirs(repository)

	populateRepoDescribe()

	// // Generate and print the final report
	// table := generateMD(repoNames, year)
	// fmt.Println(table)
}

// describeInGit function to be executed if '.git' is found
func describeInGit() {

	// // Vectors to hold stats and repo names
	// var repoNames []string

	// // Change repo name report if the repo is "."
	// if repository == "." {
	// 	repository = currentDir()
	// }

	describe, err := populateRepoDescribe()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(describe)

	// Generate the Markdown header using RepoDescribe.
	headerempty := generateHeader(RepoDescribe{})
	fmt.Print(headerempty)

	// Generate the Markdown header using RepoDescribe.
	header := generateHeader(describe)
	fmt.Print(header)

	repos := []RepoDescribe{
		describe,
	}

	generateMDDescribe(repos)

	fmt.Println(repos)

	// // Sample data for repositories.
	// describes := []RepoDescribe{
	// 	{"MIT", "A permissive license", "github.com/user/repo1"},
	// 	{"Apache-2.0", "Enterprise-friendly", "github.com/user/repo2"},
	// }

	// stats := []RepoStats{
	// 	{150, "2 years", "Go", 10000, "2MB", 6, 40, 35, 50, 25},
	// 	{200, "3 years", "Python", 15000, "3MB", 7, 60, 45, 55, 40},
	// }

	// // Generate and output the Markdown table.
	// table := generateMarkdownTable(describes, stats)
	// fmt.Print(table)

	// // collect repo
	// repoNames = append(repoNames, repository)

	// // Generate and print the final report
	// table := generateMD(repoNames, year)
	// // println("here")
	// fmt.Println(table)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(describeCmd)

	// flags
}

////////////////////////////////////////////////////////////////////////////////////////////////////
