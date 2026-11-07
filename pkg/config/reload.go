package config

import (
	"fmt"
	"path/filepath"
	"slices"
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

// WatchPaths are what a config watcher follows: the config file and its
// includes as files, and every resolved pack dir (extends included)
// recursively. Recorded by LoadFile, so a pack added to the config is
// watched after the reload that loads it.
type WatchPaths struct {
	Files []string
	Dirs  []string
	// CUEDirs watch direct CUE siblings, without walking entire checkouts.
	CUEDirs []string
}

func (w *WatchPaths) addFile(p string) {
	if p = filepath.Clean(p); !slices.Contains(w.Files, p) {
		w.Files = append(w.Files, p)
	}
}

func (w *WatchPaths) addDir(p string) {
	if p = filepath.Clean(p); !slices.Contains(w.Dirs, p) {
		w.Dirs = append(w.Dirs, p)
	}
}

func (w *WatchPaths) addCUEDir(p string) {
	if p = filepath.Clean(p); !slices.Contains(w.CUEDirs, p) {
		w.CUEDirs = append(w.CUEDirs, p)
	}
}

// IgnoredChange reports editor temp and swap files, which never trigger
// a reload: vim's .x.swp and 4913 probe, emacs' x~ and .#x lock.
func IgnoredChange(path string) bool {
	b := filepath.Base(path)
	return b == "4913" || filepath.Ext(b) == ".swp" || filepath.Ext(b) == ".swx" ||
		len(b) > 0 && b[len(b)-1] == '~' || len(b) > 1 && b[:2] == ".#"
}
