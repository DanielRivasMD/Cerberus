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

	"github.com/DanielRivasMD/horus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// Global declarations (reserved for future variables)
var (
	verbose bool
	output  string
)

const (
	repoLen = 25

	commitLen   = 6
	ageLen      = 6
	languageLen = 15
	linesLen    = 6
	sizeLen     = 7
	meanLen     = 4
	qLen        = 3

	overviewLen = 93
	licenseLen  = 7

	remoteLen = 95
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// rootCmd defines the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "cerberus",
	Short: "",
	Long: chalk.Green.Color(chalk.Bold.TextStyle("Daniel Rivas")) + chalk.Dim.TextStyle(chalk.Italic.TextStyle("<danielrivasmd@gmail.com>")) + `

` + chalk.Cyan.Color("cerberus") + chalk.Blue.Color(`

`) + ``,

	Example: `
` + chalk.Cyan.Color("cerberus") + ` help`,

	////////////////////////////////////////////////////////////////////////////////////////////////////

}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Execute is the entry point for executing the command.
// It wraps the root command execution and handles any errors using Horus's checkErr function.
func Execute() {
	err := rootCmd.Execute()
	horus.CheckErr(err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// initializeConfig sets up configuration using Viper.
// It also leverages the Domovoi library for any additional configuration management.
func initializeConfig(cmd *cobra.Command, configPath string, configName string) error {
	// Create a new Viper instance for configuration management.
	vConfig := viper.New()

	// Set the path and name of the configuration file.
	vConfig.AddConfigPath(configPath)
	vConfig.SetConfigName(configName)

	// Attempt to read the configuration file.
	err := vConfig.ReadInConfig()
	if err != nil {
		// If the config file is not found, that's acceptable.
		_, notFound := err.(viper.ConfigFileNotFoundError)
		if !notFound {
			// Return the error for any other issue.
			return err
		}
	}

	// Bind command flags with configuration values.
	bindFlags(cmd, vConfig)

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// bindFlags synchronizes each Cobra flag with the corresponding Viper configuration value.
// If the flag is unset and a configuration value is available, the flag is updated.
func bindFlags(cmd *cobra.Command, vConfig *viper.Viper) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// If the flag wasn't explicitly set but has a value in the config, apply that value.
		if !flag.Changed && vConfig.IsSet(flag.Name) {
			value := vConfig.Get(flag.Name)
			cmd.Flags().Set(flag.Name, fmt.Sprintf("%v", value))
		}
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Execute prior main.
// init registers persistent flags and performs additional initialization tasks.
func init() {
	// Set up persistent flags.
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "File output")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
