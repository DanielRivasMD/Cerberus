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
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: refactor as struct
var (
	repository  string
	year        int
	aggregation string
	plot        bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	statsCmd := MakeCmd("stats", runStats)

	statsCmd.Flags().StringVarP(&repository, "repo", "r", ".", "Repository")
	statsCmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "Year for commit frequency calculation")
	statsCmd.Flags().StringVarP(&aggregation, "time", "t", "yearly", "Time aggregation: quarterly or yearly")
	statsCmd.Flags().BoolVarP(&plot, "plot", "p", true, "Render as graph or markdown")

	rootCmd.AddCommand(statsCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runStats(cmd *cobra.Command, args []string) {
	err := handleGit("stats", RootFlags.verbose)
	horus.CheckErr(err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
