package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolveScriptEntries preserves function names while making script paths
// independent of the controller's working directory.
func resolveScriptEntries(m map[string]any, dir, file string, sources map[string][]string) error {
	for section, fields := range map[string][]string{
		"workflows": {"on_start", "on_step_completed"},
		"triggers":  {"handler"},
	} {
		entries, _ := m[section].(map[string]any)
		for _, name := range sortedKeys(entries) {
			entry, _ := entries[name].(map[string]any)
			for _, field := range fields {
				raw, ok := entry[field].(string)
				if !ok || raw == "" {
					continue
				}
				key := section + "." + name + "." + field
				path, fn, ok := strings.Cut(raw, ":")
				if !ok || path == "" || fn == "" || strings.Contains(fn, ":") {
					return fmt.Errorf("%s: %s: expected file.star:function", file, key)
				}
				if !filepath.IsAbs(path) {
					path = filepath.Join(dir, path)
				}
				path, err := filepath.Abs(path)
				if err != nil {
					return fmt.Errorf("%s: %s: %w", file, key, err)
				}
				info, err := os.Stat(path)
				if err != nil {
					return fmt.Errorf("%s: %s: %w", file, key, err)
				}
				if !info.Mode().IsRegular() {
					return fmt.Errorf("%s: %s: script must be a regular file", file, key)
				}
				entry[field] = path + ":" + fn
				if sources != nil {
					sources[key] = []string{path}
					if target, err := filepath.EvalSymlinks(path); err == nil && target != path {
						sources[key] = append(sources[key], target)
					}
				}
			}
		}
	}
	return nil
}
