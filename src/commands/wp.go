package commands

import (
	"github.com/fatih/color"
	"github.com/ploycloud/ploy-server-cli/src/utils"

	"github.com/ploycloud/ploy-server-cli/src/docker"
	"github.com/spf13/cobra"
)

var runWpCli = docker.RunWpCli

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

		if err := runWpCli(composePath, args); err != nil {
			color.Red("Error running wp-cli: %s", err)
		}
	},
}
