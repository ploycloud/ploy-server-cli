package commands

import (
	"fmt"
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
	MockGOOS = "linux" // Default to Linux
}

func TestGlobalStartCmd(t *testing.T) {
	setupTest()

	output := CaptureOutput(func() {
		globalStartCmd.Run(&cobra.Command{}, []string{})
	})

	t.Logf("Full output:\n%s", output)
	assert.Contains(t, output, "Starting global services (mysql, redis, nginx-proxy)")
	assert.Contains(t, output, "Mock docker-compose command executed")
	assert.Contains(t, output, "Global services started successfully")
}

func TestGlobalStopCmd(t *testing.T) {
	setupTest()

	output := CaptureOutput(func() {
		globalStopCmd.Run(&cobra.Command{}, []string{})
	})

	t.Logf("Full output:\n%s", output)
	assert.Contains(t, output, "Stopping global services")
	assert.Contains(t, output, "Mock docker-compose command executed")
	assert.Contains(t, output, "Global services stopped successfully")
}

func TestGlobalRestartCmd(t *testing.T) {
	setupTest()

	output := CaptureOutput(func() {
		globalRestartCmd.Run(&cobra.Command{}, []string{})
	})

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
		t.Run(tc.name, func(t *testing.T) {
			MockGOOS = tc.goos
			mockExecCommand = tc.nginxCmd

			output := CaptureOutput(func() {
				installNginxProxyCmd.Run(&cobra.Command{}, []string{})
			})

			t.Logf("Full output:\n%s", output)
			assert.Contains(t, output, tc.expected)
		})
	}
}
