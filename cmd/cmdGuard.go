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

	// TODO: extract using onefetch & tokei
	Run: func(cmd *cobra.Command, args []string) {
		files, err := listFiles(repo)
		checkErr(err)

		readmeFound := false
		licenseFound := false
		for _, file := range files {
			if file == "README.md" {
				readmeFound = true
				description, err := parseReadme(repo+"/"+file)
				checkErr(err)
				fmt.Println("Extracted Description:\n", description)
			}

			if file == "LICENSE" {
				licenseFound = true
				licenseType, err := detectLicense(repo+"/"+file)
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

		cmdTokei := `tokei`
		execCmd(cmdTokei)

		cmdOnefetch := `onefetch`
		execCmd(cmdOnefetch)
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

// Calculates the age of the repository by fetching the first commit date
func calculateRepoAge(repo string) (string, error) {
    cmd := exec.Command("git", "-C", repo, "log", "--reverse", "--format=%ci")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil // Parse and calculate age
}

// Counts the total commits in the repository
func countCommits(repo string) (int, error) {
    cmd := exec.Command("git", "-C", repo, "rev-list", "--count", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return 0, err
    }
    commits, err := strconv.Atoi(strings.TrimSpace(string(output)))
    return commits, err
}

// Extracts repository remote URL
func getRemoteURL(repo string) (string, error) {
    cmd := exec.Command("git", "-C", repo, "remote", "-v")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return parseRemoteURL(string(output)), nil // Implement helper parseRemoteURL
}

// Parses repo size by calculating file sizes recursively
func calculateRepoSize(repo string) (int64, error) {
    var size int64
    err := filepath.Walk(repo, func(_ string, info os.FileInfo, err error) error {
        if err == nil && !info.IsDir() {
            size += info.Size()
        }
        return err
    })
    return size, err
}

