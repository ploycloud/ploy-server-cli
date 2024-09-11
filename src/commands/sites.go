package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var SitesCmd = &cobra.Command{
	Use:   "sites",
	Short: "Manage all sites",
	Long:  `Start, stop, or restart all sites on the server.`,
}

func init() {
	SitesCmd.AddCommand(sitesStartCmd)
	SitesCmd.AddCommand(sitesStopCmd)
	SitesCmd.AddCommand(sitesRestartCmd)
}

var sitesStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all sites",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting all sites...")
		// Add logic to start all sites
	},
}

var sitesStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all sites",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping all sites...")
		// Add logic to stop all sites
	},
}

var sitesRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart all sites",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting all sites...")
		// Add logic to restart all sites
	},
}
