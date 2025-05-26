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

import (
	"fmt"

	"github.com/DanielRivasMD/domovoi"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// declarations
var (
	iden = `In Greek mythology, ` + chalk.Yellow.Color("Cerberus") + `, ` + chalk.Dim.TextStyle("Κέρβερος") + `, often referred to as the hound of Hades, is a multi-headed dog
that guards the gates of the underworld to prevent the dead from leaving.

He was the offspring of the monsters Echidna and Typhon, and was usually described as having three heads,
a serpent for a tail, and snakes protruding from his body.

Cerberus is primarily known for his capture by Heracles, the last of Heracles' twelve labours`
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// identityCmd
var identityCmd = &cobra.Command{
	Use:   "identity",
	Short: `Reveal identity`,
	Long:  `Reveal identity`,

	////////////////////////////////////////////////////////////////////////////////////////////////////

	Run: func(κ *cobra.Command, args []string) {

		domovoi.LineBreaks()
		fmt.Println()

		fmt.Println(iden)

		domovoi.LineBreaks()
		fmt.Println()

	},
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// execute prior main
func init() {
	rootCmd.AddCommand(identityCmd)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
