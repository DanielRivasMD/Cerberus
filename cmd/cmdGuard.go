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

	err := horus.CheckDirExist("nonexistent_dir", createDirectory)
	if err != nil {
		fmt.Println("Error:", err)
		if herr, ok := horus.AsHerror(err); ok {
			fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
			if herr.Err != nil {
				fmt.Printf("  Underlying Error: %v\n", herr.Err)
			}
		}
	}

	err = horus.CheckDirExist("another_nonexistent_dir", logNotFound)
	if err != nil {
		fmt.Println("Error:", err)
		if herr, ok := horus.AsHerror(err); ok {
			fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		}
	}

	err = horus.CheckDirExist("yet_another_nonexistent_dir", nil)
	if err != nil {
		fmt.Println("Error:", err)
		if herr, ok := horus.AsHerror(err); ok {
			fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		}
	}

	err = horus.CheckDirExist("existing_dir", createDirectory)
	if err != nil {
		fmt.Println("Error:", err)
		if herr, ok := horus.AsHerror(err); ok {
			fmt.Printf("  Operation: %s, Message: %s, Details: %v\n", herr.Op, herr.Message, herr.Details)
		}
	}

		gitPresence := dirExist(".git")
		// if err != nil {
		// 	fmt.Println("Error processing config:", err)
		// 	// if msg := horus.UserMessage(err); msg != "" {
		// 	// 	fmt.Printf("  User Message: %s\n", msg)
		// 	// }
		// 	// if step, ok := horus.Detail(err, "step"); ok {
		// 	// 	fmt.Printf("  Step: %v\n", step)
		// 	// }
		// }

		if gitPresence {

			// collect repo data
			stats, ε := populateRepoStats(repo, year)
			checkErr(ε)

			// change repo name report
			if repo == "." {
				repo = currentDir()
			}

			// generate report
			table := generateMD(stats, repo, year)
			fmt.Println(table)
		} else {

			originalDir := recallDir()
			println(originalDir)
			dirs, _ := listFiles(originalDir)
			println(dirs)

		}
	},
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
