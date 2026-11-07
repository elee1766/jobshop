// Package packcue evaluates local CUE packages without accessing module registries.
package packcue

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/mod/modfile"
	"cuelang.org/go/mod/module"
)

// Result holds concrete exported data and the local inputs used to evaluate it.
// It does not retain the CUE evaluator. Callers may modify Data independently
// of the recorded sources.
type Result struct {
	Data map[string]any
	// Files includes module manifests, even an absent local-module.cue, so
	// callers can detect newly added replacements.
	Files []string
	// Dirs contains package directories whose direct CUE siblings affect loading.
	Dirs []string

	sources source
}

// Sources returns the files contributing to a value and its descendants.
// Path elements are field names or decimal list indices, not dotted paths.
// An unknown path returns nil.
func (r *Result) Sources(path ...string) []string {
	s := r.sources
	for _, part := range path {
		var ok bool
		s, ok = s.children[part]
		if !ok {
			return nil
		}
	}
	return slices.Clone(s.files)
}

// Load evaluates the package named by file, including siblings and native
// imports, from dir. The caller chooses dir to apply its symlink policy.
// Only same-module packages and checked-in local replacements are available.
func Load(file, dir string) (*Result, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	parsed, err := parser.ParseFile(file, b, parser.PackageClauseOnly)
	if err != nil {
		return nil, diagnostic(file, err)
	}
	pkg := parsed.PackageName()
	if pkg == "" {
		pkg = "_"
	}
	instances := load.Instances([]string{"."}, &load.Config{
		Dir: dir, Package: pkg, Registry: offlineRegistry{}, Env: []string{},
	})
	if len(instances) != 1 {
		return nil, fmt.Errorf("%s: expected one CUE package, got %d", file, len(instances))
	}
	inst := instances[0]
	if inst.Err != nil {
		return nil, diagnostic(file, inst.Err)
	}
	value := cuecontext.New().BuildInstance(inst)
	if err := value.Validate(cue.Concrete(true)); err != nil {
		return nil, diagnostic(file, err)
	}
	var data map[string]any
	if err := value.Decode(&data); err != nil {
		return nil, diagnostic(file, err)
	}
	files, dirs := dependencies(inst)
	return &Result{Data: data, Files: files, Dirs: dirs, sources: valueSourceTree(value, data, file)}, nil
}

// offlineRegistry keeps loading independent of network, credentials and
// machine-local module caches. Same-module packages and local replacements work.
type offlineRegistry struct{}

func (offlineRegistry) ModFile(_ context.Context, version module.Version) (*modfile.File, error) {
	return nil, fmt.Errorf("external CUE module %s is unavailable offline; use a checked-in local package or module replacement", version)
}

func (offlineRegistry) Fetch(_ context.Context, version module.Version) (module.SourceLoc, error) {
	return module.SourceLoc{}, fmt.Errorf("external CUE module %s is unavailable offline; use a checked-in local package or module replacement", version)
}

func (offlineRegistry) ModuleVersions(_ context.Context, path string) ([]string, error) {
	return nil, fmt.Errorf("external CUE module %s is unavailable offline; use a checked-in local package or module replacement", path)
}

func diagnostic(file string, err error) error {
	return fmt.Errorf("%s: CUE: %s", file, cueerrors.Details(err, nil))
}

func dependencies(root *build.Instance) (files, dirs []string) {
	seen := map[*build.Instance]bool{}
	add := func(paths *[]string, path string) {
		if path = filepath.Clean(path); !slices.Contains(*paths, path) {
			*paths = append(*paths, path)
		}
	}
	var visit func(*build.Instance)
	visit = func(inst *build.Instance) {
		if seen[inst] {
			return
		}
		seen[inst] = true
		for _, file := range inst.BuildFiles {
			add(&files, file.Filename)
			add(&dirs, filepath.Dir(file.Filename))
		}
		if inst.Module != "" && inst.Root != "" {
			for _, name := range []string{"module.cue", "local-module.cue"} {
				add(&files, filepath.Join(inst.Root, "cue.mod", name))
			}
		}
		for _, imported := range inst.Imports {
			visit(imported)
		}
	}
	visit(root)
	return files, dirs
}

type source struct {
	files    []string
	children map[string]source
}

func valueSourceTree(value cue.Value, data any, fallback string) source {
	s := source{files: valueSources(value, fallback)}
	switch data := data.(type) {
	case map[string]any:
		s.children = make(map[string]source, len(data))
		for name, child := range data {
			v := value.LookupPath(cue.MakePath(cue.Str(name)))
			s.children[name] = valueSourceTree(v, child, fallback)
		}
	case []any:
		s.children = make(map[string]source, len(data))
		for i, child := range data {
			v := value.LookupPath(cue.MakePath(cue.Index(i)))
			s.children[strconv.Itoa(i)] = valueSourceTree(v, child, fallback)
		}
	}
	return s
}

func valueSources(value cue.Value, fallback string) []string {
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
