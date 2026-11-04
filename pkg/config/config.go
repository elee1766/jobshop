// Package config loads jobshop.yaml and locates the Jobshop root directory.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
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
	// PromptRules add operator guidance to matching sessions' prompts
	// (#205): packs' rules in pack order, then this file's.
	PromptRules []PromptRule `yaml:"prompt_rules"`
	Telemetry   Telemetry    `yaml:"telemetry"`
	// Limits are host-wide caps across pools (#281).
	Limits Limits `yaml:"limits"`
	// Resources are named slots steps hold while in progress (#400).
	Resources map[string]Resource `yaml:"resources"`
	// Ingest configures the ingest inbox (#321).
	Ingest Ingest `yaml:"ingest"`
	// Defaults hold settings agents and pools inherit (#262).
	Defaults Defaults `yaml:"defaults"`
	// SharedTools are dirs mounted read-only into every sandboxed
	// session (#262).
	SharedTools []string `yaml:"shared_tools"`
	// Attention configures `jobshop attention` (#185).
	Attention Attention `yaml:"attention"`
	// Hooks are guardrail commands (#188). Default none.
	Hooks []Hook `yaml:"hooks"`
	// Gateway configures `jobshop gateway <name>` processes, by name.
	Gateway map[string]Gateway `yaml:"gateway"`
	// Identities are GitHub Apps and Linear tokens sessions act as (#200).
	Identities map[string]Identity `yaml:"identities"`
	// Subscriptions polls GitHub threads that tickets follow (#187).
	// Absent: off.
	Subscriptions *Subscriptions `yaml:"subscriptions"`
	// Linear configures the Linear webhook endpoint (#300).
	Linear Linear `yaml:"linear"`
	// Activity polls recent GitHub and Linear activity for Desk (#255).
	// Absent: off.
	Activity *Activity `yaml:"activity"`
	// Packs are pack directories merged before this file (see LoadFile).
	Packs []PackRef `yaml:"packs"`
	// Instances summarize each Packs entry as loaded (#248). Not kept in
	// the last good config.
	Instances []PackInstance `yaml:"-"`
	// Warnings are non-fatal load findings, such as a deprecated pack var
	// name the host still sets (#261). The controller logs them.
	Warnings []string `yaml:"-"`
	// MetaPacks are extra allowed top-level ticket meta keys, beyond
	// BuiltinMetaPacks.
	MetaPacks []string `yaml:"meta_packs"`
	// Sources maps "pools.<name>", "agents.<name>", "triggers.<name>",
	// "workflows.<name>" and "prompt_rules.<name>" to the files that set
	// the entry, in merge order.
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
	// MaxRuns caps the runs a stall trigger starts per key (one stall);
	// after that the stall is left to the escalation of the last run.
	// Default 3. Stall triggers only.
	MaxRuns int `yaml:"max_runs"`
}

// Store stall signals, computed by the controller from the store and its
// own state. Any other signal name needs a check command.
const (
	StallReadyUnclaimed  = "ready_unclaimed"  // ready ticket nobody claimed for After
	StallPoolQuarantined = "pool_quarantined" // pool paused for crash-looping
	StallStepStalled     = "step_stalled"     // first unfinished step of a running run with no history or comment for After
	StallSessionSilent   = "session_silent"   // session holding a ticket with no journal output for After
	StallNeedsHuman      = "needs_human"      // needs_human run, or needs-human ticket, older than After
	StallBlocked         = "blocked"          // open ticket labeled blocked, unchanged for After
	// StallIssueNoRun is a check signal whose subjects are GitHub issue
	// numbers; the controller drops issues that have an open run
	// (meta.github.issue), so the check only lists open work issues.
	StallIssueNoRun = "issue_no_run"
	// StallLinearIssueNoRun is computed by the controller from Linear
	// (#287): open issues of Team matching States and Labels, minus
	// SkipLabels, that no open run carries in meta.linear.issue.
	StallLinearIssueNoRun = "linear_issue_no_run"
)

// DefaultLinearPoll is how often linear_issue_no_run reads Linear.
const DefaultLinearPoll = 5 * time.Minute

// StoreStallSignals lists the signals the controller computes itself.
var StoreStallSignals = []string{StallReadyUnclaimed, StallPoolQuarantined, StallStepStalled, StallSessionSilent, StallNeedsHuman, StallBlocked}

// Stall template vars, filled per stall when a stall trigger fires.
const (
	StallSignalVar  = "stall_signal"
	StallSubjectVar = "stall_subject"
	StallDetailVar  = "stall_detail"
)

// StallSignal is one stall condition. A store signal (StoreStallSignals)
// uses After. Any other name runs Check (sh -c in the root dir): exit 0
// with one stall per stdout line, "<subject> <detail>"; the subject is
// the first word.
type StallSignal struct {
	Signal string `yaml:"signal"`
	After  string `yaml:"after"` // Go duration threshold
	Check  string `yaml:"check"`
	// Repo (owner/name) is the repo an issue_no_run check lists; only
	// runs whose meta.github.repo matches hide an issue (#183).
	Repo string `yaml:"repo"`
	// linear_issue_no_run only (#287). Team is a Linear team key; empty
	// turns the signal off, so a pack can list it unconditionally.
	// States, Labels and SkipLabels are comma-separated names, matched
	// case-insensitively; empty States or Labels match any. Poll is how
	// often Linear is read (Go duration, default DefaultLinearPoll).
	Team       string `yaml:"team"`
	States     string `yaml:"states"`
	Labels     string `yaml:"labels"`
	SkipLabels string `yaml:"skip_labels"`
	Poll       string `yaml:"poll"`
}

// TriggerRun starts a workflow run, like `jobshop run start`. Vars and
// Key are templates; ${check_output} is the first line of the check's
// trimmed stdout. An open run with the same workflow and key dedups.
type TriggerRun struct {
	Workflow string            `yaml:"workflow"`
	Vars     map[string]string `yaml:"vars"`
	Key      string            `yaml:"key"`
	// Template is the declaring pack's syntax (SyntaxGo or ""), set by
	// the loader.
	Template string `yaml:"template,omitempty"`
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

	// Stall lists stall signals. Each stall (signal and subject) starts
	// its own run with key stall:<signal>:<subject>. A stall trigger has
	// no other condition and needs run:. check_timeout bounds its checks.
	Stall []StallSignal `yaml:"stall"`

	// Learnings holds on open learnings recorded since the trigger last
	// fired (#273).
	Learnings *LearningsWhen `yaml:"learnings"`

	// Edits holds when a pending memory edit is due for the curator
	// gate (#274).
	Edits *EditsWhen `yaml:"edits"`
}

// EditsWhen holds when a merged curator edit saw Runs runs in its pools
// or Age passed since its merge (defaults 20 and 72h).
type EditsWhen struct {
	Runs int    `yaml:"runs"`
	Age  string `yaml:"age"` // Go duration
}

// LearningsWhen holds when at least Min open learnings were recorded
// since the trigger last fired (or, never fired, since the controller
// started), or when Every passed since then and at least one was.
type LearningsWhen struct {
	Min   int    `yaml:"min"`
	Every string `yaml:"every"` // Go duration; empty: threshold only
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

// Subscriptions configures the GitHub thread poller (#187).
type Subscriptions struct {
	// Interval between polls (Go duration). Default 1m.
	Interval string `yaml:"interval"`
	// ResumeWindow is how long after a ticket closed a new human comment
	// still reopens it (Go duration). Older: the comment is only recorded.
	// Default 168h.
	ResumeWindow string `yaml:"resume_window"`
	// Ignore lists logins whose comments are dropped, beyond bots.
	Ignore []string `yaml:"ignore"`
	// Authors, when set, are the only logins whose comments are
	// delivered (#340); the public repo's other commenters are dropped.
	Authors []string `yaml:"authors"`
}

// Activity configures the external activity poller (#255). Repos and
// teams come from the loaded config: every workflow repo and every Linear
// identity team.
type Activity struct {
	// Interval between polls (Go duration). Default 1m; GitHub's
	// X-Poll-Interval can only lengthen it.
	Interval string `yaml:"interval"`
	// Window is how long items are kept (Go duration). Default 168h.
	Window string `yaml:"window"`
	// MaxItems bounds the items kept per source (repo or team). Default 1000.
	MaxItems int `yaml:"max_items"`
	// Repos are polled besides the workflow repos.
	Repos []string `yaml:"repos"`
}

// Gateway.Network values.
const (
	GatewayMatrix  = "matrix"
	GatewayDiscord = "discord"
)

// Gateway connects one chat network account to a chat pool. Messages
// from rooms or senders not listed are ignored.
type Gateway struct {
	Network string `yaml:"network"` // "matrix" or "discord"
	// Homeserver is the Matrix client-server base URL (matrix only).
	// For discord, rooms are channel ids (text or forum), senders user ids.
	Homeserver string `yaml:"homeserver"`
	// TokenFile holds the bot access token. Tokens are never in the yaml.
	TokenFile string   `yaml:"token_file"`
	Pool      string   `yaml:"pool"`    // a type: chat pool
	Rooms     []string `yaml:"rooms"`   // allowed room ids
	Senders   []string `yaml:"senders"` // allowed sender ids
	// InterruptMarker prefixes a message sent with interrupt. Default "!".
	InterruptMarker string `yaml:"interrupt_marker"`
	// EscalationRoom gets one new thread per escalation ticket, with a
	// session of Pool keyed by it. Must be one of Rooms so replies route.
	EscalationRoom string `yaml:"escalation_room"`
	// AttentionDigest is a 5-field cron (UTC) at which the gateway posts
	// `jobshop attention` to EscalationRoom as a new thread, only when the
	// list is not empty (#185). Empty: no digest.
	AttentionDigest string `yaml:"attention_digest"`
}

// Attention configures the attention view: what waits on a human.
type Attention struct {
	// RollbackFile is the self-deploy state file holding the commit that
	// was rolled back (the NixOS module's /var/lib/jobshop-deploy/failed).
	// The rollback is listed while it is newer than the "deployed" file
	// beside it, so a later successful deploy clears it. Empty:
	// rollbacks are not checked.
	RollbackFile string `yaml:"rollback_file"`
}

// DefaultInboxRetention is how long terminal inbox items are kept.
const DefaultInboxRetention = 14 * 24 * time.Hour

// Ingest configures the ingest inbox (#321).
type Ingest struct {
	// Retention keeps dispatched, denied and failed items this long (Go
	// duration, or Nd for days). Default 14d.
	Retention string `yaml:"retention"`
	// Listen is the TCP address `jobshop ingest serve` listens on
	// (#334). Empty: 127.0.0.1:8740.
	Listen string `yaml:"listen"`
	// Spool is the SQLite database received deliveries wait in until
	// the drain (#339). Relative to the root; empty: ingest/spool.db.
	Spool string `yaml:"spool"`
	// Endpoints are the webhook endpoints, each with its own path and
	// secret (#339).
	Endpoints []IngestEndpoint `yaml:"endpoints"`
	// GitHubReconcile is how often failed GitHub App webhook deliveries
	// are redelivered (#335): a Go duration, "off", or empty for 15m.
	// Only github-app identities have a deliveries API; with none it is
	// off.
	GitHubReconcile string `yaml:"github_reconcile"`
	// Admission hooks run on pending inbox items (#348).
	Admission []AdmissionHook `yaml:"admission"`

	// Removed in #339; set only to fail the load with a pointer to the
	// replacement instead of silently turning a source off.
	OldSpoolDir         string `yaml:"spool_dir"`
	OldGitHubSecretFile string `yaml:"github_secret_file"`
	OldLinearSecretFile string `yaml:"linear_secret_file"`
}

// Ingest endpoint sources.
const (
	IngestGitHub = "github"
	IngestLinear = "linear"
	// IngestGeneric is any other HMAC-signed JSON sender (#399).
	IngestGeneric = "generic"
)

// IngestSources are the endpoint sources.
var IngestSources = []string{IngestGitHub, IngestLinear, IngestGeneric}

// IngestEndpoint is one webhook endpoint (#339).
type IngestEndpoint struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
	// SecretFile is relative to the root. Identity instead names a
	// Linear identity (app or API key, #340) whose webhook_secret_file signs.
	SecretFile string `yaml:"secret_file"`
	Identity   string `yaml:"identity"`
	// Path defaults to /<source>/<name>.
	Path string `yaml:"path"`
	// Repos, GitHub only, allows deliveries for these owner/name repos.
	// Empty allows all, including deliveries with no repository.
	Repos []string `yaml:"repos"`
	// SignatureHeader, generic only, is the request header holding the
	// hex HMAC-SHA256 of the raw body, after SignaturePrefix (e.g.
	// "sha256=") is cut. DeliveryIDHeader, if set, names the sender's
	// delivery id; without one the body hash dedupes (#399).
	SignatureHeader  string `yaml:"signature_header"`
	SignaturePrefix  string `yaml:"signature_prefix"`
	DeliveryIDHeader string `yaml:"delivery_id_header"`
	// Map, generic only, says where the inbox subject fields are in the
	// JSON payload.
	Map IngestMap `yaml:"map"`
	// Implicit marks an endpoint made for an identity's webhook secret.
	Implicit bool `yaml:"-"`
}

// IngestMap maps a generic payload to inbox fields (#399). Each value
// is a dotted path into the JSON (`repo`, `prs.0`, `pull.head.sha`);
// NaturalKey is text with `{path}` parts filled the same way. An unset
// or missing path leaves the field empty.
type IngestMap struct {
	Kind       string `yaml:"kind" json:"kind,omitempty"`
	RepoOrTeam string `yaml:"repo_or_team" json:"repo_or_team,omitempty"`
	Number     string `yaml:"number" json:"number,omitempty"`
	Actor      string `yaml:"actor" json:"actor,omitempty"`
	CommentID  string `yaml:"comment_id" json:"comment_id,omitempty"`
	DeliveryID string `yaml:"delivery_id" json:"delivery_id,omitempty"`
	NaturalKey string `yaml:"natural_key" json:"natural_key,omitempty"`
}

// LinearWebhookPath is the shared path of implicit identity endpoints.
const LinearWebhookPath = "/linear/webhook"

var endpointName = regexp.MustCompile(`^[a-z0-9-]+$`)

// IngestEndpoints is ingest.endpoints plus an implicit endpoint
// linear-<identity> at /linear/webhook for every Linear identity
// with a webhook secret that no endpoint names, so #300 setups keep
// working.
func (c *Config) IngestEndpoints() []IngestEndpoint {
	out := slices.Clone(c.Ingest.Endpoints)
	named := map[string]bool{}
	for _, e := range out {
		if e.Identity != "" {
			named[e.Identity] = true
		}
	}
	for _, name := range c.IdentityNames() {
		id := c.Identities[name]
		if !IsLinear(id.Kind) || id.WebhookSecretFile == "" || named[name] {
			continue
		}
		out = append(out, IngestEndpoint{
			Name: "linear-" + name, Source: IngestLinear, Identity: name,
			Path: LinearWebhookPath, Implicit: true,
		})
	}
	return out
}

// EndpointPath is Path or the default /<source>/<name>.
func (e IngestEndpoint) EndpointPath() string {
	if e.Path != "" {
		return e.Path
	}
	return "/" + e.Source + "/" + e.Name
}

func (c *Config) validateIngest() error {
	i := c.Ingest
	if r := i.GitHubReconcile; r != "" && r != "off" {
		if d, err := time.ParseDuration(r); err != nil || d < time.Minute {
			return fmt.Errorf("%s: ingest.github_reconcile %q must be a duration of at least 1m, or off", FileName, r)
		}
	}
	if i.OldGitHubSecretFile != "" || i.OldLinearSecretFile != "" {
		return fmt.Errorf("%s: ingest.github_secret_file and ingest.linear_secret_file were replaced by ingest.endpoints (#339): list one endpoint per secret", FileName)
	}
	if i.OldSpoolDir != "" {
		return fmt.Errorf("%s: ingest.spool_dir was replaced by ingest.spool, a SQLite database path, and ingest.endpoints (#339)", FileName)
	}
	names, paths := map[string]bool{}, map[string]string{}
	for _, e := range i.Endpoints {
		where := fmt.Sprintf("%s: ingest.endpoints %q", FileName, e.Name)
		if !endpointName.MatchString(e.Name) {
			return fmt.Errorf("%s: name must match [a-z0-9-]+", where)
		}
		if names[e.Name] {
			return fmt.Errorf("%s: duplicate name", where)
		}
		names[e.Name] = true
		if !slices.Contains(IngestSources, e.Source) {
			return fmt.Errorf("%s: source must be github, linear or generic", where)
		}
		if e.Source == IngestGeneric {
			if e.SignatureHeader == "" {
				return fmt.Errorf("%s: a generic endpoint needs signature_header", where)
			}
		} else if e.SignatureHeader != "" || e.SignaturePrefix != "" || e.DeliveryIDHeader != "" || e.Map != (IngestMap{}) {
			return fmt.Errorf("%s: signature_header, signature_prefix, delivery_id_header and map are for generic endpoints", where)
		}
		if (e.SecretFile == "") == (e.Identity == "") {
			return fmt.Errorf("%s: set exactly one of secret_file and identity", where)
		}
		if e.Identity != "" {
			id, ok := c.Identities[e.Identity]
			if !ok || !IsLinear(id.Kind) || id.WebhookSecretFile == "" || e.Source != IngestLinear {
				return fmt.Errorf("%s: identity %q must be a Linear identity with webhook_secret_file, on a linear endpoint", where, e.Identity)
			}
		}
		if len(e.Repos) > 0 && e.Source != IngestGitHub {
			return fmt.Errorf("%s: repos is for github endpoints", where)
		}
		for _, r := range e.Repos {
			if !RepoPattern.MatchString(r) {
				return fmt.Errorf("%s: repo %q must be owner/name", where, r)
			}
		}
		p := e.EndpointPath()
		if !strings.HasPrefix(p, "/") || strings.ContainsAny(p, " {}") {
			return fmt.Errorf("%s: path %q must start with / and hold no spaces or braces", where, p)
		}
		if other, ok := paths[p]; ok {
			return fmt.Errorf("%s: path %s already used by %q", where, p, other)
		}
		paths[p] = e.Name
	}
	// Implicit endpoints share /linear/webhook; an explicit endpoint
	// there would shadow them.
	for _, e := range c.IngestEndpoints() {
		if e.Implicit {
			if other, ok := paths[LinearWebhookPath]; ok {
				return fmt.Errorf("%s: ingest.endpoints %q uses %s, which identity %q's webhook secret also serves; name the identity in that endpoint instead", FileName, other, LinearWebhookPath, e.Identity)
			}
			break
		}
	}
	return nil
}

// DefaultIngestListen is the ingest serve address when unset.
const DefaultIngestListen = "127.0.0.1:8740"

// ListenAddr is Listen or the default.
func (i Ingest) ListenAddr() string {
	if i.Listen == "" {
		return DefaultIngestListen
	}
	return i.Listen
}

// SpoolPath is the spool database path under root.
func (i Ingest) SpoolPath(r Root) string {
	if i.Spool == "" {
		return filepath.Join(r.Dir, "ingest", "spool.db")
	}
	return r.Path(i.Spool)
}

// DefaultGitHubReconcile is the reconcile interval when unset.
const DefaultGitHubReconcile = 15 * time.Minute

// ReconcileEvery is the GitHub reconcile interval; 0 is off.
func (i Ingest) ReconcileEvery() time.Duration {
	if i.GitHubReconcile == "off" {
		return 0
	}
	if d, err := time.ParseDuration(i.GitHubReconcile); err == nil && d > 0 {
		return d
	}
	return DefaultGitHubReconcile
}

// RetentionDuration is Retention parsed, or the default.
func (i Ingest) RetentionDuration() time.Duration {
	d, _ := parseDays(i.Retention)
	if d <= 0 {
		return DefaultInboxRetention
	}
	return d
}

// parseDays is time.ParseDuration that also takes "Nd".
func parseDays(s string) (time.Duration, error) {
	if n, ok := strings.CutSuffix(s, "d"); ok {
		var days int
		if _, err := fmt.Sscanf(n, "%d", &days); err != nil || fmt.Sprint(days) != n {
			return 0, fmt.Errorf("bad day count %q", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

// Limits are host-wide caps (#281).
type Limits struct {
	// MaxSessions caps supervisor-started sessions across all ticket
	// pools. 0 means no host cap.
	MaxSessions int `yaml:"max_sessions"`
}

// Linear configures Linear push delivery (#300).
type Linear struct {
	// WebhookListen is the TCP address for POST /linear/webhook. Empty:
	// off, intake and replies are polled only. Linear must reach it, so a
	// host with no public endpoint needs a reverse proxy or tunnel.
	WebhookListen string `yaml:"webhook_listen"`
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
	// Model, Effort and ConfigOptions are ACP session config options set
	// on every session of the agent (#304); see SessionOptions. They are
	// not part of the host spec, so changing them never respawns a host.
	Model         string            `yaml:"model"`
	Effort        string            `yaml:"effort"`
	ConfigOptions map[string]string `yaml:"config_options"`
	// Sandbox confines the agent (#262); pools add to it.
	Sandbox *Sandbox `yaml:"sandbox"`
}

// ACP session config option ids for the model and effort shorthands.
const (
	OptionModel  = "model"
	OptionEffort = "reasoning_effort"
)

// Options merges extra with the model and effort shorthands.
func Options(model, effort string, extra map[string]string) map[string]string {
	out := map[string]string{}
	maps.Copy(out, extra)
	if model != "" {
		out[OptionModel] = model
	}
	if effort != "" {
		out[OptionEffort] = effort
	}
	return out
}

// SessionOptions is the ACP config options a session of pool running
// agent gets: the agent's, overridden per id by the pool's. Values are
// passed to the agent as given (any jcode route), not validated.
func (c *Config) SessionOptions(pool, agent string) map[string]string {
	out := map[string]string{}
	if a, ok := c.Agents[agent]; ok {
		maps.Copy(out, Options(a.Model, a.Effort, a.ConfigOptions))
	}
	if p, ok := c.Pools[pool]; ok {
		maps.Copy(out, Options(p.Model, p.Effort, p.ConfigOptions))
	}
	return out
}

// PoolChat is Pool.Type for chat pools.
const PoolChat = "chat"

// PoolAdminMethods are the admin-socket methods a pool's sessions may be
// granted with Pool.Admin. Ticket mutation and session lifecycle methods
// that act on other work (reopen, session new/send/reap, shutdown) are not
// grantable.
var PoolAdminMethods = []string{"config-reload", "config-show", "session-list", "session-restart", "drain", "trigger-list", "pool-list", "pool-unquarantine", "pool-scale", "session-cordon", "session-config", "report-incidents", "attention", "activity"}

// Allows reports whether the pool's sessions may call admin method m.
func (p Pool) Allows(m string) bool { return slices.Contains(p.Admin, m) }

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
	// MinSessions are kept open with no ready work (a warm pool, #281).
	// Default 0. Must be <= max_sessions.
	MinSessions int `yaml:"min_sessions"`
	// ScaleDownAfter is how long demand must stay below the open count
	// before surplus sessions are cordoned (Go duration). Default 10m.
	ScaleDownAfter string `yaml:"scale_down_after"`
	// Priority orders pools when limits.max_sessions caps the host: lower
	// gets its sessions first. Default 0.
	Priority int `yaml:"priority"`
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
	// Admin lists admin methods (PoolAdminMethods) this pool's session
	// sockets may call, e.g. [config-reload, session-list] for an
	// operator chat pool. Default none.
	Admin []string `yaml:"admin"`
	// Identities the pool's sessions may act as, on top of their
	// project's (#200). Lets a pack name the identity it needs.
	Identities []string `yaml:"identities"`
	// Model, Effort and ConfigOptions override the agent's per option id
	// (#304). A reload applies a change to live sessions between turns.
	Model         string            `yaml:"model"`
	Effort        string            `yaml:"effort"`
	ConfigOptions map[string]string `yaml:"config_options"`
	// Sandbox overrides and adds to the agent's sandbox (#262).
	Sandbox *Sandbox `yaml:"sandbox"`
}

// Project is one repo / work area.
type Project struct {
	// Prefix starts every ticket id, e.g. "js" gives js-0031w02j0r012.
	// Default: the project name.
	Prefix string `yaml:"prefix"`
	// Identities the project's sessions may act as (#200).
	Identities []string `yaml:"identities"`
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// RepoPattern is a GitHub owner/name.
var RepoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

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
	for name, i := range c.Identities {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: identity name %q must match %s", FileName, name, namePattern)
		}
		if err := checkIdentity(i); err != nil {
			return nil, fmt.Errorf("%s: identity %s: %w", FileName, name, err)
		}
	}
	for name, p := range c.Projects {
		for _, id := range p.Identities {
			if _, ok := c.Identities[id]; !ok {
				return nil, fmt.Errorf("%s: project %s: unknown identity %q", FileName, name, id)
			}
		}
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
		if err := a.Sandbox.check(); err != nil {
			return nil, fmt.Errorf("%s: agent %s: %w", src("agents", name), name, err)
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
		if err := p.Sandbox.check(); err != nil {
			return nil, fmt.Errorf("%s: pool %s: %w", src("pools", name), name, err)
		}
		if p.Type != "" && p.Type != PoolChat {
			return nil, fmt.Errorf("%s: pool %s: type %q must be empty or %q", src("pools", name), name, p.Type, PoolChat)
		}
		if p.Attach != "" && p.Attach != AttachRW {
			return nil, fmt.Errorf("%s: pool %s: attach %q must be empty or %q", src("pools", name), name, p.Attach, AttachRW)
		}
		// An unknown admin method is dropped with a warning, not fatal: a
		// pack or last good config written for a newer binary must not stop
		// an older one from starting (#325). The grant is simply not given.
		var admin []string
		for _, m := range p.Admin {
			if !slices.Contains(PoolAdminMethods, m) {
				c.Warnings = append(c.Warnings, fmt.Sprintf("%s: pool %s: admin method %q is unknown to this binary, ignored (known: %s)", src("pools", name), name, m, strings.Join(PoolAdminMethods, ", ")))
				continue
			}
			admin = append(admin, m)
		}
		if len(admin) != len(p.Admin) {
			p.Admin = admin
		}
		for _, id := range p.Identities {
			if _, ok := c.Identities[id]; !ok {
				return nil, fmt.Errorf("%s: pool %s: unknown identity %q", src("pools", name), name, id)
			}
		}
		if p.MaxSessions < 0 {
			return nil, fmt.Errorf("%s: pool %s: max_sessions must be >= 0", src("pools", name), name)
		}
		if p.MinSessions < 0 || p.MinSessions > p.MaxSessions {
			return nil, fmt.Errorf("%s: pool %s: min_sessions must be between 0 and max_sessions", src("pools", name), name)
		}
		if p.MinSessions > 0 && p.Chat() {
			return nil, fmt.Errorf("%s: pool %s: min_sessions does not apply to chat pools", src("pools", name), name)
		}
		if p.ConversationExport && c.Telemetry.LogsURL() == "" {
			return nil, fmt.Errorf("%s: pool %s: conversation_export needs telemetry.otlp_endpoint or otlp_logs_endpoint", src("pools", name), name)
		}
		for _, d := range []struct{ key, v string }{{"idle_timeout", p.IdleTimeout}, {"turn_timeout", p.TurnTimeout}, {"scale_down_after", p.ScaleDownAfter}} {
			if d.v == "" {
				continue
			}
			if v, err := time.ParseDuration(d.v); err != nil || v <= 0 {
				return nil, fmt.Errorf("%s: pool %s: %s %q must be a positive duration", src("pools", name), name, d.key, d.v)
			}
		}
		c.Pools[name] = p
	}
	if err := c.Defaults.Sandbox.check(); err != nil {
		return nil, fmt.Errorf("%s: defaults: %w", FileName, err)
	}
	for _, d := range c.SharedTools {
		if !filepath.IsAbs(d) {
			return nil, fmt.Errorf("%s: shared_tools: %q must be absolute", FileName, d)
		}
	}
	if err := c.validateIngest(); err != nil {
		return nil, err
	}
	if err := c.validateAdmission(); err != nil {
		return nil, err
	}
	if r := c.Ingest.Retention; r != "" {
		if d, err := parseDays(r); err != nil || d <= 0 {
			return nil, fmt.Errorf("%s: ingest.retention %q must be a positive duration (e.g. 14d, 72h)", FileName, r)
		}
	}
	if c.Limits.MaxSessions < 0 {
		return nil, fmt.Errorf("%s: limits.max_sessions must be >= 0", FileName)
	}
	if s := c.Subscriptions; s != nil {
		for _, d := range []struct{ key, v string }{{"interval", s.Interval}, {"resume_window", s.ResumeWindow}} {
			if d.v == "" {
				continue
			}
			if v, err := time.ParseDuration(d.v); err != nil || v <= 0 {
				return nil, fmt.Errorf("%s: subscriptions: %s %q must be a positive duration", FileName, d.key, d.v)
			}
		}
	}
	if a := c.Activity; a != nil {
		for _, d := range []struct{ key, v string }{{"interval", a.Interval}, {"window", a.Window}} {
			if d.v == "" {
				continue
			}
			if v, err := time.ParseDuration(d.v); err != nil || v <= 0 {
				return nil, fmt.Errorf("%s: activity: %s %q must be a positive duration", FileName, d.key, d.v)
			}
		}
		if a.MaxItems < 0 {
			return nil, fmt.Errorf("%s: activity: max_items must not be negative", FileName)
		}
		for _, r := range a.Repos {
			if !RepoPattern.MatchString(r) {
				return nil, fmt.Errorf("%s: activity: repo %q must be owner/name", FileName, r)
			}
		}
	}
	for name, g := range c.Gateway {
		if err := c.checkGateway(g); err != nil {
			return nil, fmt.Errorf("%s: gateway %s: %w", FileName, name, err)
		}
	}
	for name, tr := range c.Triggers {
		if err := c.checkTrigger(name, &tr); err != nil {
			return nil, fmt.Errorf("%s: trigger %s: %w", src("triggers", name), name, err)
		}
		c.Triggers[name] = tr
	}
	for i, h := range c.Hooks {
		if err := c.checkHook(h); err != nil {
			return nil, fmt.Errorf("%s: hooks[%d]: %w", FileName, i, err)
		}
	}
	if err := c.checkResources(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	for name, w := range c.Workflows {
		if err := c.checkWorkflow(name, &w); err != nil {
			return nil, fmt.Errorf("%s: workflow %s: %w", src("workflows", name), name, err)
		}
	}
	if err := c.checkRules(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) checkGateway(g Gateway) error {
	switch g.Network {
	case GatewayMatrix:
		if g.Homeserver == "" {
			return errors.New("homeserver is required")
		}
	case GatewayDiscord:
	default:
		return fmt.Errorf("network %q must be %q or %q", g.Network, GatewayMatrix, GatewayDiscord)
	}
	if g.TokenFile == "" {
		return errors.New("token_file is required")
	}
	p, ok := c.Pools[g.Pool]
	if !ok {
		return fmt.Errorf("unknown pool %q", g.Pool)
	}
	if !p.Chat() {
		return fmt.Errorf("pool %s is not a chat pool", g.Pool)
	}
	if len(g.Rooms) == 0 || len(g.Senders) == 0 {
		return errors.New("rooms and senders are required (others are ignored)")
	}
	if g.EscalationRoom != "" && !slices.Contains(g.Rooms, g.EscalationRoom) {
		return fmt.Errorf("escalation_room %q must be listed in rooms", g.EscalationRoom)
	}
	if g.AttentionDigest != "" {
		if g.EscalationRoom == "" {
			return errors.New("attention_digest needs escalation_room")
		}
		if _, err := cron.Parse(g.AttentionDigest); err != nil {
			return fmt.Errorf("attention_digest: %w", err)
		}
	}
	return nil
}

func (c *Config) checkTrigger(name string, tr *Trigger) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("name must match %s", namePattern)
	}
	w := tr.When
	if len(w.Stall) > 0 {
		return c.checkStallTrigger(tr)
	}
	if tr.MaxRuns != 0 {
		return errors.New("max_runs needs when.stall")
	}
	if w.Cron == "" && !w.HasCounts() && w.Check == "" && w.Learnings == nil && w.Edits == nil {
		return errors.New("when needs a cron, a count threshold, a check, learnings or edits")
	}
	if e := w.Edits; e != nil {
		if e.Runs < 0 {
			return errors.New("when: edits.runs must be >= 0")
		}
		if e.Age != "" {
			if d, err := time.ParseDuration(e.Age); err != nil || d <= 0 {
				return fmt.Errorf("when: edits.age %q must be a positive duration", e.Age)
			}
		}
	}
	if l := w.Learnings; l != nil {
		if l.Min < 1 {
			return errors.New("when: learnings.min must be >= 1")
		}
		if l.Every != "" {
			if d, err := time.ParseDuration(l.Every); err != nil || d <= 0 {
				return fmt.Errorf("when: learnings.every %q must be a positive duration", l.Every)
			}
		}
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

// ActivityRepos lists the GitHub repos the activity poller reads: every
// workflow repo and activity.repos, sorted and unique.
func (c *Config) ActivityRepos() []string {
	var out []string
	for _, w := range c.Workflows {
		if RepoPattern.MatchString(w.Repo) {
			out = append(out, w.Repo)
		}
	}
	if c.Activity != nil {
		out = append(out, c.Activity.Repos...)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// ActivityTeams lists the Linear team keys of all Linear identities,
// sorted and unique. "*" is skipped: it names no team to poll.
func (c *Config) ActivityTeams() []string {
	var out []string
	for _, i := range c.Identities {
		if !IsLinear(i.Kind) {
			continue
		}
		for _, t := range i.Teams {
			if t != "*" {
				out = append(out, t)
			}
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
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

// BuiltinMetaPacks are the meta pack names jobshop itself writes (the
// controller, and the gateway on escalation tickets).
var BuiltinMetaPacks = []string{"github", "linear", "trigger", "run", "gateway", "pool", "hook", "subscribe", "report", "question"}

// KnownMetaPack reports whether name may be a top-level meta key: a
// built-in pack or one listed under meta_packs.
func (c *Config) KnownMetaPack(name string) bool {
	return slices.Contains(BuiltinMetaPacks, name) || slices.Contains(c.MetaPacks, name)
}
