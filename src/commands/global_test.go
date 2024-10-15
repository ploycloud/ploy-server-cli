package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEchoCmd(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", "\n"},
		{"Hello World", "Hello World", "Hello World\n"},
		{"Multiple words", "This is a test", "This is a test\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := CaptureOutput(func() {
				EchoCmd.Run(EchoCmd, []string{tt.input})
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}
