package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for empty DATABASE_URL")
	}
}

func TestLoadReadsDatabaseURLAndAPIPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app:password@postgres:5432/app?sslmode=disable")
	t.Setenv("API_PORT", "9090")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.DatabaseURL != "postgres://app:password@postgres:5432/app?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q", config.DatabaseURL)
	}

	if config.APIAddress != ":9090" {
		t.Fatalf("APIAddress = %q, want %q", config.APIAddress, ":9090")
	}
}

func TestLoadDefaultsAPIPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app:password@postgres:5432/app?sslmode=disable")
	t.Setenv("API_PORT", "")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.APIAddress != ":8080" {
		t.Fatalf("APIAddress = %q, want %q", config.APIAddress, ":8080")
	}
}
