package commands

import (
	"github.com/fatih/color"
	"github.com/ploycloud/ploy-cli/src/docker"
	"github.com/ploycloud/ploy-cli/src/utils"
	"github.com/spf13/cobra"
	"os"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the site in the current directory",
	Long:  "Start the Docker container for the site in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		composePath := utils.FindComposeFile()
		if composePath == "" {
			color.Red("No docker-compose.yml file found.")
			return
		}

		if err := docker.RunCompose(composePath, "up", "-d"); err != nil {
			color.Red("Error starting container:", err)
		}
	},
}

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the site in the current directory",
	Long:  "Stop the Docker container for the site in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		composePath := utils.FindComposeFile()
		if composePath == "" {
			color.Red("No docker-compose.yml file found.")
			return
		}

		if err := docker.RunCompose(composePath, "down"); err != nil {
			color.Red("Error stopping container:", err)
		}
	},
}

var RestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the site in the current directory",
	Long:  "Restart the Docker containers for the site in the current directory.",
	Run: func(cmd *cobra.Command, args []string) {
		composePath := utils.FindComposeFile()
		if composePath == "" {
			color.Red("No docker-compose.yml file found")
			return
		}

		if err := docker.RunCompose(composePath, "restart"); err != nil {
			color.Red("Error restarting site using Docker Compose setup:", err)
		}

		color.Green("Site restarted successfully")
	},
}

var ExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command in the Docker container",
	Run: func(cmd *cobra.Command, args []string) {
		composePath := utils.FindComposeFile()
		if composePath == "" {
			color.Red("No docker-compose.yml file found")
			return
		}

		if len(args) == 0 {
			color.Yellow("No command provided")
			return
		}

		// if the next argument is "php", "nginx" or "litespeed", use it as the service name
		// otherwise, use "php" as the default service name
		composeArgs := []string{"exec"}
		if args[0] == "php" || args[0] == "nginx" || args[0] == "litespeed" {
			composeArgs = append(composeArgs, args[0])
			args = args[1:]
		} else {
			composeArgs = append(composeArgs, "php")
		}

		composeArgs = append(composeArgs, args...)

		if err := docker.RunCompose(composePath, composeArgs...); err != nil {
			color.Red("Error executing command: %v\n", err)
		}
	},
}

var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show logs of the site in the current directory",
	Long:  `Show logs of the Docker container for the site in the current directory.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		composePath := utils.FindComposeFile()
		if composePath == "" {
			color.Red("No docker-compose.yml file found")
			os.Exit(1)
		}

		composeArgs := []string{"logs"}
		if len(args) == 1 {
			composeArgs = append(composeArgs, args[0])
		}

		if err := docker.RunCompose(composePath, composeArgs...); err != nil {
			color.Red("Error showing logs: %v\n", err)
			os.Exit(1)
		}
	},
}
