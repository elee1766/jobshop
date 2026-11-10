package pack

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/gfx-labs/jobshop/pkg/pack/tmpl"
)

// maxDepth bounds extends chains so a cycle fails instead of looping.
const maxDepth = 8

// Input is the service file as the caller read it.
type Input struct {
	// Path is the service file (jobshop.yaml). Its directory, after
	// following links (#532), is what relative pack paths resolve against.
	Path string
	// Top is the service file's document with its includes already
	// applied. Assemble consumes it.
	Top map[string]any
	// Watch lists files the caller read (the service file and its
	// includes), so the result watches them too.
	Watch []string
	// Rewrite passes pack and packs[].path through before they are read
	// (nil keeps them). `jobshop config check` uses it to load the host's
	// config against another checkout's packs before a deploy switches to
	// it (#261).
	Rewrite func(string) string
}

// Options are the facts of the service configuration that assembly needs,
// passed as data so this package does not import the configuration types.
type Options struct {
	// CheckLayer strictly decodes one pack file's merged-section content
	// (root is true for the root pack, which also carries projects and
	// resources) and returns the first schema error. It is how an unknown
	// key in a pack fails at that pack's file instead of in the merged
	// document.
	CheckLayer func(layer map[string]any, root bool) error
	// LateVars are template vars filled after load (a trigger's check
	// output, its event, scan findings). A pack's own text sees them as
	// the references they will be rendered from, not as missing vars.
	LateVars map[string]any
	// EscalationVars are the vars an escalation chat's first message is
	// rendered with at run time. Escalation prompts keep them and take
	// the pack vars at load.
	EscalationVars []string
	// StepTemplateKeys are the workflow step fields kept as templates:
	// pack substitution skips them so run vars survive until run start.
	StepTemplateKeys []string
}

// Result is the assembled document and what is known about where it came
// from. Document is YAML for the caller to decode and validate.
type Result struct {
	Document []byte
	// Sources maps "<section>.<name>" (and "<section>.<name>.<field>" for
	// a *_file field) to the files that set it.
	Sources map[string][]string
	// Pack is the root pack file named by the service file's pack key.
	Pack string
	// Packs are the instances listed by the service file and the root pack.
	Packs     []Ref
	Instances []Instance
	// Warnings are compatibility diagnostics for the operator.
	Warnings []string
	// Watch is what a config watcher follows. The caller may add files it
	// read later with Watch.AddFile.
	Watch WatchPaths

	refs       []Ref
	rootLayers []layer
}

// Blame marks err as a pack error when it names a pack file, so the
// controller keeps its last good config instead of exiting (#227). An
// error that names only the service file is returned as is.
func (r *Result) Blame(err error) error {
	if len(r.refs) > 0 && strings.Contains(err.Error(), FileName) {
		return packErr(err)
	}
	for _, l := range r.rootLayers {
		if strings.Contains(err.Error(), l.file) {
			return packErr(err)
		}
		for _, files := range l.entrySources {
			for _, file := range files {
				if strings.Contains(err.Error(), file) {
					return packErr(err)
				}
			}
		}
	}
	return err
}

// loader holds the caller's options for one assembly.
type loader struct{ opts Options }

// lateVars returns a copy of the late-filled placeholders, to add a pack's
// own vars to.
func (ld loader) lateVars() map[string]any {
	out := make(map[string]any, len(ld.opts.LateVars))
	maps.Copy(out, ld.opts.LateVars)
	return out
}

// Assemble resolves the service file's root pack and its reusable packs.
// Children merge in list order, then the root pack, then the service
// file's own definitions.
func Assemble(in Input, opts Options) (*Result, error) {
	ld := loader{opts}
	path, top := in.Path, in.Top
	name := filepath.Base(path)
	watch := &WatchPaths{}
	for _, f := range in.Watch {
		watch.AddFile(f)
	}
	// Includes name host files next to the root's jobshop.yaml; packs and
	// prompt files live with the config, which may be a link into a
	// checkout (#532).
	dir := FileDir(path)
	// Validate the packs list on its own so its errors are precise.
	var head struct {
		Pack  string `yaml:"pack"`
		Packs []Ref  `yaml:"packs"`
	}
	raw, err := yaml.Marshal(map[string]any{"pack": top["pack"], "packs": top["packs"]})
	if err != nil {
		return nil, fmt.Errorf("%s: packs: %w", name, err)
	}
	if err := yaml.UnmarshalWithOptions(raw, &head, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("%s: packs: %w", name, err)
	}
	if _, ok := top["pack"]; ok && head.Pack == "" {
		return nil, fmt.Errorf("%s: pack: file path is required", name)
	}
	merged := map[string]any{}
	sources := map[string][]string{}
	var instances []Instance
	var warnings []string
	if head.Pack != "" {
		var legacy []string
		for _, key := range []string{"default_project", "projects", "agents", "resources", "pools", "triggers", "workflows", "prompt_rules", "hooks", "escalation", "learnings", "desk_commands", "sandbox_profiles", "isolation_profiles", "meta_packs", "packs"} {
			if _, ok := top[key]; ok {
				legacy = append(legacy, key)
			}
		}
		if len(legacy) > 0 {
			warnings = append(warnings, fmt.Sprintf("%s: move work definitions (%s) to %s; service work definitions are deprecated", name, strings.Join(legacy, ", "), head.Pack))
		}
	}
	rootLayers, rootRefs, err := ld.loadRoot(head.Pack, dir, in.Rewrite)
	if err != nil {
		return nil, err
	}
	head.Packs = append(rootRefs, head.Packs...)
	for i, ref := range head.Packs {
		if ref.Path == "" {
			return nil, fmt.Errorf("%s: packs[%d]: path is required", name, i)
		}
		p := ref.Path
		if in.Rewrite != nil {
			p = in.Rewrite(p)
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		if ref.Prefix != "" && !prefixPattern.MatchString(ref.Prefix) {
			return nil, fmt.Errorf("%s: packs[%d]: prefix %q must match %s", name, i, ref.Prefix, prefixPattern)
		}
		layers, err := ld.loadPack(p, ref.Vars, ref.Prefix, 0)
		if err != nil {
			return nil, packErr(err)
		}
		if ref.Prefix != "" {
			prefixLayers(layers, ref.Prefix)
			for i := range layers {
				layers[i].textSources = prefixSources(layers[i].textSources, ref.Prefix)
			}
		}
		instances = append(instances, summarize(ref.Path, ref.Prefix, layers))
		for _, l := range layers {
			watch.addDir(filepath.Dir(l.file))
		}
		for _, l := range layers {
			mergeLayer(merged, l, sources)
			warnings = append(warnings, l.warnings...)
		}
	}
	for _, l := range rootLayers {
		// Watch only the root file and its text files, not the entire checkout.
		watch.AddFile(l.file)
		if target, err := filepath.EvalSymlinks(l.file); err == nil {
			watch.AddFile(target)
		}
		mergeLayer(merged, l, sources)
		instances = append(instances, l.instances...)
		for _, file := range l.watch.Files {
			watch.AddFile(file)
		}
		for _, d := range l.watch.CUEDirs {
			watch.addCUEDir(d)
		}
		warnings = append(warnings, l.warnings...)
	}
	delete(top, "pack")
	delete(top, "packs")
	if err := setPackVars(top, nil, name); err != nil {
		return nil, err
	}
	if err := ld.resolveTextFiles(top, textFiles{dir: dir, file: name, sources: sources}); err != nil {
		return nil, err
	}
	if err := resolveScriptEntries(top, dir, name, sources); err != nil {
		return nil, err
	}
	mergeInto(merged, top, path, sources)
	out, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	// Text files set by a path outside every pack dir are watched alone.
	for _, files := range sources {
		for _, f := range files {
			if filepath.IsAbs(f) && !slices.ContainsFunc(watch.Dirs, func(d string) bool { return f == d || strings.HasPrefix(f, d+string(filepath.Separator)) }) {
				watch.AddFile(f)
			}
		}
	}
	return &Result{
		Document: out, Sources: sources, Pack: head.Pack, Packs: head.Packs,
		Instances: instances, Warnings: warnings, Watch: *watch,
		refs: head.Packs, rootLayers: rootLayers,
	}, nil
}

// loadPack reads dir/pack.yaml, substitutes vars, resolves prompt_file and
// checks the schema. It returns the extended packs' layers first, then its
// own. Errors name the pack file.
func (ld loader) loadPack(dir string, given map[string]any, prefix string, depth int) ([]layer, error) {
	return ld.loadPackFile(filepath.Join(dir, FileName), given, prefix, depth, false)
}

func (ld loader) loadPackFile(file string, given map[string]any, prefix string, depth int, root bool) ([]layer, error) {
	dir := FileDir(file)
	if depth > maxDepth {
		return nil, fmt.Errorf("%s: extends nested more than %d deep (a cycle?)", file, maxDepth)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("pack %s: %w", dir, err)
	}
	m, err := DecodeMap(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	var h head
	raw, err := yaml.Marshal(map[string]any{"defaults": m["defaults"], "extends": m["extends"], "renamed_vars": m["renamed_vars"], "removed_vars": m["removed_vars"], "template": m["template"]})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if err := yaml.UnmarshalWithOptions(raw, &h, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	delete(m, "defaults")
	delete(m, "extends")
	delete(m, "renamed_vars")
	delete(m, "removed_vars")
	delete(m, "template")
	if err := checkSyntax(h.Template); err != nil {
		return nil, fmt.Errorf("%s: template: %w", file, err)
	}
	if err := checkRemovedVars(given, h.RemovedVars, file); err != nil {
		return nil, err
	}
	given, warnings := renameVars(given, h.RenamedVars, file)
	warnings = append(warnings, fmt.Sprintf("%s: YAML packs are deprecated; migrate this pack to CUE (see docs/configuration/migration.md)", file))
	// Given vars win over defaults.
	vars := map[string]any{}
	maps.Copy(vars, h.Defaults)
	maps.Copy(vars, given)
	vars[PrefixVar] = prefix
	if err := renderDefaults(vars, h.Defaults, given); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if root {
		if err := rootPackRefs(m, vars, &h); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
	}
	var layers []layer
	for i, ref := range h.Extends {
		if ref.Path == "" {
			return nil, fmt.Errorf("%s: extends[%d]: path is required", file, i)
		}
		if ref.Prefix != "" {
			return nil, fmt.Errorf("%s: extends[%d]: prefix is only allowed in packs entries", file, i)
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
		base, err := ld.loadPack(p, bvars, prefix, depth+1)
		if err != nil {
			return nil, err
		}
		layers = append(layers, base...)
	}
	// ${name} is checked on the raw pack, before vars fill in: the
	// removed syntax would otherwise pass through as literal text (#566).
	// A var's value may hold ${...} (shell in a check command).
	if left := leftoverVars(m, file); len(left) > 0 {
		return nil, errors.New(strings.Join(left, "\n"))
	}
	// Step titles and prompts keep {{.run var}} references for run start.
	restore := ld.holdTemplates(m)
	restoreEsc := holdEscalationPrompts(m)
	// Late-filled vars (check output, event, scan findings) keep their
	// own references.
	subVars := ld.lateVars()
	maps.Copy(subVars, vars)
	v, err := substitute(m, subVars, "")
	restore()
	restoreEsc()
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
	textSources := map[string][]string{}
	if err := ld.resolveTextFiles(m, textFiles{dir: dir, file: file, vars: subVars, sources: textSources}); err != nil {
		return nil, err
	}
	if err := resolveScriptEntries(m, dir, file, textSources); err != nil {
		return nil, err
	}
	if err := ld.checkLayer(m, root); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return append(layers, layer{file: file, m: m, vars: vars, warnings: warnings, textSources: textSources, packs: h.Packs}), nil
}

func (ld loader) checkLayer(m map[string]any, root bool) error {
	if ld.opts.CheckLayer == nil {
		return nil
	}
	return ld.opts.CheckLayer(m, root)
}

// checkRemovedVars fails when the includer sets a var the pack removed.
func checkRemovedVars(given map[string]any, removed map[string]string, file string) error {
	for _, k := range slices.Sorted(maps.Keys(removed)) {
		if _, ok := given[k]; ok {
			return fmt.Errorf("%s: var %q is removed: %s", file, k, removed[k])
		}
	}
	return nil
}

// renderDefaults fills the string defaults that hold a template, so a
// default can name the instance's own pool: dev_pool:
// "{{.pack_prefix}}dev" (#880). Templates see the given vars, the
// defaults and pack_prefix; another templated default reads as its
// template text. A var the includer set is left as given.
func renderDefaults(vars, defaults, given map[string]any) error {
	src := maps.Clone(vars)
	for _, k := range slices.Sorted(maps.Keys(defaults)) {
		s, ok := defaults[k].(string)
		if _, set := given[k]; set || !ok || !strings.Contains(s, "{{") {
			continue
		}
		r, err := tmpl.Value(s, src)
		if err != nil {
			return fmt.Errorf("defaults.%s: %w", k, err)
		}
		vars[k] = r
	}
	return nil
}

// renameVars returns given with old names from renamed moved to their new
// names, and a warning per old name used. A new name set as well wins.
func renameVars(given map[string]any, renamed map[string]string, file string) (map[string]any, []string) {
	if len(renamed) == 0 {
		return given, nil
	}
	out := maps.Clone(given)
	if out == nil {
		out = map[string]any{}
	}
	var warnings []string
	for _, o := range slices.Sorted(maps.Keys(renamed)) {
		v, ok := out[o]
		if !ok {
			continue
		}
		n := renamed[o]
		delete(out, o)
		if _, set := out[n]; !set {
			out[n] = v
		}
		warnings = append(warnings, fmt.Sprintf("%s: var %q is renamed to %q; set %q in the host config, the old name will stop working", file, o, n, n))
	}
	return out, warnings
}

// holdTemplates removes step templates from the raw workflows section so
// pack var substitution leaves them alone. The returned func puts them
// back.
func (ld loader) holdTemplates(m map[string]any) func() {
	type held struct {
		step map[string]any
		key  string
		v    any
	}
	var hs []held
	wfs, _ := m["workflows"].(map[string]any)
	for _, w := range wfs {
		wm, _ := w.(map[string]any)
		steps, _ := wm["steps"].([]any)
		for _, s := range steps {
			sm, _ := s.(map[string]any)
			for _, k := range ld.opts.StepTemplateKeys {
				if v, ok := sm[k]; ok {
					hs = append(hs, held{sm, k, v})
					delete(sm, k)
				}
			}
		}
	}
	return func() {
		for _, h := range hs {
			h.step[h.key] = h.v
		}
	}
}

// setPackVars records the declaring pack's vars on each workflow it
// declares, and rejects a user-set pack_vars.
func setPackVars(m map[string]any, vars map[string]any, file string) error {
	wfs, _ := m["workflows"].(map[string]any)
	for name, w := range wfs {
		wm, _ := w.(map[string]any)
		if wm == nil {
			continue
		}
		if _, ok := wm["pack_vars"]; ok {
			return fmt.Errorf("%s: workflows.%s: pack_vars is set by the loader", file, name)
		}
		if vars != nil {
			wm["pack_vars"] = vars
		}
	}
	return nil
}

// escalationPrompts are the inline prompt fields of an escalation block.
var escalationPrompts = []string{"prompt", "pool_prompt", "hold_prompt"}

// escalationPassthrough returns vars plus each escalation var mapped to
// its own reference, so pack substitution leaves them for the controller.
// Escalation vars win over pack vars of the same name.
func escalationPassthrough(vars map[string]any, escalation []string) map[string]any {
	out := make(map[string]any, len(vars)+len(escalation))
	maps.Copy(out, vars)
	for _, k := range escalation {
		out[k] = Placeholder(k)
	}
	return out
}

// holdEscalationPrompts removes inline escalation prompts so pack var
// substitution of the whole file leaves them alone; the returned func
// puts them back. resolveTextFiles then substitutes them with the
// escalation vars passed through.
func holdEscalationPrompts(m map[string]any) func() {
	type held struct {
		e map[string]any
		k string
		v any
	}
	var hs []held
	for _, b := range escalationBlocks(m) {
		e, _ := b.(map[string]any)
		for _, k := range escalationPrompts {
			if v, ok := e[k]; ok {
				hs = append(hs, held{e, k, v})
				delete(e, k)
			}
		}
	}
	return func() {
		for _, h := range hs {
			h.e[h.k] = h.v
		}
	}
}

// escalationBlocks returns the raw escalation maps of m: the top-level
// escalation block and each workflow's escalate block.
func escalationBlocks(m map[string]any) map[string]any {
	out := map[string]any{}
	if e, ok := m["escalation"].(map[string]any); ok {
		out["escalation"] = e
	}
	wfs, _ := m["workflows"].(map[string]any)
	for n, w := range wfs {
		wm, _ := w.(map[string]any)
		if e, ok := wm["escalate"].(map[string]any); ok {
			out["workflows."+n+".escalate"] = e
		}
	}
	return out
}
