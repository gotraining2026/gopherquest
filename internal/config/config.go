package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contains the small set of environment-backed service settings.
type Config struct {
	Port            string
	DatabaseURL     string
	ProxyBaseURL    string
	HTTPTimeout     time.Duration
	DatabaseTimeout time.Duration
}

// Load reads service configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		Port:            envOrDefault("PORT", "8080"),
		DatabaseURL:     strings.TrimSpace(os.Getenv("DATABASE_URL")),
		ProxyBaseURL:    envOrDefault("GOPROXY_BASE_URL", "https://proxy.golang.org"),
		HTTPTimeout:     5 * time.Second,
		DatabaseTimeout: 3 * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if _, err := strconv.Atoi(cfg.Port); err != nil {
		return Config{}, fmt.Errorf("PORT must be numeric: %w", err)
	}

	var err error
	if cfg.HTTPTimeout, err = durationFromEnv("HTTP_TIMEOUT", cfg.HTTPTimeout); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseTimeout, err = durationFromEnv("DB_TIMEOUT", cfg.DatabaseTimeout); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func durationFromEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		if err == nil {
			err = fmt.Errorf("duration must be positive")
		}
		return 0, fmt.Errorf("%s is invalid: %w", name, err)
	}
	return duration, nil
}
