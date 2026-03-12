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

// TODO: rework logic & functionality. at the moment, this command is a draft
func LsCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("ls", runLs))

	cmd.Flags().BoolVarP(&lsFlags.all, "all", "a", false, "list entries starting with .")
	cmd.Flags().BoolVarP(&lsFlags.almostAll, "almost-all", "A", false, "list all except . and ..")
	cmd.Flags().BoolVarP(&lsFlags.sortByCtime, "ctime", "c", false, "sort by ctime (inode change time)")
	cmd.Flags().BoolVarP(&lsFlags.directory, "directory", "d", false, "list only directories")
	cmd.Flags().BoolVarP(&lsFlags.noDirectory, "no-directory", "n", false, "do not list directories")
	cmd.Flags().BoolVarP(&lsFlags.human, "human", "h", false, "show filesizes in human-readable format")
	cmd.Flags().BoolVar(&lsFlags.si, "si", false, "with -h, use powers of 1000 not 1024")
	cmd.Flags().BoolVarP(&lsFlags.reverse, "reverse", "r", false, "reverse sort order")
	cmd.Flags().BoolVarP(&lsFlags.sortBySize, "size", "S", false, "sort by size")
	cmd.Flags().BoolVarP(&lsFlags.sortByTime, "time", "t", false, "sort by time (modification time)")
	cmd.Flags().BoolVarP(&lsFlags.sortByAtime, "atime", "u", false, "sort by atime (access time)")
	cmd.Flags().BoolVarP(&lsFlags.unsorted, "unsorted", "U", false, "unsorted")
	cmd.Flags().StringVar(&lsFlags.sortWord, "sort", "", "sort by WORD: none, size, time, ctime, atime")
	cmd.Flags().BoolVar(&lsFlags.noVCS, "no-vcs", false, "do not get VCS status (much faster)")
	cmd.Flags().BoolVar(&lsFlags.help, "help", false, "show help")

	// Mark --sort flag as not an argument that consumes following args
	cmd.Flags().SetInterspersed(false)

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runLs(cmd *cobra.Command, args []string) {
	if lsFlags.help {
		cmd.Help()
		return
	}
	// Delegate to implementation in utilLs.go
	executeLs(args)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type lsFlag struct {
	all         bool
	almostAll   bool
	sortByCtime bool
	directory   bool
	noDirectory bool
	human       bool
	si          bool
	reverse     bool
	sortBySize  bool
	sortByTime  bool
	sortByAtime bool
	unsorted    bool
	sortWord    string
	noVCS       bool
	help        bool
}

var lsFlags lsFlag

////////////////////////////////////////////////////////////////////////////////////////////////////
