/*
Copyright © 2026 Daniel Rivas <danielrivasmd@gmail.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// Flags for the ls command
type lsFlags struct {
	all             bool
	almostAll       bool
	sortByCtime     bool
	directory       bool
	noDirectory     bool
	human           bool
	si              bool
	reverse         bool
	sortBySize      bool
	sortByTime      bool
	sortByAtime     bool
	unsorted        bool
	sortWord        string
	noVCS           bool
	help            bool
}

var lsOpts lsFlags

func init() {
	lsCmd := MakeCmd("ls", runLs)

	// Define flags
	lsCmd.Flags().BoolVarP(&lsOpts.all, "all", "a", false, "list entries starting with .")
	lsCmd.Flags().BoolVarP(&lsOpts.almostAll, "almost-all", "A", false, "list all except . and ..")
	lsCmd.Flags().BoolVarP(&lsOpts.sortByCtime, "ctime", "c", false, "sort by ctime (inode change time)")
	lsCmd.Flags().BoolVarP(&lsOpts.directory, "directory", "d", false, "list only directories")
	lsCmd.Flags().BoolVarP(&lsOpts.noDirectory, "no-directory", "n", false, "do not list directories")
	lsCmd.Flags().BoolVarP(&lsOpts.human, "human", "h", false, "show filesizes in human-readable format")
	lsCmd.Flags().BoolVar(&lsOpts.si, "si", false, "with -h, use powers of 1000 not 1024")
	lsCmd.Flags().BoolVarP(&lsOpts.reverse, "reverse", "r", false, "reverse sort order")
	lsCmd.Flags().BoolVarP(&lsOpts.sortBySize, "size", "S", false, "sort by size")
	lsCmd.Flags().BoolVarP(&lsOpts.sortByTime, "time", "t", false, "sort by time (modification time)")
	lsCmd.Flags().BoolVarP(&lsOpts.sortByAtime, "atime", "u", false, "sort by atime (access time)")
	lsCmd.Flags().BoolVarP(&lsOpts.unsorted, "unsorted", "U", false, "unsorted")
	lsCmd.Flags().StringVar(&lsOpts.sortWord, "sort", "", "sort by WORD: none, size, time, ctime, atime")
	lsCmd.Flags().BoolVar(&lsOpts.noVCS, "no-vcs", false, "do not get VCS status (much faster)")
	lsCmd.Flags().BoolVar(&lsOpts.help, "help", false, "show help")

	// Mark --sort flag as not an argument that consumes following args
	lsCmd.Flags().SetInterspersed(false)

	rootCmd.AddCommand(lsCmd)
}

func runLs(cmd *cobra.Command, args []string) {
	if lsOpts.help {
		cmd.Help()
		return
	}
	// Delegate to implementation in utilLs.go
	executeLs(args)
}
