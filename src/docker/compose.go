package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ComposeFile represents the Docker Compose file
type ComposeFile struct {
	Path string
}

// NewComposeFile creates a new ComposeFile instance
func NewComposeFile(path string) *ComposeFile {
	return &ComposeFile{Path: path}
}

// Up starts the Docker Compose services
func (c *ComposeFile) Up(services ...string) error {
	args := []string{"compose", "-f", c.Path, "up", "-d"}
	args = append(args, services...)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Down stops and removes the Docker Compose services
func (c *ComposeFile) Down() error {
	cmd := exec.Command("docker", "compose", "-f", c.Path, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Restart restarts the Docker Compose services
func (c *ComposeFile) Restart(services ...string) error {
	args := []string{"compose", "-f", c.Path, "restart"}
	args = append(args, services...)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Logs shows the logs of the Docker Compose services
func (c *ComposeFile) Logs(services ...string) error {
	args := []string{"compose", "-f", c.Path, "logs", "--follow"}
	args = append(args, services...)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Exec executes a command in a running container
func (c *ComposeFile) Exec(service string, command ...string) error {
	args := []string{"compose", "-f", c.Path, "exec", service}
	args = append(args, command...)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// GetComposeFile returns the ComposeFile for the current directory
func GetComposeFile() (*ComposeFile, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	composePath := filepath.Join(cwd, "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("docker-compose.yml not found in the current directory")
	}

	return NewComposeFile(composePath), nil
}

// GetGlobalComposeFile returns the ComposeFile for the base services
func GetGlobalComposeFile() (*ComposeFile, error) {
	// Assuming the base docker-compose.yml is located in a specific directory
	basePath := "/path/to/base/docker-compose.yml"
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("base docker-compose.yml not found")
	}

	return NewComposeFile(basePath), nil
}
