package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var EchoCmd = &cobra.Command{
	Use:   "echo [text]",
	Short: "Echo the input text",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args[0])
	},
}
