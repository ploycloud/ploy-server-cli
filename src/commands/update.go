package commands

import (
	"fmt"
	"github.com/ploycloud/ploy-server-cli/src/utils"
	"github.com/spf13/cobra"
	"os"
)

var yesFlag bool

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ploy cli to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		if os.Geteuid() != 0 {
			fmt.Println("Error: The update command must be run as root.")
			fmt.Println("Please run 'sudo ploy update'")
			return
		}

		latestVersion, hasUpdate, err := utils.CheckForUpdates()
		if err != nil {
			fmt.Println("Error checking for updates:", err)
			return
		}

		if !hasUpdate {
			fmt.Println("You are already running the latest version.")
			return
		}

		fmt.Printf("New version available: %s\n", latestVersion)

		if !yesFlag {
			fmt.Print("Do you want to update? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Update cancelled.")
				return
			}
		}

		fmt.Println("Updating...")
		if _, err := utils.SelfUpdate(); err != nil {
			fmt.Println("Error updating:", err)
		} else {
			fmt.Println("Update successful. Please restart ploy cli.")
			os.Exit(0)
		}
	},
}

func init() {
	UpdateCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically answer yes to update confirmation")
}
