package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
	oldHomeDir := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHomeDir)

	// Create mock Docker Compose templates
	dockerDir := filepath.Join(tempDir, "docker", "wp")
	err = os.MkdirAll(dockerDir, 0755)
	assert.NoError(t, err)

	staticTemplatePath := filepath.Join(dockerDir, "wp-compose-static.yml")
	dynamicTemplatePath := filepath.Join(dockerDir, "wp-compose-dynamic.yml")

	staticTemplateContent := `
version: '3'
services:
  wordpress:
    image: wordpress:latest
    environment:
      WORDPRESS_DB_HOST: ${DB_HOST}:${DB_PORT}
      WORDPRESS_DB_NAME: ${DB_NAME}
      WORDPRESS_DB_USER: ${DB_USER}
      WORDPRESS_DB_PASSWORD: ${DB_PASSWORD}
    deploy:
      replicas: ${REPLICAS}
    labels:
      - "traefik.http.routers.${DOMAIN}.rule=Host(` + "`${DOMAIN}`" + `)"
`
	err = ioutil.WriteFile(staticTemplatePath, []byte(staticTemplateContent), 0644)
	assert.NoError(t, err)

	dynamicTemplateContent := `
version: '3'
services:
  wordpress:
    image: wordpress:latest
    environment:
      WORDPRESS_DB_HOST: ${DB_HOST}:${DB_PORT}
      WORDPRESS_DB_NAME: ${DB_NAME}
      WORDPRESS_DB_USER: ${DB_USER}
      WORDPRESS_DB_PASSWORD: ${DB_PASSWORD}
    deploy:
      replicas: ${REPLICAS}
      update_config:
        parallelism: 1
    labels:
      - "traefik.http.routers.${DOMAIN}.rule=Host(` + "`${DOMAIN}`" + `)"
`
	err = ioutil.WriteFile(dynamicTemplatePath, []byte(dynamicTemplateContent), 0644)
	assert.NoError(t, err)

	// Mock the GitHub fetching function
	oldGetDockerComposeTemplate := getDockerComposeTemplate
	getDockerComposeTemplate = func(filename string) ([]byte, error) {
		if filename == "wp/wp-compose-static.yml" {
			return []byte(staticTemplateContent), nil
		} else if filename == "wp/wp-compose-dynamic.yml" {
			return []byte(dynamicTemplateContent), nil
		}
		return nil, fmt.Errorf("unknown template: %s", filename)
	}
	defer func() {
		getDockerComposeTemplate = oldGetDockerComposeTemplate
	}()

	// Mock the execCommand function
	oldExecCommand := execCommand
	defer func() { execCommand = oldExecCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "Mock command executed")
	}

	// Test static scaling
	err = launchSite("wp", "example.com", "external", "db.example.com", "3306", "wordpress", "user", "password", "static", 2, 0, "", "")
	assert.NoError(t, err)

	composePath := filepath.Join(tempDir, "example.com", "docker-compose.yml")
	content, err := ioutil.ReadFile(composePath)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "image: wordpress:latest")
	assert.Contains(t, string(content), "WORDPRESS_DB_HOST: db.example.com:3306")
	assert.Contains(t, string(content), "replicas: 2")
	assert.Contains(t, string(content), "traefik.http.routers.example.com.rule=Host(`example.com`)")

	// Test static scaling with siteID and hostname
	err = launchSite("wp", "example.com", "external", "db.example.com", "3306", "wordpress", "user", "password", "static", 2, 0, "site123", "host.example.com")
	assert.NoError(t, err)

	composePath = filepath.Join(tempDir, "example.com", "docker-compose.yml")
	content, err = ioutil.ReadFile(composePath)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "image: wordpress:latest")
	assert.Contains(t, string(content), "WORDPRESS_DB_HOST: db.example.com:3306")
	assert.Contains(t, string(content), "replicas: 2")
	assert.Contains(t, string(content), "traefik.http.routers.example.com.rule=Host(`example.com`)")
	assert.Contains(t, string(content), "SITE_ID: site123")
	assert.Contains(t, string(content), "HOSTNAME: host.example.com")

	// Test dynamic scaling
	err = launchSite("wp", "dynamic.com", "external", "db.example.com", "3306", "wordpress", "user", "password", "dynamic", 2, 5, "", "")
	assert.NoError(t, err)

	composePath = filepath.Join(tempDir, "dynamic.com", "docker-compose.yml")
	content, err = ioutil.ReadFile(composePath)
	assert.NoError(t, err)

	assert.Contains(t, string(content), "image: wordpress:latest")
	assert.Contains(t, string(content), "WORDPRESS_DB_HOST: db.example.com:3306")
	assert.Contains(t, string(content), "replicas: 2")
	assert.Contains(t, string(content), "update_config:")
	assert.Contains(t, string(content), "traefik.http.routers.dynamic.com.rule=Host(`dynamic.com`)")
}

// Add more tests for other functions as needed...
