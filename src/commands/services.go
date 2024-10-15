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

var execCommand = exec.Command

var osExit = os.Exit

var ServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage Global Docker Compose services",
	Long:  `Manage Global Docker Compose services including MySQL, Redis, Ofelia, and Nginx Proxy.`,
}

func init() {
	// Disable color output for tests
	if os.Getenv("GO_TEST") == "1" {
		color.NoColor = true
	}

	ServicesCmd.AddCommand(globalStartCmd)
	ServicesCmd.AddCommand(globalStopCmd)
	ServicesCmd.AddCommand(globalRestartCmd)
	ServicesCmd.AddCommand(installCmd)
	installCmd.AddCommand(installNginxProxyCmd)
}

var globalStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting global services (mysql, redis, nginx-proxy)...")
		if err := docker.RunCompose(globalCompose, "up", "-d"); err != nil {
			fmt.Printf("Error starting global services: %v\n", err)
			osExit(1)
			return
		}

		if err := docker.RunCompose(globalCompose, "ps"); err != nil {
			fmt.Printf("Error checking status of global services: %v\n", err)
		}

		fmt.Println("Global services started successfully")
	},
}

var globalStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping global services...")
		if err := docker.RunCompose(globalCompose, "down"); err != nil {
			fmt.Printf("Error stopping global services: %v\n", err)
			return
		}

		fmt.Println("Global services stopped successfully")
	},
}

var globalRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting global services...")
		if err := docker.RunCompose(globalCompose, "down"); err != nil {
			fmt.Printf("Error stopping global services: %v\n", err)
			return
		}

		if err := docker.RunCompose(globalCompose, "up", "-d"); err != nil {
			fmt.Printf("Error starting global services: %v\n", err)
			return
		}

		if err := docker.RunCompose(globalCompose, "ps"); err != nil {
			fmt.Printf("Error checking status of global services: %v\n", err)
		}

		fmt.Println("Global services restarted successfully")
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install various services",
	Long:  `Install services such as nginx-proxy, and potentially others in the future.`,
}

var installNginxProxyCmd = &cobra.Command{
	Use:   "nginx-proxy",
	Short: "Install Nginx as a proxy on the host machine",
	Run: func(cmd *cobra.Command, args []string) {
		if GetGOOS() == "darwin" {
			fmt.Println("Nginx installation is not supported on macOS. Please install Nginx manually.")
			return
		}

		fmt.Println("Checking if Nginx is already installed...")

		// Check if Nginx is already installed
		checkCmd := execCommand("nginx", "-v")
		if err := checkCmd.Run(); err == nil {
			fmt.Println("Nginx is already installed.")
			return
		}

		fmt.Println("Installing Nginx as a proxy...")

		// Install Nginx (this assumes a Debian-based system like Ubuntu)
		installCmd := execCommand("sudo", "apt-get", "update")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			color.Red("Error updating package list: %v", err)
			return
		}

		installCmd = execCommand("sudo", "apt-get", "install", "-y", "nginx")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			color.Red("Error installing Nginx: %v", err)
			return
		}

		// Start Nginx service
		startCmd := execCommand("sudo", "systemctl", "start", "nginx")
		if err := startCmd.Run(); err != nil {
			color.Red("Error starting Nginx service: %v", err)
			return
		}

		// Enable Nginx to start on boot
		enableCmd := execCommand("sudo", "systemctl", "enable", "nginx")
		if err := enableCmd.Run(); err != nil {
			color.Red("Error enabling Nginx service: %v", err)
			return
		}

		fmt.Println("Nginx installed and configured successfully as a proxy")
		fmt.Println("You can now configure Nginx as a proxy for your Docker containers.")
		fmt.Println("Don't forget to configure your Nginx configuration file to proxy requests to your Docker containers.")
	},
}
