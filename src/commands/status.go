package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status [deployment]",
	Short: "Check the status of a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deployment := args[0]
		fmt.Printf("Checking status of deployment: %s\n", deployment)
		// Add your status check logic here
	},
}
