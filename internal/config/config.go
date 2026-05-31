package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all user-configurable settings for Twirl.
type Config struct {
	Cursor string `toml:"cursor"`
	Blink  bool   `toml:"blink"`
}

// Default returns the default configuration.
func Default() Config {
	return Config{
		Cursor: "bar",
		Blink:  true,
	}
}

// dir returns the config directory, respecting XDG_CONFIG_HOME.
func dir() string {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg != "" {
		return filepath.Join(xdg, "twirl")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	return filepath.Join(home, ".config", "twirl")
}

// Path returns the full path to the config file.
func Path() string {
	return filepath.Join(dir(), "config.toml")
}

// Load reads the config file from disk. Returns Default if
// the file is missing or unreadable.
func Load() Config {
	data, err := os.ReadFile(Path())
	if err != nil {
		return Default()
	}
	cfg := Default()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Default()
	}
	return cfg
}

// Save writes the config file to disk, creating the directory
// if needed.
func Save(cfg Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(&cfg)
}
