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
		files, err := listFiles(repo)
		checkErr(err)

		readmeFound := false
		licenseFound := false
		for _, file := range files {
			if file == "README.md" {
				readmeFound = true
				description, err := parseReadme(repo + "/" + file)
				checkErr(err)
				fmt.Println("Extracted Description:\n", description)
			}

			if file == "LICENSE" {
				licenseFound = true
				licenseType, err := detectLicense(repo + "/" + file)
				checkErr(err)
				fmt.Println("License is: ", licenseType)
			}
		}

		if !readmeFound {
			fmt.Println("README.md not found in the directory.")
		}
		if !licenseFound {
			fmt.Println("LICENSE not found in the directory.")
		}

		// tokei
		tokeiOut, tokeiErr, ε := execCmdCapture("tokei", "-C")
		checkErr(ε)

		fmt.Println(tokeiOut)
		fmt.Println(tokeiErr)

		// Parse and retrieve the most common language
		result, err := parseTokei(tokeiOut)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println(result)
		}

		// Fetch additional metrics
		repoAge, err := repoAge(repo)
		checkErr(err)
		fmt.Println("Repo Age: ", repoAge)

		commitCount, err := countCommits(repo)
		checkErr(err)
		fmt.Println("Number of Commits: ", commitCount)

		remoteURL, err := getRemote(repo)
		checkErr(err)
		fmt.Println("Repo Remote: ", remoteURL)

		size, err := repoSize(repo)
		fmt.Println("Human-readable repo size:", size)

		commitFrequency, err := commitFrequency(repo, year)
		checkErr(err)
		fmt.Println("Commit Frequency: ", commitFrequency)

		averageCommits := averageCommits(commitFrequency)
		fmt.Printf("Average Commits Per Month: %.2f\n", averageCommits)
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(guardCmd)

	// flags
	guardCmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository")
	guardCmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "Year for commit frequency calculation (default: current year)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
