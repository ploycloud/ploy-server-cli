package main

import (
	"fmt"
	"os"

	"github.com/ploycloud/ploy-server-cli/cmd"
)

var osExit = os.Exit

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		osExit(1)
	}
}
