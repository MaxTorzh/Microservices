package config

import "os"

type Config struct {
	Port     string
	LogLevel string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		Port:     port,
		LogLevel: logLevel,
	}
}

func (c *Config) GetPort() string {
	return ":" + c.Port
}

func (c *Config) GetLogLevel() string {
	return c.LogLevel
}
