package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	dirName  = ".nautikube"
	fileName = "config.yaml"
)

// Config holds user preferences persisted to ~/.nautikube/config.yaml.
type Config struct {
	Language         string `yaml:"language,omitempty"`          // en | pt
	Severity         string `yaml:"severity,omitempty"`          // critical | high | medium | low | info
	Output           string `yaml:"output,omitempty"`            // table | yaml | csv | json
	HistoryRetention int    `yaml:"history_retention,omitempty"` // days; 0=disabled, -1=unlimited, default 90
}

// Defaults returns a Config with the built-in default values.
func Defaults() Config {
	return Config{
		Language:         "en",
		Severity:         "",
		Output:           "table",
		HistoryRetention: 90,
	}
}

// Path returns the absolute path to the config file.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, dirName, fileName), nil
}

// Load reads the config file and merges it over defaults.
// Returns defaults silently if the file is missing or malformed.
func Load() Config {
	cfg := Defaults()

	p, err := Path()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return cfg
	}

	var file Config
	if err := yaml.Unmarshal(data, &file); err != nil {
		return cfg
	}

	if file.Language != "" {
		cfg.Language = strings.ToLower(file.Language)
	}
	if file.Severity != "" {
		cfg.Severity = strings.ToLower(file.Severity)
	}
	if file.Output != "" {
		cfg.Output = strings.ToLower(file.Output)
	}
	if file.HistoryRetention != 0 {
		cfg.HistoryRetention = file.HistoryRetention
	}

	return cfg
}

// Save writes the config to ~/.nautikube/config.yaml, creating the directory if needed.
func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(p, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// HistoryDir returns the absolute path to the history directory.
func HistoryDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, dirName, "history"), nil
}

// RulesDir returns the absolute path to the custom rules directory.
func RulesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, dirName, "rules"), nil
}
