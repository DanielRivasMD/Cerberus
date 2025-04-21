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

	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// declarations
var (
	repo string
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
		for _, file := range files {
			if file == "README.md" {
				readmeFound = true
				description, err := parseReadme(repo+"/"+file)
				checkErr(err)
				fmt.Println("Extracted Description:\n", description)
				break
			}
		}

		if !readmeFound {
			fmt.Println("README.md not found in the directory.")
		}
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(guardCmd)

	// flags
	guardCmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
