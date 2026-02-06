package config

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("READ_TIMEOUT", "5s")
	t.Setenv("WRITE_TIMEOUT", "6s")
	t.Setenv("IDLE_TIMEOUT", "7s")
	t.Setenv("SHUTDOWN_TIMEOUT", "8s")
	t.Setenv("REQUEST_TIMEOUT", "9s")

	cfg := LoadConfig()
	if cfg.JWTSecret != "secret" || cfg.DBPassword != "pass" {
		t.Fatalf("expected secrets to be set")
	}
	if cfg.DBSSLMode != "require" {
		t.Fatalf("expected sslmode require")
	}
	if cfg.ReadTimeout != 5*time.Second || cfg.WriteTimeout != 6*time.Second || cfg.IdleTimeout != 7*time.Second || cfg.ShutdownTimeout != 8*time.Second || cfg.RequestTimeout != 9*time.Second {
		t.Fatalf("expected timeouts to be parsed")
	}
}
