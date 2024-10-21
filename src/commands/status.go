package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/ploycloud/ploy-server-cli/src/common"

	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of all services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s directory %s\n", common.ServicesDir, dirStatus(common.ServicesDir))
		fmt.Printf("%s %s\n", common.GlobalCompose, fileStatus(common.GlobalCompose))
		fmt.Printf("%s directory %s\n", common.ProvisionsDir, dirStatus(common.ProvisionsDir))
		fmt.Printf("MySQL directory %s\n", dirStatus(common.MysqlDir))
		fmt.Printf("Redis directory %s\n", dirStatus(common.RedisDir))
		fmt.Printf("Nginx directory %s\n", dirStatus(common.NginxDir))

		dockerVersion, err := getDockerVersion()
		if err != nil {
			fmt.Println("Docker is not installed")
		} else {
			fmt.Printf("Docker is installed, version: %s\n", dockerVersion)
		}

		if isDockerRunning() {
			fmt.Println("Docker is running")
		} else {
			fmt.Println("Docker is not running")
		}
	},
}

func dirStatus(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "does not exist"
	}
	return "exists"
}

func fileStatus(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "does not exist"
	}
	return "exists"
}

func getDockerVersion() (string, error) {
	cmd := execCommand("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func isDockerRunning() bool {
	cmd := execCommand("docker", "info")
	return cmd.Run() == nil
}
