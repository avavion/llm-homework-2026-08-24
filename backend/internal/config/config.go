package config

import (
	"fmt"
	"os"
)

const defaultAPIPort = "8080"

type Config struct {
	DatabaseURL string
	APIAddress  string
}

func Load() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = defaultAPIPort
	}

	return Config{
		DatabaseURL: databaseURL,
		APIAddress:  ":" + port,
	}, nil
}
