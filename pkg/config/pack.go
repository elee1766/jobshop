package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/goccy/go-yaml"
)

// PackFileName is the file in a pack directory that declares its pools,
// agents and triggers.
const PackFileName = "pack.yaml"

// PackRef includes a pack directory from jobshop.yaml.
type PackRef struct {
	// Path is the pack directory. Relative paths resolve against the
	// directory holding jobshop.yaml (the Jobshop root).
	Path string `yaml:"path"`
	// Vars fill ${name} references in the pack's string values.
	Vars map[string]any `yaml:"vars"`
	// Prefix renames everything the pack and its bases declare (see
	// prefixLayers), so one pack can be included once per repo. Only
	// in jobshop.yaml's packs list; bases inherit it.
	Prefix string `yaml:"prefix,omitempty"`
}

// packFile is the schema of pack.yaml: the mergeable jobshop.yaml sections,
// plus defaults for the pack's vars and the base packs it extends.
type packFile struct {
	Agents    map[string]Agent    `yaml:"agents"`
	Pools     map[string]Pool     `yaml:"pools"`
	Triggers  map[string]Trigger  `yaml:"triggers"`
	Workflows map[string]Workflow `yaml:"workflows"`
}

// packHead is the part of pack.yaml read before substitution.
type packHead struct {
	// Defaults are var values used when the including file does not set
	// them.
	Defaults map[string]any `yaml:"defaults"`
	// Extends lists base packs merged before this one, like packs: in
	// jobshop.yaml. Their vars may reference this pack's vars.
	Extends []PackRef `yaml:"extends"`
}

// packLayer is one pack file's contribution after substitution.
type packLayer struct {
	file string
	m    map[string]any
}

// maxPackDepth bounds extends chains so a cycle fails instead of looping.
const maxPackDepth = 8

// mergedSections are the top-level keys packs contribute to. Their
// entries merge per key, and inside an entry per field.
var mergedSections = []string{"agents", "pools", "triggers", "workflows"}

// LoadFile reads a config file, merges the packs it lists, and validates
// the result. Packs merge in list order, then the file itself.
func LoadFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseFile(path, b)
}

func parseFile(path string, b []byte) (*Config, error) {
	name := filepath.Base(path)
	dir := filepath.Dir(path)
	top, err := decodeMap(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	// Validate the packs list on its own so its errors are precise.
	var head struct {
		Packs []PackRef `yaml:"packs"`
	}
	raw, err := yaml.Marshal(map[string]any{"packs": top["packs"]})
	if err != nil {
		return nil, fmt.Errorf("%s: packs: %w", name, err)
	}
	if err := yaml.UnmarshalWithOptions(raw, &head, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("%s: packs: %w", name, err)
	}
	merged := map[string]any{}
	sources := map[string][]string{}
	for i, ref := range head.Packs {
		if ref.Path == "" {
			return nil, fmt.Errorf("%s: packs[%d]: path is required", name, i)
		}
		p := ref.Path
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		if ref.Prefix != "" && !prefixPattern.MatchString(ref.Prefix) {
			return nil, fmt.Errorf("%s: packs[%d]: prefix %q must match %s", name, i, ref.Prefix, prefixPattern)
		}
		layers, err := loadPack(p, ref.Vars, ref.Prefix, 0)
		if err != nil {
			return nil, err
		}
		if ref.Prefix != "" {
			prefixLayers(layers, ref.Prefix)
		}
		for _, l := range layers {
			mergeInto(merged, l.m, l.file, sources)
		}
	}
	delete(top, "packs")
	if err := setPackVars(top, nil, name); err != nil {
		return nil, err
	}
	if err := resolvePromptFiles(top, dir, name, nil); err != nil {
		return nil, err
	}
	mergeInto(merged, top, path, sources)
	out, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	c, err := parse(out, sources)
	if err != nil {
		return nil, err
	}
	c.Packs = head.Packs
	return c, nil
}

// loadPack reads dir/pack.yaml, substitutes vars, resolves prompt_file and
// checks the schema. It returns the extended packs' layers first, then its
// own. Errors name the pack file.
func loadPack(dir string, given map[string]any, prefix string, depth int) ([]packLayer, error) {
	file := filepath.Join(dir, PackFileName)
	if depth > maxPackDepth {
		return nil, fmt.Errorf("%s: extends nested more than %d deep (a cycle?)", file, maxPackDepth)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("pack %s: %w", dir, err)
	}
	m, err := decodeMap(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	var head packHead
	raw, err := yaml.Marshal(map[string]any{"defaults": m["defaults"], "extends": m["extends"]})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if err := yaml.UnmarshalWithOptions(raw, &head, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	delete(m, "defaults")
	delete(m, "extends")
	// Given vars win over defaults.
	vars := map[string]any{}
	for k, v := range head.Defaults {
		vars[k] = v
	}
	for k, v := range given {
		vars[k] = v
	}
	vars[PackPrefixVar] = prefix
	var layers []packLayer
	for i, ref := range head.Extends {
		if ref.Path == "" {
			return nil, fmt.Errorf("%s: extends[%d]: path is required", file, i)
		}
		if ref.Prefix != "" {
			return nil, fmt.Errorf("%s: extends[%d]: prefix is only allowed in jobshop.yaml packs", file, i)
		}
		p := ref.Path
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		// The base's vars are templates over this pack's vars.
		bv, err := substitute(cloneValue(ref.Vars), vars, fmt.Sprintf("extends[%d].vars", i))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		bvars, _ := bv.(map[string]any)
		base, err := loadPack(p, bvars, prefix, depth+1)
		if err != nil {
			return nil, err
		}
		layers = append(layers, base...)
	}
	// Step titles and prompts keep ${run var} references for run start.
	restore := holdTemplates(m)
	// ${check_output} is filled when a trigger fires, not at load.
	subVars := map[string]any{CheckOutputVar: "${" + CheckOutputVar + "}"}
	for _, k := range StallVars { // likewise, per stall
		subVars[k] = "${" + k + "}"
	}
	for k, v := range vars {
		subVars[k] = v
	}
	v, err := substitute(m, subVars, "")
	restore()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	m, _ = v.(map[string]any)
	if m == nil {
		m = map[string]any{}
	}
	if err := setPackVars(m, vars, file); err != nil {
		return nil, err
	}
	if err := resolvePromptFiles(m, dir, file, subVars); err != nil {
		return nil, err
	}
	raw, err = yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	var pf packFile
	if err := yaml.UnmarshalWithOptions(raw, &pf, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return append(layers, packLayer{file: file, m: m}), nil
}

// cloneValue deep-copies maps and lists so substitution does not edit
// the caller's config.
func cloneValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = cloneValue(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = cloneValue(e)
		}
		return out
	}
	return v
}

func decodeMap(b []byte) (map[string]any, error) {
	m := map[string]any{}
	if len(bytes.TrimSpace(b)) == 0 {
		return m, nil
	}
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

var varPattern = regexp.MustCompile(`\$\$|\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// substitute replaces ${name} in string values. A string that is exactly
// one reference takes the var's value with its type, so numbers stay
// numbers. "$$" is a literal "$" (for ${SESSION_DIR} style agent args).
func substitute(v any, vars map[string]any, key string) (any, error) {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			r, err := substitute(e, vars, joinKey(key, k))
			if err != nil {
				return nil, err
			}
			x[k] = r
		}
		return x, nil
	case []any:
		for i, e := range x {
			r, err := substitute(e, vars, fmt.Sprintf("%s[%d]", key, i))
			if err != nil {
				return nil, err
			}
			x[i] = r
		}
		return x, nil
	case string:
		if m := varPattern.FindStringSubmatch(x); m != nil && m[0] == x && m[1] != "" {
			val, ok := vars[m[1]]
			if !ok {
				return nil, fmt.Errorf("%s: var %q is not set", key, m[1])
			}
			return val, nil
		}
		var missing string
		out := varPattern.ReplaceAllStringFunc(x, func(s string) string {
			if s == "$$" {
				return "$"
			}
			n := s[2 : len(s)-1]
			val, ok := vars[n]
			if !ok {
				if missing == "" {
					missing = n
				}
				return s
			}
			return fmt.Sprint(val)
		})
		if missing != "" {
			return nil, fmt.Errorf("%s: var %q is not set", key, missing)
		}
		return out, nil
	}
	return v, nil
}

func joinKey(a, b string) string {
	if a == "" {
		return b
	}
	return a + "." + b
}

// resolvePromptFiles replaces pools.*.prompt_file with the file's content
// as prompt, reading relative to dir (the declaring file's directory).
// In a pack (vars non-nil) the content gets pack var substitution like an
// inline prompt; host prompt files stay literal.
func resolvePromptFiles(m map[string]any, dir, file string, vars map[string]any) error {
	pools, _ := m["pools"].(map[string]any)
	names := make([]string, 0, len(pools))
	for n := range pools {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		p, _ := pools[n].(map[string]any)
		pf, ok := p["prompt_file"]
		if !ok {
			continue
		}
		key := "pools." + n
		if _, both := p["prompt"]; both {
			return fmt.Errorf("%s: %s: set prompt or prompt_file, not both", file, key)
		}
		s, ok := pf.(string)
		if !ok || s == "" {
			return fmt.Errorf("%s: %s.prompt_file must be a path", file, key)
		}
		if !filepath.IsAbs(s) {
			s = filepath.Join(dir, s)
		}
		b, err := os.ReadFile(s)
		if err != nil {
			return fmt.Errorf("%s: %s.prompt_file: %w", file, key, err)
		}
		delete(p, "prompt_file")
		prompt := string(b)
		if vars != nil {
			v, err := substitute(prompt, vars, key+".prompt_file")
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			prompt = v.(string)
		}
		p["prompt"] = prompt
	}
	return resolveStepPromptFiles(m, dir, file)
}

// mergeInto applies src over dst. In the merged sections each entry merges
// recursively: maps merge per key, other values (lists, scalars) replace.
// Other top-level keys replace. sources records which files set each
// section entry, as "section.name".
func mergeInto(dst, src map[string]any, file string, sources map[string][]string) {
	for k, v := range src {
		if !isMergedSection(k) {
			dst[k] = v
			continue
		}
		sm, ok := v.(map[string]any)
		if !ok {
			dst[k] = v
			continue
		}
		dm, _ := dst[k].(map[string]any)
		if dm == nil {
			dm = map[string]any{}
			dst[k] = dm
		}
		for name, entry := range sm {
			dm[name] = mergeValue(dm[name], entry)
			key := k + "." + name
			sources[key] = append(sources[key], file)
		}
	}
}

func mergeValue(old, v any) any {
	om, ok1 := old.(map[string]any)
	nm, ok2 := v.(map[string]any)
	if !ok1 || !ok2 {
		return v
	}
	out := make(map[string]any, len(om)+len(nm))
	for k, e := range om {
		out[k] = e
	}
	for k, e := range nm {
		out[k] = mergeValue(om[k], e)
	}
	return out
}

func isMergedSection(k string) bool {
	for _, s := range mergedSections {
		if s == k {
			return true
		}
	}
	return false
}

// PackPrefixVar is set in every pack layer to the including reference's
// prefix ("" without one), so prompts name the pack's own pools and
// workflows (#184).
const PackPrefixVar = "pack_prefix"

// prefixPattern keeps prefixed names within namePattern's charset.
var prefixPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,15}$`)

// prefixLayers renames the pools, agents, triggers and workflows the
// layers declare to prefix+name, and rewrites the references between
// them (pool agent, step pool, escalate_pool, trigger when.pool,
// ticket.pool and run.workflow). Names the pack does not declare (host
// pools and agents) are left alone. This lets one pack, such as
// packs/github, be included once per repo (#184).
func prefixLayers(layers []packLayer, prefix string) {
	declared := map[string]map[string]bool{}
	for _, s := range mergedSections {
		declared[s] = map[string]bool{}
	}
	for _, l := range layers {
		for _, s := range mergedSections {
			sm, _ := l.m[s].(map[string]any)
			for n := range sm {
				declared[s][n] = true
			}
		}
	}
	ren := func(m map[string]any, key, section string) {
		if v, ok := m[key].(string); ok && declared[section][v] {
			m[key] = prefix + v
		}
	}
	for _, l := range layers {
		for _, s := range mergedSections {
			sm, ok := l.m[s].(map[string]any)
			if !ok {
				continue
			}
			out := make(map[string]any, len(sm))
			for n, e := range sm {
				em, _ := e.(map[string]any)
				switch {
				case em == nil:
				case s == "pools":
					ren(em, "agent", "agents")
				case s == "workflows":
					ren(em, "escalate_pool", "pools")
					steps, _ := em["steps"].([]any)
					for _, st := range steps {
						if sm, ok := st.(map[string]any); ok {
							ren(sm, "pool", "pools")
						}
					}
				case s == "triggers":
					for _, k := range []string{"when", "ticket"} {
						if m, ok := em[k].(map[string]any); ok {
							ren(m, "pool", "pools")
						}
					}
					if m, ok := em["run"].(map[string]any); ok {
						ren(m, "workflow", "workflows")
					}
				}
				out[prefix+n] = e
			}
			l.m[s] = out
		}
	}
}
