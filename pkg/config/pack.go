package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/goccy/go-yaml"

	"github.com/elee1766/jobshop/pkg/pack"
)

// packFile is the schema of pack.yaml: the mergeable jobshop.yaml sections,
// plus defaults for the pack's vars and the base packs it extends.
type packFile struct {
	Agents      map[string]Agent    `yaml:"agents"`
	Pools       map[string]Pool     `yaml:"pools"`
	Triggers    map[string]Trigger  `yaml:"triggers"`
	Workflows   map[string]Workflow `yaml:"workflows"`
	PromptRules []PromptRule        `yaml:"prompt_rules"`
	Hooks       []Hook              `yaml:"hooks"`
	Escalation  *EscalationChat     `yaml:"escalation"`
	// Learnings.Kinds are appended to the host's (#524).
	Learnings *LearningsConfig `yaml:"learnings"`
	// DeskCommands merge per name like pools (#542).
	DeskCommands map[string]DeskCommand `yaml:"desk_commands"`
	// SandboxProfiles merge per name like pools (#625).
	SandboxProfiles map[string]Sandbox `yaml:"sandbox_profiles"`
	// IsolationProfiles merge per name like pools (JOB-83).
	IsolationProfiles map[string]Isolation `yaml:"isolation_profiles"`
}

// rootPackFile adds shared work definitions to the reusable pack schema.
type rootPackFile struct {
	Work           packFile            `yaml:",inline"`
	DefaultProject string              `yaml:"default_project"`
	Projects       map[string]Project  `yaml:"projects"`
	Resources      map[string]Resource `yaml:"resources"`
	MetaPacks      []string            `yaml:"meta_packs"`
}

// checkPackLayer decodes one pack file's content strictly, so an unknown
// key fails at that file. It is pack.Options.CheckLayer.
func checkPackLayer(m map[string]any, root bool) error {
	b, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	var schema any = &packFile{}
	if root {
		schema = &rootPackFile{}
	}
	return yaml.UnmarshalWithOptions(b, schema, yaml.Strict())
}

// lateVars are the template vars filled after load, so a pack's text keeps
// them as references instead of failing on a missing var: a trigger's check
// output, its event fields and a scan finding's signal, subject and detail.
func lateVars() map[string]any {
	event := map[string]any{}
	for _, f := range EventFields {
		event[f] = pack.Placeholder(EventVar + "." + f)
	}
	vars := map[string]any{CheckOutputVar: pack.Placeholder(CheckOutputVar), EventVar: event}
	for _, name := range ScanVars {
		vars[name] = pack.Placeholder(name)
	}
	return vars
}

func packOptions() pack.Options {
	return pack.Options{
		CheckLayer:       checkPackLayer,
		LateVars:         lateVars(),
		EscalationVars:   EscalationVars,
		StepTemplateKeys: templateKeys,
	}
}

// LoadFile resolves the service file's root pack and its reusable packs.
// Children merge in list order, then the root pack, then service overrides.
func LoadFile(path string) (*Config, error) {
	return LoadFileWith(path, nil)
}

// LoadFileWith is LoadFile with pack and packs[].path passed through
// rewrite (nil keeps them). `jobshop config check` uses it to load
// the host's config against another checkout's packs before a deploy
// switches to it (#261).
func LoadFileWith(path string, rewrite func(string) string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseFile(path, b, rewrite)
}

func parseFile(path string, b []byte, rewrite func(string) string) (*Config, error) {
	name := filepath.Base(path)
	dir := filepath.Dir(path)
	top, err := pack.DecodeMap(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	watched := []string{path}
	if list, ok := top[IncludeKey].([]any); ok {
		for _, e := range list {
			if p, ok := e.(string); ok && p != "" {
				if !filepath.IsAbs(p) {
					p = filepath.Join(dir, p)
				}
				watched = append(watched, p)
			}
		}
	}
	if top, err = applyIncludes(top, dir, name); err != nil {
		return nil, err
	}
	res, err := pack.Assemble(pack.Input{Path: path, Top: top, Watch: watched, Rewrite: rewrite}, packOptions())
	if err != nil {
		return nil, err
	}
	c, err := parse(res.Document, res.Sources)
	if err != nil {
		// Validation errors name the files that set the entry; one naming
		// a pack file is the pack's.
		return nil, res.Blame(err)
	}
	c.Pack = res.Pack
	c.Packs = res.Packs
	c.Instances = res.Instances
	c.addInstanceLabels()
	// Label rules match the instance labels just added, so the merged
	// sandboxes can only be checked now.
	if err := c.checkSandboxes(); err != nil {
		return nil, res.Blame(err)
	}
	watch := res.Watch
	if err := c.checkHookPrograms(&watch); err != nil {
		return nil, err
	}
	slices.Sort(watch.Files)
	c.Watch = watch
	c.Warnings = append(res.Warnings, c.Warnings...)
	return c, nil
}

// CheckInstancePool rejects pool for a step of workflow when another packs
// entry declares the pool and the workflow's own entry does not (#880).
func (c *Config) CheckInstancePool(workflow, pool string) error {
	return pack.CheckInstancePool(c.Instances, workflow, pool)
}
