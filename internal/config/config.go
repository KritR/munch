package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Safety struct {
		Level string
	}
	Provider struct {
		BaseURL    string
		Model      string
		APIKeyEnv  string
		TimeoutMS  int
		MaxRetries int
	}
	UI struct {
		VisibleSuggestionCount int
	}
	Shell struct {
		Zsh struct {
			Enabled bool
		}
		Fish struct {
			Enabled bool
		}
	}
	Telemetry struct {
		Enabled bool
	}
}

type Warning string

func (c Config) HasProviderConfig() bool {
	return c.Provider.BaseURL != "" && c.Provider.Model != "" && c.Provider.APIKeyEnv != ""
}

func Defaults() Config {
	var cfg Config
	cfg.Safety.Level = "balanced"
	cfg.Provider.TimeoutMS = 4000
	cfg.Provider.MaxRetries = 1
	cfg.UI.VisibleSuggestionCount = 5
	cfg.Shell.Zsh.Enabled = true
	cfg.Shell.Fish.Enabled = true
	return cfg
}

func DefaultPath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "munch", "config.toml"), nil
}

func Load(path string) (Config, []Warning, error) {
	cfg := Defaults()
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return cfg, nil, err
		}
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil, nil
		}
		return cfg, nil, err
	}
	defer file.Close()

	warnings, err := parseInto(file, &cfg)
	if err != nil {
		return Defaults(), warnings, err
	}
	return cfg, warnings, nil
}

func parseInto(f *os.File, cfg *Config) ([]Warning, error) {
	var warnings []Warning
	section := ""
	scanner := bufio.NewScanner(f)
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}

		key, rawValue, ok := strings.Cut(line, "=")
		if !ok {
			warnings = append(warnings, Warning(fmt.Sprintf("line %d ignored: not a key/value pair", lineNo)))
			continue
		}

		key = strings.TrimSpace(key)
		rawValue = strings.TrimSpace(rawValue)
		if idx := strings.Index(rawValue, "#"); idx >= 0 {
			rawValue = strings.TrimSpace(rawValue[:idx])
		}

		fullKey := section + "." + key
		if section == "" {
			fullKey = key
		}

		if err := applyValue(cfg, fullKey, rawValue, &warnings); err != nil {
			warnings = append(warnings, Warning(fmt.Sprintf("%s: using default (%v)", fullKey, err)))
		}
	}

	if err := scanner.Err(); err != nil {
		return warnings, err
	}
	return warnings, nil
}

func applyValue(cfg *Config, key string, raw string, warnings *[]Warning) error {
	switch key {
	case "safety.level":
		v, err := parseString(raw)
		if err != nil {
			return err
		}
		switch v {
		case "low", "balanced", "strict":
			cfg.Safety.Level = v
			return nil
		default:
			return fmt.Errorf("unsupported safety level %q", v)
		}
	case "provider.base_url":
		v, err := parseString(raw)
		if err != nil {
			return err
		}
		cfg.Provider.BaseURL = v
		return nil
	case "provider.model":
		v, err := parseString(raw)
		if err != nil {
			return err
		}
		cfg.Provider.Model = v
		return nil
	case "provider.api_key_env":
		v, err := parseString(raw)
		if err != nil {
			return err
		}
		cfg.Provider.APIKeyEnv = v
		return nil
	case "provider.timeout_ms":
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 {
			return fmt.Errorf("timeout must be > 0")
		}
		cfg.Provider.TimeoutMS = v
		return nil
	case "provider.max_retries":
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			return fmt.Errorf("max_retries must be >= 0")
		}
		cfg.Provider.MaxRetries = v
		return nil
	case "ui.visible_suggestion_count":
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 {
			return fmt.Errorf("visible_suggestion_count must be > 0")
		}
		cfg.UI.VisibleSuggestionCount = v
		return nil
	case "shell.zsh.enabled":
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		cfg.Shell.Zsh.Enabled = v
		return nil
	case "shell.fish.enabled":
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		cfg.Shell.Fish.Enabled = v
		return nil
	case "telemetry.enabled":
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		cfg.Telemetry.Enabled = v
		return nil
	default:
		*warnings = append(*warnings, Warning(fmt.Sprintf("unknown key ignored: %s", key)))
		return nil
	}
}

func parseString(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("empty string")
	}
	return strconv.Unquote(raw)
}
