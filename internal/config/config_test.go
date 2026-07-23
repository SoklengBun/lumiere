package config

import "testing"

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "lumiere")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASS", "postgres")
	t.Setenv("JWT_SECRET", "test-secret")
}

func TestNewFromEnvFrontendURLDefaults(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("FRONTEND_URL", "")

	cfg, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv() error = %v", err)
	}

	if want := "https://anella.vercel.app/"; cfg.FrontendURL != want {
		t.Fatalf("FrontendURL = %q, want %q", cfg.FrontendURL, want)
	}
}

func TestNewFromEnvFrontendURLCanBeConfigured(t *testing.T) {
	setRequiredEnv(t)
	want := "https://example.com/music"
	t.Setenv("FRONTEND_URL", want)

	cfg, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv() error = %v", err)
	}

	if cfg.FrontendURL != want {
		t.Fatalf("FrontendURL = %q, want %q", cfg.FrontendURL, want)
	}
}
