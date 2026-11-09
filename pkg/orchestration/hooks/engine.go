package hooks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/execution/automation"
	"github.com/elee1766/jobshop/pkg/execution/automation/api"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

var (
	hookInvocations = store.CollectionSpec{Name: "hook_invocations", Index: []store.Field{{Name: "run", Path: "run"}, {Name: "trigger", Path: "trigger"}, {Name: "event", Path: "event.delivery"}}}
	hookCalls       = store.CollectionSpec{Name: "hook_calls", Index: []store.Field{{Name: "invocation", Path: "invocation"}}}
	hookAttempts    = store.CollectionSpec{Name: "hook_attempts", Index: []store.Field{{Name: "invocation", Path: "invocation"}}}
)

// Runtime is an immutable, accepted configuration and its loaded programs.
type Runtime struct {
	Config   *config.Config
	Programs map[string]*Program
}

// JournalPlan persists accepted graph evidence before the host applies it.
// CommitHook calls it while holding the host's graph serialization lock.
type JournalPlan func(Invocation) error

// Host owns admission policy and ticket graph mechanics, not hook persistence.
type Host interface {
	CurrentHooks() *Runtime
	AdmitHook(context.Context, *Runtime, Invocation) (string, error)
	CommitHook(context.Context, *Runtime, string, Invocation, JournalPlan) (Invocation, error)
	ApplyHookPlan(context.Context, string, Invocation) error
	HookFailed(context.Context, string, Invocation)
	HookApplied(context.Context, Invocation)
	// GitHubClient returns the shared client of the GitHub identity
	// covering repo, for the github built-in.
	GitHubClient(ctx context.Context, repo string) (*github.Client, error)
	// LinearClient returns the shared client of the Linear identity
	// covering team, for the linear built-in.
	LinearClient(ctx context.Context, team string) (*linear.Client, error)
	// ReadTickets serves tickets(), runs() and sessions().
	ReadTickets(ctx context.Context, q api.TicketQuery) ([]any, error)
}

// Collections is the persistence surface needed by the hook journal.
type Collections interface {
	Collection(context.Context, store.CollectionSpec) (store.Collection, error)
}

// Engine owns durable invocations, attempts, capability replay, and audit evidence.
// The host schedules one Process call at a time after taking its root lock.
type Engine struct {
	store            Collections
	host             Host
	root             string
	now              func() time.Time
	log              *slog.Logger
	recovered        bool
	hookAuditMu      sync.Mutex
	hookAuditSecrets []string
}

// New binds durable hook execution to a host and its root-local evidence store.
func New(st Collections, host Host, root string, now func() time.Time, log *slog.Logger) *Engine {
	return &Engine{store: st, host: host, root: root, now: now, log: log}
}

// ScanSource is Event.Source of a finding a trigger handler is called for.
// Delivery is the finding id, <signal>:<subject>; Payload holds signal,
// subject and detail.
const ScanSource = "scan"

// Event is an immutable event-time snapshot, not a mutable inbox read.
type Event struct {
	Source     string `json:"source"`
	Event      string `json:"event"`
	Delivery   string `json:"delivery"`
	Repository string `json:"repository"`
	// Trigger names the trigger the handler runs for.
	Trigger string         `json:"trigger"`
	Payload map[string]any `json:"payload"`
	Fields  map[string]any `json:"fields"`
}

// Step is the accepted event-time state of one workflow step.
type Step struct {
	ID         string         `json:"id"`
	TicketID   string         `json:"ticket_id"`
	Pool       string         `json:"pool"`
	Status     string         `json:"status"`
	Resolution string         `json:"resolution"`
	Outcome    string         `json:"outcome"`
	Output     map[string]any `json:"output"`
	Inputs     map[string]any `json:"inputs"`
	Failures   int            `json:"failures"`
	Retries    int            `json:"retries"`
}

// Run is the accepted event-time workflow snapshot passed to a callback.
type Run struct {
	ID       string            `json:"id"`
	Workflow string            `json:"workflow"`
	Key      string            `json:"key"`
	Project  string            `json:"project"`
	Inputs   map[string]any    `json:"inputs"`
	Vars     map[string]string `json:"vars"`
	State    string            `json:"state"`
	Steps    []Step            `json:"steps"`
}

// Invocation records accepted inputs and the selected graph before application.
type Invocation struct {
	ResultMetadata   *automation.ResultMetadata `json:"result_metadata,omitempty"`
	ResultOutput     map[string]any             `json:"result_output,omitempty"`
	Public           json.RawMessage            `json:"public,omitempty"`
	AcceptedAt       time.Time                  `json:"accepted_at"`
	ParentInvocation string                     `json:"parent_invocation,omitempty"`
	Targets          []string                   `json:"targets,omitempty"`
	Trigger          string                     `json:"trigger,omitempty"`
	// Sync names the guard hook or desk command a sync call ran for.
	Sync         string                `json:"sync,omitempty"`
	Run          string                `json:"run,omitempty"`
	Step         string                `json:"step,omitempty"`
	StepRevision int64                 `json:"step_revision,omitempty"`
	Entry        string                `json:"entry"`
	Digest       string                `json:"digest"`
	ReadSet      map[string]int64      `json:"read_set,omitempty"`
	Event        *Event                `json:"event,omitempty"`
	Args         map[string]any        `json:"args,omitempty"`
	RunInput     *Run                  `json:"run_input,omitempty"`
	StepInput    *Step                 `json:"step_input,omitempty"`
	Actions      []automation.Action   `json:"actions,omitempty"`
	Plan         *store.Graph          `json:"plan,omitempty"`
	Error        string                `json:"error,omitempty"`
	Output       string                `json:"output,omitempty"`
	Decisions    []automation.Decision `json:"decisions,omitempty"`
}

func hookKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// Put deduplicates an event and refreshes only superseded or explicitly retried work.
func (c *Engine) Put(ctx context.Context, key string, in Invocation) error {
	if in.AcceptedAt.IsZero() {
		in.AcceptedAt = c.now()
	}
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return err
	}
	in.Public = c.hookInvocationPublic(in)
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	rec, created, err := col.Put(ctx, store.Record{Key: hookKey(key), State: "pending", Value: data})
	if err == nil && !created && (rec.State == "superseded" || rec.State == "retry") {
		rec.State, rec.Value = "pending", data
		_, err = col.Update(ctx, rec)
	}
	return err
}

// LastFire returns the latest accepted trigger execution for scheduling policy.
func (c *Engine) LastFire(ctx context.Context, name string) (time.Time, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return time.Time{}, err
	}
	records, err := col.List(ctx, store.RecordFilter{Where: map[string]string{"trigger": name}, States: []string{"done", "planned", "failed"}, Order: store.OrderCreatedDesc, Limit: 1})
	if err != nil || len(records) == 0 {
		return time.Time{}, err
	}
	var in Invocation
	if err := DecodeJSON(records[0].Value, &in); err != nil {
		return time.Time{}, err
	}
	return in.AcceptedAt, nil
}

// Process recovers prior-process leases and evaluates one durable invocation.
func (c *Engine) Process(ctx context.Context) (bool, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return false, err
	}
	if !c.recovered {
		if err := c.recoverHookClaims(ctx, col); err != nil {
			return false, err
		}
		c.recovered = true
	}
	lease := config.Duration(c.host.CurrentHooks().Config.Automation.Timeout, 30*time.Second) + time.Minute
	rec, err := col.Claim(ctx, store.RecordFilter{States: []string{"pending", "planned"}}, "controller-hooks", lease)
	if errors.Is(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	c.executeHookAttempt(ctx, col, rec)
	return true, nil
}

func (c *Engine) saveHook(ctx context.Context, col store.Collection, rec store.Record, in Invocation, state string) (store.Record, error) {
	// Prepared evidence was sanitized with credentials that may have rotated.
	data, err := json.Marshal(in)
	if err != nil {
		return rec, err
	}
	rec.Value, rec.State = data, state
	rec.LeaseUntil = time.Time{}
	return col.Update(ctx, rec)
}

func (c *Engine) hookFailure(ctx context.Context, col store.Collection, rec store.Record, in Invocation, err error) {
	in.Error = c.hookAuditText(err.Error())
	in.Public = c.hookPublicError(in.Public, in.Error)
	_, saveErr := c.saveHook(ctx, col, rec, in, "failed")
	c.log.ErrorContext(ctx, "automation hook failed", "invocation", rec.Key, "run", in.Run, "trigger", in.Trigger, "err", in.Error, "record_error", saveErr)
	c.host.HookFailed(ctx, rec.Key, in)
}
