package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("GOPROXY_BASE_URL", "https://proxy.example.test")
	t.Setenv("HTTP_TIMEOUT", "2s")
	t.Setenv("DB_TIMEOUT", "1s")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got.Port != "9090" || got.DatabaseURL != "postgres://example" || got.ProxyBaseURL != "https://proxy.example.test" {
		t.Fatalf("Load returned config: %#v", got)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load returned nil error without DATABASE_URL")
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("HTTP_TIMEOUT", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load returned nil error for invalid HTTP_TIMEOUT")
	}
}
