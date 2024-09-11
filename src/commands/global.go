package commands

import (
	"fmt"

	"github.com/cloudoploy/ploy-cli/src/docker"
	"github.com/spf13/cobra"
)

var GlobalCmd = &cobra.Command{
	Use:   "global",
	Short: "Manage Global Docker Compose services",
	Long:  `Manage Global Docker Compose services including MySQL, Redis, Ofelia, and Nginx Proxy.`,
}

func init() {
	GlobalCmd.AddCommand(globalStartCmd)
	GlobalCmd.AddCommand(globalStopCmd)
	GlobalCmd.AddCommand(globalRestartCmd)
}

var globalStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting global services (mysql, redis, nginx-proxy)...")
		compose, err := docker.GetGlobalComposeFile()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := compose.Up(); err != nil {
			fmt.Printf("Error starting global services: %v\n", err)
		}
	},
}

var globalStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping global services...")
		compose, err := docker.GetGlobalComposeFile()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := compose.Down(); err != nil {
			fmt.Printf("Error stopping global services: %v\n", err)
		}
	},
}

var globalRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting global services...")
		compose, err := docker.GetGlobalComposeFile()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := compose.Restart(); err != nil {
			fmt.Printf("Error restarting global services: %v\n", err)
		}
	},
}
