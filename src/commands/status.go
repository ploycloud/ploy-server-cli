package commands

import (
	"github.com/cloudoploy/ploy-cli/src/common"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of all services",
	Run: func(cmd *cobra.Command, args []string) {
		servicesDir := common.ServicesDir
		if _, err := os.Stat(servicesDir); os.IsNotExist(err) {
			color.Red(servicesDir + " directory does not exist")
		} else {
			color.Green(servicesDir + " directory exists")
		}

		globalCompose := common.GlobalCompose
		if _, err := os.Stat(globalCompose); os.IsNotExist(err) {
			color.Red(globalCompose + " does not exist")
		} else {
			color.Green(globalCompose + " exists")
		}

		provisionsDir := common.ProvisionsDir
		if _, err := os.Stat(provisionsDir); os.IsNotExist(err) {
			color.Red(provisionsDir + " directory does not exist")
		} else {
			color.Green(provisionsDir + " directory exists")
		}

		if _, err := os.Stat(common.MysqlDir); os.IsNotExist(err) {
			color.Red("MySQL directory does not exist")
		} else {
			color.Green("MySQL directory exists")
		}

		if _, err := os.Stat(common.RedisDir); os.IsNotExist(err) {
			color.Red("Redis directory does not exist")
		} else {
			color.Green("Redis directory exists")
		}

		// check if nginx directory exists
		if _, err := os.Stat(common.NginxDir); os.IsNotExist(err) {
			color.Red("Nginx directory does not exist")
		} else {
			color.Green("Nginx directory exists")
		}

		if output, err := exec.Command("docker", "version", "--format", "{{.Server.Version}}").CombinedOutput(); err != nil {
			color.Red("Docker is not installed")
		} else {
			color.Green("Docker is installed, version: %s", strings.TrimSpace(string(output)))
		}

		if _, err := exec.Command("docker", "version").CombinedOutput(); err != nil {
			color.Red("Docker is not running")
		} else {
			color.Green("Docker is running")
		}
	},
}
