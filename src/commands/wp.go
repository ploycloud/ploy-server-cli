package commands

import (
	"fmt"

	"github.com/cloudoploy/ploy-cli/src/docker"
	"github.com/spf13/cobra"
)

var WpCmd = &cobra.Command{
	Use:   "wp",
	Short: "Execute WP-CLI commands",
	Long:  `Execute WP-CLI commands for the current WordPress site.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executing WP-CLI command:", args)
		compose, err := docker.GetComposeFile()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := docker.RunWpCli(compose.Path, args); err != nil {
			fmt.Printf("Error executing WP-CLI command: %v\n", err)
		}
	},
}
