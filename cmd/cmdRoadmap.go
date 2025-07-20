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
	"github.com/DanielRivasMD/domovoi"
	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// declarations
var ()

////////////////////////////////////////////////////////////////////////////////////////////////////

// roadmapCmd
var roadmapCmd = &cobra.Command{
	Use:   "roadmap",
	Short: "" + chalk.Yellow.Color("") + ".",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas ")) + chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `
`,

	Example: `
` + chalk.Cyan.Color("") + ` help ` + chalk.Yellow.Color("") + chalk.Yellow.Color("roadmap"),

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(cmd *cobra.Command, args []string) {

		// Preview & open ROADMAP.txt in a floating Zellij pane
		cmdReadme := `
zellij run --name roadmap \
	--close-on-exit --floating \
	--height 100 --width 130 --x 15 --y 0 \
	-- zsh -c '
		file=$(
			fd --type f --glob 'ROADMAP.txt' . \
			| fzf \
				--preview="bat --style=plain --color=always {}" \
				--preview-window="right:70%" \
				--height=100% \
				--reverse
		)
	[[ -n $file ]] && hx "$file"
	'`

		// execute command
		err := domovoi.ExecCmd("bash", "-c", cmdReadme)
		horus.CheckErr(err)
	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(roadmapCmd)

	// flags
}

////////////////////////////////////////////////////////////////////////////////////////////////////
