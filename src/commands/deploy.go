package commands

import (
	"fmt"

	"github.com/ploycloud/ploy-server-cli/src/utils"
	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [repo]",
	Short: "Deploy a repository to PloyCloud server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := args[0]
		fmt.Printf("Deploying repository: %s\n", repo)

		if err := utils.CloneRepo(repo); err != nil {
			fmt.Printf("Error cloning repository: %v\n", err)
			return
		}

		// Add your deployment logic here
		fmt.Println("Deployment successful!")
	},
}
