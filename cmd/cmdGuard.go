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
	"os"
	"os/exec"
	"fmt"
	"path/filepath"
	"strings"
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

    remoteURL, err := getRemoteURL(repo)
    checkErr(err)
    fmt.Println("Repo Remote: ", remoteURL)

		repoSize, err := calculateRepoSize(repo)
		checkErr(err)
		fmt.Printf("Repo Size: %d bytes\n", repoSize)

		humanReadableSize := formatRepoSize(repoSize)
		fmt.Println("Human-readable repo size:", humanReadableSize)

    commitFrequency, err := calculateCommitFrequency(repo, year)
    checkErr(err)
    fmt.Println("Commit Frequency: ", commitFrequency)

        averageCommits := calculateAverageCommits(commitFrequency)
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

func formatRepoSize(size int64) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
    )

    switch {
    case size >= GB:
        return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
    case size >= MB:
        return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
    case size >= KB:
        return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
    default:
        return fmt.Sprintf("%d bytes", size)
    }
}

func parseRemoteURL(remoteOutput string) string {
    // Split output by lines
    lines := strings.Split(remoteOutput, "\n")
    for _, line := range lines {
        // Check for 'origin' remote and extract URL
        if strings.HasPrefix(line, "origin") {
            parts := strings.Fields(line) // Split line into components
            if len(parts) >= 2 {
                // Remote URL is typically the second field
                return parts[1]
            }
        }
    }
    // Return empty string if no URL is found
    return ""
}

func calculateCommitFrequency(repoPath string, year int) (map[string]int, error) {
    // Initialize a map with all months set to 0
    commitFrequency := map[string]int{
        fmt.Sprintf("%d-01", year): 0,
        fmt.Sprintf("%d-02", year): 0,
        fmt.Sprintf("%d-03", year): 0,
        fmt.Sprintf("%d-04", year): 0,
        fmt.Sprintf("%d-05", year): 0,
        fmt.Sprintf("%d-06", year): 0,
        fmt.Sprintf("%d-07", year): 0,
        fmt.Sprintf("%d-08", year): 0,
        fmt.Sprintf("%d-09", year): 0,
        fmt.Sprintf("%d-10", year): 0,
        fmt.Sprintf("%d-11", year): 0,
        fmt.Sprintf("%d-12", year): 0,
    }

    // Use git log to get commit dates within the specified year
    cmd := exec.Command("git", "-C", repoPath, "log", "--since", fmt.Sprintf("%d-01-01", year), "--until", fmt.Sprintf("%d-12-31", year), "--format=%ci")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to fetch commit dates: %w", err)
    }

    // Process the output and group by month
    commitDates := strings.Split(string(output), "\n")
    layout := "2006-01-02 15:04:05 -0700" // Git date format

    for _, dateStr := range commitDates {
        if strings.TrimSpace(dateStr) == "" {
            continue // Skip empty lines
        }

        commitTime, err := time.Parse(layout, dateStr)
        if err != nil {
            fmt.Println("Error parsing date:", err)
            continue
        }

        // Use "YYYY-MM" format for grouping by month
        month := commitTime.Format("2006-01")
        commitFrequency[month]++
    }

    return commitFrequency, nil
}

// Calculates the average number of commits per month
func calculateAverageCommits(commitFrequency map[string]int) float64 {
    totalCommits := 0
    for _, count := range commitFrequency {
        totalCommits += count
    }

    // Total months in a year
    return float64(totalCommits) / 12.0
}

////////////////////////////////////////////////////////////////////////////////////////////////////
