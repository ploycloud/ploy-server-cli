package cmd

import (
	"github.com/cloudoploy/cloudoploy-cli/internal/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ploy",
	Short: "Ploy CLI - Manage your cloud deployments",
	Long:  `Ploy CLI is a powerful tool for managing and deploying your cloud applications.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(commands.DeployCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.StatusCmd)
}
