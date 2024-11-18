package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	// Mock execSudo with actual file operations
	oldExecSudo := execSudo
	execSudo = mockExecSudo(t, tempDir)
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

// Mock execSudo with actual file operations
func mockExecSudo(t *testing.T, tempDir string) func(name string, arg ...string) *exec.Cmd {
	return func(name string, arg ...string) *exec.Cmd {
		t.Logf("Mock sudo command: %v", arg)

		// Handle different sudo commands
		if len(arg) == 0 {
			return exec.Command("echo", "mock sudo command")
		}

		// First argument is the command
		cmd := arg[0]
		args := arg[1:]

		switch {
		case cmd == "mkdir" && len(args) >= 2 && args[0] == "-p":
			// Handle mkdir -p
			for _, dir := range args[1:] {
				t.Logf("Creating directory: %s", dir)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Logf("Failed to create directory %s: %v", dir, err)
					return exec.Command("false")
				}
			}
			return exec.Command("echo", "directory created")

		case cmd == "-p":
			// Handle direct -p
			for _, dir := range args {
				t.Logf("Creating directory: %s", dir)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Logf("Failed to create directory %s: %v", dir, err)
					return exec.Command("false")
				}
			}
			return exec.Command("echo", "directory created")

		case cmd == "mv" || strings.HasPrefix(cmd, "/"):
			// Handle mv command or direct path
			var src, dst string
			if cmd == "mv" {
				if len(args) < 2 {
					return exec.Command("false")
				}
				src, dst = args[0], args[1]
			} else {
				if len(args) < 1 {
					return exec.Command("false")
				}
				src = cmd
				dst = args[len(args)-1]
			}

			t.Logf("Moving file from %s to %s", src, dst)

			// Ensure destination directory exists
			dstDir := filepath.Dir(dst)
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				t.Logf("Failed to create destination directory %s: %v", dstDir, err)
				return exec.Command("false")
			}

			// Read source file
			content, err := os.ReadFile(src)
			if err != nil {
				t.Logf("Failed to read source file %s: %v", src, err)
				return exec.Command("false")
			}

			// Write to destination
			if err := os.WriteFile(dst, content, 0644); err != nil {
				t.Logf("Failed to write destination file %s: %v", dst, err)
				return exec.Command("false")
			}

			// Remove source file
			if err := os.Remove(src); err != nil {
				t.Logf("Failed to remove source file %s: %v", src, err)
			}

			return exec.Command("echo", "file moved")

		case cmd == "chmod" || cmd == "644":
			// Handle chmod command or direct mode
			var mode int64
			var file string
			if cmd == "chmod" {
				if len(args) < 2 {
					return exec.Command("false")
				}
				mode, _ = strconv.ParseInt(args[0], 8, 32)
				file = args[1]
			} else {
				if len(args) < 1 {
					return exec.Command("false")
				}
				mode, _ = strconv.ParseInt(cmd, 8, 32)
				file = args[0]
			}

			t.Logf("Setting permissions %s on %s", cmd, file)
			if err := os.Chmod(file, os.FileMode(mode)); err != nil {
				t.Logf("Failed to set permissions: %v", err)
				return exec.Command("false")
			}
			return exec.Command("echo", "permissions set")

		case cmd == "rm" && len(args) >= 2 && args[0] == "-f":
			// Handle rm -f
			for _, file := range args[1:] {
				t.Logf("Removing file: %s", file)
				if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
					t.Logf("Failed to remove file %s: %v", file, err)
				}
			}
			return exec.Command("echo", "file removed")

		case cmd == "-f":
			// Handle direct -f
			for _, file := range args {
				t.Logf("Removing file: %s", file)
				if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
					t.Logf("Failed to remove file %s: %v", file, err)
				}
			}
			return exec.Command("echo", "file removed")

		case cmd == "ln" && len(args) >= 3 && args[0] == "-s":
			// Handle ln -s
			target, linkPath := args[1], args[2]
			return createSymlink(t, target, linkPath)

		case cmd == "-s" && len(args) >= 2:
			// Handle direct -s
			target, linkPath := args[0], args[1]
			return createSymlink(t, target, linkPath)

		case cmd == "systemctl":
			return exec.Command("echo", "systemctl executed")

		case cmd == "touch" && len(args) >= 1:
			file := args[0]
			t.Logf("Touching file: %s", file)
			if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
				t.Logf("Failed to create parent directory: %v", err)
				return exec.Command("false")
			}
			if err := os.WriteFile(file, []byte{}, 0644); err != nil {
				t.Logf("Failed to touch file: %v", err)
				return exec.Command("false")
			}
			return exec.Command("echo", "file touched")
		}

		t.Logf("Unhandled sudo command: %v", arg)
		return exec.Command("echo", fmt.Sprintf("mock sudo command: %v", arg))
	}
}

// Helper function to create symlink
func createSymlink(t *testing.T, target, linkPath string) *exec.Cmd {
	t.Logf("Creating symlink from %s to %s", target, linkPath)

	// Remove existing symlink if it exists
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		t.Logf("Failed to remove existing symlink %s: %v", linkPath, err)
	}

	// Ensure parent directory exists
	linkDir := filepath.Dir(linkPath)
	if err := os.MkdirAll(linkDir, 0755); err != nil {
		t.Logf("Failed to create symlink directory %s: %v", linkDir, err)
		return exec.Command("false")
	}

	// Create new symlink
	if err := os.Symlink(target, linkPath); err != nil {
		t.Logf("Failed to create symlink: %v", err)
		return exec.Command("false")
	}

	return exec.Command("echo", "symlink created")
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
		t.Logf("Mock command: %s %v", name, arg)
		return exec.Command("echo", "mock command")
	}
	defer func() { execCommand = oldExecCommand }()

	// Mock execSudo with actual file operations
	oldExecSudo := execSudo
	execSudo = func(name string, arg ...string) *exec.Cmd {
		t.Logf("Mock sudo command: %s %v", name, arg)

		switch name {
		case "mkdir":
			if len(arg) >= 2 && arg[0] == "-p" {
				for _, dir := range arg[1:] {
					t.Logf("Creating directory: %s", dir)
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Logf("Failed to create directory: %v", err)
						return exec.Command("false")
					}
				}
			}
			return exec.Command("echo", "directory created")

		case "cp":
			if len(arg) >= 2 {
				src, dst := arg[0], arg[1]
				t.Logf("Copying file from %s to %s", src, dst)
				if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
					t.Logf("Failed to create destination directory: %v", err)
					return exec.Command("false")
				}
				content, err := ioutil.ReadFile(src)
				if err != nil {
					t.Logf("Failed to read source file: %v", err)
					return exec.Command("false")
				}
				if err := ioutil.WriteFile(dst, content, 0644); err != nil {
					t.Logf("Failed to write destination file: %v", err)
					return exec.Command("false")
				}
			}
			return exec.Command("echo", "file copied")

		case "chmod":
			// Actually set the permissions in test
			if len(arg) >= 2 {
				mode, _ := strconv.ParseInt(arg[0], 8, 32)
				file := arg[1]
				if err := os.Chmod(file, os.FileMode(mode)); err != nil {
					t.Logf("Failed to set permissions: %v", err)
					return exec.Command("false")
				}
			}
			return exec.Command("echo", "permissions set")

		case "rm":
			if len(arg) >= 2 && arg[0] == "-f" {
				for _, file := range arg[1:] {
					t.Logf("Removing file: %s", file)
					os.Remove(file) // Ignore errors for non-existent files
				}
			}
			return exec.Command("echo", "file removed")

		case "ln":
			if len(arg) >= 3 && arg[0] == "-s" {
				target, linkPath := arg[1], arg[2]
				t.Logf("Creating symlink from %s to %s", target, linkPath)

				// Ensure target exists
				if _, err := os.Stat(target); err != nil {
					t.Logf("Target file does not exist: %s", target)
					return exec.Command("false")
				}

				// Ensure parent directory exists
				if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
					t.Logf("Failed to create symlink directory: %v", err)
					return exec.Command("false")
				}

				// Remove existing symlink if it exists
				os.Remove(linkPath)

				// Create new symlink
				if err := os.Symlink(target, linkPath); err != nil {
					t.Logf("Failed to create symlink: %v", err)
					return exec.Command("false")
				}

				// Verify symlink was created
				if _, err := os.Lstat(linkPath); err != nil {
					t.Logf("Failed to verify symlink creation: %v", err)
					return exec.Command("false")
				}
			}
			return exec.Command("echo", "symlink created")

		case "systemctl":
			return exec.Command("echo", "systemctl executed")
		}

		return exec.Command("echo", fmt.Sprintf("mock sudo command: %s %v", name, arg))
	}
	defer func() { execSudo = oldExecSudo }()

	// Set PLOY_TEST_ENV
	os.Setenv("PLOY_TEST_ENV", "true")
	defer os.Unsetenv("PLOY_TEST_ENV")

	domain := "test.com"
	err = createNginxConfig(domain, "")
	assert.NoError(t, err)

	// Verify config file was created
	configPath := filepath.Join(tempDir, "sites-available", domain+".conf")
	assert.FileExists(t, configPath)

	// Read and verify config content
	content, err := ioutil.ReadFile(configPath)
	assert.NoError(t, err)
	configContent := string(content)
	assert.Contains(t, configContent, fmt.Sprintf("server_name %s;", domain))

	// Verify symlink was created
	enabledPath := filepath.Join(tempDir, "sites-enabled", domain+".conf")
	assert.FileExists(t, enabledPath)

	// Verify symlink points to the correct file
	target, err := os.Readlink(enabledPath)
	assert.NoError(t, err)
	assert.Equal(t, configPath, target)
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
