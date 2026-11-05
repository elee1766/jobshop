// Package rpc defines the JSON-RPC protocol between the jobshop CLI and the
// controller: method names, parameter and result types, error mapping, and
// a client. Transport is NDJSON over a unix socket, using the JSON-RPC
// connection from github.com/coder/acp-go-sdk (the same one ACP uses).
package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	acp "github.com/coder/acp-go-sdk"
	"github.com/elee1766/acphost"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Method names.
const (
	MethodTicketCreate     = "ticket.create"
	MethodTicketShow       = "ticket.show"
	MethodTicketTrace      = "ticket.trace"
	MethodTicketList       = "ticket.list"
	MethodTicketUpdate     = "ticket.update"
	MethodTicketClose      = "ticket.close"
	MethodTicketReopen     = "ticket.reopen"
	MethodReady            = "ticket.ready"
	MethodCommentAdd       = "comment.add"
	MethodEdgeAdd          = "edge.add"
	MethodEdgeRemove       = "edge.remove"
	MethodLabelAdd         = "label.add"
	MethodLabelRemove      = "label.remove"
	MethodClaim            = "claim"
	MethodHeartbeat        = "heartbeat"
	MethodRelease          = "release"
	MethodHandoff          = "handoff"
	MethodSessionNew       = "session.new"
	MethodSessionList      = "session.list"
	MethodSessionReap      = "session.reap"
	MethodSessionRestart   = "session.restart"
	MethodSessionPrompt    = "session.prompt"
	MethodSessionSandbox   = "session.sandbox"
	MethodSessionSend      = "session.send"
	MethodSessionSubscribe = "session.subscribe"
	MethodWhoami           = "whoami"
	MethodShutdown         = "controller.shutdown"
	MethodDrain            = "controller.drain"
	MethodReload           = "config.reload"
	MethodTriggerList      = "trigger.list"
	MethodConfigShow       = "config.show"
	MethodRunStart         = "run.start"
	MethodRunShow          = "run.show"
	MethodRunList          = "run.list"
	MethodRunCancel        = "run.cancel"
	MethodRunResume        = "run.resume"
	MethodBatchStart       = "batch.start"
	MethodBatchShow        = "batch.show"
	MethodBatchList        = "batch.list"
	MethodPoolList         = "pool.list"
	MethodPoolScale        = "pool.scale"
	MethodSessionCordon    = "session.cordon"
	MethodSessionConfig    = "session.config"
	MethodPoolUnquarantine = "pool.unquarantine"
	MethodReportIncidents  = "report.incidents"
	MethodGuard            = "guard"
	MethodVersion          = "version"
	MethodRecall           = "recall"
	MethodAttention        = "attention"
	MethodIdentityList     = "identity.list"
	MethodIdentityCheck    = "identity.check"
	MethodIdentityToken    = "identity.token"
	MethodIngestDrain      = "ingest.drain"
	MethodSubscribe        = "subscribe"
	MethodReport           = "report"
	MethodQuestionAnswer   = "question.answer"
	MethodActivityList     = "activity.list"
	MethodActivityRecord   = "activity.record"
	MethodLearnAdd         = "learn.add"
	MethodLearnList        = "learn.list"
	MethodLearnUpdate      = "learn.update"
	MethodLearnEvaluate    = "learn.evaluate"
	MethodLearnEdits       = "learn.edits"
	MethodWriteKeyClaim    = "writekey.claim"
	MethodWriteKeyFinish   = "writekey.finish"
)

// IngestDrainParams is ingest.drain. Spool, if set, drains that spool
// database instead of the configured one.
type IngestDrainParams struct {
	Spool string `json:"spool,omitempty"`
}

// IngestDrainResult counts one `ingest drain` pass (#335): spool rows
// moved, how many were new inbox items, and rows moved to spool_bad.
type IngestDrainResult struct {
	Moved int `json:"moved"`
	Fresh int `json:"fresh"`
	Bad   int `json:"bad"`
}

// Ingest worker methods (#443): an external `jobshop ingest work`
// process claims pending inbox items, runs the admission hooks itself
// and reports the outcome. Each call is also the worker's heartbeat.
const (
	MethodIngestClaim   = "ingest.claim"
	MethodIngestAdmit   = "ingest.admit"
	MethodIngestWorkers = "ingest.workers"
)

// IngestClaimParams is ingest.claim. Worker is a stable id per process.
type IngestClaimParams struct {
	Worker string `json:"worker"`
	Host   string `json:"host,omitempty"`
}

// IngestClaimResult is a leased pending item with the hooks to run and
// the event's project; Item nil when nothing is pending. Answered: a
// human already closed a hold for the item, so validating hooks are
// skipped.
type IngestClaimResult struct {
	Item     *store.InboxItem       `json:"item,omitempty"`
	Hooks    []config.AdmissionHook `json:"hooks,omitempty"`
	Project  string                 `json:"project,omitempty"`
	Answered bool                   `json:"answered,omitempty"`
	LeaseMS  int64                  `json:"lease_ms,omitempty"` // lease length; clocks may differ
}

// IngestAdmitParams reports a run of the hooks on a claimed item. Run
// is the worker's hook run (events, results, denies and asks); the
// controller decides the state from it.
type IngestAdmitParams struct {
	Worker string          `json:"worker"`
	ID     int64           `json:"id"`
	Run    json.RawMessage `json:"run"`
}

// IngestWorker is one external worker's last heartbeat.
type IngestWorker struct {
	ID        string    `json:"id"`
	Host      string    `json:"host,omitempty"`
	Seen      time.Time `json:"seen"`
	LastClaim time.Time `json:"last_claim,omitzero"`
	Admitted  int64     `json:"admitted"`
}

// IngestWorkersResult lists external workers and inbox counts by state.
type IngestWorkersResult struct {
	Mode    string           `json:"mode"`
	Workers []IngestWorker   `json:"workers"`
	Counts  map[string]int64 `json:"counts"`
}

// Activity sources (#255).
const (
	ActivityGitHub = "github"
	ActivityLinear = "linear"
)

// ActivityItem is one external event: a push, PR, review, comment, CI
// result or Linear change. Ref is the issue or PR number (GitHub) or the
// issue key (Linear) it refers to; Tickets are the Jobshop runs and
// tickets whose meta.github or meta.linear names that ref.
type ActivityItem struct {
	ID      string    `json:"id"`
	Source  string    `json:"source"`
	Repo    string    `json:"repo,omitempty"` // owner/name, or the Linear team key
	Kind    string    `json:"kind"`
	Actor   string    `json:"actor"`
	Bot     bool      `json:"bot,omitempty"`
	At      time.Time `json:"at"`
	Title   string    `json:"title"`
	Body    string    `json:"body,omitempty"`
	URL     string    `json:"url,omitempty"`
	Ref     string    `json:"ref,omitempty"`
	PR      bool      `json:"pr,omitempty"`
	Tickets []string  `json:"tickets,omitempty"`
}

// ActivityListParams filters activity.list. Bots is "humans", "bots" or
// "" (all). Ticket keeps items linked to that run or ticket.
type ActivityListParams struct {
	Since  time.Duration `json:"since,omitempty"`
	Source string        `json:"source,omitempty"`
	Repo   string        `json:"repo,omitempty"`
	Kind   string        `json:"kind,omitempty"`
	Actor  string        `json:"actor,omitempty"`
	Bots   string        `json:"bots,omitempty"`
	Ticket string        `json:"ticket,omitempty"`
	Limit  int           `json:"limit,omitempty"`
}

// ActivityListResult is newest first. Sources lists every polled repo and
// team, for filter chips.
type ActivityListResult struct {
	Items   []ActivityItem `json:"items"`
	Sources []string       `json:"sources"`
}

// ActivityRecordParams records one outside call a session made through
// `jobshop gh` (#349). Action is the guard action name; Target is
// owner/name, with #n for a PR or issue.
type ActivityRecordParams struct {
	Action   string `json:"action"`
	Repo     string `json:"repo,omitempty"`
	Target   string `json:"target,omitempty"`
	Decision string `json:"decision"`
	ExitCode int    `json:"exit_code"`
	Command  string `json:"command,omitempty"`
}

// CreateParams creates a ticket. Project defaults to the caller's project.
type CreateParams struct {
	Project  string          `json:"project,omitempty"`
	Kind     store.Kind      `json:"kind,omitempty"` // default task
	Title    string          `json:"title"`
	Body     string          `json:"body,omitempty"`
	Priority *int            `json:"priority,omitempty"`
	Pool     string          `json:"pool,omitempty"`
	Holder   string          `json:"holder,omitempty"`
	ParentID string          `json:"parent_id,omitempty"`
	Labels   []string        `json:"labels,omitempty"`
	Fields   json.RawMessage `json:"fields,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	// Force pre-assigns to a cordoned session (#281).
	Force bool `json:"force,omitempty"`
}

// IDParams names one ticket. An empty ID in a session scope means the
// session's current ticket.
type IDParams struct {
	ID string `json:"id,omitempty"`
}

// ShowResult is a ticket with its relations and timeline.
type ShowResult struct {
	Ticket   store.Ticket         `json:"ticket"`
	Edges    []store.Edge         `json:"edges"`
	Comments []store.Comment      `json:"comments"`
	History  []store.HistoryEntry `json:"history"`
}

// ListParams filters tickets.
type ListParams struct {
	Project  string       `json:"project,omitempty"`
	Kind     store.Kind   `json:"kind,omitempty"`
	Status   store.Status `json:"status,omitempty"`
	Pool     string       `json:"pool,omitempty"`
	Holder   string       `json:"holder,omitempty"`
	ParentID string       `json:"parent_id,omitempty"`
	Label    string       `json:"label,omitempty"`
	// Newest lists the most recently created first.
	Newest bool `json:"newest,omitempty"`
	// Report matches tickets whose last worker report has this state (#186).
	Report string `json:"report,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	// Orphaned lists open, unclaimed tickets with no pool or a pool not in
	// the loaded config. Other filters except Project and Limit are ignored.
	Orphaned bool `json:"orphaned,omitempty"`
}

// ReadyParams selects the ready frontier.
type ReadyParams struct {
	Project string `json:"project,omitempty"`
	Pool    string `json:"pool,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// UpdateParams changes a ticket. Nil fields are unchanged.
type UpdateParams struct {
	ID         string          `json:"id"`
	Title      *string         `json:"title,omitempty"`
	Body       *string         `json:"body,omitempty"`
	Priority   *int            `json:"priority,omitempty"`
	Pool       *string         `json:"pool,omitempty"`
	Holder     *string         `json:"holder,omitempty"`
	DeferUntil *time.Time      `json:"defer_until,omitempty"`
	Fields     json.RawMessage `json:"fields,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	IfRevision int64           `json:"if_revision,omitempty"`
	// Force pre-assigns to a cordoned session, which otherwise refuses
	// new work (#281).
	Force bool `json:"force,omitempty"`
}

// CloseParams closes a ticket. In a session scope the session must hold it.
type CloseParams struct {
	ID         string `json:"id,omitempty"`
	Resolution string `json:"resolution,omitempty"` // default "done"
	// Outcome is a step's result word (approve, reject, ...), stored in
	// its fields for later steps' when conditions (#385). Steps only.
	Outcome string `json:"outcome,omitempty"`
}

// CommentParams adds a comment.
type CommentParams struct {
	ID   string `json:"id,omitempty"`
	Body string `json:"body"`
}

// EdgeParams adds or removes an edge.
type EdgeParams struct {
	From string         `json:"from"`
	To   string         `json:"to"`
	Kind store.EdgeKind `json:"kind"`
}

// LabelParams adds or removes a label.
type LabelParams struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// ClaimParams claims a ticket: ID for a specific one, else the next one.
// Holder and Pool are fixed by a session scope; admin callers must set Holder.
type ClaimParams struct {
	ID      string `json:"id,omitempty"`
	Project string `json:"project,omitempty"`
	Pool    string `json:"pool,omitempty"`
	Holder  string `json:"holder,omitempty"`
}

// ClaimResult is the claimed ticket and its fence.
type ClaimResult struct {
	Ticket store.Ticket `json:"ticket"`
	Claim  store.Claim  `json:"claim"`
}

// SubscribeParams makes a ticket (default: the session's current one)
// follow a GitHub issue or PR thread (#187).
type SubscribeParams struct {
	ID     string `json:"id,omitempty"`
	Repo   string `json:"repo"`
	Number int    `json:"number"`
	PR     bool   `json:"pr,omitempty"`
}

// HandoffParams moves a ticket to another pool, releasing its holder and
// adding Note as a comment, in one transaction (#492).
type HandoffParams struct {
	ID   string `json:"id,omitempty"`
	Pool string `json:"pool"`
	Note string `json:"note,omitempty"`
}

// ReleaseParams releases a held ticket.
type ReleaseParams struct {
	ID     string `json:"id,omitempty"`
	Reason string `json:"reason,omitempty"`
	// Until defers the ticket; such a release is not a step failure.
	Until *time.Time `json:"until,omitempty"`
}

// SessionNewParams creates a session ticket and its CLI socket. With a
// configured agent the controller starts and drives it; otherwise the
// session is external (the caller runs an agent with JOBSHOP_SOCKET).
type SessionNewParams struct {
	Project string `json:"project,omitempty"`
	Pool    string `json:"pool,omitempty"`
	Agent   string `json:"agent,omitempty"` // default: the pool's agent
	Key     string `json:"key,omitempty"`
	Title   string `json:"title,omitempty"`
}

// VersionResult is the controller binary's version and commit. Sessions
// read it to see what is deployed: their own binary is fixed when they
// start (#180).
type VersionResult struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// IdentityParams names a target: an owner/repo or a Linear team key.
type IdentityParams struct {
	Target string `json:"target"`
}

// IdentityInfo describes one identity, without secrets.
type IdentityInfo struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	AppID     int64    `json:"app_id,omitempty"`
	Owners    []string `json:"owners,omitempty"`
	Repos     []string `json:"repos,omitempty"`
	Workspace string   `json:"workspace,omitempty"`
	Teams     []string `json:"teams,omitempty"`
	// Projects and Pools that may act as the identity.
	Projects []string `json:"projects,omitempty"`
	Pools    []string `json:"pools,omitempty"`
}

// IdentityCheckResult says which identity covers a target; a check also
// obtains a credential, so a bad key or installation fails here.
type IdentityCheckResult struct {
	Target   string    `json:"target"`
	Identity string    `json:"identity"`
	Kind     string    `json:"kind"`
	Expires  time.Time `json:"expires,omitzero"`
	// Viewer is the user the token acts as (Linear kinds and github-token).
	Viewer string `json:"viewer,omitempty"`
}

// IdentityTokenResult is a credential for a target.
type IdentityTokenResult struct {
	Identity string    `json:"identity"`
	Kind     string    `json:"kind"`
	Token    string    `json:"token"`
	Expires  time.Time `json:"expires,omitzero"`
}

// ShutdownParams stops the controller. All also ends every session; without
// it, sessions keep running and the next controller adopts them.
type ShutdownParams struct {
	All bool `json:"all,omitempty"`
}

// SessionNewResult names the new session and where its socket is.
type SessionNewResult struct {
	Ticket store.Ticket `json:"ticket"`
	Socket string       `json:"socket"`
}

// SessionListParams selects sessions. All includes closed session tickets.
// State keeps only sessions with that resolution (for example "crashed"),
// and Since only sessions closed within that duration. Either one selects
// closed sessions only and leaves orphan hosts out, so the result is
// evidence about ended sessions alone.
type SessionListParams struct {
	All   bool          `json:"all,omitempty"`
	State string        `json:"state,omitempty"`
	Since time.Duration `json:"since,omitempty"`
}

// SessionRestartParams restarts one session's host. Without Now a busy
// session restarts after its current turn; Now cancels the turn.
type SessionRestartParams struct {
	ID  string `json:"id"`
	Now bool   `json:"now,omitempty"`
}

// SessionSandboxParams names the session whose sandbox to show.
type SessionSandboxParams struct {
	ID string `json:"id"`
}

// SessionSandboxResult is the sandbox the session's next spawn would get.
// Sandbox is nil when its pool is not sandboxed. Error is set when the
// bwrap argv cannot be built (bwrap missing, a mount source missing).
type SessionSandboxResult struct {
	Session string           `json:"session"`
	Pool    string           `json:"pool"`
	Sandbox *acphost.Sandbox `json:"sandbox,omitempty"`
	Argv    []string         `json:"argv,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// SessionPromptParams names the session whose next prompt to show.
type SessionPromptParams struct {
	ID string `json:"id"`
}

// PromptPart is one labeled piece of a prompt. Source is "pool",
// "rule:<name>", "ticket" or "messages".
type PromptPart struct {
	Source string `json:"source"`
	Text   string `json:"text"`
	// Status, on rule parts, says why the part is or is not sent (#397):
	// new, changed, always, removed, or unchanged (skipped: the session
	// already holds this text).
	Status string `json:"status,omitempty"`
}

// SessionPromptResult is the prompt the session's next turn would get,
// in order, with each part labeled. Turn is "ticket" or "message";
// Ticket is the ticket it held or would claim. Text is the parts
// rendered under "=== <source> ===" lines.
type SessionPromptResult struct {
	Session string       `json:"session"`
	Turn    string       `json:"turn"`
	Ticket  string       `json:"ticket,omitempty"`
	Rules   []string     `json:"rules"`
	Parts   []PromptPart `json:"parts"`
	Text    string       `json:"text"`
}

// SessionRestartResult reports the host that will be replaced.
type SessionRestartResult struct {
	ID      string `json:"id"`
	HostPid int    `json:"host_pid,omitempty"`
	Now     bool   `json:"now,omitempty"`
}

// SessionInfo is a session ticket joined with the live state of its
// session directory. Ticket is nil for an orphan: a directory with a live
// host but no open session ticket.
type SessionInfo struct {
	ID       string        `json:"id"`
	Ticket   *store.Ticket `json:"ticket,omitempty"`
	Orphan   bool          `json:"orphan,omitempty"`
	Alive    bool          `json:"alive"`
	HostPid  int           `json:"host_pid,omitempty"`
	AgentPid int           `json:"agent_pid,omitempty"` // known once a driver connected
	Current  string        `json:"current,omitempty"`   // ticket the session holds
	Seq      int64         `json:"seq,omitempty"`       // last journal sequence number
	SeqTime  *time.Time    `json:"seq_time,omitempty"`
	Exit     string        `json:"exit,omitempty"`    // exit.json reason, for dead hosts
	Drained  bool          `json:"drained,omitempty"` // takes no new tickets
	// Cordon is the session's cordon reason (scale, manual, drain), empty
	// while in rotation (#281).
	Cordon string `json:"cordon,omitempty"`
	// Crashes counts unplanned host restarts in the crash window.
	Crashes int `json:"crashes,omitempty"`
	// BackoffUntil is set while the session waits out a provider rate limit.
	BackoffUntil *time.Time `json:"backoff_until,omitempty"`
	// Model and Effort are the agent's current values (#304); Override is
	// the session's `session config` override, if any.
	Model    string            `json:"model,omitempty"`
	Effort   string            `json:"effort,omitempty"`
	Override map[string]string `json:"override,omitempty"`
}

// SessionReapParams cleans up session leftovers. Hosts with no open session
// ticket are always stopped. Closed removes the directories (and git
// worktrees) of sessions closed more than OlderThan ago. DryRun reports
// without acting.
type SessionReapParams struct {
	Closed    bool          `json:"closed,omitempty"`
	OlderThan time.Duration `json:"older_than,omitempty"` // default 24h
	DryRun    bool          `json:"dry_run,omitempty"`
}

// SessionReapResult names what was (or, with DryRun, would be) stopped and
// removed.
type SessionReapResult struct {
	Stopped []string `json:"stopped"`
	Removed []string `json:"removed"`
	DryRun  bool     `json:"dry_run,omitempty"`
}

// Whoami describes the caller's scope.
type Whoami struct {
	Actor   string `json:"actor"`
	Admin   bool   `json:"admin"`
	Session string `json:"session,omitempty"`
	Project string `json:"project,omitempty"`
	Pool    string `json:"pool,omitempty"`
}

// Error codes in the JSON-RPC server-defined range, one per store error.
const (
	CodeNotFound     = -32001
	CodeConflict     = -32002
	CodeNotHolder    = -32003
	CodeNoWork       = -32004
	CodeInvalid      = -32005
	CodeCycle        = -32006
	CodeClosed       = -32007
	CodeAlreadyExist = -32008
	CodeForbidden    = -32009
)

// ErrForbidden means the caller's scope does not allow the operation.
var ErrForbidden = errors.New("forbidden")

var codes = []struct {
	code int
	err  error
}{
	{CodeNotFound, store.ErrNotFound},
	{CodeConflict, store.ErrConflict},
	{CodeNotHolder, store.ErrNotHolder},
	{CodeNoWork, store.ErrNoWork},
	{CodeInvalid, store.ErrInvalid},
	{CodeCycle, store.ErrCycle},
	{CodeClosed, store.ErrClosed},
	{CodeAlreadyExist, store.ErrAlreadyExist},
	{CodeForbidden, ErrForbidden},
}

// ToWire maps an error to a JSON-RPC error that keeps its kind.
func ToWire(err error) *acp.RequestError {
	if err == nil {
		return nil
	}
	var re *acp.RequestError
	if errors.As(err, &re) {
		return re
	}
	for _, c := range codes {
		if errors.Is(err, c.err) {
			return &acp.RequestError{Code: c.code, Message: err.Error()}
		}
	}
	return &acp.RequestError{Code: -32603, Message: err.Error()}
}

// remoteError is a server error that still matches the store sentinel.
type remoteError struct {
	msg  string
	kind error
}

func (e *remoteError) Error() string { return e.msg }
func (e *remoteError) Unwrap() error { return e.kind }

// FromWire maps a JSON-RPC error back so callers can use errors.Is with
// store sentinels.
func FromWire(err error) error {
	var re *acp.RequestError
	if !errors.As(err, &re) {
		return err
	}
	for _, c := range codes {
		if re.Code == c.code {
			return &remoteError{msg: re.Message, kind: c.err}
		}
	}
	if re.Data != nil {
		return fmt.Errorf("%s: %v", re.Message, re.Data)
	}
	return errors.New(re.Message)
}

// Client is a connection to a controller or session socket.
type Client struct {
	nc       net.Conn
	conn     *acp.Connection
	onEvent  atomic.Pointer[func(SessionEvent)]
	onChange atomic.Pointer[func(Event)]
}

// Dial connects to a unix socket.
func Dial(ctx context.Context, path string) (*Client, error) {
	var d net.Dialer
	nc, err := d.DialContext(ctx, "unix", path)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w (is the controller running?)", path, err)
	}
	// The controller only sends session.subscribe events back.
	gate := NewGate(nc)
	c := &Client{nc: nc}
	c.conn = acp.NewConnection(c.handle, nc, gate)
	c.conn.SetLogger(QuietLogger(slog.New(slog.DiscardHandler)))
	gate.Open()
	return c, nil
}

// Gate is a reader that blocks until Open. The SDK starts reading in
// NewConnection, and its receive loop reads the logger that SetLogger
// writes; holding reads until after SetLogger avoids that data race.
type Gate struct {
	r    io.Reader
	open chan struct{}
	once sync.Once
}

// NewGate wraps r closed.
func NewGate(r io.Reader) *Gate { return &Gate{r: r, open: make(chan struct{})} }

// Open lets reads through.
func (g *Gate) Open() { g.once.Do(func() { close(g.open) }) }

func (g *Gate) Read(p []byte) (int, error) {
	<-g.open
	return g.r.Read(p)
}

// QuietLogger returns l raised to WARN: the SDK logs every closed
// connection at INFO, which is one line per CLI call.
func QuietLogger(l *slog.Logger) *slog.Logger {
	return slog.New(minLevel{l.Handler(), slog.LevelWarn})
}

type minLevel struct {
	slog.Handler
	min slog.Level
}

func (h minLevel) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= h.min && h.Handler.Enabled(ctx, l)
}

func (h minLevel) WithAttrs(as []slog.Attr) slog.Handler {
	return minLevel{h.Handler.WithAttrs(as), h.min}
}

func (h minLevel) WithGroup(n string) slog.Handler { return minLevel{h.Handler.WithGroup(n), h.min} }

// Call invokes method and decodes the result into out (may be nil).
func (c *Client) Call(ctx context.Context, method string, params, out any) error {
	params, err := withMeta(ctx, params)
	if err != nil {
		return err
	}
	res, err := acp.SendRequest[json.RawMessage](c.conn, ctx, method, params)
	if err != nil {
		return FromWire(err)
	}
	if out == nil || len(res) == 0 {
		return nil
	}
	return json.Unmarshal(res, out)
}

// MetaKey carries request metadata (trace context) inside params, as ACP
// does with _meta.
const MetaKey = "_meta"

// Meta is request metadata.
type Meta struct {
	TraceParent string `json:"traceparent,omitempty"`
	// Actor names the person behind an admin-socket call, such as
	// "desk:alice" (#515). The controller honors it only on the admin
	// socket; a session socket's actor is fixed by the socket.
	Actor string `json:"actor,omitempty"`
}

type actorKey struct{}

// WithActor makes calls under ctx carry actor in _meta.actor.
func WithActor(ctx context.Context, actor string) context.Context {
	if actor == "" {
		return ctx
	}
	return context.WithValue(ctx, actorKey{}, actor)
}

// ActorFrom returns the actor set by WithActor.
func ActorFrom(ctx context.Context) string {
	a, _ := ctx.Value(actorKey{}).(string)
	return a
}

// MaxActor is the longest _meta.actor accepted, in bytes.
const MaxActor = 256

// ValidActor checks an actor name: at most MaxActor bytes of printable
// text, no control characters.
func ValidActor(a string) error {
	if a == "" || len(a) > MaxActor || !utf8.ValidString(a) {
		return fmt.Errorf("%w: actor must be 1 to %d bytes of UTF-8", store.ErrInvalid, MaxActor)
	}
	for _, r := range a {
		if !unicode.IsPrint(r) {
			return fmt.Errorf("%w: actor has a non-printable character", store.ErrInvalid)
		}
	}
	return nil
}

// withMeta adds _meta (traceparent, actor) from ctx to object params.
func withMeta(ctx context.Context, params any) (any, error) {
	tp := telemetry.TraceParent(ctx)
	actor := ActorFrom(ctx)
	if tp == "" && actor == "" {
		return params, nil
	}
	obj := map[string]json.RawMessage{}
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		if string(b) != "null" && json.Unmarshal(b, &obj) != nil {
			return params, nil // not an object: send without metadata
		}
	}
	m, err := json.Marshal(Meta{TraceParent: tp, Actor: actor})
	if err != nil {
		return nil, err
	}
	obj[MetaKey] = m
	return obj, nil
}

// SplitMeta removes _meta from object params.
func SplitMeta(raw json.RawMessage) (json.RawMessage, Meta) {
	var meta Meta
	if !bytes.Contains(raw, []byte(MetaKey)) {
		return raw, meta
	}
	obj := map[string]json.RawMessage{}
	if json.Unmarshal(raw, &obj) != nil {
		return raw, meta
	}
	m, ok := obj[MetaKey]
	if !ok {
		return raw, meta
	}
	_ = json.Unmarshal(m, &meta)
	delete(obj, MetaKey)
	out, err := json.Marshal(obj)
	if err != nil {
		return raw, meta
	}
	return out, meta
}

// Close closes the connection.
func (c *Client) Close() error {
	if err := c.nc.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}
	return nil
}

// DrainParams drains (or with Undrain, undrains) one pool, or every pool
// and session when Pool is empty. Undrain with no Pool clears all drains.
type DrainParams struct {
	Pool    string `json:"pool,omitempty"`
	Undrain bool   `json:"undrain,omitempty"`
}

// ReloadResult counts what the reloaded config defines.
type ReloadResult struct {
	Pools     int `json:"pools"`
	Agents    int `json:"agents"`
	Triggers  int `json:"triggers"`
	Workflows int `json:"workflows"`
}

// PoolParams names one pool.
type PoolParams struct {
	Pool string `json:"pool"`
}

// PoolStatusResult lists pools that start no sessions: Quarantined until
// reload or unquarantine (with the spawn error), Paused until a time
// after crash-looping.
type PoolStatusResult struct {
	Quarantined map[string]string    `json:"quarantined"`
	Paused      map[string]time.Time `json:"paused"`
	// Pools are the ticket pools' scale columns (#281).
	Pools []PoolScale `json:"pools,omitempty"`
	// HostMax is limits.max_sessions; 0 means no host cap.
	HostMax int `json:"host_max,omitempty"`
}

// PoolScale is one ticket pool's scaling state. Target is the manual
// target from pool scale; nil means auto.
type PoolScale struct {
	Name     string `json:"name"`
	Min      int    `json:"min"`
	Max      int    `json:"max"`
	Priority int    `json:"priority,omitempty"`
	Target   *int   `json:"target,omitempty"`
	Desired  int    `json:"desired"`
	Open     int    `json:"open"`
	Working  int    `json:"working"`
	Idle     int    `json:"idle"`
	Cordoned int    `json:"cordoned"`
	Drained  bool   `json:"drained,omitempty"`
}

// PoolScaleParams sets a pool's manual target, or with Auto returns it to
// demand-based scaling.
type PoolScaleParams struct {
	Pool   string `json:"pool"`
	Target *int   `json:"target,omitempty"`
	Auto   bool   `json:"auto,omitempty"`
}

// SessionCordonParams cordons one session (no new claims), with Drain
// also ends it once idle, or with Uncordon returns it to rotation.
type SessionCordonParams struct {
	ID       string `json:"id"`
	Drain    bool   `json:"drain,omitempty"`
	Uncordon bool   `json:"uncordon,omitempty"`
}

// SessionConfigParams sets ACP config options on one session (#304),
// applied between turns. Options keys are option ids; "effort" means
// reasoning_effort. Reset drops the override first, back to the pool's
// values.
type SessionConfigParams struct {
	ID      string            `json:"id"`
	Options map[string]string `json:"options,omitempty"`
	Reset   bool              `json:"reset,omitempty"`
}

// DrainResult lists drained pools after the call; "*" means all.
type DrainResult struct {
	Drained []string `json:"drained"`
}

// RecallParams is a recall search (#265). Sessions always search their own
// project; admin may set Project or leave it empty for all.
type RecallParams struct {
	Query    string `json:"query"`
	Limit    int    `json:"limit,omitempty"`
	Project  string `json:"project,omitempty"`
	Kind     string `json:"kind,omitempty"` // comment, journal or learning
	TicketID string `json:"ticket_id,omitempty"`
}

// TraceParams selects the ticket to trace.
type TraceParams struct {
	ID    string `json:"id,omitempty"`
	Depth int    `json:"depth,omitempty"` // 0 means 5
}

// TraceNode is one ticket in a trace: how it came to be, what it links to
// and the tickets it led to.
type TraceNode struct {
	ID         string       `json:"id"`
	Title      string       `json:"title,omitempty"`
	Pool       string       `json:"pool,omitempty"`
	Status     store.Status `json:"status"`
	Resolution string       `json:"resolution,omitempty"`
	CreatedBy  string       `json:"created_by,omitempty"`
	Trigger    string       `json:"trigger,omitempty"` // meta.trigger.name
	// Sessions that claimed the ticket, oldest first.
	Holders []string     `json:"holders,omitempty"`
	Edges   []store.Edge `json:"edges,omitempty"`
	// GitHub issue and PR URLs found in the body and comments.
	Links    []string    `json:"links,omitempty"`
	Children []TraceNode `json:"children,omitempty"`
}

// TriggerInfo is one configured trigger's state now.
type TriggerInfo struct {
	Name      string     `json:"name"`
	Holds     bool       `json:"holds"`          // would fire on the next tick
	Why       string     `json:"why,omitempty"`  // first condition that does not hold
	Open      string     `json:"open,omitempty"` // open ticket it created
	LastFired *time.Time `json:"last_fired,omitempty"`
	NextCron  *time.Time `json:"next_cron,omitempty"`
	// Stalls are the stall keys a stall trigger would start runs for.
	Stalls []string `json:"stalls,omitempty"`
}

// ConfigShowResult is the active merged config, keyed as in jobshop.yaml,
// and for each "agents.<name>", "pools.<name>" and "triggers.<name>" the
// files that set it in merge order (packs first, jobshop.yaml last).
type ConfigShowResult struct {
	Config  map[string]any      `json:"config"`
	Sources map[string][]string `json:"sources"`
	// Instances are the packs entries with their vars and the pools,
	// workflows and triggers each declares, prefix applied (#248).
	Instances []config.PackInstance `json:"instances"`
}

// RunStartParams starts a workflow run. Key, when set, dedups: an open run
// of the same workflow and key is returned instead of a new one.
type RunStartParams struct {
	Workflow string            `json:"workflow"`
	Vars     map[string]string `json:"vars,omitempty"`
	Key      string            `json:"key,omitempty"`
	Project  string            `json:"project,omitempty"`
	// After lists tickets (often other runs) the first step waits on. The
	// edges are created with the steps, so no step is claimable before
	// them (#149). A repeated start with a key adds no edges.
	After []string `json:"after,omitempty"`
}

// RunResult is a run ticket and its steps in order.
type RunResult struct {
	Run   store.Ticket   `json:"run"`
	Steps []store.Ticket `json:"steps"`
	// Existing is true when start returned an open run for the key.
	Existing bool `json:"existing,omitempty"`
	// Waits are steps held back by a full resource (#408).
	Waits []store.ResourceWait `json:"waits,omitempty"`
}

// BatchStartParams starts one run per issue under a batch ticket (#169).
// Order holds "a>b" or "a>b>c" chains: b's first step waits on a's last.
// A repeated start with the same key returns the open batch.
type BatchStartParams struct {
	Key      string   `json:"key"`
	Issues   []string `json:"issues"`
	Order    []string `json:"order,omitempty"`
	Workflow string   `json:"workflow,omitempty"` // default pr-work
	Project  string   `json:"project,omitempty"`
}

// BatchResult is a batch ticket and its member runs in issue order.
type BatchResult struct {
	Batch store.Ticket `json:"batch"`
	Runs  []RunResult  `json:"runs"`
	// Existing is true when start returned an open batch for the key.
	Existing bool `json:"existing,omitempty"`
	// Adopted lists runs that existed and were re-parented.
	Adopted []string `json:"adopted,omitempty"`
	// BlockedBy maps a member run id to the issues of other members whose
	// open steps block its current step (batch.show only).
	BlockedBy map[string][]string `json:"blocked_by,omitempty"`
}

// BatchListParams selects open batches.
type BatchListParams struct {
	Project string `json:"project,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// RunListParams selects runs. State is a store.Run* value.
type RunListParams struct {
	Project string `json:"project,omitempty"`
	State   string `json:"state,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// RunResumeParams reopens a run's step (default: its first unfinished
// step) with failures reset.
type RunResumeParams struct {
	ID   string `json:"id"`
	Step string `json:"step,omitempty"`
}

// SessionSendParams delivers text to a session as a message. Name the
// session by ID, or by Pool and Key; with Pool and Key and no open session
// for the key, a chat pool starts one. A busy session gets the message
// after its turn; Interrupt cancels the turn first.
type SessionSendParams struct {
	ID        string `json:"id,omitempty"`
	Pool      string `json:"pool,omitempty"`
	Key       string `json:"key,omitempty"`
	Text      string `json:"text"`
	Interrupt bool   `json:"interrupt,omitempty"`
	// Context is background for the session's model, delivered with the
	// message in its own delimited block of the prompt. Clients that
	// show the conversation get it apart from the visible text
	// (SessionEvent.Context).
	Context string `json:"context,omitempty"`
	// Attach marks a send from `session attach --rw`; it is refused
	// unless the session's pool sets attach: rw.
	Attach bool `json:"attach,omitempty"`
}

// SessionSendResult names the session and the message ticket.
type SessionSendResult struct {
	Session string       `json:"session"`
	Message store.Ticket `json:"message"`
	Started bool         `json:"started,omitempty"` // the send started the session
}

// SessionSubscribeParams streams a session's events after SinceSeq. The
// call stays open, sending NotifySessionEvent notifications, until the
// session closes or the client hangs up. Reconnect with the last seq seen.
type SessionSubscribeParams struct {
	ID       string `json:"id"`
	SinceSeq int64  `json:"since_seq,omitempty"`
}

// SessionSubscribeResult is sent when the stream ends.
type SessionSubscribeResult struct {
	Seq        int64  `json:"seq"` // last journal seq sent
	Closed     bool   `json:"closed,omitempty"`
	Resolution string `json:"resolution,omitempty"`
}

// NotifySessionEvent is the notification session.subscribe sends.
const NotifySessionEvent = "session.event"

// Session event types.
const (
	EventTurnStart = "turn_start" // Text is the prompt
	EventText      = "text"       // Text is an agent message chunk
	EventTool      = "tool"       // Tool is set
	EventTurnEnd   = "turn_end"   // StopReason, or Error
	EventState     = "state"      // Status and Resolution; Seq is 0
)

// SessionEvent is one item of a session's output, read from its journal.
type SessionEvent struct {
	Session string     `json:"session"`
	Seq     int64      `json:"seq,omitempty"`
	Time    *time.Time `json:"time,omitempty"`
	Type    string     `json:"type"`
	Text    string     `json:"text,omitempty"`
	// Context is the part of a turn_start prompt that was addressed to
	// the model only (session.send context), kept out of Text.
	Context    string       `json:"context,omitempty"`
	Tool       *ToolSummary `json:"tool,omitempty"`
	StopReason string       `json:"stop_reason,omitempty"`
	Error      string       `json:"error,omitempty"`
	Status     store.Status `json:"status,omitempty"`
	Resolution string       `json:"resolution,omitempty"`
}

// ToolSummary is one tool call state change.
type ToolSummary struct {
	ID     string `json:"id"`
	Title  string `json:"title,omitempty"`
	Status string `json:"status"`
}

// Subscribe calls session.subscribe and passes each event to fn until the
// stream ends or ctx is done. fn runs on the connection's reader, in order.
// One subscription per client at a time. The result's Seq is the last seq
// seen even when ctx ends the call, so it can be passed to the next one.
func (c *Client) Subscribe(ctx context.Context, p SessionSubscribeParams, fn func(SessionEvent)) (SessionSubscribeResult, error) {
	var seen atomic.Int64
	seen.Store(p.SinceSeq)
	wrap := func(ev SessionEvent) {
		if ev.Seq > seen.Load() {
			seen.Store(ev.Seq)
		}
		fn(ev)
	}
	c.onEvent.Store(&wrap)
	defer c.onEvent.Store(nil)
	var res SessionSubscribeResult
	err := c.Call(ctx, MethodSessionSubscribe, p, &res)
	res.Seq = max(res.Seq, seen.Load())
	return res, err
}

// MethodEventsSubscribe streams store change events (#147).
const MethodEventsSubscribe = "events.subscribe"

// NotifyEvent is the notification events.subscribe sends.
const NotifyEvent = "events.event"

// Change event kinds. Ready and Lagged are stream control: Ready is sent
// once replay is done and live events follow; Lagged means the subscriber
// fell behind, the stream ends and the client must resync.
const (
	EventKindTicket  = "ticket"
	EventKindRun     = "run"
	EventKindSession = "session"
	EventKindReady   = "ready"
	// EventKindPool: a pool scale or session cordon change (#281); Op is
	// up, cordon, uncordon, manual or auto.
	EventKindPool   = "pool"
	EventKindLagged = "lagged"
	// EventKindActivity: new external activity items (#255); refetch
	// activity.list. Seq is 0: it is not a history row.
	EventKindActivity = "activity"
)

// Session event ops; other session history ops (claim, update) keep their
// history name.
const (
	OpSessionStart = "start"
	OpSessionStop  = "stop"
	OpSessionCrash = "crash"
	// OpSessionConfig: the controller set an ACP config option on a live
	// session (#304); Option, Old and New say which.
	OpSessionConfig = "config"
	OpRunState      = "state"
)

// EventsSubscribeParams selects change events. SinceSeq > 0 replays
// history rows after that seq first, so a reconnect misses nothing.
// Kinds and Labels filter: an event matches if its kind is in Kinds and
// its ticket has (or the label op names) one of Labels; empty matches all.
type EventsSubscribeParams struct {
	SinceSeq int64    `json:"since_seq,omitempty"`
	Kinds    []string `json:"kinds,omitempty"`
	Labels   []string `json:"labels,omitempty"`
}

// EventsSubscribeResult is sent when the stream ends.
type EventsSubscribeResult struct {
	Seq    int64 `json:"seq"` // last history seq sent
	Lagged bool  `json:"lagged,omitempty"`
}

// Event is one store change, derived from a ticket history row. Seq is the
// history seq. Ticket fields (Status, Labels, State) are read when the
// event is sent, so they can be newer than the row.
type Event struct {
	Seq        int64        `json:"seq"`
	Time       time.Time    `json:"time,omitzero"`
	Kind       string       `json:"kind"`
	Op         string       `json:"op,omitempty"`
	ID         string       `json:"id,omitempty"`
	Revision   int64        `json:"revision,omitempty"`
	Actor      string       `json:"actor,omitempty"`
	TicketKind store.Kind   `json:"ticket_kind,omitempty"`
	Label      string       `json:"label,omitempty"` // label_add, label_remove
	Status     store.Status `json:"status,omitempty"`
	Resolution string       `json:"resolution,omitempty"`
	State      string       `json:"state,omitempty"` // run state
	Labels     []string     `json:"labels,omitempty"`
	Pool       string       `json:"pool,omitempty"`    // pool events
	Desired    int          `json:"desired,omitempty"` // pool events
	Option     string       `json:"option,omitempty"`  // session config events
	Old        string       `json:"old,omitempty"`
	New        string       `json:"new,omitempty"`
}

// Events calls events.subscribe and passes each event to fn until the
// stream ends or ctx is done. fn runs on the connection's reader, in
// order. The result's Seq is the last seq seen.
func (c *Client) Events(ctx context.Context, p EventsSubscribeParams, fn func(Event)) (EventsSubscribeResult, error) {
	var seen atomic.Int64
	seen.Store(p.SinceSeq)
	wrap := func(ev Event) {
		if ev.Seq > seen.Load() {
			seen.Store(ev.Seq)
		}
		fn(ev)
	}
	c.onChange.Store(&wrap)
	defer c.onChange.Store(nil)
	var res EventsSubscribeResult
	err := c.Call(ctx, MethodEventsSubscribe, p, &res)
	res.Seq = max(res.Seq, seen.Load())
	return res, err
}

// handle receives notifications from the controller.
func (c *Client) handle(_ context.Context, method string, params json.RawMessage) (any, *acp.RequestError) {
	if method == NotifyEvent {
		fn := c.onChange.Load()
		if fn == nil {
			return nil, nil
		}
		var ev Event
		if err := json.Unmarshal(params, &ev); err != nil {
			return nil, acp.NewInvalidParams(err.Error())
		}
		(*fn)(ev)
		return nil, nil
	}
	if method != NotifySessionEvent {
		return nil, acp.NewMethodNotFound(method)
	}
	fn := c.onEvent.Load()
	if fn == nil {
		return nil, nil
	}
	var ev SessionEvent
	if err := json.Unmarshal(params, &ev); err != nil {
		return nil, acp.NewInvalidParams(err.Error())
	}
	(*fn)(ev)
	return nil, nil
}

// Attention item kinds (#185).
const (
	AttentionRun      = "run"      // run in needs_human with no open escalation ticket
	AttentionTicket   = "ticket"   // open ticket labeled needs-human (escalations, pool quarantines)
	AttentionQuestion = "question" // open ticket labeled question: a worker asks a human
	AttentionPool     = "pool"     // quarantined pool with no open escalation
	AttentionRollback = "rollback" // self-deploy rolled back a commit
	AttentionConfig   = "config"   // controller serves its last good config: a live pack failed to load (#227)
	AttentionResource = "resource" // step waiting on a full resource past attention.resource_wait (#408)
	AttentionIngest   = "ingest"   // no live external ingest worker, or an admission backlog (#443)
	AttentionDecision = "decision" // open ticket labeled proposed: a planner-proposed target to approve (#488)
	AttentionStale    = "stale"    // open ticket no pool will claim, older than attention.orphan_age (#495)
)

// AttentionItem is one thing waiting on a human.
type AttentionItem struct {
	Kind    string    `json:"kind"`
	ID      string    `json:"id"` // ticket id, pool name or rolled-back commit
	Project string    `json:"project,omitempty"`
	Title   string    `json:"title"`
	Since   time.Time `json:"since"`
	// Repo and Ref are meta.github's repo and issue or PR, when set.
	Repo string `json:"repo,omitempty"`
	Ref  string `json:"ref,omitempty"`
	// Owner is who must act, for a wait past its timeout (#381).
	Owner string `json:"owner,omitempty"`
}

// AttentionResult lists what waits on a human, oldest first. Empty when
// healthy.
type AttentionResult struct {
	Items []AttentionItem `json:"items"`
}

// IncidentsParams selects the report window. Zero Since means 24h.
type IncidentsParams struct {
	Since time.Duration `json:"since,omitempty"`
}

// IncidentsResult is the evidence for a planner reliability pass: grouped
// controller WARN/ERROR lines (since the controller started, at most),
// sessions that ended crashed or stuck, and steps with failed attempts.
type IncidentsResult struct {
	Since    time.Time         `json:"since"`
	Logs     []LogGroup        `json:"logs"`
	Sessions []SessionIncident `json:"sessions"`
	Steps    []StepIncident    `json:"steps"`
}

// LogGroup is the controller log lines with one level and message.
type LogGroup struct {
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Count   int               `json:"count"`
	First   time.Time         `json:"first"`
	Last    time.Time         `json:"last"`
	Sample  map[string]string `json:"sample,omitempty"` // attrs of the last line
}

// SessionIncident is a session that ended crashed or stuck.
type SessionIncident struct {
	ID         string    `json:"id"`
	Pool       string    `json:"pool,omitempty"`
	Resolution string    `json:"resolution"`
	ClosedAt   time.Time `json:"closed_at"`
	Reason     string    `json:"reason,omitempty"`
}

// StepIncident is a run step with at least one failed attempt.
type StepIncident struct {
	ID         string       `json:"id"`
	Run        string       `json:"run"`
	Step       string       `json:"step"`
	Title      string       `json:"title,omitempty"`
	Status     store.Status `json:"status"`
	Resolution string       `json:"resolution,omitempty"`
	Failures   int          `json:"failures"`
	Retries    int          `json:"retries"`
	LastError  string       `json:"last_error,omitempty"` // last comment
}

// Guard decisions (#188).
const (
	GuardAllow    = "allow"
	GuardDeny     = "deny"
	GuardAskHuman = "ask_human"
)

// GuardParams asks the guardrail hooks before an outside action such as
// push, pr-create, merge, comment or message.
type GuardParams struct {
	Action string          `json:"action"`
	Target string          `json:"target,omitempty"` // e.g. a PR url or branch
	Detail string          `json:"detail,omitempty"`
	Extra  json.RawMessage `json:"extra,omitempty"`
}

// GuardResult is the hooks' answer. Ticket is the needs-human ticket an
// ask_human answer filed or read.
type GuardResult struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
	Ticket   string `json:"ticket,omitempty"`
}

// Report states (#186). failed and dropped record why before the worker
// closes the ticket; report does not close.
const (
	ReportProgress = "progress"
	ReportWaiting  = "waiting"
	ReportReady    = "ready"
	ReportBlocked  = "blocked"
	ReportFailed   = "failed"
	ReportDropped  = "dropped"
)

// ReportStates lists the valid report states.
var ReportStates = []string{ReportProgress, ReportWaiting, ReportReady, ReportBlocked, ReportFailed, ReportDropped}

// ReportParams records a worker's state on a ticket it holds (meta.report).
type ReportParams struct {
	ID    string `json:"id,omitempty"`
	State string `json:"state"`
	Note  string `json:"note,omitempty"`
	// Question files a needs-human question whose answer is delivered to
	// the reporting session.
	Question string `json:"question,omitempty"`
	// Until (waiting only) releases the ticket deferred until then.
	Until *time.Time `json:"until,omitempty"`
}

// ReportResult is the reported ticket and the question filed, if any.
type ReportResult struct {
	Ticket   store.Ticket  `json:"ticket"`
	Question *store.Ticket `json:"question,omitempty"`
}

// AnswerParams answers a question ticket (admin).
type AnswerParams struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	From string `json:"from,omitempty"` // who answered, e.g. a chat sender; default the caller
}

// AnswerResult is the closed question and the message to the asking
// session (nil when that session has ended).
type AnswerResult struct {
	Question store.Ticket  `json:"question"`
	Message  *store.Ticket `json:"message,omitempty"`
}

// LearnParams records a learning (#264). Kind is correction, failure or
// workaround; Scope is global, repo:<owner/name> or pool:<name> (default:
// the ticket's repo or pool, else global). Ticket defaults to the
// session's current ticket.
type LearnParams struct {
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
	Fix     string `json:"fix,omitempty"`
	Scope   string `json:"scope,omitempty"`
	Ticket  string `json:"ticket,omitempty"`
}

// LearnListParams filters learn.list. The result is []store.Learning,
// newest first.
type LearnListParams struct {
	Since  time.Duration `json:"since,omitempty"`
	Scope  string        `json:"scope,omitempty"`
	Kind   string        `json:"kind,omitempty"`
	Status string        `json:"status,omitempty"`
	Ticket string        `json:"ticket,omitempty"`
	Limit  int           `json:"limit,omitempty"`
}

// LearnUpdateParams sets the status and PR of learnings (#273). The
// curator sets PR while they stay open; whoever sees the PR merge or
// close sets applied or rejected. Empty Status keeps each one's status,
// empty PR keeps its PR. The result is the updated []store.Learning.
type LearnUpdateParams struct {
	IDs    []int64 `json:"ids"`
	Status string  `json:"status,omitempty"`
	PR     string  `json:"pr,omitempty"`
	// MergedAt is when the PR merged; with status applied it starts the
	// curator gate's window (#274). Zero means now.
	MergedAt time.Time `json:"merged_at,omitzero"`
}

// LearnEvaluateParams runs the curator gate (#274) on pending memory
// edits that are due: Runs runs in their pools since the merge, or Age
// since it (defaults 20 and 72h). A nonzero ID evaluates that edit even
// when not due. The result is the evaluated []store.MemoryEdit.
type LearnEvaluateParams struct {
	ID   int64         `json:"id,omitempty"`
	Runs int           `json:"runs,omitempty"`
	Age  time.Duration `json:"age,omitempty"`
}

// LearnEditsParams filters learn.edits. The result is []store.MemoryEdit.
type LearnEditsParams struct {
	State string `json:"state,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// WriteKeyParams claims or finishes an idempotency key for a tracker
// write (#280). Claim needs Key and Op; Finish needs Key and Result. The
// result of both is the stored store.WriteKey. Keys are namespaced by
// project, so two projects may use the same key.
type WriteKeyParams struct {
	Key    string `json:"key"`
	Op     string `json:"op,omitempty"`
	Result string `json:"result,omitempty"`
}

// Waits (#381).
const (
	MethodWaitAdd   = "wait.add"
	MethodWaitList  = "wait.list"
	MethodWaitClear = "wait.clear"
	// AttentionWait is a wait past its timeout.
	AttentionWait = "wait"
)

// WaitAddParams parks a ticket (default: the session's current one)
// until the conditions hold. A held ticket is released without counting
// a step failure.
type WaitAddParams struct {
	ID      string           `json:"id,omitempty"`
	Conds   []store.WaitCond `json:"conds"`
	Any     bool             `json:"any,omitempty"`
	Reason  string           `json:"reason"`
	Owner   string           `json:"owner,omitempty"`
	Resume  bool             `json:"resume,omitempty"`
	Timeout *time.Time       `json:"timeout,omitempty"`
}

// WaitListParams lists waiting tickets; Project empty is all (admin) or
// the session's.
type WaitListParams struct {
	Project string `json:"project,omitempty"`
}

// WaitItem is one waiting ticket.
type WaitItem struct {
	ID      string     `json:"id"`
	Project string     `json:"project"`
	Title   string     `json:"title"`
	Wait    store.Wait `json:"wait"`
}

// WaitClearParams removes a ticket's wait; the ticket becomes ready.
type WaitClearParams struct {
	ID     string `json:"id,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// Resource slot methods (#407): a resource provider adds, removes and
// fails its slots; list shows slots and occupancy.
const (
	MethodResourceAdd       = "resource.add"
	MethodResourceRemove    = "resource.remove"
	MethodResourceUnhealthy = "resource.unhealthy"
	MethodResourceList      = "resource.list"
)

// SlotParams names one slot of a provider resource.
type SlotParams struct {
	Resource string            `json:"resource"`
	Slot     string            `json:"slot"`
	Env      map[string]string `json:"env,omitempty"`
	Reason   string            `json:"reason,omitempty"`
}

// SlotResult is resource.unhealthy: the slot and the ticket requeued.
type SlotResult struct {
	Slot     store.Slot `json:"slot"`
	Requeued string     `json:"requeued,omitempty"`
}

// ResourceListResult is resource.list.
type ResourceListResult struct {
	Slots []store.Slot        `json:"slots"`
	Use   []store.ResourceUse `json:"use"`
}

// Artifact methods (#508): deliverables a human reads and acks in Desk.
const (
	MethodArtifactDeliver = "artifact.deliver"
	MethodArtifactList    = "artifact.list"
	MethodArtifactShow    = "artifact.show"
	MethodArtifactAck     = "artifact.ack"
)

// ArtifactDeliverParams is artifact.deliver. Files are absolute paths
// inside the session's dir or worktree; the controller copies them.
type ArtifactDeliverParams struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Files  []string `json:"files,omitempty"`
	Ticket string   `json:"ticket,omitempty"` // source; default the session's held ticket
}

// ArtifactFile is one attachment in meta.artifact.files.
type ArtifactFile struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Type   string `json:"type,omitempty"`
	Pruned bool   `json:"pruned,omitempty"`
}

// ArtifactMeta is meta.artifact on an artifact ticket.
type ArtifactMeta struct {
	Ticket  string         `json:"ticket,omitempty"`
	Run     string         `json:"run,omitempty"`
	Trigger string         `json:"trigger,omitempty"`
	Pool    string         `json:"pool,omitempty"`
	Session string         `json:"session,omitempty"`
	Files   []ArtifactFile `json:"files,omitempty"`
}

// ArtifactListParams filters artifact.list.
type ArtifactListParams struct {
	Unread  bool   `json:"unread,omitempty"`
	Trigger string `json:"trigger,omitempty"`
	Pool    string `json:"pool,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// Artifact is one artifact ticket with its parsed meta.
type Artifact struct {
	Ticket store.Ticket `json:"ticket"`
	Meta   ArtifactMeta `json:"artifact"`
	Read   bool         `json:"read"`
}

// ArtifactShowResult is artifact.show: the artifact, its comments and
// history, and the attachment dir (admin only; Desk serves files from it).
type ArtifactShowResult struct {
	Artifact
	Comments []store.Comment      `json:"comments"`
	History  []store.HistoryEntry `json:"history"`
	Dir      string               `json:"dir,omitempty"`
}

// ArtifactAckParams is artifact.ack.
type ArtifactAckParams struct {
	ID   string `json:"id"`
	Note string `json:"note,omitempty"`
}

// MethodSessionJournal reads a session's ACP journal (#529).
const MethodSessionJournal = "session.journal"

// SessionJournalParams pages a session's journal by seq. Kinds filters
// on msg.method or params.update.sessionUpdate; From on "agent" or
// "controller". Limit defaults to 500, at most 5000.
type SessionJournalParams struct {
	ID       string   `json:"id"`
	SinceSeq int64    `json:"since_seq,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Kinds    []string `json:"kinds,omitempty"`
	From     string   `json:"from,omitempty"`
}

// JournalFrame is one redacted journal frame.
type JournalFrame struct {
	Seq  int64           `json:"seq"`
	T    int64           `json:"t"`
	From string          `json:"from"`
	Msg  json.RawMessage `json:"msg"`
}

// SessionJournalResult is one page. NextSeq is the since_seq for the
// next call; Done means the page reached the end of the journal.
type SessionJournalResult struct {
	Frames  []JournalFrame `json:"frames"`
	NextSeq int64          `json:"next_seq"`
	Done    bool           `json:"done"`
}
