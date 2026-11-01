// Package rpc defines the JSON-RPC protocol between the jobshop CLI and the
// controller: method names, parameter and result types, error mapping, and
// a client. Transport is NDJSON over a unix socket, using the JSON-RPC
// connection from github.com/coder/acp-go-sdk (the same one ACP uses).
package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	acp "github.com/coder/acp-go-sdk"

	"github.com/elee1766/jobshop/pkg/store"
)

// Method names.
const (
	MethodTicketCreate = "ticket.create"
	MethodTicketShow   = "ticket.show"
	MethodTicketList   = "ticket.list"
	MethodTicketUpdate = "ticket.update"
	MethodTicketClose  = "ticket.close"
	MethodTicketReopen = "ticket.reopen"
	MethodReady        = "ticket.ready"
	MethodCommentAdd   = "comment.add"
	MethodEdgeAdd      = "edge.add"
	MethodEdgeRemove   = "edge.remove"
	MethodLabelAdd     = "label.add"
	MethodLabelRemove  = "label.remove"
	MethodClaim        = "claim"
	MethodHeartbeat    = "heartbeat"
	MethodRelease      = "release"
	MethodSessionNew   = "session.new"
	MethodSessionList  = "session.list"
	MethodWhoami       = "whoami"
	MethodShutdown     = "controller.shutdown"
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
type SessionListParams struct {
	All bool `json:"all,omitempty"`
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
	Exit     string        `json:"exit,omitempty"` // exit.json reason, for dead hosts
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
	nc   net.Conn
	conn *acp.Connection
}

// Dial connects to a unix socket.
func Dial(ctx context.Context, path string) (*Client, error) {
	var d net.Dialer
	nc, err := d.DialContext(ctx, "unix", path)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w (is the controller running?)", path, err)
	}
	// The controller never calls back, so the client has no handler.
	conn := acp.NewConnection(nil, nc, nc)
	conn.SetLogger(QuietLogger(slog.New(slog.DiscardHandler)))
	return &Client{nc: nc, conn: conn}, nil
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
	res, err := acp.SendRequest[json.RawMessage](c.conn, ctx, method, params)
	if err != nil {
		return FromWire(err)
	}
	if out == nil || len(res) == 0 {
		return nil
	}
	return json.Unmarshal(res, out)
}

// Close closes the connection.
func (c *Client) Close() error {
	if err := c.nc.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}
	return nil
}
