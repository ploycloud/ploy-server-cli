package commands

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ploycloud/ploy-server-cli/src/common"
	"github.com/stretchr/testify/assert"
)

// Existing imports and test setup...

func TestValidateInputs(t *testing.T) {
	tests := []struct {
		name        string
		siteType    string
		domain      string
		dbSource    string
		scalingType string
		replicas    int
		maxReplicas int
		expectError bool
	}{
		{"Valid WP site", "wp", "example.com", "internal", "static", 1, 0, false},
		{"Invalid site type", "invalid", "example.com", "internal", "static", 1, 0, true},
		{"Missing domain", "wp", "", "internal", "static", 1, 0, true},
		{"Invalid DB source", "wp", "example.com", "invalid", "static", 1, 0, true},
		{"Invalid scaling type", "wp", "example.com", "internal", "invalid", 1, 0, true},
		{"Invalid replicas", "wp", "example.com", "internal", "static", 0, 0, true},
		{"Invalid max replicas", "wp", "example.com", "internal", "dynamic", 2, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInputs(tt.siteType, tt.domain, tt.dbSource, tt.scalingType, tt.replicas, tt.maxReplicas)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLaunchSite(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "test_launch_site")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up test environment
	oldSitesDir := common.SitesDir
	common.SitesDir = tempDir
	defer func() { common.SitesDir = oldSitesDir }()

	// Mock the getDockerComposeTemplate function
	oldGetDockerComposeTemplate := getDockerComposeTemplate
	getDockerComposeTemplate = func(filename string) ([]byte, error) {
		return []byte("version: '3'\nservices:\n  wordpress:\n    image: wordpress:${PHP_VERSION}-fpm-alpine"), nil
	}
	defer func() { getDockerComposeTemplate = oldGetDockerComposeTemplate }()

	// Mock the execCommand function
	oldExecCommand := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "Mock command executed")
	}
	defer func() { execCommand = oldExecCommand }()

	// Test launching a site
	err = launchSite("wp", "example.com", "external", "db.example.com", "3306", "wordpress", "user", "password", "static", 2, 0, "site123", "host.example.com", "8.3")
	assert.NoError(t, err)

	// Check if the site directory was created
	siteDir := filepath.Join(tempDir, "host.example.com")
	assert.DirExists(t, siteDir, "Site directory should exist")

	// Check if the Docker Compose file was created in the site directory
	composePath := filepath.Join(siteDir, "docker-compose.yml")
	assert.FileExists(t, composePath, "Docker Compose file should exist in the site directory")

	// Read the content of the Docker Compose file
	content, err := ioutil.ReadFile(composePath)
	assert.NoError(t, err)

	// Check if the PHP version is correctly set
	assert.Contains(t, string(content), "wordpress:8.3-fpm-alpine", "PHP version should be set correctly")

	// Check if the Docker Compose command was executed
	assert.Contains(t, CaptureOutput(func() {
		launchSite("wp", "example.com", "external", "db.example.com", "3306", "wordpress", "user", "password", "static", 2, 0, "site123", "host.example.com", "8.3")
	}), "Mock command executed", "Docker Compose command should be executed")
}

// Add more tests for other functions as needed...
