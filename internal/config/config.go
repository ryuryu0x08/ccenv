// Package config loads, saves, and mutates the ccenv TOML config.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"ccenv/internal/profile"
	"github.com/pelletier/go-toml/v2"
)

// Config is the top-level TOML document.
type Config struct {
	Profiles map[string]profile.Profile `toml:"profiles"`
}

// DefaultPath returns ~/.ccenv/config.toml.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".ccenv", "config.toml"), nil
}

// Load reads config from path. A missing file yields an empty config (not an error).
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{Profiles: map[string]profile.Profile{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if c.Profiles == nil {
		c.Profiles = map[string]profile.Profile{}
	}
	return &c, nil
}

// Save writes config to path, creating parent dirs.
func Save(path string, c *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir config dir: %w", err)
	}
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

func (c *Config) Get(name string) (profile.Profile, bool) {
	p, ok := c.Profiles[name]
	return p, ok
}

func (c *Config) Has(name string) bool {
	_, ok := c.Profiles[name]
	return ok
}

func (c *Config) Set(name string, p profile.Profile) {
	if c.Profiles == nil {
		c.Profiles = map[string]profile.Profile{}
	}
	c.Profiles[name] = p
}

// Remove deletes a profile, erroring if absent.
func (c *Config) Remove(name string) error {
	if !c.Has(name) {
		return fmt.Errorf("profile %q does not exist", name)
	}
	delete(c.Profiles, name)
	return nil
}

// Names returns profile names sorted alphabetically.
func (c *Config) Names() []string {
	names := make([]string, 0, len(c.Profiles))
	for n := range c.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
