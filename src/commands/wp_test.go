package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWpCmd(t *testing.T) {
	// Create a temporary directory with a mock docker-compose.yml file
	tempDir, err := os.MkdirTemp("", "test_wp_cmd")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	composePath := filepath.Join(tempDir, "docker-compose.yml")
	err = os.WriteFile(composePath, []byte("version: '3'"), 0644)
	assert.NoError(t, err)

	// Change to the temporary directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Mock the runWpCli function
	oldRunWpCli := runWpCli
	runWpCli = func(composePath string, args []string) error {
		assert.Equal(t, []string{"plugin", "list"}, args)
		return nil
	}
	defer func() { runWpCli = oldRunWpCli }()

	// Run the command
	output := CaptureOutput(func() {
		WpCmd.Run(WpCmd, []string{"plugin", "list"})
	})

	assert.Empty(t, output)
}
