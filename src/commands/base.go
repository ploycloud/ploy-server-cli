package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var BaseCmd = &cobra.Command{
	Use:   "base",
	Short: "Manage base Docker Compose services",
	Long:  `Manage base Docker Compose services including MySQL, Redis, Ofelia, and Nginx Proxy.`,
}

func init() {
	BaseCmd.AddCommand(baseStartCmd)
	BaseCmd.AddCommand(baseStopCmd)
	BaseCmd.AddCommand(baseRestartCmd)
}

var baseStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start base services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting base services (mysql, redis, nginx-proxy)...")
		// Add logic to start base services
	},
}

var baseStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop base services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping base services...")
		// Add logic to stop base services
	},
}

var baseRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart base services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting base services...")
		// Add logic to restart base services
	},
}
