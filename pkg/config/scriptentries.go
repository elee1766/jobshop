package config

// ScriptEntry is one configured Starlark entry and the number of
// positional arguments its host passes.
type ScriptEntry struct {
	Entry      string
	Positional int
}

// ScriptEntries lists every configured Starlark entry, empty ones
// included, so callers share one inventory of callbacks and handlers.
func (c *Config) ScriptEntries() []ScriptEntry {
	var out []ScriptEntry
	for _, w := range c.Workflows {
		out = append(out, ScriptEntry{w.OnStart, 1}, ScriptEntry{w.OnStepCompleted, 2})
	}
	for _, tr := range c.Triggers {
		out = append(out, ScriptEntry{tr.Handler, 2})
		for _, s := range tr.When.Scan {
			out = append(out, ScriptEntry{s.Script, 1})
		}
	}
	for _, h := range c.Hooks {
		out = append(out, ScriptEntry{h.Handler, 2})
	}
	for _, d := range c.DeskCommands {
		out = append(out, ScriptEntry{d.Handler, 2})
	}
	return out
}
