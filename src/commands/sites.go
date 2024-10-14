package commands

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ploycloud/ploy-server-cli/src/common"
	"github.com/ploycloud/ploy-server-cli/src/docker"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
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
	Long:  `Start all sites on the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting all sites...")
		startAllSites()
	},
}

var sitesStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all sites",
	Long:  `Stop all sites on the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping all sites...")
		stopAllSites()
	},
}

var sitesRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart all sites",
	Long:  `Restart all sites on the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting all sites...")
		stopAllSites()
		startAllSites()
	},
}

func startAllSites() {
	sitesDir := common.HomeDir
	foundSite := false

	entries, err := os.ReadDir(sitesDir)
	if err != nil {
		color.Red("Error reading directory %s: %v\n", sitesDir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			path := filepath.Join(sitesDir, entry.Name())
			composePath := filepath.Join(path, "docker-compose.yml")
			if _, err := os.Stat(composePath); err == nil {
				color.Yellow("Starting site in %s\n", filepath.Base(path))
				err := docker.RunCompose(composePath, "up", "-d")

				if nil != err {
					continue
				}

				foundSite = true
			}
		}
	}

	if !foundSite {
		fmt.Println("No sites found to start.")
	}
}

func stopAllSites() {
	sitesDir := common.HomeDir
	foundSite := false

	entries, err := os.ReadDir(sitesDir)
	if err != nil {
		color.Red("Error reading directory %s: %v\n", sitesDir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			path := filepath.Join(sitesDir, entry.Name())
			composePath := filepath.Join(path, "docker-compose.yml")
			if _, err := os.Stat(composePath); err == nil {
				color.Yellow("Stopping site in %s\n", filepath.Base(path))
				err := docker.RunCompose(composePath, "down")

				if nil != err {
					continue
				}

				foundSite = true
			}
		}
	}

	if !foundSite {
		fmt.Println("No sites found to stop.")
	}
}
