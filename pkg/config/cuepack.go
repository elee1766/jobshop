package config

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/mod/modfile"
	"cuelang.org/go/mod/module"
	"github.com/goccy/go-yaml"
)

// CUEPackFileName is the native pack entry point. Its package, not only
// this file, is evaluated so CUE imports and sibling files retain their meaning.
const CUEPackFileName = "pack.cue"

// cueOfflineRegistry keeps loading independent of network, credentials and
// machine-local module caches. Same-module packages and local replacements work.
type cueOfflineRegistry struct{}

func (cueOfflineRegistry) ModFile(_ context.Context, version module.Version) (*modfile.File, error) {
	return nil, fmt.Errorf("external CUE module %s is unavailable offline; use a checked-in local package or module replacement", version)
}

func (cueOfflineRegistry) Fetch(_ context.Context, version module.Version) (module.SourceLoc, error) {
	return module.SourceLoc{}, fmt.Errorf("external CUE module %s is unavailable offline; use a checked-in local package or module replacement", version)
}

func (cueOfflineRegistry) ModuleVersions(_ context.Context, path string) ([]string, error) {
	return nil, fmt.Errorf("external CUE module %s is unavailable offline; use a checked-in local package or module replacement", path)
}

func loadCUEPack(file string) ([]packLayer, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	parsed, err := parser.ParseFile(file, b, parser.PackageClauseOnly)
	if err != nil {
		return nil, cuePackError(file, err)
	}
	pkg := parsed.PackageName()
	if pkg == "" {
		pkg = "_"
	}
	instances := load.Instances([]string{"."}, &load.Config{
		Dir: fileDir(file), Package: pkg, Registry: cueOfflineRegistry{}, Env: []string{},
	})
	if len(instances) != 1 {
		return nil, fmt.Errorf("%s: expected one CUE package, got %d", file, len(instances))
	}
	inst := instances[0]
	if inst.Err != nil {
		return nil, cuePackError(file, inst.Err)
	}
	value := cuecontext.New().BuildInstance(inst)
	if err := value.Validate(cue.Concrete(true)); err != nil {
		return nil, cuePackError(file, err)
	}
	var exported map[string]any
	if err := value.Decode(&exported); err != nil {
		return nil, cuePackError(file, err)
	}
	watch := cuePackageWatch(inst)
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
			v := value.LookupPath(cue.MakePath(cue.Str("packs"), cue.Str(name)))
			layer, err := cuePackLayer(file, path, instance, vars, v, false)
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
	root, err := cuePackLayer(file, fileDir(file), exported, nil, value, true)
	if err != nil {
		return nil, err
	}
	root.watch, root.instances = watch, summaries
	return append(layers, root), nil
}

func cuePackLayer(file, dir string, m, vars map[string]any, value cue.Value, root bool) (packLayer, error) {
	if err := setPackVars(m, vars, file); err != nil {
		return packLayer{}, err
	}
	textSources := map[string][]string{}
	if err := resolveTextFiles(m, textFiles{dir: dir, file: file, sources: textSources}); err != nil {
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
		entrySources: cueEntrySources(value, m, file),
	}, nil
}

func cuePackError(file string, err error) error {
	return fmt.Errorf("%s: CUE: %s", file, cueerrors.Details(err, nil))
}

func cuePackageWatch(root *build.Instance) WatchPaths {
	var watch WatchPaths
	seen := map[*build.Instance]bool{}
	var visit func(*build.Instance)
	visit = func(inst *build.Instance) {
		if seen[inst] {
			return
		}
		seen[inst] = true
		for _, file := range inst.BuildFiles {
			watch.addFile(file.Filename)
			watch.addCUEDir(filepath.Dir(file.Filename))
		}
		if inst.Module != "" && inst.Root != "" {
			for _, name := range []string{"module.cue", "local-module.cue"} {
				watch.addFile(filepath.Join(inst.Root, "cue.mod", name))
			}
		}
		for _, imported := range inst.Imports {
			visit(imported)
		}
	}
	visit(root)
	return watch
}

func cueEntrySources(value cue.Value, m map[string]any, fallback string) map[string][]string {
	sources := map[string][]string{}
	for _, section := range mergedSections {
		entries, _ := m[section].(map[string]any)
		for name := range entries {
			v := value.LookupPath(cue.MakePath(cue.Str(section), cue.Str(name)))
			sources[section+"."+name] = cueValueSources(v, fallback)
		}
	}
	if _, ok := m["default_project"]; ok {
		sources["default_project."] = cueValueSources(value.LookupPath(cue.MakePath(cue.Str("default_project"))), fallback)
	}
	rules, _ := m["prompt_rules"].([]any)
	for i, raw := range rules {
		rule, _ := raw.(map[string]any)
		if name, _ := rule["name"].(string); name != "" {
			v := value.LookupPath(cue.MakePath(cue.Str("prompt_rules"), cue.Index(i)))
			sources["prompt_rules."+name] = cueValueSources(v, fallback)
		}
	}
	return sources
}

func cueValueSources(value cue.Value, fallback string) []string {
	files := map[string]bool{}
	var visit func(cue.Value, int)
	visit = func(v cue.Value, depth int) {
		if depth > 32 {
			return
		}
		if p := v.Pos().Filename(); filepath.IsAbs(p) {
			files[p] = true
		}
		if op, args := v.Expr(); op == cue.AndOp {
			for _, arg := range args {
				visit(arg, depth+1)
			}
		}
		v.Walk(func(child cue.Value) bool {
			if p := child.Pos().Filename(); filepath.IsAbs(p) {
				files[p] = true
			}
			return true
		}, nil)
	}
	visit(value, 0)
	if len(files) == 0 {
		files[fallback] = true
	}
	return slices.Sorted(maps.Keys(files))
}
