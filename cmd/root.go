package cmd

import (
	"fmt"

	"github.com/cloudoploy/ploy-cli/src/commands"
	"github.com/cloudoploy/ploy-cli/src/docker"
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
	rootCmd.AddCommand(commands.GlobalCmd)
	rootCmd.AddCommand(commands.SitesCmd)
	rootCmd.AddCommand(commands.WpCmd)

	// Add other site-specific commands
	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start the website",
		Run: func(cmd *cobra.Command, args []string) {
			compose, err := docker.GetComposeFile()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			if err := compose.Up(); err != nil {
				fmt.Printf("Error starting the website: %v\n", err)
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop the website",
		Run: func(cmd *cobra.Command, args []string) {
			compose, err := docker.GetComposeFile()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			if err := compose.Down(); err != nil {
				fmt.Printf("Error stopping the website: %v\n", err)
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "restart",
		Short: "Restart the website",
		Run: func(cmd *cobra.Command, args []string) {
			compose, err := docker.GetComposeFile()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			if err := compose.Restart(); err != nil {
				fmt.Printf("Error restarting the website: %v\n", err)
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "logs [container]",
		Short: "View logs from containers",
		Run: func(cmd *cobra.Command, args []string) {
			compose, err := docker.GetComposeFile()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			if err := compose.Logs(args...); err != nil {
				fmt.Printf("Error viewing logs: %v\n", err)
			}
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "exec [container] [command]",
		Short: "Execute commands inside a container",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				fmt.Println("Usage: ploy exec [container] [command]")
				return
			}
			compose, err := docker.GetComposeFile()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			if err := compose.Exec(args[0], args[1:]...); err != nil {
				fmt.Printf("Error executing command: %v\n", err)
			}
		},
	})
}
