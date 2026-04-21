package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cfg, warnings, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(warnings))
	}
	if cfg.Safety.Level != "balanced" {
		t.Fatalf("unexpected default safety level: %q", cfg.Safety.Level)
	}
}

func TestLoadAppliesKnownKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := strings.TrimSpace(`
[safety]
level = "strict"

[provider]
model = "cerebras-test"
timeout_ms = 2500
max_retries = 2

[ui]
visible_suggestion_count = 7

[telemetry]
enabled = true
`)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, warnings, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(warnings))
	}
	if cfg.Safety.Level != "strict" {
		t.Fatalf("unexpected safety level: %q", cfg.Safety.Level)
	}
	if cfg.Provider.Model != "cerebras-test" {
		t.Fatalf("unexpected provider model: %q", cfg.Provider.Model)
	}
	if cfg.Provider.TimeoutMS != 2500 {
		t.Fatalf("unexpected timeout: %d", cfg.Provider.TimeoutMS)
	}
	if cfg.UI.VisibleSuggestionCount != 7 {
		t.Fatalf("unexpected suggestion count: %d", cfg.UI.VisibleSuggestionCount)
	}
	if !cfg.Telemetry.Enabled {
		t.Fatal("expected telemetry enabled")
	}
}

func TestLoadInvalidKnownValueFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := strings.TrimSpace(`
[provider]
timeout_ms = -1
`)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, warnings, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Provider.TimeoutMS != 4000 {
		t.Fatalf("expected default timeout, got %d", cfg.Provider.TimeoutMS)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning for invalid timeout")
	}
}

func TestLoadUnknownKeyIgnoredWithWarning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := strings.TrimSpace(`
[provider]
unknown = "value"
`)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, warnings, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning for unknown key")
	}
}
