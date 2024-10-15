package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloneRepo(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-repo-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to the temporary directory
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Test with a non-existent repo (should fail)
	err = CloneRepo("https://github.com/nonexistent/repo.git")
	assert.Error(t, err)

	// Test with a valid repo (you may want to use a mock or a known public repo)
	// err = CloneRepo("https://github.com/octocat/Hello-World.git")
	// assert.NoError(t, err)
}
