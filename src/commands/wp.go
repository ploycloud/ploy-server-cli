package commands

import (
	"github.com/cloudoploy/ploy-cli/src/utils"
	"github.com/fatih/color"

	"github.com/cloudoploy/ploy-cli/src/docker"
	"github.com/spf13/cobra"
)

var WpCmd = &cobra.Command{
	Use:   "wp",
	Short: "Execute WP-CLI commands",
	Long:  `Execute WP-CLI commands for the current WordPress site.`,
	Run: func(cmd *cobra.Command, args []string) {
		composePath := utils.FindComposeFile()
		if composePath == "" {
			color.Red("No docker-compose.yml file found.")
			return
		}

		if err := docker.RunWpCli(composePath, args); err != nil {
			color.Red("Error running wp-cli: %s", err)
		}
	},
}
