package docker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDockerComposeTemplate(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock template content"))
	}))
	defer server.Close()

	// Mock the getGitHubURL function
	oldGetGitHubURL := getGitHubURL
	getGitHubURL = func() string { return server.URL + "/" }
	defer func() { getGitHubURL = oldGetGitHubURL }()

	content, err := GetDockerComposeTemplate("test-template.yml")
	assert.NoError(t, err)
	assert.Equal(t, "mock template content", string(content))
}

func TestGetDockerComposeTemplateError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Mock the getGitHubURL function
	oldGetGitHubURL := getGitHubURL
	getGitHubURL = func() string { return server.URL + "/" }
	defer func() { getGitHubURL = oldGetGitHubURL }()

	_, err := GetDockerComposeTemplate("non-existent-template.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch template from GitHub: status code 404")
}
