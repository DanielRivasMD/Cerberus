/*
Copyright © 2024 Daniel Rivas <danielrivasmd@gmail.com>

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
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

const APP = "cerberus"
const VERSION = "v0.1.0"
const NAME = "Daniel Rivas"
const EMAIL = "<danielrivasmd@gmail.com>"

////////////////////////////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:     GetUse("root"),
	Long:    formatLongHelp(GetHelp("root")),
	Example: GetExample("root"),
	Version: VERSION,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func Execute() {
	horus.CheckErr(rootCmd.Execute())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&RootFlags.verbose, "verbose", "v", false, "Verbose")
	rootCmd.PersistentFlags().StringVarP(&RootFlags.output, "output", "o", "", "File output")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var RootFlags struct {
	verbose bool
	output  string
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var Defaults = struct {
	repoLen     int
	commitLen   int
	ageLen      int
	languageLen int
	linesLen    int
	sizeLen     int
	meanLen     int
	qLen        int
	overviewLen int
	licenseLen  int
	remoteLen   int
}{
	repoLen:     25,
	commitLen:   6,
	ageLen:      6,
	languageLen: 15,
	linesLen:    6,
	sizeLen:     7,
	meanLen:     4,
	qLen:        3,
	overviewLen: 92,
	licenseLen:  7,
	remoteLen:   95,
}

////////////////////////////////////////////////////////////////////////////////////////////////////
