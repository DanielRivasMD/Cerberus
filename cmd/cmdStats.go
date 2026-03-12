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

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////////////////////////

func StatsCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("stats", runStats))

	cmd.Flags().StringVarP(&repository, "repo", "r", ".", "Repository")
	cmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "Year for commit frequency calculation")
	cmd.Flags().StringVarP(&aggregation, "time", "t", "yearly", "Time aggregation: quarterly or yearly")
	cmd.Flags().BoolVarP(&plot, "plot", "p", true, "Render as graph or markdown")

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runStats(cmd *cobra.Command, args []string) {
	err := handleGit("stats", rootFlags.verbose)
	horus.CheckErr(err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: repository is declared here globally and used within functions by leaking. the intend implementation is to use it as a flag struct, probably pass a generic struct to the function handleGit
var (
	repository  string
	year        int
	aggregation string
	plot        bool
)

type statsFlag struct {
}

////////////////////////////////////////////////////////////////////////////////////////////////////
