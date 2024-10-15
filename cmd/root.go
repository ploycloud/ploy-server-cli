package cmd

import (
	"fmt"

	"github.com/ploycloud/ploy-server-cli/src/commands"
	"github.com/ploycloud/ploy-server-cli/src/common"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "ploy",
	Short:   "Ploy CLI - Manage your cloud deployments",
	Long:    `Ploy CLI is a powerful tool for managing and deploying your cloud applications. You are using ploy version: ` + common.CurrentCliVersion,
	Version: common.CurrentCliVersion,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(commands.DeployCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.StatusCmd)
	rootCmd.AddCommand(commands.ServicesCmd)
	rootCmd.AddCommand(commands.SitesCmd)
	rootCmd.AddCommand(commands.WpCmd)
	rootCmd.AddCommand(commands.StartCmd)
	rootCmd.AddCommand(commands.StopCmd)
	rootCmd.AddCommand(commands.RestartCmd)
	rootCmd.AddCommand(commands.ExecCmd)
	rootCmd.AddCommand(commands.LogsCmd)
	rootCmd.AddCommand(commands.UpdateCmd)
	rootCmd.AddCommand(commands.EchoCmd)

	// Add a custom version command
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Print the version number of ploy cli",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(common.CurrentCliVersion)
			},
		},
	)
}
