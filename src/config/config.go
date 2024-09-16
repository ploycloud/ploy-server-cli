package config

// Config holds the configuration for the PloyCloud CLI
type Config struct {
	APIKey string
	Region string
}

// LoadConfig loads the configuration from a file or environment variables
func LoadConfig() (*Config, error) {
	// Implement configuration loading logic
	return &Config{
		APIKey: "your-api-key",
		Region: "us-west-2",
	}, nil
}
