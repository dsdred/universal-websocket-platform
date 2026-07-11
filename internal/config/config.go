package config

import "os"

// Config contains the Control Service startup configuration.
type Config struct {
	Port string
}

// Load reads configuration from the environment and applies defaults.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{Port: port}
}
