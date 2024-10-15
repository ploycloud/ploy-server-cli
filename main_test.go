package main

import (
	"os"
	"strings"
	"testing"

	"github.com/ploycloud/ploy-server-cli/cmd"
	"github.com/ploycloud/ploy-server-cli/src/commands"
	"github.com/ploycloud/ploy-server-cli/src/common"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	// Save original args and osExit
	oldArgs := os.Args
	oldOsExit := osExit

	// Restore original args and osExit after the test
	defer func() {
		os.Args = oldArgs
		osExit = oldOsExit
	}()

	// Mock osExit
	var exitCode int
	osExit = func(code int) {
		exitCode = code
		panic("osExit") // Use panic to stop execution
	}

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedExit   int
	}{
		{
			name:           "No arguments",
			args:           []string{"ploy"},
			expectedOutput: "Ploy CLI is a powerful tool",
			expectedExit:   0,
		},
		{
			name:           "Valid command",
			args:           []string{"ploy", "version"},
			expectedOutput: common.CurrentCliVersion,
			expectedExit:   0,
		},
		{
			name:           "Invalid command",
			args:           []string{"ploy", "invalidcommand"},
			expectedOutput: "unknown command \"invalidcommand\" for \"ploy\"",
			expectedExit:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			exitCode = 0

			output := commands.CaptureOutput(func() {
				defer func() {
					if r := recover(); r != nil {
						if r != "osExit" {
							panic(r) // re-panic if it's not our expected panic
						}
					}
				}()
				err := cmd.Execute()
				if err != nil {
					osExit(1)
				}
			})

			t.Logf("Full output:\n%s", output)
			assert.Contains(t, strings.ToLower(output), strings.ToLower(tt.expectedOutput))
			assert.Equal(t, tt.expectedExit, exitCode)
		})
	}
}
