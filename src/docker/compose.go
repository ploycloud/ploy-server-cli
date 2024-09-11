package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"
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
	return RunCompose(c.Path, append([]string{"up", "-d"}, services...)...)
}

// Down stops and removes the Docker Compose services
func (c *ComposeFile) Down() error {
	return RunCompose(c.Path, "down")
}

// Restart restarts the Docker Compose services
func (c *ComposeFile) Restart(services ...string) error {
	return RunCompose(c.Path, append([]string{"restart"}, services...)...)
}

// Logs shows the logs of the Docker Compose services
func (c *ComposeFile) Logs(services ...string) error {
	return RunCompose(c.Path, append([]string{"logs", "--follow"}, services...)...)
}

// Exec executes a command in a running container
func (c *ComposeFile) Exec(service string, command ...string) error {
	return RunCompose(c.Path, append([]string{"exec", service}, command...)...)
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

// ComposeConfig represents the Docker Compose configuration
type ComposeConfig struct {
	Services map[string]interface{} `yaml:"services"`
}

func isInteractive() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func RunCompose(composePath string, args ...string) error {
	baseArgs := []string{"compose", "-f", composePath}

	// Check if 'exec' is the first argument and add -T if not interactive
	if len(args) > 0 && args[0] == "exec" && !isInteractive() {
		baseArgs = append(baseArgs, "exec", "-T")
		args = args[1:]
	}

	fullArgs := append(baseArgs, args...)

	cmd := exec.Command("docker", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func getContainerName(composePath string) (string, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return "", err
	}

	var config ComposeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", err
	}

	if _, exists := config.Services["php"]; exists {
		return "php", nil
	} else if _, exists := config.Services["litespeed"]; exists {
		return "litespeed", nil
	}

	return "", fmt.Errorf("no suitable container found")
}

func RunWpCli(composePath string, args []string) error {
	containerName, err := getContainerName(composePath)
	if err != nil {
		return err
	}

	wpArgs := append([]string{"exec", containerName, "wp"}, args...)
	return RunCompose(composePath, wpArgs...)
}
