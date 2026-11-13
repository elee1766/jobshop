package pack

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/gfx-labs/jobshop/pkg/pack/cue"
)

// CUEFileName is the native pack entry point. Its package, not only
// this file, is evaluated so CUE imports and sibling files retain their meaning.
const CUEFileName = "pack.cue"

func (ld loader) loadCUE(file string) ([]layer, error) {
	result, err := cue.Load(file, FileDir(file))
	if err != nil {
		return nil, err
	}
	exported := result.Data
	var layers []layer
	var summaries []Instance
	if raw, exists := exported["packs"]; exists {
		packs, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: packs must be a mapping of named CUE instances", file)
		}
		// Every entry's prefix first: pools_from (JOB-207) names another
		// entry and takes its prefix for the pools.
		prefixes := map[string]string{}
		for _, name := range slices.Sorted(maps.Keys(packs)) {
			if !NamePattern.MatchString(name) {
				return nil, fmt.Errorf("%s: packs.%s: invalid instance name", file, name)
			}
			instance, ok := packs[name].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s: packs.%s must be a mapping", file, name)
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
			prefixes[name] = prefix
		}
		owners := map[string]int{} // entry name -> index in summaries
		type entry struct {
			name, path, prefix, poolsFrom string
			l                             layer
		}
		var entries []entry
		for _, name := range slices.Sorted(maps.Keys(packs)) {
			instance := packs[name].(map[string]any)
			path, ok := instance["path"].(string)
			if !ok || path == "" {
				return nil, fmt.Errorf("%s: packs.%s.path: asset directory is required", file, name)
			}
			if !filepath.IsAbs(path) {
				path = filepath.Join(FileDir(file), path)
			}
			if info, err := os.Stat(path); err != nil {
				return nil, fmt.Errorf("%s: packs.%s.path: %w", file, name, err)
			} else if !info.IsDir() {
				return nil, fmt.Errorf("%s: packs.%s.path must name a directory", file, name)
			}
			vars := map[string]any{}
			if raw, exists := instance["vars"]; exists {
				var ok bool
				vars, ok = raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("%s: packs.%s.vars must be a mapping", file, name)
				}
			}
			poolsFrom := ""
			if raw, exists := instance["pools_from"]; exists {
				var ok bool
				if poolsFrom, ok = raw.(string); !ok {
					return nil, fmt.Errorf("%s: packs.%s.pools_from must be a string", file, name)
				}
				if _, known := prefixes[poolsFrom]; poolsFrom != "" && (!known || poolsFrom == name) {
					return nil, fmt.Errorf("%s: packs.%s.pools_from %q must name another packs entry", file, name, poolsFrom)
				}
			}
			vars[PoolPrefixVar] = prefixes[name]
			if poolsFrom != "" {
				vars[PoolPrefixVar] = prefixes[poolsFrom]
			}
			delete(instance, "path")
			delete(instance, "prefix")
			delete(instance, "vars")
			delete(instance, "pools_from")
			l, err := ld.cueLayer(file, path, instance, vars, cueEntrySources(result, instance, "packs", name), false)
			if err != nil {
				return nil, fmt.Errorf("%s: packs.%s: %w", file, name, err)
			}
			entries = append(entries, entry{name, path, prefixes[name], poolsFrom, l})
		}
		// Entries that take their pools from another (pools_from) are renamed
		// after it: every pool name it declares maps to its pool.
		summaries = make([]Instance, len(entries))
		for i, e := range entries {
			owners[e.name] = i
		}
		for _, sharing := range []bool{false, true} {
			for i, e := range entries {
				if (e.poolsFrom != "") != sharing {
					continue
				}
				var pools map[string]string
				if sharing {
					pools = sharedPools(summaries[owners[e.poolsFrom]], prefixes[e.poolsFrom])
				}
				child := []layer{e.l}
				if e.prefix != "" || pools != nil {
					prefixLayers(child, e.prefix, pools)
					child[0].textSources = renameSources(e.l.textSources, e.prefix, pools)
					child[0].entrySources = renameSources(e.l.entrySources, e.prefix, pools)
				}
				entries[i].l = child[0]
				summaries[i] = summarize(e.path, e.prefix, child)
				if sharing {
					summaries[i].PoolsFrom, summaries[i].PoolPrefix = e.poolsFrom, prefixes[e.poolsFrom]
				}
			}
		}
		if err := checkPoolsFrom(summaries, owners); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		// A sharing entry uses every pool of its owner, declared or not:
		// its hooks resolve pool names (start_step(pool = "triage")) and
		// its runs' steps are checked against this list.
		for i := range summaries {
			if from := summaries[i].PoolsFrom; from != "" {
				summaries[i].Pools = slices.Clone(summaries[owners[from]].Pools)
			}
		}
		for _, e := range entries {
			layers = append(layers, e.l)
		}
		delete(exported, "packs")
	}
	root, err := ld.cueLayer(file, FileDir(file), exported, nil, cueEntrySources(result, exported), true)
	if err != nil {
		return nil, err
	}
	root.watch = WatchPaths{Files: result.Files, CUEDirs: result.Dirs}
	root.instances = summaries
	return append(layers, root), nil
}

func (ld loader) cueLayer(file, dir string, m, vars map[string]any, sources map[string][]string, root bool) (layer, error) {
	if err := setPackVars(m, vars, file); err != nil {
		return layer{}, err
	}
	textSources := map[string][]string{}
	var textVars map[string]any
	if vars != nil {
		textVars = ld.lateVars()
		maps.Copy(textVars, vars)
	}
	if err := ld.resolveTextFiles(m, textFiles{dir: dir, file: file, vars: textVars, sources: textSources}); err != nil {
		return layer{}, err
	}
	if err := resolveScriptEntries(m, dir, file, textSources); err != nil {
		return layer{}, err
	}
	if err := ld.checkLayer(m, root); err != nil {
		return layer{}, fmt.Errorf("%s: %w", file, err)
	}
	return layer{
		file: file, m: m, vars: vars, textSources: textSources,
		entrySources: sources,
	}, nil
}

func cueEntrySources(result *cue.Result, m map[string]any, path ...string) map[string][]string {
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
