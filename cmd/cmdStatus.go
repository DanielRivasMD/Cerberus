/*
Copyright © 2026 Daniel Rivas <danielrivasmd@gmail.com>

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
	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	statusRepo  string
	statusFetch bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func StatusCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("status", runStatus))

	cmd.Flags().StringVarP(&statusRepo, "repo", "r", "", "Specific repository path (default: scan subdirectories)")
	cmd.Flags().BoolVarP(&statusFetch, "fetch", "f", false, "Run git fetch before checking upstream")

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runStatus(cmd *cobra.Command, args []string) {
	err := runStatusMulti(statusRepo, statusFetch, rootFlags.verbose)
	horus.CheckErr(err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
