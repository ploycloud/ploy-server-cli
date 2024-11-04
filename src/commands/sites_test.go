package commands

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ploycloud/ploy-server-cli/src/common"
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
	// Create temporary directory for test
	tempDir, err := ioutil.TempDir("", "test_launch_site")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up test environment
	oldSitesDir := common.SitesDir
	common.SitesDir = tempDir
	defer func() { common.SitesDir = oldSitesDir }()

	// Save original log base path and restore after test
	oldLogBasePath := logBasePath
	logBasePath = tempDir
	defer func() { logBasePath = oldLogBasePath }()

	// Save original nginx base path and restore after test
	oldNginxBasePath := nginxBasePath
	nginxBasePath = tempDir
	defer func() { nginxBasePath = oldNginxBasePath }()

	// Create nginx directories
	nginxSitesDir := filepath.Join(tempDir, "sites-available")
	nginxEnabledDir := filepath.Join(tempDir, "sites-enabled")
	err = os.MkdirAll(nginxSitesDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(nginxEnabledDir, 0755)
	assert.NoError(t, err)

	// Mock the getDockerComposeTemplate function
	oldGetDockerComposeTemplate := getDockerComposeTemplate
	getDockerComposeTemplate = func(filename string) ([]byte, error) {
		return []byte("version: '3'\nservices:\n  wordpress:\n    image: wordpress:${PHP_VERSION}-fpm-alpine"), nil
	}
	defer func() { getDockerComposeTemplate = oldGetDockerComposeTemplate }()

	// Mock the execCommand function
	oldExecCommand := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if name == "systemctl" && arg[0] == "reload" && arg[1] == "nginx" {
			return exec.Command("echo", "nginx reloaded")
		}
		return exec.Command("echo", "Mock command executed")
	}
	defer func() { execCommand = oldExecCommand }()

	// Test with default domain
	err = launchSite("wp", "", "external", "db.example.com", "3306", "wordpress", "user", "password", "static", 2, 0, "site123", "host.example.com", "8.3", "")
	assert.NoError(t, err)

	// Check if the site directory was created
	siteDir := filepath.Join(tempDir, "host.example.com")
	assert.DirExists(t, siteDir, "Site directory should exist")

	// Check if the Docker Compose file was created
	composePath := filepath.Join(siteDir, "docker-compose-wp-php8.3.yml")
	assert.FileExists(t, composePath, "Docker Compose file should exist")

	// Check Docker Compose content
	content, err := ioutil.ReadFile(composePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "wordpress:8.3-fpm-alpine")

	// Check if nginx config was created
	nginxConfigPath := filepath.Join(nginxSitesDir, "host.example.com.localhost.conf")
	assert.FileExists(t, nginxConfigPath, "Nginx config file should exist")

	// Check nginx config content
	content, err = ioutil.ReadFile(nginxConfigPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "server_name host.example.com.localhost;")

	// Check if nginx config is enabled (symlinked)
	enabledPath := filepath.Join(nginxEnabledDir, "host.example.com.localhost.conf")
	assert.FileExists(t, enabledPath, "Nginx enabled config should exist")

	// Verify symlink
	linkTarget, err := os.Readlink(enabledPath)
	assert.NoError(t, err)
	assert.Equal(t, nginxConfigPath, linkTarget)

	// Check command execution
	output := CaptureOutput(func() {
		launchSite("wp", "example.com", "external", "db.example.com", "3306", "wordpress", "user", "password", "static", 2, 0, "site123", "host.example.com", "8.3", "")
	})
	assert.Contains(t, output, "Mock command executed")

	// Check if log directory and file were created
	logDir := filepath.Join(tempDir, "sites", "host.example.com")
	assert.DirExists(t, logDir, "Log directory should exist")

	logFile := filepath.Join(logDir, "creating.log")
	assert.FileExists(t, logFile, "Log file should exist")

	// Check log content
	content, err = ioutil.ReadFile(logFile)
	assert.NoError(t, err)
	logContent := string(content)
	assert.Contains(t, logContent, "Starting site creation process")
	assert.Contains(t, logContent, "Using default domain: host.example.com.localhost")
	assert.Contains(t, logContent, "Creating nginx configuration")
	assert.Contains(t, logContent, "Site launched successfully")
}

// Add more tests for other functions as needed...

func TestSetupNginxProxy(t *testing.T) {
	// Save original execCommand and restore after test
	oldExecCommand := execCommand
	defer func() { execCommand = oldExecCommand }()

	tests := []struct {
		name          string
		nginxStatus   string
		installOutput string
		expectError   bool
		expectedError string
	}{
		{
			name:        "nginx-proxy already running",
			nginxStatus: "nginx-proxy is running",
			expectError: false,
		},
		{
			name:          "nginx-proxy install success",
			nginxStatus:   "nginx-proxy is not running",
			installOutput: "Installation successful",
			expectError:   false,
		},
		{
			name:          "nginx-proxy install failure",
			nginxStatus:   "nginx-proxy is not running",
			installOutput: "",
			expectError:   true,
			expectedError: "failed to install nginx-proxy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommand = func(name string, arg ...string) *exec.Cmd {
				if arg[0] == "services" && arg[1] == "status" {
					return exec.Command("echo", tt.nginxStatus)
				}
				if arg[0] == "services" && arg[1] == "install" {
					if tt.installOutput != "" {
						return exec.Command("echo", tt.installOutput)
					}
					return exec.Command("false")
				}
				return exec.Command("echo", "unexpected command")
			}

			err := setupNginxProxy("")
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateNginxConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := ioutil.TempDir("", "nginx-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Save original nginx base path and restore after test
	oldNginxBasePath := nginxBasePath
	nginxBasePath = tmpDir
	defer func() { nginxBasePath = oldNginxBasePath }()

	// Save original execCommand and restore after test
	oldExecCommand := execCommand
	defer func() { execCommand = oldExecCommand }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "mock command")
	}

	// Test creating nginx config
	domain := "test.localhost"
	err = createNginxConfig(domain, "")
	assert.NoError(t, err)

	// Check if config file was created
	configPath := filepath.Join(tmpDir, "sites-available", domain+".conf")
	assert.FileExists(t, configPath)

	// Check if enabled symlink was created
	enabledPath := filepath.Join(tmpDir, "sites-enabled", domain+".conf")
	assert.FileExists(t, enabledPath)

	// Check config content
	content, err := ioutil.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "server_name test.localhost;")
	assert.Contains(t, string(content), "proxy_pass http://test-localhost:80;")
	assert.Contains(t, string(content), "proxy_set_header Upgrade $http_upgrade;")
	assert.Contains(t, string(content), "proxy_set_header Connection \"upgrade\";")

	// Verify symlink
	linkTarget, err := os.Readlink(enabledPath)
	assert.NoError(t, err)
	assert.Equal(t, configPath, linkTarget)

	// Test with subdomain
	domain = "app.example.com"
	err = createNginxConfig(domain, "")
	assert.NoError(t, err)

	// Check config content for subdomain
	configPath = filepath.Join(tmpDir, "sites-available", domain+".conf")
	content, err = ioutil.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "server_name app.example.com;")
	assert.Contains(t, string(content), "proxy_pass http://app-example-com:80;")
}

func TestCreateSiteLog(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := ioutil.TempDir("", "test_site_log")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original log base path and restore after test
	oldLogBasePath := logBasePath
	logBasePath = tempDir
	defer func() { logBasePath = oldLogBasePath }()

	hostname := "test.example.com"
	message := "Test log message"

	err = createSiteLog(hostname, message)
	assert.NoError(t, err)

	// Check if log file was created
	logFile := filepath.Join(tempDir, "sites", hostname, "creating.log")
	assert.FileExists(t, logFile)

	// Check log content
	content, err := ioutil.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), message)
	assert.Regexp(t, `\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] Test log message`, string(content))
}
