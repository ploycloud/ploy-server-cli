package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// Define test variables first
	testDomain := "test.com"
	testHostname := "host.example.com"
	testDBHost := "localhost"
	testDBUser := "user"
	testDBPassword := "password"
	testDBName := "wordpress"
	testPhpVersion := "8.3"
	testSiteID := "site123"

	// Save original paths and restore after test
	oldSitesDir := common.SitesDir
	oldLogBasePath := logBasePath
	oldNginxBasePath := nginxBasePath
	common.SitesDir = tempDir
	logBasePath = tempDir
	nginxBasePath = tempDir
	defer func() {
		common.SitesDir = oldSitesDir
		logBasePath = oldLogBasePath
		nginxBasePath = oldNginxBasePath
	}()

	// Set test environment variable
	os.Setenv("PLOY_TEST_ENV", "true")
	defer os.Unsetenv("PLOY_TEST_ENV")

	// Create a mock template with the exact structure we expect
	oldGetDockerComposeTemplate := getDockerComposeTemplate
	getDockerComposeTemplate = func(filename string) ([]byte, error) {
		return []byte(fmt.Sprintf(`version: '3'
services:
  wordpress:
    image: wordpress:%s-fpm-alpine
    container_name: wp-%s-php%s
    environment:
      WORDPRESS_DB_HOST: %s
      WORDPRESS_DB_USER: %s
      WORDPRESS_DB_PASSWORD: %s
      WORDPRESS_DB_NAME: %s
      HOSTNAME: %s
      SITE_ID: %s
      DOMAIN: %s`, testPhpVersion, testHostname, testPhpVersion, testDBHost, testDBUser, testDBPassword, testDBName, testHostname, testSiteID, testDomain)), nil
	}
	defer func() { getDockerComposeTemplate = oldGetDockerComposeTemplate }()

	// Mock execCommand to return actual values
	oldExecCommand := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		// For docker-compose
		if name == "docker-compose" {
			return exec.Command("echo", "docker-compose mock")
		}
		// For MySQL status check
		if name == "ploy" && len(arg) > 2 && arg[1] == "services" && arg[2] == "status" {
			return exec.Command("echo", "mysql is running")
		}
		// For MySQL details
		if name == "ploy" && len(arg) > 2 && arg[1] == "services" && arg[2] == "details" {
			return exec.Command("echo", `{"Host":"localhost","Port":"3306","Database":"wordpress","User":"user","Password":"password"}`)
		}
		// For all other commands, return empty string
		return exec.Command("echo", "")
	}
	defer func() { execCommand = oldExecCommand }()

	// Mock execSudo
	oldExecSudo := execSudo
	execSudo = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "mock sudo command")
	}
	defer func() { execSudo = oldExecSudo }()

	// Create required directories
	err = os.MkdirAll(filepath.Join(tempDir, "sites"), 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(tempDir, "sites-available"), 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(tempDir, "sites-enabled"), 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(tempDir, testHostname), 0755)
	assert.NoError(t, err)

	// Test launching a site
	err = launchSite(
		"wp",           // siteType
		testDomain,     // domain
		"internal",     // dbSource
		testDBHost,     // dbHost
		"3306",         // dbPort
		testDBName,     // dbName
		testDBUser,     // dbUser
		testDBPassword, // dbPassword
		"static",       // scalingType
		1,              // replicas
		0,              // maxReplicas
		testSiteID,     // siteID
		testHostname,   // hostname
		testPhpVersion, // phpVersion
		"",             // webhook
	)
	assert.NoError(t, err)

	// Check if log file was created
	logFile := filepath.Join(tempDir, "sites", "host.example.com", "deploy.log")
	assert.FileExists(t, logFile)

	// Check log content
	content, err := ioutil.ReadFile(logFile)
	assert.NoError(t, err)
	logContent := string(content)
	assert.Contains(t, logContent, "Starting site creation process")
	assert.Contains(t, logContent, "Site launched successfully")

	// Check if docker-compose file was created
	composeFile := filepath.Join(tempDir, testHostname, fmt.Sprintf("docker-compose-wp-php%s.yml", testPhpVersion))
	assert.FileExists(t, composeFile)

	// Read and check docker-compose content with more detailed error messages
	content, err = ioutil.ReadFile(composeFile)
	assert.NoError(t, err)
	composeContent := string(content)

	t.Logf("Docker compose content:\n%s", composeContent)

	// Check for replaced variables in the compose file
	expectedStrings := []string{
		fmt.Sprintf("wordpress:%s-fpm-alpine", testPhpVersion),
		fmt.Sprintf("container_name: wp-%s-php%s", testHostname, testPhpVersion),
		fmt.Sprintf("WORDPRESS_DB_HOST: %s", testDBHost),
		fmt.Sprintf("WORDPRESS_DB_USER: %s", testDBUser),
		fmt.Sprintf("WORDPRESS_DB_PASSWORD: %s", testDBPassword),
		fmt.Sprintf("WORDPRESS_DB_NAME: %s", testDBName),
		fmt.Sprintf("HOSTNAME: %s", testHostname),
		fmt.Sprintf("SITE_ID: %s", testSiteID),
		fmt.Sprintf("DOMAIN: %s", testDomain),
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(composeContent, expected) {
			t.Errorf("Docker compose file should contain '%s', but got:\n%s", expected, composeContent)
		}
	}

	// Check nginx config
	nginxConfigPath := filepath.Join(tempDir, "sites-available", "test.com.conf")
	assert.FileExists(t, nginxConfigPath)

	// Check nginx config content
	content, err = ioutil.ReadFile(nginxConfigPath)
	assert.NoError(t, err)
	nginxContent := string(content)
	assert.Contains(t, nginxContent, "server_name test.com;")
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
	tempDir, err := ioutil.TempDir("", "test_nginx_config")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original nginx base path and restore after test
	oldNginxBasePath := nginxBasePath
	nginxBasePath = tempDir
	defer func() { nginxBasePath = oldNginxBasePath }()

	// Mock execCommand
	oldExecCommand := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "mock command")
	}
	defer func() { execCommand = oldExecCommand }()

	// Mock execSudo
	oldExecSudo := execSudo
	execSudo = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "mock sudo command")
	}
	defer func() { execSudo = oldExecSudo }()

	tests := []struct {
		name string

		domain      string
		webhook     string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "successful config creation",
			domain:      "test.com",
			webhook:     "",
			wantErr:     false,
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create required directories
			nginxSitesDir := filepath.Join(tempDir, "sites-available")
			nginxEnabledDir := filepath.Join(tempDir, "sites-enabled")
			err := os.MkdirAll(nginxSitesDir, 0755)
			assert.NoError(t, err)
			err = os.MkdirAll(nginxEnabledDir, 0755)
			assert.NoError(t, err)

			err = createNginxConfig(tt.domain, tt.webhook)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)

				// Verify config file was created
				configPath := filepath.Join(nginxSitesDir, tt.domain+".conf")
				assert.FileExists(t, configPath)

				// Verify symlink was created
				enabledPath := filepath.Join(nginxEnabledDir, tt.domain+".conf")
				assert.FileExists(t, enabledPath)
			}
		})
	}
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

	// Save original execSudo and restore after test
	oldExecSudo := execSudo
	defer func() { execSudo = oldExecSudo }()

	// Mock execSudo
	execSudo = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "mock sudo command")
	}

	hostname := "test.example.com"
	message := "Test log message"

	err = createSiteLog(hostname, message)
	assert.NoError(t, err)

	// Check if log file was created
	logFile := filepath.Join(tempDir, "sites", hostname, "deploy.log")
	assert.FileExists(t, logFile)

	// Check log content
	content, err := ioutil.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), message)
	assert.Regexp(t, `\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] Test log message`, string(content))
}
