package commands

import (
	"fmt"
	"github.com/fatih/color"
	"os"

	"github.com/cloudoploy/ploy-cli/src/common"
	"github.com/cloudoploy/ploy-cli/src/docker"
	"github.com/spf13/cobra"
)

const globalCompose = common.GlobalCompose

var ServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage Global Docker Compose services",
	Long:  `Manage Global Docker Compose services including MySQL, Redis, Ofelia, and Nginx Proxy.`,
}

func init() {
	ServicesCmd.AddCommand(globalStartCmd)
	ServicesCmd.AddCommand(globalStopCmd)
	ServicesCmd.AddCommand(globalRestartCmd)
}

var globalStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting global services (mysql, redis, nginx-proxy)...")
		if err := docker.RunCompose(globalCompose, "up", "-d"); err != nil {
			color.Red("Error starting global services: %v", err)
			os.Exit(1)
			return
		}

		if err := docker.RunCompose(globalCompose, "ps"); err != nil {
			color.Red("Error checking status of global services: %v", err)
		}

		color.Green("Global services started successfully")
	},
}

var globalStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping global services...")
		if err := docker.RunCompose(globalCompose, "down"); err != nil {
			color.Red("Error stopping global services:", err)
		}

		color.Green("Global services stopped successfully")
	},
}

var globalRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting global services...")
		if err := docker.RunCompose(globalCompose, "down"); err != nil {
			color.Red("Error stopping global services:", err)
		}

		if err := docker.RunCompose(globalCompose, "up", "-d"); err != nil {
			color.Red("Error starting global services:", err)
		}

		if err := docker.RunCompose(globalCompose, "ps"); err != nil {
			color.Red("Error checking status of global services:", err)
		}

		color.Green("Global services restarted successfully")
	},
}
