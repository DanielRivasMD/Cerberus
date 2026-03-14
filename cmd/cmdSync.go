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
	"fmt"

	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	syncRepo string
	syncPush bool
	syncPull bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func SyncCmd() *cobra.Command {
	d := horus.Must(domovoi.GlobalDocs())
	cmd := horus.Must(d.MakeCmd("sync", runSync))

	cmd.Flags().StringVarP(&syncRepo, "repo", "r", "", "Specific repository path (default: scan subdirectories)")
	cmd.Flags().BoolVarP(&syncPush, "push", "", false, "Push commits to remote")
	cmd.Flags().BoolVarP(&syncPull, "pull", "", false, "Pull commits from remote")

	return cmd
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func runSync(cmd *cobra.Command, args []string) {
	if !syncPush && !syncPull {
		horus.CheckErr(fmt.Errorf("either --push or --pull must be specified"))
	}
	if syncPush && syncPull {
		horus.CheckErr(fmt.Errorf("--push and --pull cannot be used together"))
	}
	err := runSyncMulti(syncRepo, syncPush, syncPull, rootFlags.verbose)
	horus.CheckErr(err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
