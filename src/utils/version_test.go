package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Add any version-related functions and their tests here
// For example:

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	assert.NotEmpty(t, version)
}
