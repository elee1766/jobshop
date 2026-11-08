// Package config loads jobshop.yaml and locates the Jobshop root directory.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/elee1766/jobshop/pkg/cron"
	"github.com/elee1766/jobshop/pkg/ids"
	"github.com/elee1766/jobshop/pkg/store"
)

// Config is the resolved service and work configuration.
type Config struct {
	// HookDigests are populated by the controller when it compiles a live snapshot.
	HookDigests map[string]string `yaml:"-" json:"-"`
	// HookRunPins are the run pins of callback workflows, by workflow name,
	// populated with HookDigests (see hooks.RunPin).
	HookRunPins map[string]string `yaml:"-" json:"-"`
	// Pack names the root work-definition file, relative to jobshop.yaml.
	Pack string `yaml:"pack,omitempty"`
	// Automation grants host capabilities to trusted Starlark hooks.
	Automation Automation `yaml:"automation,omitempty"`
	// Jev configures the optional model-decision hook capability.
	Jev Jev `yaml:"jev,omitempty"`
	// DefaultProject is used when a request names no project.
	DefaultProject string             `yaml:"default_project"`
	Projects       map[string]Project `yaml:"projects"`
	// IDs picks the ticket id generator (JOB-57); a project's ids block
	// overrides it. Default sonyflake.
	IDs ids.Spec `yaml:"ids"`
	// Lease is the claim lease (Go duration). Default 30m.
	Lease     string              `yaml:"lease"`
	Agents    map[string]Agent    `yaml:"agents"`
	Pools     map[string]Pool     `yaml:"pools"`
	Triggers  map[string]Trigger  `yaml:"triggers"`
	Workflows map[string]Workflow `yaml:"workflows"`
	// Escalation is the escalation chat for workflows without an
	// escalate block and for pool quarantines (#379). Absent: no chat.
	Escalation *EscalationChat `yaml:"escalation"`
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
	// WriteKeys configures the local idempotency key cache (#415).
	WriteKeys WriteKeys `yaml:"write_keys"`
	// Redact configures secret scanning (#578). Host-only.
	Redact Redact `yaml:"redact"`
	// Reload configures watch-and-reload of this file and its packs (#533).
	Reload ReloadConfig `yaml:"reload"`
	// Artifacts configures delivered artifacts' attachments (#508).
	Artifacts Artifacts `yaml:"artifacts"`
	// Journal configures journal shipping and search (#530).
	Journal Journal `yaml:"journal"`
	// Store selects the ticket store backend (#493).
	Store Store `yaml:"store"`
	// Defaults hold settings agents and pools inherit (#262).
	Defaults Defaults `yaml:"defaults"`
	// SharedTools are dirs mounted read-only into every sandboxed
	// session (#262).
	SharedTools []string `yaml:"shared_tools"`
	// SandboxProfiles are named sandbox fragments (system mounts, env,
	// network) that sandbox profiles: lists apply (#625). Packs
	// contribute entries like pools.
	SandboxProfiles map[string]Sandbox `yaml:"sandbox_profiles,omitempty"`
	// IsolationProfiles are named isolation fragments (user mapping,
	// groups, limits) that isolation profiles: lists apply (JOB-83).
	// Packs contribute entries like pools.
	IsolationProfiles map[string]Isolation `yaml:"isolation_profiles,omitempty"`
	// Attention configures `jobshop attention` (#185).
	Attention Attention `yaml:"attention"`
	// Hooks are guardrail commands (#188). Default none.
	Hooks []Hook `yaml:"hooks"`
	// Learnings configures the learning loop (#524).
	Learnings LearningsConfig `yaml:"learnings,omitempty"`
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
	// Desk configures the first-party Desk and its generic RPC endpoint.
	Desk Desk `yaml:"desk,omitempty"`
	// DeskCommands are commands Desk plugins run by name (#542). Packs
	// contribute entries like pools.
	DeskCommands map[string]DeskCommand `yaml:"desk_commands,omitempty"`
	// Packs are pack directories merged before this file (see LoadFile).
	Packs []PackRef `yaml:"packs"`
	// Instances summarize each Packs entry as loaded (#248). Not kept in
	// the last good config.
	Instances []PackInstance `yaml:"-"`
	// Warnings are non-fatal load findings, such as a deprecated pack var
	// name the host still sets (#261). The controller logs them.
	Warnings []string `yaml:"-"`
	// Watch is what a config watcher follows (#533), recorded at load.
	Watch WatchPaths `yaml:"-"`
	// MetaPacks are extra allowed top-level ticket meta keys, beyond
	// BuiltinMetaPacks.
	MetaPacks []string `yaml:"meta_packs"`
	// Sources maps "pools.<name>", "agents.<name>", "triggers.<name>",
	// "workflows.<name>" and "prompt_rules.<name>" to the files that set
	// the entry, in merge order.
	Sources map[string][]string `yaml:"-"`
	// Disabled lists the entries set enabled: false (JOB-24). They were
	// validated, then removed from Pools, Workflows and Triggers.
	Disabled Disabled `yaml:"-"`
}

// Trigger creates a ticket from Ticket, or starts a workflow run from Run,
// while every condition in When holds and nothing it created is open.
type Trigger struct {
	Handler string         `yaml:"handler,omitempty"`
	Args    map[string]any `yaml:"args,omitempty"`
	// Enabled false keeps the trigger validated but not loaded (JOB-24).
	Enabled *bool          `yaml:"enabled,omitempty"`
	When    TriggerWhen    `yaml:"when"`
	Ticket  *TriggerTicket `yaml:"ticket"`
	Run     *TriggerRun    `yaml:"run"`
	// Cooldown is the minimum time after its last ticket closed before the
	// trigger fires again (Go duration).
	Cooldown string `yaml:"cooldown"`
	// MaxRuns caps the runs a stall trigger starts per key (one stall);
	// after that the stall is left to the escalation of the last run.
	// Default 3. Stall triggers only.
	MaxRuns int `yaml:"max_runs"`
	// MaxOpen caps the open runs of a stall trigger across all keys, so a
	// backlog of stalls is worked a few at a time (JOB-93). Default 3.
	// Stall triggers only.
	MaxOpen int `yaml:"max_open"`
	// On fires the trigger when an admitted ingest delivery matches one
	// entry (JOB-62), between ticks. When conditions still apply, and
	// dedup is the same as a poll's, so an event and a poll of the same
	// thing fire once. Event fields are the template var event.
	On []TriggerEvent `yaml:"on"`
	// Fallback is how often the trigger is polled (Go duration). Empty:
	// every tick, or DefaultTriggerFallback with on:. A trigger with on:
	// and no when conditions is never polled.
	Fallback string `yaml:"fallback"`
	// Command runs a host command instead of creating a ticket or run.
	Command *TriggerCommand `yaml:"command"`
}

// DefaultTriggerFallback is the poll interval of a trigger with on: and
// no fallback: a lost webhook only delays it.
const DefaultTriggerFallback = 10 * time.Minute

// TriggerEvent matches ingest deliveries from one source. Set exactly
// one of GitHub and Linear.
type TriggerEvent struct {
	Source string      `yaml:"source,omitempty"`
	Event  string      `yaml:"event,omitempty"`
	GitHub *EventMatch `yaml:"github"`
	Linear *EventMatch `yaml:"linear"`
}

// EventMatch lists fields a delivery must have; empty fields match any.
// Values compare case-insensitively. Workflow matches the workflow name
// or its file name without extension (ci for .github/workflows/ci.yml);
// Label matches any label on the delivery. Repo is owner/name on GitHub,
// the team key on Linear.
type EventMatch struct {
	Event      string `yaml:"event"`
	Action     string `yaml:"action"`
	Repo       string `yaml:"repo"`
	Branch     string `yaml:"branch"`
	Workflow   string `yaml:"workflow"`
	Conclusion string `yaml:"conclusion"`
	Label      string `yaml:"label"`
}

// EventVar is the template var holding a firing event's fields.
const EventVar = "event"

// EventFields are the keys of the event var.
var EventFields = []string{"source", "event", "action", "repo", "branch", "workflow", "conclusion", "head_sha", "pr", "label", "delivery"}

// TriggerCommand is the command action: sh -c Run in the root dir. Run
// is a template over check_output and event. Each fire is recorded as a
// message ticket that holds the trigger while the command runs; a
// failure keeps it open with the needs-human label, so the trigger stays
// held until a human closes it.
type TriggerCommand struct {
	Run     string `yaml:"run"`
	Timeout string `yaml:"timeout"` // Go duration, default 1m
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
	StallResourceWait    = "resource_wait"    // open, unblocked ticket waiting on a full resource for After (#408)
	StallIngestWorker    = "ingest_worker"    // ingest.workers external and no worker heartbeat for ingest.worker_missing (#443)
	StallIngestBacklog   = "ingest_backlog"   // over ingest.backlog_max pending inbox items older than ingest.backlog_after (#443)
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
var StoreStallSignals = []string{StallReadyUnclaimed, StallPoolQuarantined, StallStepStalled, StallSessionSilent, StallNeedsHuman, StallBlocked, StallResourceWait, StallIngestWorker, StallIngestBacklog}

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
	// Grace skips issues updated (moved to Todo, edited) within it (JOB-26).
	// Go duration; empty is off, so a fresh Todo issue starts at once. The
	// fire-time store recheck already covers a run the planner just started.
	Grace string `yaml:"grace"`
}

// TriggerRun starts a workflow run, like `jobshop run start`. Vars and
// Key are templates; {{.check_output}} is the first line of the check's
// trimmed stdout. An open run with the same workflow and key dedups.
type TriggerRun struct {
	Workflow string            `yaml:"workflow"`
	Vars     map[string]string `yaml:"vars"`
	Key      string            `yaml:"key"`
	// KeyPrefix replaces "stall" in a stall trigger's run keys,
	// <prefix>:<signal>:<subject>, for triggers that do intake rather
	// than stuck-work checks (JOB-95). Stall triggers only, not with Key.
	KeyPrefix string `yaml:"key_prefix"`
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
	// its own run with key stall:<signal>:<subject> (run.key_prefix
	// replaces stall, JOB-95). A stall trigger has
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
	// Filters (#524), all optional and ANDed: kind globs, scope globs
	// (global, repo:<owner/name>, pool:<name>) and source (agent or
	// controller). Min and Every count only matching learnings.
	Kinds  []string `yaml:"kinds,omitempty"`
	Scope  []string `yaml:"scope,omitempty"`
	Source string   `yaml:"source,omitempty"`
}

// LearningsConfig is the top-level learnings section (#524).
type LearningsConfig struct {
	// Kinds are the learning kinds agents may record: names or
	// path.Match globs (review-*, *). Packs add to the list. Empty: the
	// defaults, correction, failure and workaround.
	Kinds []string `yaml:"kinds,omitempty"`
}

// LearningKinds is the resolved kind patterns: Learnings.Kinds, or the
// defaults when unset.
func (c *Config) LearningKinds() []string {
	if len(c.Learnings.Kinds) == 0 {
		return store.LearningKinds
	}
	return c.Learnings.Kinds
}

// HasConditions reports whether w has any condition a poll evaluates.
func (w TriggerWhen) HasConditions() bool {
	return w.Cron != "" || w.HasCounts() || w.Check != "" || w.Learnings != nil || w.Edits != nil || len(w.Stall) > 0
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
	// Bots lists bot logins (exact, with the [bot] suffix on GitHub, the
	// app actor id or name on Linear) whose comments and reviews are delivered
	// like a human's (#402). The host's own identities are dropped even
	// when listed. Authors does not apply to them.
	Bots []string `yaml:"bots"`
	// Authors, when set, are the only logins whose comments are
	// delivered (#340); the public repo's other commenters are dropped.
	// On Linear an entry is a user id, displayName or email; a full name
	// still matches but logs a warning (#595).
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

// Gateway.Network values. Matrix and Discord were removed (2026-11-06):
// Desk is the operator UI.
const (
	// GatewayA2A routes escalations to another agent over A2A (#547).
	GatewayA2A = "a2a"
)

// EscalationCauses are meta.escalation.cause values, the kinds an A2A
// gateway's escalations list names (the cause var of EscalationVars).
var EscalationCauses = []string{"run", "pool", "wait", "guard", "label"}

// Gateway is one A2A peer that escalations are sent to (#547).
type Gateway struct {
	Network string `yaml:"network"` // "a2a"
	// TokenFile holds the peer's bearer token. Tokens are never in the yaml.
	TokenFile Secret `yaml:"token_file"`
	Pool      string `yaml:"pool"` // a type: chat pool
	// DeskURL is Desk's base URL for links in messages. Empty: no link.
	DeskURL string `yaml:"desk_url"`
	// URL is the A2A peer's base URL; its agent card is fetched from
	// /.well-known/agent-card.json under it.
	URL string `yaml:"url,omitempty"`
	// Escalations are the escalation causes sent to the A2A peer: "all"
	// or a list of EscalationCauses. Empty: none.
	Escalations []string `yaml:"escalations,omitempty"`
	// Poll is the GetTask interval for a peer task (Go duration).
	// Default 30s.
	Poll string `yaml:"poll,omitempty"`
}

// Desk is the desk section. Desk reads it through config.show, so a
// reload applies without restarting Desk.
type Desk struct {
	// Plugins is rejected with a migration message rather than silently ignored.
	Plugins map[string]any `yaml:"plugins,omitempty" json:"plugins,omitempty"`
	// RPC narrows (deny) or widens (allow) methods API clients may call.
	RPC DeskRPC `yaml:"rpc,omitempty" json:"rpc"`
}

// DeskRPC holds path.Match globs over admin method names.
type DeskRPC struct {
	Allow []string `yaml:"allow,omitempty" json:"allow,omitempty"`
	Deny  []string `yaml:"deny,omitempty" json:"deny,omitempty"`
}

func (d Desk) validate() error {
	if len(d.Plugins) > 0 {
		return errors.New("desk.plugins is no longer supported: remove it and use the controller API with packages/jobshop-client for custom frontends")
	}
	for _, g := range slices.Concat(d.RPC.Allow, d.RPC.Deny) {
		if _, err := path.Match(g, ""); err != nil {
			return fmt.Errorf("desk.rpc: bad glob %q: %w", g, err)
		}
	}
	return nil
}

// Attention configures the attention view: what waits on a human.
type Attention struct {
	// RollbackFile and ConfigRepoFile were read for deploy attention
	// items, removed by #577: deploy tools report through needs-human
	// tickets. Still parsed so a load can name the replacement.
	RollbackFile   string `yaml:"rollback_file"`
	ConfigRepoFile string `yaml:"config_repo_file"`
	// ResourceWait lists a step waiting on a full resource longer than
	// this Go duration (#408). Empty: not listed.
	ResourceWait string `yaml:"resource_wait"`
	// OrphanAge lists an open ticket no pool will claim (no pool, or a
	// pool not in the config) once it is older than this Go duration
	// (#495). Default 72h; "off" disables it.
	OrphanAge string `yaml:"orphan_age"`
	// Prestart starts an item's resolving chat when it enters attention,
	// so a click in Desk attaches to a session that already answered
	// (#606). Default true when PrestartPool resolves to a chat pool.
	Prestart *bool `yaml:"prestart"`
	// PrestartPool is the chat pool pre-started and clicked attention
	// chats run in. Default "operator" when that pool is a chat pool.
	PrestartPool string `yaml:"prestart_pool"`
	// PrestartMax caps pre-started chats open at once; items past it
	// start on click. Default 4.
	PrestartMax int `yaml:"prestart_max"`
}

// DefaultPrestartMax is attention.prestart_max when unset.
const DefaultPrestartMax = 4

// AttentionChatPool is the pool attention chats run in, "" when none:
// prestart_pool, else operator when it is a chat pool.
func (c *Config) AttentionChatPool() string {
	if c.Attention.PrestartPool != "" {
		return c.Attention.PrestartPool
	}
	if p, ok := c.Pools["operator"]; ok && p.Chat() {
		return "operator"
	}
	return ""
}

// PrestartOn reports whether attention chats are pre-started.
func (c *Config) PrestartOn() bool {
	if c.Attention.Prestart != nil && !*c.Attention.Prestart {
		return false
	}
	return c.AttentionChatPool() != ""
}

// PrestartLimit is attention.prestart_max with its default.
func (a Attention) PrestartLimit() int {
	if a.PrestartMax > 0 {
		return a.PrestartMax
	}
	return DefaultPrestartMax
}

// DefaultOrphanAge is attention.orphan_age when unset.
const DefaultOrphanAge = 72 * time.Hour

// OrphanAgeDuration is attention.orphan_age, 0 when off.
func (a Attention) OrphanAgeDuration() time.Duration {
	if a.OrphanAge == "off" {
		return 0
	}
	return Duration(a.OrphanAge, DefaultOrphanAge)
}

// DefaultInboxRetention is how long terminal inbox items are kept.
const DefaultInboxRetention = 7 * 24 * time.Hour

// Ingest configures the ingest inbox (#321).
type Ingest struct {
	// Retention keeps dispatched, denied and failed items this long
	// after their last update (Go duration, or Nd for days). Default 7d
	// (JOB-63). Deliveries are deduped only within it.
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
	// Pollers read GitHub repos that send no webhook (#341) and write
	// inbox rows shaped like webhook deliveries.
	Pollers []IngestPoller `yaml:"pollers"`
	// Workers is where admission runs (#443): internal (default) in the
	// controller, external in `jobshop ingest work` processes.
	Workers string `yaml:"workers"`
	// Dispatch is where admitted items are dispatched (#633): internal
	// (default) in the controller, external in `jobshop ingest work
	// --dispatch` processes. Several dispatchers are safe either way.
	Dispatch string `yaml:"dispatch"`
	// WorkerMissing lists external mode with no worker heartbeat for
	// this Go duration as attention. Empty: 5m.
	WorkerMissing string `yaml:"worker_missing"`
	// BacklogAfter and BacklogMax list attention when more than
	// BacklogMax pending items are older than BacklogAfter. Empty: 10m
	// and 0.
	BacklogAfter string `yaml:"backlog_after"`
	BacklogMax   int    `yaml:"backlog_max"`

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
	SecretFile Secret `yaml:"secret_file"`
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
		if !IsLinear(id.Kind) || id.WebhookSecretFile.IsZero() || named[name] {
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
	if i.Workers != "" && i.Workers != IngestWorkersInternal && i.Workers != IngestWorkersExternal {
		return fmt.Errorf("%s: ingest.workers %q must be internal or external", FileName, i.Workers)
	}
	if i.Dispatch != "" && i.Dispatch != IngestWorkersInternal && i.Dispatch != IngestWorkersExternal {
		return fmt.Errorf("%s: ingest.dispatch %q must be internal or external", FileName, i.Dispatch)
	}
	for k, v := range map[string]string{"worker_missing": i.WorkerMissing, "backlog_after": i.BacklogAfter} {
		if d, err := time.ParseDuration(v); v != "" && (err != nil || d <= 0) {
			return fmt.Errorf("%s: ingest.%s %q must be a positive duration", FileName, k, v)
		}
	}
	if i.BacklogMax < 0 {
		return fmt.Errorf("%s: ingest.backlog_max must not be negative", FileName)
	}
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
		if e.SecretFile.IsZero() == (e.Identity == "") {
			return fmt.Errorf("%s: set exactly one of secret_file and identity", where)
		}
		if e.Identity != "" {
			id, ok := c.Identities[e.Identity]
			if !ok || !IsLinear(id.Kind) || id.WebhookSecretFile.IsZero() || e.Source != IngestLinear {
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
	if err := c.validatePollers(names); err != nil {
		return err
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

// IngestPoller polls GitHub repos into the inbox (#341).
type IngestPoller struct {
	Name   string   `yaml:"name"`
	Source string   `yaml:"source"` // github
	Repos  []string `yaml:"repos"`
	// Events are webhook event names: issues, issue_comment,
	// pull_request, pull_request_review. Empty: all four.
	Events []string `yaml:"events"`
	// Interval between polls, default 1m, at least 10s.
	Interval string `yaml:"interval"`
	// Identity is a github-app or github-token identity covering the
	// repos. Empty: the identity that covers each repo.
	Identity string `yaml:"identity"`
}

// PollerEvents are the events an ingest poller can read.
var PollerEvents = []string{"issues", "issue_comment", "pull_request", "pull_request_review"}

// DefaultPollInterval is a poller's interval when unset.
const DefaultPollInterval = time.Minute

// Every is the poll interval.
func (p IngestPoller) Every() time.Duration {
	if d, err := time.ParseDuration(p.Interval); err == nil && d > 0 {
		return d
	}
	return DefaultPollInterval
}

// EventSet is Events, or all of them when empty.
func (p IngestPoller) EventSet() []string {
	if len(p.Events) == 0 {
		return PollerEvents
	}
	return p.Events
}

func (c *Config) validatePollers(endpoints map[string]bool) error {
	names := map[string]bool{}
	for _, p := range c.Ingest.Pollers {
		where := fmt.Sprintf("%s: ingest.pollers %q", FileName, p.Name)
		if !endpointName.MatchString(p.Name) {
			return fmt.Errorf("%s: name must match [a-z0-9-]+", where)
		}
		if names[p.Name] || endpoints[p.Name] {
			return fmt.Errorf("%s: name already used by a poller or endpoint", where)
		}
		names[p.Name] = true
		if p.Source != IngestGitHub {
			return fmt.Errorf("%s: source must be github", where)
		}
		if len(p.Repos) == 0 {
			return fmt.Errorf("%s: needs repos", where)
		}
		for _, r := range p.Repos {
			if !RepoPattern.MatchString(r) {
				return fmt.Errorf("%s: repo %q must be owner/name", where, r)
			}
		}
		for _, e := range p.Events {
			if !slices.Contains(PollerEvents, e) {
				return fmt.Errorf("%s: event %q must be one of %s", where, e, strings.Join(PollerEvents, ", "))
			}
		}
		if p.Interval != "" {
			if d, err := time.ParseDuration(p.Interval); err != nil || d < 10*time.Second {
				return fmt.Errorf("%s: interval %q must be a duration of at least 10s", where, p.Interval)
			}
		}
		if p.Identity != "" {
			id, ok := c.Identities[p.Identity]
			if !ok || (id.Kind != IdentityGitHubApp && id.Kind != IdentityGitHubToken) {
				return fmt.Errorf("%s: identity %q must be a github-app or github-token identity", where, p.Identity)
			}
			for _, r := range p.Repos {
				if !id.CoversRepo(r) {
					return fmt.Errorf("%s: identity %q does not cover %s", where, p.Identity, r)
				}
			}
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

// Ingest worker modes (#443).
const (
	IngestWorkersInternal = "internal"
	IngestWorkersExternal = "external"
)

// External reports that admission runs in `jobshop ingest work`, not the controller.
func (i Ingest) External() bool { return i.Workers == IngestWorkersExternal }

// ExternalDispatch reports that the controller does not dispatch
// admitted items: `jobshop ingest work --dispatch` does (#633).
func (i Ingest) ExternalDispatch() bool { return i.Dispatch == IngestWorkersExternal }

// WorkerMissingAfter is WorkerMissing or 5m.
func (i Ingest) WorkerMissingAfter() time.Duration { return Duration(i.WorkerMissing, 5*time.Minute) }

// BacklogAge is BacklogAfter or 10m.
func (i Ingest) BacklogAge() time.Duration { return Duration(i.BacklogAfter, 10*time.Minute) }

// RetentionDuration is Retention parsed, or the default.
func (i Ingest) RetentionDuration() time.Duration {
	d, _ := parseDays(i.Retention)
	if d <= 0 {
		return DefaultInboxRetention
	}
	return d
}

// DefaultWriteKeyRetention is how long local write keys are kept (#415).
const DefaultWriteKeyRetention = 48 * time.Hour

// WriteKeys configures the write key cache (#415). The remote key or
// marker is the record; the local row only saves a remote check.
type WriteKeys struct {
	// Retention keeps keys this long, done or pending (Go duration or
	// Nd). Default 48h.
	Retention string `yaml:"retention"`
}

// RetentionDuration is Retention parsed, or the default.
func (w WriteKeys) RetentionDuration() time.Duration {
	d, _ := parseDays(w.Retention)
	if d <= 0 {
		return DefaultWriteKeyRetention
	}
	return d
}

// DefaultArtifactMaxBytes caps one artifact's attachments (#508).
const DefaultArtifactMaxBytes = 10 << 20

// DefaultArtifactRetention is how long attachments are kept after ack.
const DefaultArtifactRetention = 90 * 24 * time.Hour

// Artifacts configures artifact attachments under <root>/artifacts (#508).
type Artifacts struct {
	// MaxBytes caps the total attachment size of one artifact. Default 10MB.
	MaxBytes int64 `yaml:"max_bytes"`
	// Retention deletes an acked artifact's attachments this long after
	// the ack (Go duration or Nd). The ticket stays. Default 90d.
	Retention string `yaml:"retention"`
}

// MaxBytesOrDefault is MaxBytes, or the default when unset.
func (a Artifacts) MaxBytesOrDefault() int64 {
	if a.MaxBytes <= 0 {
		return DefaultArtifactMaxBytes
	}
	return a.MaxBytes
}

// RetentionDuration is Retention parsed, or the default.
func (a Artifacts) RetentionDuration() time.Duration {
	d, _ := parseDays(a.Retention)
	if d <= 0 {
		return DefaultArtifactRetention
	}
	return d
}

// DefaultJournalLagMax is when unshipped journal frames become an
// attention item (#530).
const DefaultJournalLagMax = 10 * time.Minute

// Journal configures journal shipping and search (#530).
type Journal struct {
	// VictoriaLogs is where frames are shipped and searched. Absent: no
	// shipping, and journal.search errors.
	VictoriaLogs *VictoriaLogs `yaml:"victorialogs"`
	// LagMax is the oldest unshipped frame age that raises attention
	// journal-lag (Go duration). Default 10m.
	LagMax string `yaml:"lag_max"`
	// Compress compresses closed sessions' journals with zstd (#631).
	Compress JournalCompress `yaml:"compress"`
}

// JournalCompress compresses a closed session's journal.ndjson to
// journal.ndjson.zst (acphost#74). Readers serve both forms.
type JournalCompress struct {
	// After is how long after a session closed its journal is compressed
	// (Go duration or Nd). Empty: off.
	After string `yaml:"after"`
	// Level is the zstd level, 1-22. 0: 3.
	Level int `yaml:"level"`
}

// AfterDuration is After parsed; 0 means off.
func (j JournalCompress) AfterDuration() time.Duration {
	d, _ := parseDays(j.After)
	return d
}

// VictoriaLogs is one VictoriaLogs instance.
type VictoriaLogs struct {
	// URL is the base URL, e.g. http://127.0.0.1:9428.
	URL string `yaml:"url"`
	// Headers go on every request. ${VAR} expands from the controller env.
	Headers map[string]string `yaml:"headers"`
	// RetentionHint is the instance's retention as set on its side, for
	// display only: jobshop does not delete rows.
	RetentionHint string `yaml:"retention_hint"`
}

// HeaderValues returns Headers with ${VAR} expanded by getenv.
func (v VictoriaLogs) HeaderValues(getenv func(string) string) map[string]string {
	h := make(map[string]string, len(v.Headers))
	for k, x := range v.Headers {
		h[k] = os.Expand(x, getenv)
	}
	return h
}

// LagMaxDuration is LagMax parsed, or the default.
func (j Journal) LagMaxDuration() time.Duration {
	d, _ := parseDays(j.LagMax)
	if d <= 0 {
		return DefaultJournalLagMax
	}
	return d
}

func (j Journal) check() error {
	if r := j.LagMax; r != "" {
		if d, err := parseDays(r); err != nil || d <= 0 {
			return fmt.Errorf("journal.lag_max %q must be a positive duration (e.g. 10m)", r)
		}
	}
	if r := j.Compress.After; r != "" {
		if d, err := parseDays(r); err != nil || d <= 0 {
			return fmt.Errorf("journal.compress.after %q must be a positive duration (e.g. 24h)", r)
		}
	}
	if l := j.Compress.Level; l < 0 || l > 22 {
		return fmt.Errorf("journal.compress.level %d must be 1-22 (0: 3)", l)
	}
	if v := j.VictoriaLogs; v != nil {
		u, err := url.Parse(v.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("journal.victorialogs.url %q must be an http(s) URL", v.URL)
		}
	}
	return nil
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
	// Cgroup splits the controller's delegated cgroup into the controller
	// and its sessions, so session load cannot starve the controller
	// (JOB-91). Nil: one cgroup, as before.
	Cgroup *HostCgroup `yaml:"cgroup"`
	// CgroupDriver is how session cgroups are made (#944): "cgroupfs"
	// (the default) under the controller's delegated cgroup, or "systemd":
	// a transient scope per session in SessionsSlice, from the manager the
	// controller runs under. systemd needs no Delegate= on the unit, so a
	// restart with live limited sessions works; the controller must be in
	// that manager's tree (a user unit, or root).
	CgroupDriver string `yaml:"cgroup_driver"`
	// SessionsSlice is the slice session scopes go in with the systemd
	// driver. Default DefaultSessionsSlice.
	SessionsSlice string `yaml:"sessions_slice"`
}

// Cgroup drivers for limits.cgroup_driver.
const (
	CgroupDriverCgroupfs = "cgroupfs"
	CgroupDriverSystemd  = "systemd"
	DefaultSessionsSlice = "jobshop-sessions.slice"
)

// Systemd reports whether session cgroups are systemd scopes.
func (l Limits) Systemd() bool { return l.CgroupDriver == CgroupDriverSystemd }

// Slice is SessionsSlice or its default.
func (l Limits) Slice() string {
	if l.SessionsSlice == "" {
		return DefaultSessionsSlice
	}
	return l.SessionsSlice
}

// checkDriver validates the driver and slice. It returns a warning when
// limits.cgroup sets values the controller cannot apply with systemd:
// those belong on the slice unit.
func (l Limits) checkDriver() (string, error) {
	switch l.CgroupDriver {
	case "", CgroupDriverCgroupfs, CgroupDriverSystemd:
	default:
		return "", fmt.Errorf("limits.cgroup_driver %q: want %s or %s", l.CgroupDriver, CgroupDriverCgroupfs, CgroupDriverSystemd)
	}
	if l.SessionsSlice != "" {
		if !l.Systemd() {
			return "", fmt.Errorf("limits.sessions_slice needs limits.cgroup_driver: systemd")
		}
		name, ok := strings.CutSuffix(l.SessionsSlice, ".slice")
		if !ok || name == "" || strings.ContainsAny(name, "/ ") {
			return "", fmt.Errorf("limits.sessions_slice %q: want a slice unit name such as %s", l.SessionsSlice, DefaultSessionsSlice)
		}
	}
	if hc := l.Cgroup; l.Systemd() && hc != nil && (hc.SessionsCPU != "" || hc.SessionsCPUWeight != 0 || hc.ControllerCPUWeight != 0) {
		return "limits.cgroup sessions_cpu and the cpu weights are ignored with cgroup_driver systemd: set CPUQuota= and CPUWeight= on " + l.Slice() + " and the controller's unit", nil
	}
	return "", nil
}

// Cgroup weight bounds and defaults (cpu.weight, kernel range 1-10000).
const (
	cgroupWeightMax         = 10000
	DefaultSessionsWeight   = 100
	DefaultControllerWeight = 1000
)

// HostCgroup tunes the two children of the controller's delegated cgroup:
// <own>/controller holds the controller and what it runs itself (admission
// and other hooks, git), <own>/sessions every session host and agent. It
// needs a delegated cgroup (Delegate= on the unit); without one the
// controller logs a warning and runs unsplit.
type HostCgroup struct {
	// SessionsCPU caps all sessions together, in CPUs ("5", "0.5"), as
	// cpu.max. Unset: no cap below the slice's own. Leave the slice room
	// above it: the gap is what the controller can always use.
	SessionsCPU string `yaml:"sessions_cpu"`
	// SessionsCPUWeight and ControllerCPUWeight are cpu.weight of the two
	// children (1-10000), the share each gets when the slice is full.
	// Defaults 100 and 1000.
	SessionsCPUWeight   int `yaml:"sessions_cpu_weight"`
	ControllerCPUWeight int `yaml:"controller_cpu_weight"`
}

// SessionsWeight is SessionsCPUWeight or its default.
func (h HostCgroup) SessionsWeight() int {
	if h.SessionsCPUWeight == 0 {
		return DefaultSessionsWeight
	}
	return h.SessionsCPUWeight
}

// ControllerWeight is ControllerCPUWeight or its default.
func (h HostCgroup) ControllerWeight() int {
	if h.ControllerCPUWeight == 0 {
		return DefaultControllerWeight
	}
	return h.ControllerCPUWeight
}

// SessionsCPUMax is SessionsCPU as a cpu.max value, "max" when unset.
func (h HostCgroup) SessionsCPUMax() (string, error) {
	v, err := CgroupLimits{CPU: h.SessionsCPU}.CPUMax()
	if err != nil {
		return "", fmt.Errorf("limits.cgroup.sessions_cpu: %w", err)
	}
	if v == "" {
		return "max", nil
	}
	return v, nil
}

func (h *HostCgroup) check() error {
	if h == nil {
		return nil
	}
	if _, err := h.SessionsCPUMax(); err != nil {
		return err
	}
	for _, w := range []struct {
		k string
		v int
	}{{"sessions_cpu_weight", h.SessionsCPUWeight}, {"controller_cpu_weight", h.ControllerCPUWeight}} {
		if w.v < 0 || w.v > cgroupWeightMax {
			return fmt.Errorf("limits.cgroup.%s %d must be 1-%d", w.k, w.v, cgroupWeightMax)
		}
	}
	return nil
}

// Linear configures Linear push delivery (#300).
type Linear struct {
	// WebhookListen is the TCP address for POST /linear/webhook. Empty:
	// off, intake and replies are polled only. Linear must reach it, so a
	// host with no public endpoint needs a reverse proxy or tunnel.
	WebhookListen string `yaml:"webhook_listen"`
	// AskOnIssue posts worker questions and run escalations on a run's
	// Linear issue as a comment, adds Blocked for a question, and takes a
	// reply as the answer (#293). Default false (JOB-114): they stay on
	// Desk and Linear replies are ordinary comments.
	AskOnIssue bool `yaml:"ask_on_issue"`
}

// AsksOnIssue reports whether questions and escalations go to the Linear
// issue. Opt-in.
func (l Linear) AsksOnIssue() bool { return l.AskOnIssue }

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
	// Instance names this jobshop on every signal (service.instance.id
	// and jobshop.instance, JOB-106). Empty: the root directory's base
	// name, or hostname:root when that is "jobshop".
	Instance string `yaml:"instance"`
	// Resource adds resource attributes to traces, logs and metrics, e.g.
	// deployment.environment. Values expand ${VAR} like OTLPHeaders.
	Resource map[string]string `yaml:"resource"`
}

// ResourceAttrs returns Resource with ${VAR} expanded by getenv.
func (t Telemetry) ResourceAttrs(getenv func(string) string) map[string]string {
	return expandValues(t.Resource, getenv)
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
	return expandValues(t.OTLPHeaders, getenv)
}

func expandValues(m map[string]string, getenv func(string) string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	h := make(map[string]string, len(m))
	for k, v := range m {
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
	// Isolation runs the agent as a mapped user with cgroup limits (#388).
	Isolation *Isolation `yaml:"isolation"`
	// Credentials lists the agent's login files ("~/..."). For a pool with
	// isolation.user they are copied into the mapped user's home before
	// each spawn, since that user cannot read the controller's (JOB-19).
	Credentials []string `yaml:"credentials"`
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
var PoolAdminMethods = []string{"config-reload", "config-show", "session-list", "session-restart", "drain", "trigger-list", "pool-list", "pool-unquarantine", "pool-scale", "session-cordon", "session-config", "session-rename", "session-close", "report-incidents", "attention", "attention-answer", "desk-command", "activity", "handoff", "ticket-release", "ticket-undefer", "decision-resolve", "question-answer", "session-journal", "journal-search", "memory-write"}

// Allows reports whether the pool's sessions may call admin method m.
func (p Pool) Allows(m string) bool { return slices.Contains(p.Admin, m) }

// Chat reports whether p is a chat pool.
func (p Pool) Chat() bool { return p.Type == PoolChat }

// AttachRW is Pool.Attach for pools whose sessions accept
// `session attach --rw`.
const AttachRW = "rw"

// Pool is a queue of work in one project served by sessions of one agent.
type Pool struct {
	// Enabled false keeps the pool validated but not loaded (JOB-24).
	Enabled *bool  `yaml:"enabled,omitempty"`
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
	// ConversationExport was OTLP conversation export (#45), removed by
	// #530. Still parsed so an older config loads, with a warning.
	ConversationExport bool `yaml:"conversation_export"`
	// Admin lists admin methods (PoolAdminMethods) this pool's session
	// sockets may call, e.g. [config-reload, session-list] for an
	// operator chat pool. Default none.
	Admin []string `yaml:"admin"`
	// DisplayName labels the pool's new sessions in lists (#626). Not
	// unique; a session may rename itself. Default: pool and ticket title.
	DisplayName string `yaml:"display_name"`
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
	// Isolation overrides the agent's user and limits (#388).
	Isolation *Isolation `yaml:"isolation"`
	// Labels group pools in lists and filters (JOB-81). Cosmetic: they do
	// not affect scheduling. A pool a packs entry declares also gets
	// instance:<name> and pack:<pack> (see addInstanceLabels). A host's
	// labels add to the pack's.
	Labels []string `yaml:"labels"`
}

// Project is one repo / work area.
type Project struct {
	// Prefix starts every ticket id, e.g. "js" gives js-0031w02j0r012.
	// Default: the project name.
	Prefix string `yaml:"prefix"`
	// IDs overrides the global ids block for this project (JOB-57).
	IDs *ids.Spec `yaml:"ids"`
	// Identities the project's sessions may act as (#200).
	Identities []string `yaml:"identities"`
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// RepoPattern is a GitHub owner/name.
var RepoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// varName is a valid template var name.
var varName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Load reads and validates the root's config.
func Load(r Root) (*Config, error) {
	return LoadFile(r.Config())
}

// Parse decodes and validates config bytes with no packs or prompt files
// (they need a file location; see LoadFile). Unknown keys are rejected.
func Parse(b []byte) (*Config, error) {
	c, err := parse(b, nil)
	if err != nil {
		return nil, err
	}
	if err := c.checkSandboxes(); err != nil {
		return nil, err
	}
	return c, nil
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
		// The typed decode below turns a number into a string, so pool
		// labels are checked as written.
		if err := checkRawPoolLabels(b, src); err != nil {
			return nil, err
		}
		if err := yaml.UnmarshalWithOptions(b, &c, yaml.Strict()); err != nil {
			return nil, fmt.Errorf("%s: %w", FileName, err)
		}
	}
	c.Sources = sources
	if err := c.checkAutomation(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	if c.Pack != "" || len(c.Packs) > 0 {
		return nil, fmt.Errorf("%s: pack and packs need a config file location", FileName)
	}
	if c.Projects == nil {
		c.Projects = map[string]Project{}
	}
	seen := map[string]string{}
	for name, p := range c.Projects {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: project name %q must match %s", src("projects", name), name, namePattern)
		}
		if p.Prefix == "" {
			p.Prefix = name
			c.Projects[name] = p
		}
		if !namePattern.MatchString(p.Prefix) {
			return nil, fmt.Errorf("%s: project %s: prefix %q must match %s", src("projects", name), name, p.Prefix, namePattern)
		}
		if other, dup := seen[p.Prefix]; dup {
			return nil, fmt.Errorf("%s: projects %s and %s share prefix %q", src("projects", name), other, name, p.Prefix)
		}
		seen[p.Prefix] = name
		if p.IDs != nil {
			if err := p.IDs.Check(); err != nil {
				return nil, fmt.Errorf("%s: project %s: %w", src("projects", name), name, err)
			}
		}
	}
	if err := c.IDs.Check(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
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
				return nil, fmt.Errorf("%s: project %s: unknown identity %q", src("projects", name), name, id)
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
			return nil, fmt.Errorf("%s: default_project %q is not defined", src("default_project", ""), c.DefaultProject)
		}
	}
	for name, s := range c.SandboxProfiles {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: sandbox_profiles: name %q must match %s", src("sandbox_profiles", name), name, namePattern)
		}
		if len(s.Profiles) > 0 {
			return nil, fmt.Errorf("%s: sandbox_profiles.%s: a profile cannot name profiles", src("sandbox_profiles", name), name)
		}
		if err := s.check(nil); err != nil {
			return nil, fmt.Errorf("%s: sandbox_profiles.%s: %w", src("sandbox_profiles", name), name, err)
		}
	}
	for name, s := range c.IsolationProfiles {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: isolation_profiles: name %q must match %s", src("isolation_profiles", name), name, namePattern)
		}
		if len(s.Profiles) > 0 {
			return nil, fmt.Errorf("%s: isolation_profiles.%s: a profile cannot name profiles", src("isolation_profiles", name), name)
		}
		if err := s.check(nil); err != nil {
			return nil, fmt.Errorf("%s: isolation_profiles.%s: %w", src("isolation_profiles", name), name, err)
		}
	}
	for name, a := range c.Agents {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: agent name %q must match %s", src("agents", name), name, namePattern)
		}
		if a.Command == "" {
			return nil, fmt.Errorf("%s: agent %s: command is required", src("agents", name), name)
		}
		if err := a.Sandbox.check(c.SandboxProfiles); err != nil {
			return nil, fmt.Errorf("%s: agent %s: %w", src("agents", name), name, err)
		}
		if err := a.Isolation.check(c.IsolationProfiles); err != nil {
			return nil, fmt.Errorf("%s: agent %s: %w", src("agents", name), name, err)
		}
		for _, p := range a.Credentials {
			if rest, ok := strings.CutPrefix(p, "~/"); !ok || rest == "" || strings.Contains(rest, "..") {
				return nil, fmt.Errorf("%s: agent %s: credentials path %q must start with ~/ and stay under it", src("agents", name), name, p)
			}
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
		if err := p.Sandbox.check(c.SandboxProfiles); err != nil {
			return nil, fmt.Errorf("%s: pool %s: %w", src("pools", name), name, err)
		}
		if err := p.Isolation.check(c.IsolationProfiles); err != nil {
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
		if err := checkLabels(p.Labels); err != nil {
			return nil, fmt.Errorf("%s: pool %s: labels: %w", src("pools", name), name, err)
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
		if p.ConversationExport {
			// A warning for one release, then an error (#530, as #325).
			c.Warnings = append(c.Warnings, fmt.Sprintf("%s: pool %s: conversation_export is removed, see journal.victorialogs; ignored", src("pools", name), name))
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
	if err := c.Defaults.Sandbox.check(c.SandboxProfiles); err != nil {
		return nil, fmt.Errorf("%s: defaults: %w", FileName, err)
	}
	if err := c.Defaults.Isolation.check(c.IsolationProfiles); err != nil {
		return nil, fmt.Errorf("%s: defaults: %w", FileName, err)
	}
	if err := c.Defaults.checkLabelRules(&c); err != nil {
		return nil, err
	}
	for _, d := range c.SharedTools {
		if !filepath.IsAbs(d) {
			return nil, fmt.Errorf("%s: shared_tools: %q must be absolute", FileName, d)
		}
	}
	if err := c.Store.check(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	if err := c.validateIngest(); err != nil {
		return nil, err
	}
	if err := c.Desk.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	if err := c.validateAdmission(); err != nil {
		return nil, err
	}
	if err := c.Redact.compile(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	if r := c.Ingest.Retention; r != "" {
		if d, err := parseDays(r); err != nil || d <= 0 {
			return nil, fmt.Errorf("%s: ingest.retention %q must be a positive duration (e.g. 14d, 72h)", FileName, r)
		}
	}
	if r := c.WriteKeys.Retention; r != "" {
		if d, err := parseDays(r); err != nil || d <= 0 {
			return nil, fmt.Errorf("%s: write_keys.retention %q must be a positive duration (e.g. 48h, 2d)", FileName, r)
		}
	}
	if err := c.Reload.check(); err != nil {
		return nil, err
	}
	if r := c.Artifacts.Retention; r != "" {
		if d, err := parseDays(r); err != nil || d <= 0 {
			return nil, fmt.Errorf("%s: artifacts.retention %q must be a positive duration (e.g. 90d)", FileName, r)
		}
	}
	if c.Artifacts.MaxBytes < 0 {
		return nil, fmt.Errorf("%s: artifacts.max_bytes must be >= 0", FileName)
	}
	if err := c.Journal.check(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	if c.Limits.MaxSessions < 0 {
		return nil, fmt.Errorf("%s: limits.max_sessions must be >= 0", FileName)
	}
	if err := c.Limits.Cgroup.check(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	if w, err := c.Limits.checkDriver(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	} else if w != "" {
		c.Warnings = append(c.Warnings, FileName+": "+w)
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
	for name, d := range c.DeskCommands {
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: desk_commands: name %q must match %s", FileName, name, namePattern)
		}
		if err := checkDeskCommand(d); err != nil {
			return nil, fmt.Errorf("%s: desk_commands.%s: %w", src("desk_commands", name), name, err)
		}
	}
	for _, k := range c.Learnings.Kinds {
		if err := store.ValidKindPattern(k); err != nil {
			return nil, fmt.Errorf("%s: learnings.kinds: %w", FileName, err)
		}
	}
	if err := c.checkResources(); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	// A warning for one release, then an error (as conversation_export):
	// self-deploy checks the new binary against the old module's generated
	// yaml, which still sets these keys, so an error blocks every deploy.
	for key, v := range map[string]string{"rollback_file": c.Attention.RollbackFile, "config_repo_file": c.Attention.ConfigRepoFile} {
		if v != "" {
			c.Warnings = append(c.Warnings, fmt.Sprintf("%s: attention.%s is removed (#577) and ignored: deploy tools report a problem as a needs-human ticket, `jobshop ticket upsert --key deploy:rollback -l needs-human ...`, and close it when it clears", FileName, key))
		}
	}
	if w := c.Attention.ResourceWait; w != "" {
		if d, err := time.ParseDuration(w); err != nil || d <= 0 {
			return nil, fmt.Errorf("%s: attention.resource_wait %q must be a positive duration", FileName, w)
		}
	}
	if c.Attention.PrestartMax < 0 {
		return nil, fmt.Errorf("%s: attention.prestart_max must be >= 0", FileName)
	}
	if p := c.Attention.PrestartPool; p != "" {
		if pl, ok := c.Pools[p]; !ok || !pl.Chat() {
			return nil, fmt.Errorf("%s: attention.prestart_pool %q is not a chat pool (type: chat)", FileName, p)
		}
	}
	if w := c.Attention.OrphanAge; w != "" && w != "off" {
		if d, err := time.ParseDuration(w); err != nil || d <= 0 {
			return nil, fmt.Errorf("%s: attention.orphan_age %q must be a positive duration or off", FileName, w)
		}
	}
	for name, w := range c.Workflows {
		if err := c.checkWorkflow(name, &w); err != nil {
			return nil, fmt.Errorf("%s: workflow %s: %w", src("workflows", name), name, err)
		}
	}
	if c.Escalation != nil {
		if err := c.checkEscalation(c.Escalation); err != nil {
			return nil, fmt.Errorf("%s: escalation: %w", FileName, err)
		}
	}
	if err := c.checkRules(); err != nil {
		return nil, err
	}
	if err := c.dropDisabled(src); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) checkGateway(g Gateway) error {
	if g.Network != GatewayA2A {
		return fmt.Errorf("network %q must be %q (Matrix and Discord were removed; use Desk)", g.Network, GatewayA2A)
	}
	return c.checkA2AGateway(g)
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
	if tr.MaxOpen < 0 {
		return errors.New("max_open must not be negative")
	}
	if tr.MaxOpen != 0 && tr.Handler == "" {
		return errors.New("max_open needs when.stall")
	}
	if err := checkTriggerOn(tr); err != nil {
		return err
	}
	if len(tr.On) == 0 && !w.HasConditions() {
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
		for _, k := range l.Kinds {
			if err := store.ValidKindPattern(k); err != nil {
				return fmt.Errorf("when: learnings.kinds: %w", err)
			}
		}
		for _, s := range l.Scope {
			if _, err := path.Match(s, ""); err != nil || s == "" {
				return fmt.Errorf("when: learnings.scope %q is not a valid glob", s)
			}
		}
		if l.Source != "" && l.Source != store.LearningSourceAgent && l.Source != store.LearningSourceController {
			return fmt.Errorf("when: learnings.source %q must be agent or controller", l.Source)
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
	actions := 0
	for _, set := range []bool{tr.Ticket != nil, tr.Run != nil, tr.Command != nil, tr.Handler != ""} {
		if set {
			actions++
		}
	}
	if actions != 1 {
		return errors.New("set exactly one of ticket, run, command and handler")
	}
	if tr.Handler != "" {
		return nil
	}
	if tr.Run != nil {
		return c.checkTriggerRun(tr)
	}
	if cm := tr.Command; cm != nil {
		if strings.TrimSpace(cm.Run) == "" {
			return errors.New("command: run is required")
		}
		if cm.Timeout != "" {
			if d, err := time.ParseDuration(cm.Timeout); err != nil || d <= 0 {
				return fmt.Errorf("command: timeout %q must be a positive duration", cm.Timeout)
			}
		}
		return checkTriggerTemplates(tr, map[string]string{"command.run": cm.Run})
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

// IDSpec is the id generator for project: its own ids block, else the
// global one (JOB-57).
func (c *Config) IDSpec(project string) ids.Spec {
	if p, ok := c.Projects[project]; ok && p.IDs != nil {
		return p.IDs.Normal()
	}
	return c.IDs.Normal()
}

// BuiltinMetaPacks are the meta pack names jobshop itself writes (the
// controller, the gateway on escalation tickets, and Desk's attention
// chat key, meta.desk.attention_chat, #535).
var BuiltinMetaPacks = []string{"github", "linear", "trigger", "run", "gateway", "pool", "hook", "subscribe", "report", "question", "escalation", "upsert", "desk"}

// KnownMetaPack reports whether name may be a top-level meta key: a
// built-in pack or one listed under meta_packs.
func (c *Config) KnownMetaPack(name string) bool {
	return slices.Contains(BuiltinMetaPacks, name) || slices.Contains(c.MetaPacks, name)
}
