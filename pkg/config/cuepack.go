package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/goccy/go-yaml"

	"github.com/elee1766/jobshop/pkg/packcue"
)

// CUEPackFileName is the native pack entry point. Its package, not only
// this file, is evaluated so CUE imports and sibling files retain their meaning.
const CUEPackFileName = "pack.cue"

func loadCUEPack(file string) ([]packLayer, error) {
	result, err := packcue.Load(file, fileDir(file))
	if err != nil {
		return nil, err
	}
	exported := result.Data
	var layers []packLayer
	var summaries []PackInstance
	if raw, exists := exported["packs"]; exists {
		packs, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: packs must be a mapping of named CUE instances", file)
		}
		for _, name := range slices.Sorted(maps.Keys(packs)) {
			if !namePattern.MatchString(name) {
				return nil, fmt.Errorf("%s: packs.%s: invalid instance name", file, name)
			}
			instance, ok := packs[name].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s: packs.%s must be a mapping", file, name)
			}
			path, ok := instance["path"].(string)
			if !ok || path == "" {
				return nil, fmt.Errorf("%s: packs.%s.path: asset directory is required", file, name)
			}
			if !filepath.IsAbs(path) {
				path = filepath.Join(fileDir(file), path)
			}
			if info, err := os.Stat(path); err != nil {
				return nil, fmt.Errorf("%s: packs.%s.path: %w", file, name, err)
			} else if !info.IsDir() {
				return nil, fmt.Errorf("%s: packs.%s.path must name a directory", file, name)
			}
			prefix := name + "-"
			if raw, exists := instance["prefix"]; exists {
				var ok bool
				prefix, ok = raw.(string)
				if !ok {
					return nil, fmt.Errorf("%s: packs.%s.prefix must be a string", file, name)
				}
			}
			if prefix != "" && !prefixPattern.MatchString(prefix) {
				return nil, fmt.Errorf("%s: packs.%s.prefix %q must match %s", file, name, prefix, prefixPattern)
			}
			vars := map[string]any{}
			if raw, exists := instance["vars"]; exists {
				var ok bool
				vars, ok = raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("%s: packs.%s.vars must be a mapping", file, name)
				}
			}
			delete(instance, "path")
			delete(instance, "prefix")
			delete(instance, "vars")
			layer, err := cuePackLayer(file, path, instance, vars, cueEntrySources(result, instance, "packs", name), false)
			if err != nil {
				return nil, fmt.Errorf("%s: packs.%s: %w", file, name, err)
			}
			child := []packLayer{layer}
			if prefix != "" {
				prefixLayers(child, prefix)
				child[0].textSources = prefixSources(layer.textSources, prefix)
				child[0].entrySources = prefixSources(layer.entrySources, prefix)
			}
			summaries = append(summaries, packInstance(path, prefix, child))
			layers = append(layers, child...)
		}
		delete(exported, "packs")
	}
	root, err := cuePackLayer(file, fileDir(file), exported, nil, cueEntrySources(result, exported), true)
	if err != nil {
		return nil, err
	}
	root.watch = WatchPaths{Files: result.Files, CUEDirs: result.Dirs}
	root.instances = summaries
	return append(layers, root), nil
}

func cuePackLayer(file, dir string, m, vars map[string]any, sources map[string][]string, root bool) (packLayer, error) {
	if err := setPackVars(m, vars, file); err != nil {
		return packLayer{}, err
	}
	textSources := map[string][]string{}
	var textVars map[string]any
	if vars != nil {
		textVars = map[string]any{CheckOutputVar: placeholder(CheckOutputVar)}
		event := map[string]any{}
		for _, field := range EventFields {
			event[field] = placeholder(EventVar + "." + field)
		}
		textVars[EventVar] = event
		for _, name := range ScanVars {
			textVars[name] = placeholder(name)
		}
		maps.Copy(textVars, vars)
	}
	if err := resolveTextFiles(m, textFiles{dir: dir, file: file, vars: textVars, sources: textSources}); err != nil {
		return packLayer{}, err
	}
	if err := resolveScriptEntries(m, dir, file, textSources); err != nil {
		return packLayer{}, err
	}
	b, err := yaml.Marshal(m)
	if err != nil {
		return packLayer{}, fmt.Errorf("%s: %w", file, err)
	}
	var schema any = &packFile{}
	if root {
		schema = &rootPackFile{}
	}
	if err := yaml.UnmarshalWithOptions(b, schema, yaml.Strict()); err != nil {
		return packLayer{}, fmt.Errorf("%s: %w", file, err)
	}
	return packLayer{
		file: file, m: m, vars: vars, textSources: textSources,
		entrySources: sources,
	}, nil
}

func cueEntrySources(result *packcue.Result, m map[string]any, path ...string) map[string][]string {
	sources := map[string][]string{}
	for _, section := range mergedSections {
		entries, _ := m[section].(map[string]any)
		for name := range entries {
			sources[section+"."+name] = result.Sources(append(path, section, name)...)
		}
	}
	if _, ok := m["default_project"]; ok {
		sources["default_project."] = result.Sources(append(path, "default_project")...)
	}
	rules, _ := m["prompt_rules"].([]any)
	for i, raw := range rules {
		rule, _ := raw.(map[string]any)
		if name, _ := rule["name"].(string); name != "" {
			sources["prompt_rules."+name] = result.Sources(append(path, "prompt_rules", strconv.Itoa(i))...)
		}
	}
	return sources
}
