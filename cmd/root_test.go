package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	// This is a basic test to ensure Execute doesn't panic
	assert.NotPanics(t, func() {
		Execute()
	})
}
