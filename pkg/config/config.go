// Package config loads jobshop.yaml and locates the Jobshop root directory.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/elee1766/jobshop/pkg/cron"
)

// Config is the parsed jobshop.yaml.
type Config struct {
	// DefaultProject is used when a request names no project.
	DefaultProject string             `yaml:"default_project"`
	Projects       map[string]Project `yaml:"projects"`
	// Lease is the claim lease (Go duration). Default 30m.
	Lease     string             `yaml:"lease"`
	Agents    map[string]Agent   `yaml:"agents"`
	Pools     map[string]Pool    `yaml:"pools"`
	Triggers  map[string]Trigger `yaml:"triggers"`
	Telemetry Telemetry          `yaml:"telemetry"`
}

// Trigger creates a ticket from Ticket while every condition in When
// holds and no ticket it created is still open.
type Trigger struct {
	When   TriggerWhen   `yaml:"when"`
	Ticket TriggerTicket `yaml:"ticket"`
	// Cooldown is the minimum time after its last ticket closed before the
	// trigger fires again (Go duration).
	Cooldown string `yaml:"cooldown"`
}

// TriggerWhen holds the conditions; all set ones must hold. Count
// thresholds compare the number of tickets matching Pool, Kind and Label.
type TriggerWhen struct {
	Pool  string `yaml:"pool"`
	Kind  string `yaml:"kind"` // default task
	Label string `yaml:"label"`

	ReadyBelow      *int `yaml:"ready_below"`
	ReadyAbove      *int `yaml:"ready_above"`
	OpenBelow       *int `yaml:"open_below"`
	OpenAbove       *int `yaml:"open_above"`
	InProgressBelow *int `yaml:"in_progress_below"`
	InProgressAbove *int `yaml:"in_progress_above"`

	// Cron is a 5-field expression; the condition holds from each
	// scheduled time until the trigger fires for it.
	Cron     string `yaml:"cron"`
	Timezone string `yaml:"timezone"` // IANA name, default UTC
}

// HasCounts reports whether any count threshold is set.
func (w TriggerWhen) HasCounts() bool {
	for _, p := range []*int{w.ReadyBelow, w.ReadyAbove, w.OpenBelow, w.OpenAbove, w.InProgressBelow, w.InProgressAbove} {
		if p != nil {
			return true
		}
	}
	return false
}

// TriggerTicket is the template for created tickets.
type TriggerTicket struct {
	Project  string   `yaml:"project"` // default: the pool's project
	Pool     string   `yaml:"pool"`
	Title    string   `yaml:"title"`
	Body     string   `yaml:"body"`
	Priority *int     `yaml:"priority"`
	Labels   []string `yaml:"labels"`
}

// Telemetry configures observability.
type Telemetry struct {
	// MetricsListen is the TCP address for Prometheus /metrics, e.g.
	// 127.0.0.1:9464. Empty: no metrics server.
	MetricsListen string `yaml:"metrics_listen"`
	// OTLPEndpoint is an OTLP/HTTP base URL (e.g. http://127.0.0.1:4318);
	// traces go to /v1/traces and logs to /v1/logs. Empty: no export.
	OTLPEndpoint string `yaml:"otlp_endpoint"`
}

// Agent is how to start an ACP agent process.
type Agent struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
}

// Pool is a queue of work in one project served by sessions of one agent.
type Pool struct {
	Project string `yaml:"project"`
	Agent   string `yaml:"agent"`
	// MaxSessions caps concurrent sessions. 0 means the pool never starts
	// sessions on its own (manual sessions only).
	MaxSessions int `yaml:"max_sessions"`
	// Dir is the working directory for agents. With Repo set, each session
	// instead gets its own git worktree of Repo under its session directory.
	Dir  string `yaml:"dir"`
	Repo string `yaml:"repo"`
	// Ref is the commit-ish new worktrees start from. Default HEAD.
	Ref string `yaml:"ref"`
	// Prompt is prepended to every ticket prompt (pool instructions).
	Prompt string `yaml:"prompt"`
	// IdleTimeout ends a session that found no work for this long. Default 5m.
	IdleTimeout string `yaml:"idle_timeout"`
	// TurnTimeout cancels a single agent turn. Default 2h.
	TurnTimeout string `yaml:"turn_timeout"`
}

// Project is one repo / work area.
type Project struct {
	// Prefix starts every ticket id, e.g. "js" gives js-0031w02j0r012.
	// Default: the project name.
	Prefix string `yaml:"prefix"`
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// Load reads and validates the root's config.
func Load(r Root) (*Config, error) {
	b, err := os.ReadFile(r.Config())
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse decodes and validates config bytes. Unknown keys are rejected.
func Parse(b []byte) (*Config, error) {
	var c Config
	if len(bytes.TrimSpace(b)) > 0 {
		if err := yaml.UnmarshalWithOptions(b, &c, yaml.Strict()); err != nil {
			return nil, fmt.Errorf("%s: %w", FileName, err)
		}
	}
	if c.Projects == nil {
		c.Projects = map[string]Project{}
	}
	seen := map[string]string{}
	for name, p := range c.Projects {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: project name %q must match %s", FileName, name, namePattern)
		}
		if p.Prefix == "" {
			p.Prefix = name
			c.Projects[name] = p
		}
		if !namePattern.MatchString(p.Prefix) {
			return nil, fmt.Errorf("%s: project %s: prefix %q must match %s", FileName, name, p.Prefix, namePattern)
		}
		if other, dup := seen[p.Prefix]; dup {
			return nil, fmt.Errorf("%s: projects %s and %s share prefix %q", FileName, other, name, p.Prefix)
		}
		seen[p.Prefix] = name
	}
	if c.DefaultProject == "" && len(c.Projects) == 1 {
		for name := range c.Projects {
			c.DefaultProject = name
		}
	}
	if c.DefaultProject != "" {
		if _, ok := c.Projects[c.DefaultProject]; !ok {
			return nil, fmt.Errorf("%s: default_project %q is not defined", FileName, c.DefaultProject)
		}
	}
	for name, a := range c.Agents {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: agent name %q must match %s", FileName, name, namePattern)
		}
		if a.Command == "" {
			return nil, fmt.Errorf("%s: agent %s: command is required", FileName, name)
		}
	}
	for name, p := range c.Pools {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: pool name %q must match %s", FileName, name, namePattern)
		}
		if p.Project == "" {
			p.Project = c.DefaultProject
		}
		if _, ok := c.Projects[p.Project]; !ok {
			return nil, fmt.Errorf("%s: pool %s: unknown project %q", FileName, name, p.Project)
		}
		if _, ok := c.Agents[p.Agent]; !ok {
			return nil, fmt.Errorf("%s: pool %s: unknown agent %q", FileName, name, p.Agent)
		}
		if p.Dir == "" && p.Repo == "" {
			return nil, fmt.Errorf("%s: pool %s: set dir or repo", FileName, name)
		}
		if p.MaxSessions < 0 {
			return nil, fmt.Errorf("%s: pool %s: max_sessions must be >= 0", FileName, name)
		}
		for _, d := range []struct{ key, v string }{{"idle_timeout", p.IdleTimeout}, {"turn_timeout", p.TurnTimeout}} {
			if d.v == "" {
				continue
			}
			if v, err := time.ParseDuration(d.v); err != nil || v <= 0 {
				return nil, fmt.Errorf("%s: pool %s: %s %q must be a positive duration", FileName, name, d.key, d.v)
			}
		}
		c.Pools[name] = p
	}
	for name, tr := range c.Triggers {
		if err := c.checkTrigger(name, &tr); err != nil {
			return nil, fmt.Errorf("%s: trigger %s: %w", FileName, name, err)
		}
		c.Triggers[name] = tr
	}
	return &c, nil
}

func (c *Config) checkTrigger(name string, tr *Trigger) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("name must match %s", namePattern)
	}
	w := tr.When
	if w.Cron == "" && !w.HasCounts() {
		return errors.New("when needs a cron or a count threshold")
	}
	if w.Pool != "" {
		if _, ok := c.Pools[w.Pool]; !ok {
			return fmt.Errorf("when: unknown pool %q", w.Pool)
		}
	}
	if w.Kind != "" && w.Kind != "task" && w.Kind != "message" {
		return fmt.Errorf("when: kind %q must be task or message", w.Kind)
	}
	for _, p := range []*int{w.ReadyBelow, w.ReadyAbove, w.OpenBelow, w.OpenAbove, w.InProgressBelow, w.InProgressAbove} {
		if p != nil && *p < 0 {
			return errors.New("when: thresholds must be >= 0")
		}
	}
	if w.Cron != "" {
		if _, err := cron.Parse(w.Cron); err != nil {
			return fmt.Errorf("when: %w", err)
		}
	}
	if w.Timezone != "" {
		if w.Cron == "" {
			return errors.New("when: timezone needs cron")
		}
		if _, err := time.LoadLocation(w.Timezone); err != nil {
			return fmt.Errorf("when: timezone: %w", err)
		}
	}
	if tr.Cooldown != "" {
		if d, err := time.ParseDuration(tr.Cooldown); err != nil || d < 0 {
			return fmt.Errorf("cooldown %q must be a non-negative duration", tr.Cooldown)
		}
	}
	t := &tr.Ticket
	if t.Title == "" {
		return errors.New("ticket: title is required")
	}
	if t.Pool != "" {
		p, ok := c.Pools[t.Pool]
		if !ok {
			return fmt.Errorf("ticket: unknown pool %q", t.Pool)
		}
		if t.Project == "" {
			t.Project = p.Project
		}
	}
	if t.Project == "" {
		t.Project = c.DefaultProject
	}
	if _, ok := c.Projects[t.Project]; !ok {
		return fmt.Errorf("ticket: unknown project %q", t.Project)
	}
	return nil
}

// Duration parses an optional duration string already checked by Parse.
func Duration(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return def
}

// Prefix returns the id prefix for a project.
func (c *Config) Prefix(project string) string {
	if p, ok := c.Projects[project]; ok {
		return p.Prefix
	}
	return project
}
