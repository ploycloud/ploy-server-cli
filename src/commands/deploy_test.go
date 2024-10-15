package commands

import (
	"testing"

	"github.com/ploycloud/ploy-server-cli/src/utils"
	"github.com/stretchr/testify/assert"
)

func TestDeployCmd(t *testing.T) {
	// Store the original CloneRepo function
	originalCloneRepo := utils.CloneRepo

	// Create a mock function
	var mockCloneRepo func(string) error

	// Replace the CloneRepo function with our mock
	utils.CloneRepo = func(repo string) error {
		return mockCloneRepo(repo)
	}

	// Restore the original function after the test
	defer func() {
		utils.CloneRepo = originalCloneRepo
	}()

	tests := []struct {
		name        string
		repo        string
		mockClone   func(string) error
		expectedOut string
	}{
		{
			name: "Successful deployment",
			repo: "https://github.com/example/repo.git",
			mockClone: func(repo string) error {
				return nil
			},
			expectedOut: "Deploying repository: https://github.com/example/repo.git\nDeployment successful!\n",
		},
		{
			name: "Failed deployment",
			repo: "https://github.com/example/fail-repo.git",
			mockClone: func(repo string) error {
				return assert.AnError
			},
			expectedOut: "Deploying repository: https://github.com/example/fail-repo.git\nError cloning repository: assert.AnError general error for testing\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloneRepo = tt.mockClone

			output := CaptureOutput(func() {
				DeployCmd.Run(DeployCmd, []string{tt.repo})
			})

			assert.Equal(t, tt.expectedOut, output)
		})
	}
}
