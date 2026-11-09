// Package pack assembles the work definitions of a service: the root pack,
// the reusable pack instances it lists, their extends chains, var
// substitution, prefixing, merge and the record of which file set what. It
// reads files and returns the merged YAML document, the instances, the
// sources and the paths to watch. It does not know the service
// configuration types: the caller passes the few facts it needs in Options
// and decodes the result itself (JOB-172).
package pack

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// FileName is the file in a pack directory that declares its pools,
// agents and triggers.
const FileName = "pack.yaml"

// ErrPack marks a load error in a pack, not in the host's own config. A
// live pack can need a newer binary than the one running (#227); the
// controller then keeps its last good config instead of exiting.
var ErrPack = errors.New("pack error")

// packErr wraps err as ErrPack, keeping its message.
func packErr(err error) error { return packError{err} }

type packError struct{ err error }

func (e packError) Error() string   { return e.err.Error() }
func (e packError) Unwrap() []error { return []error{e.err, ErrPack} }

// Marked wraps err as ErrPack for a caller that validates pack content
// itself, such as hook programs.
func Marked(err error) error { return packErr(err) }

// NamePattern is the charset of names a pack declares and of instance names.
// A prefix stays within it so a prefixed name is still a valid name.
var NamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// prefixPattern keeps prefixed names within NamePattern's charset.
var prefixPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,15}$`)

// ValidPrefix reports whether s is acceptable as an instance prefix.
func ValidPrefix(s string) bool { return prefixPattern.MatchString(s) }

// PrefixVar is set in every pack layer to the including reference's
// prefix ("" without one), so prompts name the pack's own pools and
// workflows (#184).
const PrefixVar = "pack_prefix"

// Ref includes a reusable pack directory.
type Ref struct {
	// Path is the pack directory. Relative paths resolve against the
	// directory holding the file that lists the pack.
	Path string `yaml:"path"`
	// Vars fill Go template references in the pack's string values.
	Vars map[string]any `yaml:"vars"`
	// Prefix renames everything the pack and its bases declare (see
	// prefixLayers), so one pack can be included once per repo. Only
	// in the root pack's packs list; bases inherit it.
	Prefix string `yaml:"prefix,omitempty"`
}

// head is the part of pack.yaml read before substitution.
type head struct {
	// Packs are independently prefixed instances loaded by the root pack.
	Packs []Ref `yaml:"packs"`
	// Defaults are var values used when the including file does not set
	// them.
	Defaults map[string]any `yaml:"defaults"`
	// Extends lists base packs merged before this one, like packs: in
	// jobshop.yaml. Their vars may reference this pack's vars.
	Extends []Ref `yaml:"extends"`
	// RenamedVars maps an old var name to its new one (#261). A host that
	// still sets the old name keeps working for a release, with a warning,
	// so a rename on master does not fail the deploy of hosts not yet
	// updated.
	RenamedVars map[string]string `yaml:"renamed_vars"`
	// RemovedVars maps a removed var name to what replaces it (JOB-83). A
	// host that still sets one fails to load, with that text in the error:
	// a var that carried config cannot keep working without a rendering
	// the pack no longer has.
	RemovedVars map[string]string `yaml:"removed_vars"`
	// Template is the pack's template: key. Only "go" (or unset) is
	// accepted, as a no-op for one release; ${name} was removed in #566.
	Template string `yaml:"template"`
}

// layer is one pack file's contribution after substitution.
type layer struct {
	file     string
	m        map[string]any
	vars     map[string]any // the layer's vars after defaults
	warnings []string       // compatibility diagnostics for the includer
	// textSources maps "<entry>.<field>" to the file a *_file field read.
	textSources  map[string][]string
	entrySources map[string][]string
	watch        WatchPaths
	instances    []Instance
	packs        []Ref
}

// Instance is one entry of the root pack's packs list as loaded: what
// an operator session reads to handle each repo or team the host serves
// (#248).
type Instance struct {
	Path   string `json:"path"`
	Prefix string `json:"prefix,omitempty"`
	// Vars are the pack's vars after defaults. With extends, a base's
	// vars win over the including pack's: the including pack passes
	// concrete values down (selfdev's repo is a checkout path, its
	// github base's repo is owner/name).
	Vars      map[string]any `json:"vars"`
	Pools     []string       `json:"pools"`
	Workflows []string       `json:"workflows"`
	Triggers  []string       `json:"triggers"`
}

// summarize builds one packs entry's Instance from its loaded layers.
func summarize(path, prefix string, layers []layer) Instance {
	in := Instance{Path: path, Prefix: prefix, Vars: map[string]any{}}
	names := map[string]map[string]bool{"pools": {}, "workflows": {}, "triggers": {}}
	// Layers are bases first; apply the top pack first so bases win.
	for i := len(layers) - 1; i >= 0; i-- {
		for k, v := range layers[i].vars {
			in.Vars[k] = v
		}
	}
	for _, l := range layers {
		for s, set := range names {
			sm, _ := l.m[s].(map[string]any)
			for n := range sm {
				set[n] = true
			}
		}
	}
	list := func(set map[string]bool) []string {
		out := make([]string, 0, len(set))
		for n := range set {
			out = append(out, n)
		}
		sort.Strings(out)
		return out
	}
	in.Pools, in.Workflows, in.Triggers = list(names["pools"]), list(names["workflows"]), list(names["triggers"])
	return in
}

// Labels are the labels an instance gives every pool it declares:
// instance:<name> from the entry's prefix (or the pack's name without
// one) and pack:<pack> (JOB-81).
func (in Instance) Labels() []string {
	pack := filepath.Base(in.Path)
	inst := strings.Trim(in.Prefix, "-_")
	if inst == "" {
		inst = pack
	}
	return []string{"instance:" + inst, "pack:" + pack}
}

// CheckInstancePool rejects pool for a step of workflow when another
// instance declares the pool and the workflow's own instance does not: a
// prefixed instance's run routed to the unprefixed instance's dev pool
// would work the wrong repo (#880). Host pools and workflows outside any
// pack are not checked.
func CheckInstancePool(instances []Instance, workflow, pool string) error {
	var own, other *Instance
	for i := range instances {
		in := &instances[i]
		hasPool := slices.Contains(in.Pools, pool)
		if slices.Contains(in.Workflows, workflow) {
			if hasPool {
				return nil
			}
			own = in
		} else if hasPool && other == nil {
			other = in
		}
	}
	if own == nil || other == nil {
		return nil
	}
	return fmt.Errorf("pool %q belongs to pack instance %s, not %s's %s", pool, instanceName(*other), workflow, instanceName(*own))
}

// instanceName names a packs entry in errors: its path, with the prefix.
func instanceName(in Instance) string {
	if in.Prefix == "" {
		return in.Path
	}
	return fmt.Sprintf("%s (prefix %s)", in.Path, in.Prefix)
}
