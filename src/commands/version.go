package commands

import (
	"github.com/cloudoploy/ploy-cli/src/common"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ploy cli",
	Run: func(cmd *cobra.Command, args []string) {
		color.Green("Ploy CLI version: %s\n", common.CurrentCliVersion)
	},
}
