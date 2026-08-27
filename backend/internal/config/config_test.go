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

func TestLoadDefaultsFrontendOrigins(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app:password@postgres:5432/app?sslmode=disable")
	t.Setenv("FRONTEND_ORIGINS", "")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(config.FrontendOrigins) != 2 {
		t.Fatalf("FrontendOrigins = %v, want the two local dev defaults", config.FrontendOrigins)
	}
}

func TestLoadReadsFrontendOriginsList(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app:password@postgres:5432/app?sslmode=disable")
	t.Setenv("FRONTEND_ORIGINS", "https://app.example.com, https://staging.example.com")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := []string{"https://app.example.com", "https://staging.example.com"}
	if len(config.FrontendOrigins) != len(want) || config.FrontendOrigins[0] != want[0] || config.FrontendOrigins[1] != want[1] {
		t.Fatalf("FrontendOrigins = %v, want %v", config.FrontendOrigins, want)
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

func TestLoadDefaultsRecognitionProviderEmpty(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app:password@postgres:5432/app?sslmode=disable")
	t.Setenv("RECOGNITION_PROVIDER", "")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.RecognitionProvider != "" {
		t.Fatalf("RecognitionProvider = %q, want empty by default", config.RecognitionProvider)
	}
}

func TestLoadReadsRecognitionProvider(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app:password@postgres:5432/app?sslmode=disable")
	t.Setenv("RECOGNITION_PROVIDER", "mock")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.RecognitionProvider != "mock" {
		t.Fatalf("RecognitionProvider = %q, want %q", config.RecognitionProvider, "mock")
	}
}
