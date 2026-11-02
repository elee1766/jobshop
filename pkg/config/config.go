// Package config loads jobshop.yaml and locates the Jobshop root directory.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
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
	Lease     string              `yaml:"lease"`
	Agents    map[string]Agent    `yaml:"agents"`
	Pools     map[string]Pool     `yaml:"pools"`
	Triggers  map[string]Trigger  `yaml:"triggers"`
	Workflows map[string]Workflow `yaml:"workflows"`
	Telemetry Telemetry           `yaml:"telemetry"`
	// Packs are pack directories merged before this file (see LoadFile).
	Packs []PackRef `yaml:"packs"`
	// MetaPacks are extra allowed top-level ticket meta keys, beyond
	// BuiltinMetaPacks.
	MetaPacks []string `yaml:"meta_packs"`
	// Sources maps "pools.<name>", "agents.<name>", "triggers.<name>" and
	// "workflows.<name>" to the files that set the entry, in merge order.
	Sources map[string][]string `yaml:"-"`
}

// Trigger creates a ticket from Ticket, or starts a workflow run from Run,
// while every condition in When holds and nothing it created is open.
type Trigger struct {
	When   TriggerWhen    `yaml:"when"`
	Ticket *TriggerTicket `yaml:"ticket"`
	Run    *TriggerRun    `yaml:"run"`
	// Cooldown is the minimum time after its last ticket closed before the
	// trigger fires again (Go duration).
	Cooldown string `yaml:"cooldown"`
}

// TriggerRun starts a workflow run, like `jobshop run start`. Vars and
// Key are templates; ${check_output} is the first line of the check's
// trimmed stdout. An open run with the same workflow and key dedups.
type TriggerRun struct {
	Workflow string            `yaml:"workflow"`
	Vars     map[string]string `yaml:"vars"`
	Key      string            `yaml:"key"`
}

// CheckOutputVar is the template var holding a trigger check's output.
const CheckOutputVar = "check_output"

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

	// Check is a shell command (sh -c, run in the root dir) that holds
	// when it exits 0. Its trimmed stdout, if any, is a dedup key: the
	// trigger files at most one ticket per key, ever (e.g. a failing SHA).
	// It runs only after the store conditions hold.
	Check string `yaml:"check"`
	// CheckTimeout bounds one check run (Go duration, default 1m).
	CheckTimeout string `yaml:"check_timeout"`
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
	// OTLPTracesEndpoint and OTLPLogsEndpoint are full URLs that override
	// the base for one signal (VictoriaTraces and VictoriaLogs use their own
	// paths). Either one alone enables that signal.
	OTLPTracesEndpoint string `yaml:"otlp_traces_endpoint"`
	OTLPLogsEndpoint   string `yaml:"otlp_logs_endpoint"`
	// OTLPHeaders are sent on every export. Values expand ${VAR} from the
	// controller environment so tokens stay out of the file.
	OTLPHeaders map[string]string `yaml:"otlp_headers"`
}

// TracesURL is the traces export URL, or "" for none.
func (t Telemetry) TracesURL() string { return t.signalURL(t.OTLPTracesEndpoint, "/v1/traces") }

// LogsURL is the logs export URL, or "" for none.
func (t Telemetry) LogsURL() string { return t.signalURL(t.OTLPLogsEndpoint, "/v1/logs") }

func (t Telemetry) signalURL(full, path string) string {
	if full != "" {
		return full
	}
	if t.OTLPEndpoint == "" {
		return ""
	}
	return strings.TrimSuffix(t.OTLPEndpoint, "/") + path
}

// Headers returns OTLPHeaders with ${VAR} expanded by getenv. Expansion
// happens at use, not parse, so config show never prints secrets.
func (t Telemetry) Headers(getenv func(string) string) map[string]string {
	if len(t.OTLPHeaders) == 0 {
		return nil
	}
	h := make(map[string]string, len(t.OTLPHeaders))
	for k, v := range t.OTLPHeaders {
		h[k] = os.Expand(v, getenv)
	}
	return h
}

// Agent is how to start an ACP agent process.
type Agent struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
}

// PoolChat is Pool.Type for chat pools.
const PoolChat = "chat"

// Chat reports whether p is a chat pool.
func (p Pool) Chat() bool { return p.Type == PoolChat }

// AttachRW is Pool.Attach for pools whose sessions accept
// `session attach --rw`.
const AttachRW = "rw"

// Pool is a queue of work in one project served by sessions of one agent.
type Pool struct {
	Project string `yaml:"project"`
	Agent   string `yaml:"agent"`
	// Type is "" for a ticket pool or "chat". Chat pool sessions claim no
	// tickets; they are started by session.send for a key and answer
	// messages only.
	Type string `yaml:"type"`
	// Attach "rw" lets `session attach --rw` send messages to the pool's
	// sessions. Default "" (read-only attach).
	Attach string `yaml:"attach"`
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
	// PromptFile is read into Prompt at load time, relative to the file
	// that declares it. Setting both is an error.
	PromptFile string `yaml:"prompt_file"`
	// IdleTimeout ends a session that found no work for this long. Default
	// 5m, or 1h for chat pools (a message resets it).
	IdleTimeout string `yaml:"idle_timeout"`
	// TurnTimeout cancels a single agent turn. Default 2h.
	TurnTimeout string `yaml:"turn_timeout"`
	// ConversationExport makes the session host send prompts, agent
	// messages and tool calls as OTLP log records to the telemetry logs
	// endpoint, redacted and capped. Default off.
	ConversationExport bool `yaml:"conversation_export"`
}

// Project is one repo / work area.
type Project struct {
	// Prefix starts every ticket id, e.g. "js" gives js-0031w02j0r012.
	// Default: the project name.
	Prefix string `yaml:"prefix"`
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// varName is a ${name} reference name (see varPattern).
var varName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Load reads and validates the root's config.
func Load(r Root) (*Config, error) {
	return LoadFile(r.Config())
}

// Parse decodes and validates config bytes with no packs or prompt files
// (they need a file location; see LoadFile). Unknown keys are rejected.
func Parse(b []byte) (*Config, error) {
	return parse(b, nil)
}

// parse validates merged config bytes. sources maps "section.name" to the
// files that set it, so errors can name the file to fix.
func parse(b []byte, sources map[string][]string) (*Config, error) {
	src := func(section, name string) string {
		if s := sources[section+"."+name]; len(s) > 0 {
			return strings.Join(s, ", ")
		}
		return FileName
	}
	var c Config
	if len(bytes.TrimSpace(b)) > 0 {
		if err := yaml.UnmarshalWithOptions(b, &c, yaml.Strict()); err != nil {
			return nil, fmt.Errorf("%s: %w", FileName, err)
		}
	}
	c.Sources = sources
	if len(c.Packs) > 0 {
		return nil, fmt.Errorf("%s: packs need a config file location", FileName)
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
			return nil, fmt.Errorf("%s: agent name %q must match %s", src("agents", name), name, namePattern)
		}
		if a.Command == "" {
			return nil, fmt.Errorf("%s: agent %s: command is required", src("agents", name), name)
		}
	}
	for name, p := range c.Pools {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: pool name %q must match %s", src("pools", name), name, namePattern)
		}
		if p.Project == "" {
			p.Project = c.DefaultProject
		}
		if _, ok := c.Projects[p.Project]; !ok {
			return nil, fmt.Errorf("%s: pool %s: unknown project %q", src("pools", name), name, p.Project)
		}
		if _, ok := c.Agents[p.Agent]; !ok {
			return nil, fmt.Errorf("%s: pool %s: unknown agent %q", src("pools", name), name, p.Agent)
		}
		if p.Dir == "" && p.Repo == "" {
			return nil, fmt.Errorf("%s: pool %s: set dir or repo", src("pools", name), name)
		}
		if p.PromptFile != "" {
			return nil, fmt.Errorf("%s: pool %s: prompt_file needs a config file location", src("pools", name), name)
		}
		if p.Type != "" && p.Type != PoolChat {
			return nil, fmt.Errorf("%s: pool %s: type %q must be empty or %q", src("pools", name), name, p.Type, PoolChat)
		}
		if p.Attach != "" && p.Attach != AttachRW {
			return nil, fmt.Errorf("%s: pool %s: attach %q must be empty or %q", src("pools", name), name, p.Attach, AttachRW)
		}
		if p.MaxSessions < 0 {
			return nil, fmt.Errorf("%s: pool %s: max_sessions must be >= 0", src("pools", name), name)
		}
		if p.ConversationExport && c.Telemetry.LogsURL() == "" {
			return nil, fmt.Errorf("%s: pool %s: conversation_export needs telemetry.otlp_endpoint or otlp_logs_endpoint", src("pools", name), name)
		}
		for _, d := range []struct{ key, v string }{{"idle_timeout", p.IdleTimeout}, {"turn_timeout", p.TurnTimeout}} {
			if d.v == "" {
				continue
			}
			if v, err := time.ParseDuration(d.v); err != nil || v <= 0 {
				return nil, fmt.Errorf("%s: pool %s: %s %q must be a positive duration", src("pools", name), name, d.key, d.v)
			}
		}
		c.Pools[name] = p
	}
	for name, tr := range c.Triggers {
		if err := c.checkTrigger(name, &tr); err != nil {
			return nil, fmt.Errorf("%s: trigger %s: %w", src("triggers", name), name, err)
		}
		c.Triggers[name] = tr
	}
	for name, w := range c.Workflows {
		if err := c.checkWorkflow(name, &w); err != nil {
			return nil, fmt.Errorf("%s: workflow %s: %w", src("workflows", name), name, err)
		}
	}
	return &c, nil
}

func (c *Config) checkTrigger(name string, tr *Trigger) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("name must match %s", namePattern)
	}
	w := tr.When
	if w.Cron == "" && !w.HasCounts() && w.Check == "" {
		return errors.New("when needs a cron, a count threshold or a check")
	}
	if w.CheckTimeout != "" {
		if w.Check == "" {
			return errors.New("when: check_timeout needs check")
		}
		if d, err := time.ParseDuration(w.CheckTimeout); err != nil || d <= 0 {
			return fmt.Errorf("when: check_timeout %q must be a positive duration", w.CheckTimeout)
		}
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
	if (tr.Ticket == nil) == (tr.Run == nil) {
		return errors.New("set exactly one of ticket and run")
	}
	if tr.Run != nil {
		return c.checkTriggerRun(tr)
	}
	t := tr.Ticket
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

// BuiltinMetaPacks are the meta pack names the controller itself writes.
var BuiltinMetaPacks = []string{"github", "trigger", "run"}

// KnownMetaPack reports whether name may be a top-level meta key: a
// built-in pack or one listed under meta_packs.
func (c *Config) KnownMetaPack(name string) bool {
	return slices.Contains(BuiltinMetaPacks, name) || slices.Contains(c.MetaPacks, name)
}
