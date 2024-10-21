package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForUpdates(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v1.0.1"}`))
	}))
	defer server.Close()

	// Set the ReleaseEndpoint to our mock server
	oldReleaseEndpoint := ReleaseEndpoint
	ReleaseEndpoint = server.URL
	defer func() { ReleaseEndpoint = oldReleaseEndpoint }()

	version, hasUpdate, err := CheckForUpdates()
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.1", version)
	assert.True(t, hasUpdate)
}

func TestGetLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"tag_name": "v1.0.1",
			"assets": [
				{
					"name": "ploy-linux-amd64.tar.gz",
					"browser_download_url": "https://example.com/ploy-linux-amd64.tar.gz"
				}
			]
		}`))
	}))
	defer server.Close()

	oldReleaseEndpoint := ReleaseEndpoint
	ReleaseEndpoint = server.URL
	defer func() { ReleaseEndpoint = oldReleaseEndpoint }()

	release, err := getLatestRelease()
	assert.NoError(t, err)
	assert.Equal(t, "v1.0.1", release.TagName)
	assert.Len(t, release.Assets, 1)
	assert.Equal(t, "ploy-linux-amd64.tar.gz", release.Assets[0].Name)
}

// Add more tests for SelfUpdate and getAssetURL functions
