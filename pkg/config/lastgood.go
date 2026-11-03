package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Last good config (#227). After each successful load the controller
// saves the merged config, packs already applied, so a later start whose
// live packs need a newer binary can keep serving with it.

// LastGoodName is the saved merged config in the root.
const LastGoodName = "config.last-good.yaml"

// LastGood returns the saved last good config path.
func (r Root) LastGood() string { return filepath.Join(r.Dir, LastGoodName) }

// SaveLastGood writes c (merged, no packs list) to path atomically.
func SaveLastGood(path string, c *Config) error {
	cc := *c
	cc.Packs = nil
	b, err := yaml.Marshal(&cc)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadLastGood reads a config saved by SaveLastGood. Sources are not
// kept; packs are already merged in.
func LoadLastGood(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c, err := Parse(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", LastGoodName, err)
	}
	return c, nil
}

// LoadOrLastGood loads the root's config. If only a pack fails and a
// last good config exists, it returns that config and the pack error;
// any other failure returns a nil config.
func LoadOrLastGood(r Root) (*Config, error) {
	c, err := Load(r)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, ErrPack) {
		return nil, err
	}
	lg, lerr := LoadLastGood(r.LastGood())
	if lerr != nil {
		return nil, err
	}
	return lg, err
}
