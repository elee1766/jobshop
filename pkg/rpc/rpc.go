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

	acp "github.com/coder/acp-go-sdk"

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
	MethodSessionNew       = "session.new"
	MethodSessionList      = "session.list"
	MethodSessionReap      = "session.reap"
	MethodSessionRestart   = "session.restart"
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
)

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
	Limit    int          `json:"limit,omitempty"`
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
}

// CloseParams closes a ticket. In a session scope the session must hold it.
type CloseParams struct {
	ID         string `json:"id,omitempty"`
	Resolution string `json:"resolution,omitempty"` // default "done"
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

// ReleaseParams releases a held ticket.
type ReleaseParams struct {
	ID     string `json:"id,omitempty"`
	Reason string `json:"reason,omitempty"`
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
	// Crashes counts unplanned host restarts in the crash window.
	Crashes int `json:"crashes,omitempty"`
	// BackoffUntil is set while the session waits out a provider rate limit.
	BackoffUntil *time.Time `json:"backoff_until,omitempty"`
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
	nc      net.Conn
	conn    *acp.Connection
	onEvent atomic.Pointer[func(SessionEvent)]
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
}

// withMeta adds _meta.traceparent from ctx to object params.
func withMeta(ctx context.Context, params any) (any, error) {
	tp := telemetry.TraceParent(ctx)
	if tp == "" {
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
	m, err := json.Marshal(Meta{TraceParent: tp})
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

// DrainResult lists drained pools after the call; "*" means all.
type DrainResult struct {
	Drained []string `json:"drained"`
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
}

// ConfigShowResult is the active merged config, keyed as in jobshop.yaml,
// and for each "agents.<name>", "pools.<name>" and "triggers.<name>" the
// files that set it in merge order (packs first, jobshop.yaml last).
type ConfigShowResult struct {
	Config  map[string]any      `json:"config"`
	Sources map[string][]string `json:"sources"`
}

// RunStartParams starts a workflow run. Key, when set, dedups: an open run
// of the same workflow and key is returned instead of a new one.
type RunStartParams struct {
	Workflow string            `json:"workflow"`
	Vars     map[string]string `json:"vars,omitempty"`
	Key      string            `json:"key,omitempty"`
	Project  string            `json:"project,omitempty"`
}

// RunResult is a run ticket and its steps in order.
type RunResult struct {
	Run   store.Ticket   `json:"run"`
	Steps []store.Ticket `json:"steps"`
	// Existing is true when start returned an open run for the key.
	Existing bool `json:"existing,omitempty"`
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
	Session    string       `json:"session"`
	Seq        int64        `json:"seq,omitempty"`
	Time       *time.Time   `json:"time,omitempty"`
	Type       string       `json:"type"`
	Text       string       `json:"text,omitempty"`
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

// handle receives notifications from the controller.
func (c *Client) handle(_ context.Context, method string, params json.RawMessage) (any, *acp.RequestError) {
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
