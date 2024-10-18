package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/ploycloud/ploy-server-cli/src/common"
	"github.com/ploycloud/ploy-server-cli/src/docker"
	"github.com/spf13/cobra"
)

const globalCompose = common.GlobalCompose

var execCommand = exec.Command

var osExit = os.Exit

var ServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage Global Docker Compose services",
	Long:  `Manage Global Docker Compose services including MySQL, Redis, Ofelia, and Nginx Proxy.`,
}

func init() {
	// Disable color output for tests
	if os.Getenv("GO_TEST") == "1" {
		color.NoColor = true
	}

	ServicesCmd.AddCommand(globalStartCmd)
	ServicesCmd.AddCommand(globalStopCmd)
	ServicesCmd.AddCommand(globalRestartCmd)
	ServicesCmd.AddCommand(installCmd)
	ServicesCmd.AddCommand(detailsCmd)
	installCmd.AddCommand(installNginxProxyCmd)
	installCmd.AddCommand(installMySQLCmd)
}

var globalStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting global services (mysql, redis, nginx-proxy)...")
		if err := docker.RunCompose(globalCompose, "up", "-d"); err != nil {
			fmt.Printf("Error starting global services: %v\n", err)
			osExit(1)
			return
		}

		if err := docker.RunCompose(globalCompose, "ps"); err != nil {
			fmt.Printf("Error checking status of global services: %v\n", err)
		}

		fmt.Println("Global services started successfully")
	},
}

var globalStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping global services...")
		if err := docker.RunCompose(globalCompose, "down"); err != nil {
			fmt.Printf("Error stopping global services: %v\n", err)
			return
		}

		fmt.Println("Global services stopped successfully")
	},
}

var globalRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart global services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Restarting global services...")
		if err := docker.RunCompose(globalCompose, "down"); err != nil {
			fmt.Printf("Error stopping global services: %v\n", err)
			return
		}

		if err := docker.RunCompose(globalCompose, "up", "-d"); err != nil {
			fmt.Printf("Error starting global services: %v\n", err)
			return
		}

		if err := docker.RunCompose(globalCompose, "ps"); err != nil {
			fmt.Printf("Error checking status of global services: %v\n", err)
		}

		fmt.Println("Global services restarted successfully")
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install various services",
	Long:  `Install services such as nginx-proxy, and potentially others in the future.`,
}

var installNginxProxyCmd = &cobra.Command{
	Use:   "nginx-proxy",
	Short: "Install Nginx as a proxy on the host machine",
	Run: func(cmd *cobra.Command, args []string) {
		if GetGOOS() == "darwin" {
			fmt.Println("Nginx installation is not supported on macOS. Please install Nginx manually.")
			return
		}

		fmt.Println("Checking if Nginx is already installed...")

		// Check if Nginx is already installed
		checkCmd := execCommand("nginx", "-v")
		if err := checkCmd.Run(); err == nil {
			fmt.Println("Nginx is already installed.")
			return
		}

		fmt.Println("Installing Nginx as a proxy...")

		// Install Nginx (this assumes a Debian-based system like Ubuntu)
		installCmd := execCommand("sudo", "apt-get", "update")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			color.Red("Error updating package list: %v", err)
			return
		}

		installCmd = execCommand("sudo", "apt-get", "install", "-y", "nginx")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			color.Red("Error installing Nginx: %v", err)
			return
		}

		// Start Nginx service
		startCmd := execCommand("sudo", "systemctl", "start", "nginx")
		if err := startCmd.Run(); err != nil {
			color.Red("Error starting Nginx service: %v", err)
			return
		}

		// Enable Nginx to start on boot
		enableCmd := execCommand("sudo", "systemctl", "enable", "nginx")
		if err := enableCmd.Run(); err != nil {
			color.Red("Error enabling Nginx service: %v", err)
			return
		}

		fmt.Println("Nginx installed and configured successfully as a proxy")
		fmt.Println("You can now configure Nginx as a proxy for your Docker containers.")
		fmt.Println("Don't forget to configure your Nginx configuration file to proxy requests to your Docker containers.")
	},
}

var installMySQLCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Install MySQL service",
	Long:  `Install MySQL service with optional parameters for user, password, and port.`,
	Run: func(cmd *cobra.Command, args []string) {
		user, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("password")
		port, _ := cmd.Flags().GetString("port")

		fmt.Println("Installing MySQL service...")
		if err := installMySQL(user, password, port); err != nil {
			color.Red("Error installing MySQL: %v", err)
			return
		}
		color.Green("MySQL installed successfully")
	},
}

var detailsCmd = &cobra.Command{
	Use:   "details [service]",
	Short: "Show details for a specific service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := args[0]
		details, err := getServiceDetails(service)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting %s details: %v\n", service, err)
			return
		}
		if details == nil {
			return // Info message already displayed by getServiceDetails
		}
		for k, v := range details {
			fmt.Printf("%s: %s\n", k, v)
		}
	},
}

func init() {
	// Add flags for MySQL installation
	installMySQLCmd.Flags().String("user", "default_user", "MySQL user")
	installMySQLCmd.Flags().String("password", "default_password", "MySQL password")
	installMySQLCmd.Flags().String("port", "3306", "MySQL port")
}

func installMySQL(user, password, port string) error {
	composePath := filepath.Join("docker", "databases", "mysql-compose.yml")

	// Read the compose file
	content, err := os.ReadFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to read MySQL compose file: %v", err)
	}

	// Replace placeholders with provided or default values
	replacements := map[string]string{
		"${MYSQL_USER}":     user,
		"${MYSQL_PASSWORD}": password,
		"${MYSQL_PORT}":     port,
	}

	for placeholder, value := range replacements {
		content = bytes.ReplaceAll(content, []byte(placeholder), []byte(value))
	}

	// Write the updated compose file
	tempComposePath := filepath.Join(os.TempDir(), "temp-mysql-compose.yml")
	if err := ioutil.WriteFile(tempComposePath, content, 0644); err != nil {
		return fmt.Errorf("failed to write temporary MySQL compose file: %v", err)
	}

	// Run docker-compose with the updated file
	cmd := execCommand("docker-compose", "-f", tempComposePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	// Clean up the temporary file
	os.Remove(tempComposePath)

	return err
}

func getServiceDetails(service string) (map[string]string, error) {
	switch service {
	case "mysql":
		details, err := getMySQLDetails()
		if err != nil {
			color.Yellow("MySQL info: %v", err)
			return nil, nil
		}
		return details, nil
	default:
		return nil, fmt.Errorf("unsupported service: %s", service)
	}
}

func getMySQLDetails() (map[string]string, error) {
	// Check if MySQL container is running
	cmd := execCommand("docker", "ps", "--filter", "name=mysql", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return nil, fmt.Errorf("MySQL container is not running")
	}

	// Get container ID
	containerName := strings.TrimSpace(string(output))

	// Get MySQL environment variables
	cmd = execCommand("docker", "inspect", "--format", "{{range .Config.Env}}{{println .}}{{end}}", containerName)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect MySQL container: %v", err)
	}

	// Parse environment variables
	envVars := strings.Split(string(output), "\n")
	details := make(map[string]string)
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			switch parts[0] {
			case "MYSQL_ROOT_PASSWORD":
				details["Password"] = parts[1]
			case "MYSQL_USER":
				details["User"] = parts[1]
			case "MYSQL_DATABASE":
				details["Database"] = parts[1]
			}
		}
	}

	// Get container IP address
	cmd = execCommand("docker", "inspect", "--format", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get MySQL container IP: %v", err)
	}
	details["Host"] = strings.TrimSpace(string(output))

	// Get exposed port
	cmd = execCommand("docker", "inspect", "--format", "{{range $p, $conf := .NetworkSettings.Ports}}{{if eq $p \"3306/tcp\"}}{{(index $conf 0).HostPort}}{{end}}{{end}}", containerName)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get MySQL container port: %v", err)
	}
	details["Port"] = strings.TrimSpace(string(output))

	// If Database is not set, use a default value
	if details["Database"] == "" {
		details["Database"] = "wordpress"
	}

	return details, nil
}
