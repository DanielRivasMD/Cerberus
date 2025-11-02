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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/guptarohit/asciigraph"
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

// renderCommitStats handles the formatting of commit data.
// If plot is true, it returns an ASCII graph generated with asciigraph,
// otherwise it returns a Markdown table, using aggregation type ("quarterly" or "yearly")
// and the base year provided.
func renderCommitStats(commitCounts []float64, aggregation string, year int, plot bool) string {
	if plot {
		// Generate an ASCII graph.
		graph := asciigraph.Plot(
			commitCounts,
			asciigraph.Caption("Commits Over Time ("+aggregation+")"),
			asciigraph.Height(10), // adjust as needed
			asciigraph.Width(50),  // adjust as needed
		)
		return graph
	}

	// Otherwise, build a Markdown table.
	var builder strings.Builder
	builder.WriteString("| Period    | Commits |\n")
	builder.WriteString("|-----------|---------|\n")
	if aggregation == "quarterly" {
		// Expect commitCounts to have 4 data points for the quarters.
		quarters := []string{"Q1", "Q2", "Q3", "Q4"}
		for i, count := range commitCounts {
			// If more than four data points, cycle through quarter labels.
			label := quarters[i%len(quarters)]
			builder.WriteString(fmt.Sprintf("| %-9s | %.0f     |\n", label, count))
		}
	} else {
		// For "yearly" aggregation, assume commitCounts contain data for a sequence of years
		// ending with 'year'. Calculate the start year accordingly.
		dataLen := len(commitCounts)
		startYear := year - dataLen + 1
		years := make([]int, dataLen)
		for i := 0; i < dataLen; i++ {
			years[i] = startYear + i
		}
		sort.Ints(years)
		for i, yr := range years {
			builder.WriteString(fmt.Sprintf("| %4d      | %.0f     |\n", yr, commitCounts[i]))
		}
	}
	return builder.String()
}

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

	err := handleGit("stats", verbose)
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
