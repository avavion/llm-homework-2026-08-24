package config

import (
	"fmt"
	"os"
	"strings"
)

const defaultAPIPort = "8080"

// defaultFrontendOrigins covers the two most common local dev server ports
// (Vite and Create React App/Next.js) so a frontend can call the API with
// cookies out of the box; override with FRONTEND_ORIGINS in any other setup.
var defaultFrontendOrigins = []string{"http://localhost:5173", "http://localhost:3000"}

type Config struct {
	DatabaseURL     string
	APIAddress      string
	FrontendOrigins []string
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
		DatabaseURL:     databaseURL,
		APIAddress:      ":" + port,
		FrontendOrigins: frontendOrigins(),
	}, nil
}

func frontendOrigins() []string {
	raw := os.Getenv("FRONTEND_ORIGINS")
	if raw == "" {
		return defaultFrontendOrigins
	}

	var origins []string
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
