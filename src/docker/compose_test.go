package docker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInteractive(t *testing.T) {
	result := isInteractive()
	assert.NotNil(t, result)
}

func TestRunCompose(t *testing.T) {
	// This is a basic test to ensure RunCompose doesn't panic
	assert.NotPanics(t, func() {
		RunCompose("test-compose.yml", "up", "-d")
	})
}

func TestGetContainerName(t *testing.T) {
	// Create a temporary docker-compose file
	content := `
services:
  php:
    image: php:7.4-fpm
`
	tmpfile, err := os.CreateTemp("", "docker-compose-*.yml")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)
	tmpfile.Close()

	containerName, err := getContainerName(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "php", containerName)
}
