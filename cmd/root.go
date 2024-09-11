package cmd

import (
	"github.com/cloudoploy/ploy-cli/src/commands"
	"github.com/cloudoploy/ploy-cli/src/common"
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
	rootCmd.AddCommand(commands.VersionCmd)
	rootCmd.AddCommand(commands.UpdateCmd)
	rootCmd.AddCommand(commands.EchoCmd)
}
