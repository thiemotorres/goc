package config

// Config holds application configuration
type Config struct{}

// Load reads config from file
func Load() (*Config, error) {
	return &Config{}, nil
}
