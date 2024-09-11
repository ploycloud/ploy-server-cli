package commands

import (
	"fmt"
	"github.com/cloudoploy/ploy-cli/src/docker"
	"github.com/spf13/cobra"
	"os"
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

func getAllSites() ([]string, error) {
	// This is a placeholder. You should implement a method to get all site directories
	return []string{"/path/to/site1", "/path/to/site2"}, nil
}

var sitesStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all sites",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting all sites...")
		sites, err := getAllSites()
		if err != nil {
			fmt.Printf("Error getting sites: %v\n", err)
			return
		}
		for _, site := range sites {
			if err := os.Chdir(site); err != nil {
				fmt.Printf("Error changing to directory %s: %v\n", site, err)
				continue
			}
			compose, err := docker.GetComposeFile()
			if err != nil {
				fmt.Printf("Error in site %s: %v\n", site, err)
				continue
			}
			if err := compose.Up(); err != nil {
				fmt.Printf("Error starting site %s: %v\n", site, err)
			}
		}
	},
}

var sitesStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all sites",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping all sites...")
		sites, err := getAllSites()
		if err != nil {
			fmt.Printf("Error getting sites: %v\n", err)
			return
		}
		for _, site := range sites {
			if err := os.Chdir(site); err != nil {
				fmt.Printf("Error changing to directory %s: %v\n", site, err)
				continue
			}
			compose, err := docker.GetComposeFile()
			if err != nil {
				fmt.Printf("Error in site %s: %v\n", site, err)
				continue
			}
			if err := compose.Down(); err != nil {
				fmt.Printf("Error stopping site %s: %v\n", site, err)
			}
		}
	},
}

var sitesRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart all sites",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting all sites...")
		sites, err := getAllSites()
		if err != nil {
			fmt.Printf("Error getting sites: %v\n", err)
			return
		}
		for _, site := range sites {
			if err := os.Chdir(site); err != nil {
				fmt.Printf("Error changing to directory %s: %v\n", site, err)
				continue
			}
			compose, err := docker.GetComposeFile()
			if err != nil {
				fmt.Printf("Error in site %s: %v\n", site, err)
				continue
			}
			if err := compose.Restart(); err != nil {
				fmt.Printf("Error restarting site %s: %v\n", site, err)
			}
		}
	},
}
