package cmd

import (
	"github.com/cloudoploy/ploy-cli/src/commands"
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
	rootCmd.AddCommand(commands.BaseCmd)
	rootCmd.AddCommand(commands.SitesCmd)
	rootCmd.AddCommand(commands.WpCmd)

	// Add other site-specific commands
	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the website",
		Run: func(cmd *cobra.Command, args []string) {
			// Add logic to start the website
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop the website",
		Run: func(cmd *cobra.Command, args []string) {
			// Add logic to stop the website
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "restart",
		Short: "Restart the website",
		Run: func(cmd *cobra.Command, args []string) {
			// Add logic to restart the website
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "logs [container]",
		Short: "View logs from containers",
		Run: func(cmd *cobra.Command, args []string) {
			// Add logic to view logs
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "exec [container]",
		Short: "Execute commands inside a container",
		Run: func(cmd *cobra.Command, args []string) {
			// Add logic to execute commands inside a container
		},
	})
}
