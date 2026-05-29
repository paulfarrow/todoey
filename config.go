package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type config struct {
	APIToken        string `json:"api_token"`
	AutoRefresh     bool   `json:"auto_refresh"`
	RefreshInterval int    `json:"refresh_interval_seconds"`
}

func loadConfig() config {
	var cfg config

	path := configPath()
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cfg)
	} else if os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(path), 0700)
		if data, err := json.MarshalIndent(cfg, "", "  "); err == nil {
			_ = os.WriteFile(path, data, 0600)
		}
	}

	// Environment variable takes precedence
	if token := os.Getenv("TODOIST_API_TOKEN"); token != "" {
		cfg.APIToken = token
	}

	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = 60
	}

	return cfg
}

func configPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "todoist-tui", "config.json")
}
