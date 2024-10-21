package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ploycloud/ploy-server-cli/src/common"
	"github.com/stretchr/testify/assert"
)

func TestStatusCmd(t *testing.T) {
	// Create temporary directories and files
	tempDir, err := os.MkdirTemp("", "test_status_cmd")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	os.MkdirAll(filepath.Join(tempDir, "services"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "provisions"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "mysql"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "redis"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "nginx"), 0755)
	os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte("version: '3'"), 0644)

	// Mock common package variables
	oldServicesDir := common.ServicesDir
	oldGlobalCompose := common.GlobalCompose
	oldProvisionsDir := common.ProvisionsDir
	oldMysqlDir := common.MysqlDir
	oldRedisDir := common.RedisDir
	oldNginxDir := common.NginxDir

	common.SetServicesDir(filepath.Join(tempDir, "services"))
	common.SetGlobalCompose(filepath.Join(tempDir, "docker-compose.yml"))
	common.SetProvisionsDir(filepath.Join(tempDir, "provisions"))
	common.SetMysqlDir(filepath.Join(tempDir, "mysql"))
	common.SetRedisDir(filepath.Join(tempDir, "redis"))
	common.SetNginxDir(filepath.Join(tempDir, "nginx"))

	defer func() {
		common.SetServicesDir(oldServicesDir)
		common.SetGlobalCompose(oldGlobalCompose)
		common.SetProvisionsDir(oldProvisionsDir)
		common.SetMysqlDir(oldMysqlDir)
		common.SetRedisDir(oldRedisDir)
		common.SetNginxDir(oldNginxDir)
	}()

	// Mock exec.Command
	oldExecCommand := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cmd := exec.Command("echo")
		switch {
		case name == "docker" && arg[0] == "version":
			cmd = exec.Command("echo", "20.10.14")
		case name == "docker" && arg[0] == "info":
			cmd = exec.Command("echo", "Docker info output")
		}
		return cmd
	}
	defer func() { execCommand = oldExecCommand }()

	// Capture the output
	stdout, stderr := CaptureOutputAndError(func() {
		StatusCmd.Run(StatusCmd, []string{})
	})

	// Print captured output for debugging
	t.Logf("Captured stdout: %s", stdout)
	t.Logf("Captured stderr: %s", stderr)

	output := stdout + stderr
	assert.Contains(t, output, "services directory exists")
	assert.Contains(t, output, "docker-compose.yml exists")
	assert.Contains(t, output, "provisions directory exists")
	assert.Contains(t, output, "MySQL directory exists")
	assert.Contains(t, output, "Redis directory exists")
	assert.Contains(t, output, "Nginx directory exists")
	assert.Contains(t, output, "Docker is installed, version: 20.10.14")
	assert.Contains(t, output, "Docker is running")
}
