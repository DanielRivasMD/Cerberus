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
	"strconv"
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

    // Get the tokei output
    tokeiOutput, err := getTokeiOutput(repo)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    // Parse and retrieve the most common language
    result, err := parseTokeiOutputWithPercentages(tokeiOutput)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println(result)
    }

    // Fetch additional metrics
    repoAge, err := calculateRepoAge(repo)
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

    commitFrequency, err := calculateCommitFrequency(repo)
    checkErr(err)
    fmt.Println("Commit Frequency: ", commitFrequency)

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

func getTokeiOutput(repoPath string) (string, error) {
    // Execute the 'tokei' command on the specified repository path
    cmd := exec.Command("tokei", "-C", repoPath)

    // Capture the command's output
    output, err := cmd.CombinedOutput() // Includes both stdout and stderr
    if err != nil {
        return "", fmt.Errorf("failed to execute tokei: %w. Output: %s", err, output)
    }

    // Convert the output to a string and return it
    return string(output), nil
}

func parseTokeiOutputWithPercentages(tokeiOutput string) (string, error) {
    lines := strings.Split(tokeiOutput, "\n")
    var totalFiles, totalLines, totalCode, totalComments, totalBlanks int
    var dominantLanguage string
    var dominantFiles, dominantLines, dominantCode, dominantComments, dominantBlanks int

    for _, line := range lines {
        // Skip separator lines
        if strings.HasPrefix(line, "=") || strings.TrimSpace(line) == "" {
            continue
        }

        // Split the line into parts
        parts := strings.Fields(line)
        if len(parts) < 6 { // Ensure the line has enough columns
            continue
        }

        if strings.ToLower(parts[0]) == "total" { // Parse the totals row
            totalFiles, _ = strconv.Atoi(parts[1])
            totalLines, _ = strconv.Atoi(parts[2])
            totalCode, _ = strconv.Atoi(parts[3])
            totalComments, _ = strconv.Atoi(parts[4])
            totalBlanks, _ = strconv.Atoi(parts[5])
            continue
        }

        // Parse the data for each language
        language := parts[0]
        files, _ := strconv.Atoi(parts[1])
        lines, _ := strconv.Atoi(parts[2])
        code, _ := strconv.Atoi(parts[3])
        comments, _ := strconv.Atoi(parts[4])
        blanks, _ := strconv.Atoi(parts[5])

        // Update the dominant language if this one has more files
        if files > dominantFiles {
            dominantLanguage = language
            dominantFiles = files
            dominantLines = lines
            dominantCode = code
            dominantComments = comments
            dominantBlanks = blanks
        }
    }

    // Calculate percentages for the dominant language
    if totalFiles == 0 || totalLines == 0 || totalCode == 0 || totalComments == 0 || totalBlanks == 0 {
        return "", fmt.Errorf("total counts are zero, invalid data")
    }
    filesPercentage := (float64(dominantFiles) / float64(totalFiles)) * 100
    linesPercentage := (float64(dominantLines) / float64(totalLines)) * 100
    codePercentage := (float64(dominantCode) / float64(totalCode)) * 100
    commentsPercentage := (float64(dominantComments) / float64(totalComments)) * 100
    blanksPercentage := (float64(dominantBlanks) / float64(totalBlanks)) * 100

    // Format the result
    result := fmt.Sprintf(
        "%s %d (%.2f%%) %d (%.2f%%) %d (%.2f%%) %d (%.2f%%) %d (%.2f%%)",
        dominantLanguage,
        dominantFiles, filesPercentage,
        dominantLines, linesPercentage,
        dominantCode, codePercentage,
        dominantComments, commentsPercentage,
        dominantBlanks, blanksPercentage,
    )
    return result, nil
}

func calculateRepoAge(repo string) (string, error) {
    // Get the first commit's date using git log
    cmd := exec.Command("git", "-C", repo, "log", "--reverse", "--format=%ci")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to execute git command: %w", err)
    }

    // Split the output into individual lines (one per commit date)
    commitDates := strings.Split(string(output), "\n")
    if len(commitDates) == 0 || strings.TrimSpace(commitDates[0]) == "" {
        return "", fmt.Errorf("no commit dates found in the repository")
    }

    // Use the first line (the oldest commit date)
    firstCommitDateStr := strings.TrimSpace(commitDates[0])

    // Parse the first commit date
    layout := "2006-01-02 15:04:05 -0700" // Git commit date format
    firstCommitDate, err := time.Parse(layout, firstCommitDateStr)
    if err != nil {
        return "", fmt.Errorf("failed to parse commit date: %w", err)
    }

    // Calculate the difference between the first commit date and the current date
    currentDate := time.Now()
    repoAge := currentDate.Sub(firstCommitDate)

    // Format the age as a human-readable string (e.g., "3 years and 45 days")
    years := int(repoAge.Hours() / (24 * 365))
    days := int(repoAge.Hours()/(24)) % 365
    formattedAge := fmt.Sprintf("%d years and %d days", years, days)

    return formattedAge, nil
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

// Computes commit frequency, grouping by time intervals
func calculateCommitFrequency(repo string) (map[string]int, error) {
    cmd := exec.Command("git", "-C", repo, "log", "--format=%ci")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    return groupCommitsByInterval(string(output)), nil // Implement groupCommitsByInterval
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

func groupCommitsByInterval(commitDates string) map[string]int {
    // Create a map to store the commit frequency
    commitFrequency := make(map[string]int)

    // Split the input into individual dates
    dates := strings.Split(commitDates, "\n")

    // Define a time format matching the Git commit date output
    layout := "2006-01-02 15:04:05 -0700"

    for _, dateStr := range dates {
        if strings.TrimSpace(dateStr) == "" {
            continue // Skip empty lines
        }

        // Parse the commit date
        commitTime, err := time.Parse(layout, dateStr)
        if err != nil {
            fmt.Println("Error parsing date:", err)
            continue
        }

        // Group by interval: e.g., year-week or year-month
        // Example: Use "2006-01" for year-month grouping
        interval := commitTime.Format("2006-01")

        // Increment the count for the interval
        commitFrequency[interval]++
    }

    return commitFrequency
}

////////////////////////////////////////////////////////////////////////////////////////////////////
