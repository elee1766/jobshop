package config

import (
	"context"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
)

// checkHookPrograms compiles callbacks without credentials or capability grants.
func (c *Config) checkHookPrograms(watch *WatchPaths) error {
	groups := map[string][]string{}
	arities := map[string][]int{}
	add := func(entry string, positional int) error {
		if err := addHookEntry(groups, entry); err != nil {
			return err
		}
		if entry != "" {
			arities[entry] = append(arities[entry], positional)
		}
		return nil
	}
	for _, workflow := range c.Workflows {
		if err := add(workflow.OnStart, 1); err != nil {
			return err
		}
		if err := add(workflow.OnStepCompleted, 2); err != nil {
			return err
		}
	}
	for _, trigger := range c.Triggers {
		if err := add(trigger.Handler, 2); err != nil {
			return err
		}
	}
	options := automation.Options{}
	options.Limits.Steps = c.Automation.MaxSteps
	options.Limits.OutputBytes = c.Automation.OutputBytes
	if c.Automation.Timeout != "" {
		options.Limits.Duration, _ = time.ParseDuration(c.Automation.Timeout)
	}
	for _, dir := range slices.Sorted(maps.Keys(groups)) {
		entries := groups[dir]
		slices.Sort(entries)
		entries = slices.Compact(entries)
		engine, err := automation.Load(context.Background(), dir, entries, options)
		if err != nil {
			return fmt.Errorf("%w: validate hooks in %s: %w", ErrPack, dir, err)
		}
		for _, entry := range entries {
			for _, positional := range arities[filepath.Join(dir, entry)] {
				if err := engine.CheckArity(entry, positional); err != nil {
					return fmt.Errorf("%w: validate hook %s: %w", ErrPack, filepath.Join(dir, entry), err)
				}
			}
		}
		for _, file := range engine.Files() {
			watch.addFile(file)
		}
	}
	return nil
}

func addHookEntry(groups map[string][]string, entry string) error {
	if entry == "" {
		return nil
	}
	file, function, ok := strings.Cut(entry, ":")
	if !ok || function == "" {
		return fmt.Errorf("%w: hook %q must be file.star:function", ErrPack, entry)
	}
	dir := filepath.Dir(file)
	groups[dir] = append(groups[dir], filepath.Base(file)+":"+function)
	return nil
}
