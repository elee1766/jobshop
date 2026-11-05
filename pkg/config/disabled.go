package config

import (
	"fmt"
	"maps"
	"slices"
)

// Disabled names the pools, workflows and triggers set enabled: false
// (JOB-24), sorted.
type Disabled struct {
	Pools     []string `json:"pools,omitempty"`
	Workflows []string `json:"workflows,omitempty"`
	Triggers  []string `json:"triggers,omitempty"`
}

// on reports an enabled field: absent is on.
func on(b *bool) bool { return b == nil || *b }

// dropDisabled runs after every entry, disabled or not, passed
// validation, so a disabled entry does not rot. A live entry that needs a
// disabled one is an error naming both; then the disabled entries leave
// the live maps, so no trigger fires, no run starts and no pool spawns
// sessions for them.
func (c *Config) dropDisabled(src func(section, name string) string) error {
	offPool := func(p string) bool {
		pl, ok := c.Pools[p]
		return ok && !on(pl.Enabled)
	}
	offWorkflow := func(w string) bool {
		wf, ok := c.Workflows[w]
		return ok && !on(wf.Enabled)
	}
	for _, name := range slices.Sorted(maps.Keys(c.Triggers)) {
		tr := c.Triggers[name]
		if !on(tr.Enabled) {
			continue
		}
		where := fmt.Sprintf("%s: trigger %s", src("triggers", name), name)
		if p := tr.When.Pool; p != "" && offPool(p) {
			return fmt.Errorf("%s: when: pool %s is disabled", where, p)
		}
		if tr.Ticket != nil && tr.Ticket.Pool != "" && offPool(tr.Ticket.Pool) {
			return fmt.Errorf("%s: ticket: pool %s is disabled", where, tr.Ticket.Pool)
		}
		if tr.Run != nil && offWorkflow(tr.Run.Workflow) {
			return fmt.Errorf("%s: run: workflow %s is disabled", where, tr.Run.Workflow)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(c.Workflows)) {
		w := c.Workflows[name]
		if !on(w.Enabled) {
			continue
		}
		where := fmt.Sprintf("%s: workflow %s", src("workflows", name), name)
		for _, s := range w.Steps {
			if s.Pool == "" {
				continue
			}
			// A pool naming a run var is checked at run start (ResolvePool).
			if p, err := w.Render(s.Pool, w.PackVars); err == nil && offPool(p) {
				return fmt.Errorf("%s: step %s: pool %s is disabled", where, s.ID, p)
			}
		}
		if offPool(w.EscalatePool) {
			return fmt.Errorf("%s: escalate_pool %s is disabled", where, w.EscalatePool)
		}
		if w.Escalate != nil && offPool(w.Escalate.ChatPool) {
			return fmt.Errorf("%s: escalate: chat_pool %s is disabled", where, w.Escalate.ChatPool)
		}
	}
	for i, h := range c.Hooks {
		for _, p := range h.Pools {
			if offPool(p) {
				return fmt.Errorf("%s: hooks[%d]: pool %s is disabled", FileName, i, p)
			}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(c.Gateway)) {
		if p := c.Gateway[name].Pool; offPool(p) {
			return fmt.Errorf("%s: gateway %s: pool %s is disabled", FileName, name, p)
		}
	}
	if c.Escalation != nil && offPool(c.Escalation.ChatPool) {
		return fmt.Errorf("%s: escalation: chat_pool %s is disabled", FileName, c.Escalation.ChatPool)
	}
	if p := c.Attention.PrestartPool; offPool(p) {
		return fmt.Errorf("%s: attention.prestart_pool %s is disabled", FileName, p)
	}
	var d Disabled
	for _, name := range slices.Sorted(maps.Keys(c.Pools)) {
		if !on(c.Pools[name].Enabled) {
			d.Pools = append(d.Pools, name)
			delete(c.Pools, name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(c.Workflows)) {
		if !on(c.Workflows[name].Enabled) {
			d.Workflows = append(d.Workflows, name)
			delete(c.Workflows, name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(c.Triggers)) {
		if !on(c.Triggers[name].Enabled) {
			d.Triggers = append(d.Triggers, name)
			delete(c.Triggers, name)
		}
	}
	c.Disabled = d
	return nil
}
