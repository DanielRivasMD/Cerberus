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

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"strconv"
	"time"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	repository  string
	year        int
	aggregation string // Aggregation option: "quarterly" or "yearly"
	plot        bool   // Plot flag: true = render as ASCII graph; false = output as Markdown table
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var statsCmd = &cobra.Command{
	Use:     "stats",
	Short:   "Collect repository stats",
	Long:    helpStats,
	Example: exampleStats,

	Run: runStats,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringVarP(&repository, "repo", "r", ".", "Repository")
	statsCmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "Year for commit frequency calculation (default: current year)")

	// New flags for aggregation type and plotting mode.
	statsCmd.Flags().StringVarP(&aggregation, "time", "t", "yearly", "Time aggregation for commit frequency: quarterly or yearly")
	statsCmd.Flags().BoolVarP(&plot, "plot", "p", true, "Render data as a graph (true) or as markdown (false)")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runStats(cmd *cobra.Command, args []string) {

	err := handleGit("stats", RootFlags.verbose)
	horus.CheckErr(err)

	// // Sample commit data over time.
	// // Replace these sample values with real aggregated commits data.
	// var commitCounts []float64
	// if aggregation == "quarterly" {
	// 	// Example: data for 4 quarters.
	// 	commitCounts = []float64{15, 22, 18, 30}
	// } else {
	// 	// Default to "yearly": sample data for several years.
	// 	commitCounts = []float64{50, 65, 80, 100, 90, 75}
	// }

	// // Use the helper function to render commit statistics.
	// output := renderCommitStats(commitCounts, aggregation, year, plot)
	// fmt.Println(output)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// RepoStats represents the repository statistics.
// Empty fields will use their zero values initially,
// ensuring that the automatic header generator includes these columns.
type RepoStats struct {
	Repo      string         // Repo name
	Commit    int            // Total number of commits
	Age       string         // Age of the repository (e.g., "3y 2m")
	Language  string         // Main programming language of the repository, along with details (e.g., "Go 80%")
	Lines     int            // Total lines of code
	Files     int            // Total number of files; may be populated later
	Size      string         // Size of the repository (e.g., "5MB")
	Frequency map[string]int // Commit frequency by month (e.g., "2025-01": 10)

	// The following are computed values, added to support additional table columns.
	Mean int // Average commits per month (to be computed via custom logic)
	Q1   int // Commits for Quarter 1 (computed)
	Q2   int // Commits for Quarter 2 (computed)
	Q3   int // Commits for Quarter 3 (computed)
	Q4   int // Commits for Quarter 4 (computed)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// populateRepoStats gathers statistics for a repository for the given year,
// wrapping any errors using the Horus library for additional context.
func populateRepoStats(year int) (RepoStats, error) {
	// initialize RepoStats
	stats := RepoStats{}

	// fetch repository metrics
	tokeiOut, _, err := domovoi.CaptureExecCmd("tokei", "-C")
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to capture tokei output")
	}

	// define language & lines
	tokeiStats, language, err := populateTokei(tokeiOut)
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to parse tokei output")
	}
	stats.Language = language + " " + strconv.Itoa(tokeiStats.Lines.Percentage) + "%"
	stats.Lines = tokeiStats.Lines.Number

	// define age
	age, err := repoAge()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to determine repository age")
	}
	stats.Age = age

	// define number of commits
	commitCount, err := countCommits()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to count commits")
	}
	stats.Commit = commitCount

	// define repo size
	size, err := repoSize()
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to determine repository size")
	}
	stats.Size = size

	// define commit frequency
	commitFrequency, err := commitFrequency(year)
	if err != nil {
		return stats, horus.Wrap(err, "populateRepoStats", "failed to fetch commit frequency")
	}
	stats.Frequency = commitFrequency

	return stats, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// generateStatsMD generates the Markdown table for the stats command.
func generateStatsMD(repoNames []string, year int) string {
	// Define field sizes for: Repo, Language, Age, Commit, Lines, Size, Mean, Q1, Q2, Q3, Q4.
	// (Note: Ensure that the order of fieldSizes matches the header order in your RepoStats struct.)
	fieldSizes := []int{Defaults.repoLen, Defaults.commitLen, Defaults.ageLen, Defaults.languageLen, Defaults.linesLen, Defaults.sizeLen, Defaults.meanLen, Defaults.qLen, Defaults.qLen, Defaults.qLen, Defaults.qLen}
	skip := map[string]bool{
		"Remote":    true,
		"Files":     true,
		"Frequency": true,
	}

	var sample RepoStats // Sample instance for header generation

	// populateFunc for stats command.
	populateFunc := func(repoName string) (*RepoStats, error) {
		s, err := populateRepoStats(year)
		if err != nil {
			return nil, err
		}
		s.Repo = repoName
		return &s, nil
	}

	// Create an aligners map such that only the "Repo" column is left aligned.
	// For any header not explicitly provided, the default behavior in formatCell
	// will be: if index==0 then left aligned, otherwise right aligned.
	aligners := map[string]Alignment{
		"Repo":     AlignLeft,
		"Language": AlignLeft,
	}

	// Pass the year as the extra parameter.
	return generateGenericMD(&sample, repoNames, populateFunc, fieldSizes, skip, year, aligners, RootFlags.output)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
