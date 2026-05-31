package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds all user-configurable settings for Twirl.
type Config struct {
	Cursor string `toml:"cursor"`
	Blink  bool   `toml:"blink"`
	LLM    LLM    `toml:"llm"`
}

// LLM holds the LLM provider configuration.
type LLM struct {
	Provider string `toml:"provider"`
	APIKey   string `toml:"api_key"`
	BaseURL  string `toml:"base_url"`
	Model    string `toml:"model"`
}

// IsZero returns true if no LLM fields are set.
func (l LLM) IsZero() bool {
	return l.Provider == "" && l.APIKey == "" &&
		l.BaseURL == "" && l.Model == ""
}

// ResolveAPIKey expands env var references (e.g. "$OPENAI_API_KEY")
// in the APIKey field. Returns the expanded value or an error if the
// referenced env var is unset.
func (l LLM) ResolveAPIKey() (string, error) {
	v := l.APIKey
	if strings.HasPrefix(v, "$") {
		env := os.Getenv(v[1:])
		if env == "" {
			return "", os.ErrNotExist
		}
		return env, nil
	}
	return v, nil
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
