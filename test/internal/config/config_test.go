package config

import (
	"github.com/kodelint/shell-agent/internal/config"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg == nil {
		t.Error("Config should not be nil")
	}

	// Test default values
	if cfg.AI.DefaultModel == "" {
		t.Error("Default model should not be empty")
	}

	if cfg.AI.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}
}
