package config

import (
	"fmt"
	"time"
)

// DefaultReloadDebounce is reload.debounce when unset (#533).
const DefaultReloadDebounce = 2 * time.Second

// ReloadConfig turns on watch-and-reload (#533). Off by default.
type ReloadConfig struct {
	// Watch reloads after jobshop.yaml, an include or a file in a pack
	// dir changes.
	Watch bool `yaml:"watch"`
	// Debounce waits this long after the last change (Go duration).
	// Default 2s.
	Debounce string `yaml:"debounce,omitempty"`
}

// DebounceDuration is Debounce parsed, or the default.
func (r ReloadConfig) DebounceDuration() time.Duration {
	d, err := time.ParseDuration(r.Debounce)
	if err != nil || d <= 0 {
		return DefaultReloadDebounce
	}
	return d
}

func (r ReloadConfig) check() error {
	if r.Debounce == "" {
		return nil
	}
	if d, err := time.ParseDuration(r.Debounce); err != nil || d <= 0 {
		return fmt.Errorf("%s: reload.debounce %q must be a positive duration", FileName, r.Debounce)
	}
	return nil
}
