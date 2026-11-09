package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// resolveScriptEntries preserves function names while making script paths
// independent of the controller's working directory.
func resolveScriptEntries(m map[string]any, dir, file string, sources map[string][]string) error {
	for section, fields := range map[string][]string{
		"workflows":     {"on_start", "on_step_completed"},
		"triggers":      {"handler"},
		"desk_commands": {"handler"},
	} {
		entries, _ := m[section].(map[string]any)
		for _, name := range sortedKeys(entries) {
			entry, _ := entries[name].(map[string]any)
			for _, field := range fields {
				key := section + "." + name + "." + field
				if err := resolveScriptEntry(entry, field, dir, file, key, sources); err != nil {
					return err
				}
			}
		}
	}
	// Hooks are an appended list: their sources are not keyed by name.
	hooks, _ := m["hooks"].([]any)
	for i, h := range hooks {
		entry, _ := h.(map[string]any)
		if err := resolveScriptEntry(entry, "handler", dir, file, "hooks."+strconv.Itoa(i)+".handler", nil); err != nil {
			return err
		}
	}
	triggers, _ := m["triggers"].(map[string]any)
	for _, name := range sortedKeys(triggers) {
		trigger, _ := triggers[name].(map[string]any)
		when, _ := trigger["when"].(map[string]any)
		signals, _ := when["scan"].([]any)
		for i, s := range signals {
			entry, _ := s.(map[string]any)
			if err := resolveScriptEntry(entry, "script", dir, file, "triggers."+name+".when.scan."+strconv.Itoa(i)+".script", nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveScriptEntry(entry map[string]any, field, dir, file, key string, sources map[string][]string) error {
	raw, ok := entry[field].(string)
	if !ok || raw == "" {
		return nil
	}
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
	return nil
}
