package config

import (
	"os"
	"path/filepath"
)

// Dir returns the config directory (run from backend/ or repo root).
func Dir() string {
	if _, err := os.Stat("config"); err == nil {
		return "config"
	}
	return filepath.Join("backend", "config")
}
