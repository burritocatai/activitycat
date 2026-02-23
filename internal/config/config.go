package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds application configuration for LLM provider selection
type Config struct {
	Provider   string `toml:"provider"`
	Model      string `toml:"model"`
	OllamaHost string `toml:"ollama_host"`
}

// LoadConfig reads configuration from ~/.config/activitycat/config.toml.
// Returns sensible defaults if the file is missing or unreadable.
func LoadConfig() Config {
	cfg := Config{
		Provider:   "claude",
		Model:      "claude-sonnet-4-5-20250929",
		OllamaHost: "http://localhost:11434",
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	configPath := filepath.Join(homeDir, ".config", "activitycat", "config.toml")
	_, err = toml.DecodeFile(configPath, &cfg)
	if err != nil {
		// File missing or invalid — use defaults
		return cfg
	}

	return cfg
}
