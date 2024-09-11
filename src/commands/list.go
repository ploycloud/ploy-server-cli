package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployments",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing all deployments...")
		// Add your list logic here
	},
}
