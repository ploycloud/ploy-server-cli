package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ploycloud/ploy-server-cli/src/docker"
	"github.com/stretchr/testify/assert"
)

func TestStartCmd(t *testing.T) {
	// Create a temporary directory with a mock docker-compose.yml file
	tempDir, err := os.MkdirTemp("", "test_start_cmd")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	composePath := filepath.Join(tempDir, "docker-compose.yml")
	err = os.WriteFile(composePath, []byte("version: '3'"), 0644)
	assert.NoError(t, err)

	// Change to the temporary directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Mock RunCompose function
	oldRunCompose := docker.RunCompose
	docker.RunCompose = func(composePath string, args ...string) error {
		assert.Equal(t, "up", args[0])
		assert.Equal(t, "-d", args[1])
		return nil
	}
	defer func() { docker.RunCompose = oldRunCompose }()

	// Run the command
	output := CaptureOutput(func() {
		StartCmd.Run(StartCmd, []string{})
	})

	assert.Empty(t, output)
}

func TestStopCmd(t *testing.T) {
	// Similar setup as TestStartCmd
	tempDir, err := os.MkdirTemp("", "test_stop_cmd")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	composePath := filepath.Join(tempDir, "docker-compose.yml")
	err = os.WriteFile(composePath, []byte("version: '3'"), 0644)
	assert.NoError(t, err)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	oldRunCompose := docker.RunCompose
	docker.RunCompose = func(composePath string, args ...string) error {
		assert.Equal(t, "down", args[0])
		return nil
	}
	defer func() { docker.RunCompose = oldRunCompose }()

	output := CaptureOutput(func() {
		StopCmd.Run(StopCmd, []string{})
	})

	assert.Empty(t, output)
}

// Add similar tests for RestartCmd, ExecCmd, and LogsCmd
