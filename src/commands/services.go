package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"

	"github.com/ploycloud/ploy-server-cli/src/common"
	"github.com/ploycloud/ploy-server-cli/src/docker"
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
	ServicesCmd.AddCommand(installNginxProxyCmd)
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

var installNginxProxyCmd = &cobra.Command{
	Use:   "install nginx",
	Short: "Install Nginx Proxy if not already installed",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking if Nginx Proxy is already installed...")

		// Check if Nginx Proxy is already running
		command := exec.Command("docker", "ps", "--filter", "name=nginx-proxy", "--format", "{{.Names}}")
		output, err := command.Output()
		if err != nil {
			color.Red("Error checking Nginx Proxy status: %v", err)
			return
		}

		if len(output) > 0 {
			color.Yellow("Nginx Proxy is already installed and running.")
			return
		}

		fmt.Println("Installing Nginx Proxy...")

		// Install Nginx Proxy using docker-compose
		if err := docker.RunCompose(globalCompose, "up", "-d", "nginx-proxy"); err != nil {
			color.Red("Error installing Nginx Proxy: %v", err)
			return
		}

		// Verify installation
		if err := docker.RunCompose(globalCompose, "ps", "nginx-proxy"); err != nil {
			color.Red("Error verifying Nginx Proxy installation: %v", err)
			return
		}

		color.Green("Nginx Proxy installed and configured successfully")
		fmt.Println("You can now use it as a proxy for your Docker containers.")
	},
}
