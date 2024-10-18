package commands

import (
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

func TestGenerateDockerComposeContent(t *testing.T) {
	content := generateDockerComposeContent("wp", "example.com", "db", "3306", "wordpress", "user", "password", "static", 2, 0)
	assert.Contains(t, content, "image: wordpress:latest")
	assert.Contains(t, content, "WORDPRESS_DB_HOST: db:3306")
	assert.Contains(t, content, "replicas: 2")
	assert.Contains(t, content, "traefik.http.routers.example.com.rule=Host(`example.com`)")
}

// Add more tests for other functions as needed...
