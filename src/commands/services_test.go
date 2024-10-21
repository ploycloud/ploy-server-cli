package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/fatih/color"
	"github.com/ploycloud/ploy-server-cli/src/docker"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Mock functions
var mockRunCompose func(composePath string, args ...string) error
var mockExecCommand func(name string, arg ...string) *exec.Cmd
var exitCalled bool
var exitCode int

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Set GO_TEST environment variable
	os.Setenv("GO_TEST", "1")

	// Save original functions
	originalRunCompose := docker.RunCompose
	originalExecCommand := execCommand

	// Set up mocks
	docker.RunCompose = func(composePath string, args ...string) error {
		return mockRunCompose(composePath, args...)
	}
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return mockExecCommand(name, arg...)
	}

	// Disable color output for tests
	color.NoColor = true

	// Run tests
	code := m.Run()

	// Restore original functions
	docker.RunCompose = originalRunCompose
	execCommand = originalExecCommand

	os.Exit(code)
}

func setupTest() {
	mockRunCompose = func(composePath string, args ...string) error {
		fmt.Println("Mock docker-compose command executed")
		return nil
	}
	mockExecCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "Mock command executed")
	}
	exitCalled = false
	exitCode = 0
}

func createMockMySQLComposeFile() {
	content := `version: '3'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_PASSWORD}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    ports:
      - "${MYSQL_PORT}:3306"
`
	os.MkdirAll("docker/databases", 0755)
	ioutil.WriteFile("docker/databases/mysql-compose.yml", []byte(content), 0644)
}

func TestGlobalStartCmd(t *testing.T) {
	setupTest()

	output := CaptureOutput(
		func() {
			globalStartCmd.Run(&cobra.Command{}, []string{})
		},
	)

	t.Logf("Full output:\n%s", output)
	assert.Contains(t, output, "Starting global services (mysql, redis, nginx-proxy)")
	assert.Contains(t, output, "Mock docker-compose command executed")
	assert.Contains(t, output, "Global services started successfully")
}

func TestGlobalStopCmd(t *testing.T) {
	setupTest()

	output := CaptureOutput(
		func() {
			globalStopCmd.Run(&cobra.Command{}, []string{})
		},
	)

	t.Logf("Full output:\n%s", output)
	assert.Contains(t, output, "Stopping global services")
	assert.Contains(t, output, "Mock docker-compose command executed")
	assert.Contains(t, output, "Global services stopped successfully")
}

func TestGlobalRestartCmd(t *testing.T) {
	setupTest()

	output := CaptureOutput(
		func() {
			globalRestartCmd.Run(&cobra.Command{}, []string{})
		},
	)

	t.Logf("Full output:\n%s", output)
	assert.Contains(t, output, "Restarting global services")
	assert.Contains(t, output, "Mock docker-compose command executed")
	assert.Contains(t, output, "Global services restarted successfully")
}

func TestInstallNginxProxyCmd(t *testing.T) {
	setupTest()

	testCases := []struct {
		name     string
		goos     string
		nginxCmd func(name string, arg ...string) *exec.Cmd
		expected string
	}{
		{
			name: "macOS",
			goos: "darwin",
			nginxCmd: func(name string, arg ...string) *exec.Cmd {
				return exec.Command("echo", "")
			},
			expected: "Nginx installation is not supported on macOS",
		},
		{
			name: "Linux - Nginx not installed",
			goos: "linux",
			nginxCmd: func(name string, arg ...string) *exec.Cmd {
				if name == "nginx" {
					return exec.Command("false")
				}
				return exec.Command("echo", "Mock command executed")
			},
			expected: "Nginx installed and configured successfully as a proxy",
		},
		{
			name: "Linux - Nginx already installed",
			goos: "linux",
			nginxCmd: func(name string, arg ...string) *exec.Cmd {
				return exec.Command("echo", "nginx version: nginx/1.18.0")
			},
			expected: "Nginx is already installed",
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				MockGOOS = tc.goos
				mockExecCommand = tc.nginxCmd

				output := CaptureOutput(
					func() {
						installNginxProxyCmd.Run(installNginxProxyCmd, []string{})
					},
				)

				t.Logf("Full output:\n%s", output)
				assert.Contains(t, output, tc.expected)
			},
		)
	}
}

func TestInstallMySQLCmd(t *testing.T) {
	setupTest()

	// Mock the GetDockerComposeTemplate function
	oldGetDockerComposeTemplate := getDockerComposeTemplate
	defer func() { getDockerComposeTemplate = oldGetDockerComposeTemplate }()
	getDockerComposeTemplate = func(filename string) ([]byte, error) {
		return []byte(`version: '3'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_PASSWORD}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    ports:
      - "${MYSQL_PORT}:3306"
`), nil
	}

	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "Default installation",
			args:     []string{},
			expected: "MySQL installed successfully",
		},
		{
			name:     "Custom user and password",
			args:     []string{"--user=testuser", "--password=testpass"},
			expected: "MySQL installed successfully",
		},
		{
			name:     "Custom port",
			args:     []string{"--port=3307"},
			expected: "MySQL installed successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock the docker-compose command
			mockExecCommand = func(name string, arg ...string) *exec.Cmd {
				return exec.Command("echo", "MySQL installed successfully")
			}

			// Create a new command and set flags
			cmd := &cobra.Command{}
			cmd.Flags().String("user", "default_user", "MySQL user")
			cmd.Flags().String("password", "default_password", "MySQL password")
			cmd.Flags().String("port", "3306", "MySQL port")

			// Parse flags
			cmd.ParseFlags(tc.args)

			// Capture output
			output := CaptureOutput(func() {
				installMySQLCmd.Run(cmd, []string{})
			})

			assert.Contains(t, output, "Installing MySQL service...")
			assert.Contains(t, output, tc.expected)
		})
	}
}

func TestDetailsCmd(t *testing.T) {
	setupTest()

	// Mock the execCommand function to return specific output for MySQL details
	oldExecCommand := execCommand
	defer func() { execCommand = oldExecCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cmd := exec.Command("echo", "mysql-container")
		if name == "docker" && arg[0] == "inspect" {
			switch arg[2] {
			case "{{range .Config.Env}}{{println .}}{{end}}":
				cmd = exec.Command("echo", "MYSQL_ROOT_PASSWORD=wp_password\nMYSQL_USER=wp_user\nMYSQL_DATABASE=wordpress")
			case "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}":
				cmd = exec.Command("echo", "172.17.0.2")
			case "{{range $p, $conf := .NetworkSettings.Ports}}{{if eq $p \"3306/tcp\"}}{{(index $conf 0).HostPort}}{{end}}{{end}}":
				cmd = exec.Command("echo", "3306")
			}
		}
		return cmd
	}

	// Test MySQL details
	stdout, _ := CaptureOutputAndError(func() {
		detailsCmd.Run(detailsCmd, []string{"mysql"})
	})

	assert.Contains(t, stdout, "Host: 172.17.0.2")
	assert.Contains(t, stdout, "Port: 3306")
	assert.Contains(t, stdout, "Database: wordpress")
	assert.Contains(t, stdout, "User: wp_user")
	assert.Contains(t, stdout, "Password: wp_password")

	// Test unsupported service
	_, stderr := CaptureOutputAndError(func() {
		detailsCmd.Run(detailsCmd, []string{"unsupported"})
	})

	assert.Contains(t, stderr, "Error getting unsupported details: unsupported service: unsupported")
}
