// Package rpc defines the JSON-RPC protocol between the jobshop CLI and the
// controller: method names, parameter and result types, error mapping, and
// a client. Transport is NDJSON over a unix socket (go.lsp.dev/jsonrpc2).
package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"go.lsp.dev/jsonrpc2"

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

// SessionNewParams creates a session ticket and its CLI socket.
type SessionNewParams struct {
	Project string `json:"project,omitempty"`
	Pool    string `json:"pool,omitempty"`
	Agent   string `json:"agent"`
	Key     string `json:"key,omitempty"`
	Title   string `json:"title,omitempty"`
}

// SessionNewResult names the new session and where its socket is.
type SessionNewResult struct {
	Ticket store.Ticket `json:"ticket"`
	Socket string       `json:"socket"`
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
	CodeNotFound     jsonrpc2.Code = -32001
	CodeConflict     jsonrpc2.Code = -32002
	CodeNotHolder    jsonrpc2.Code = -32003
	CodeNoWork       jsonrpc2.Code = -32004
	CodeInvalid      jsonrpc2.Code = -32005
	CodeCycle        jsonrpc2.Code = -32006
	CodeClosed       jsonrpc2.Code = -32007
	CodeAlreadyExist jsonrpc2.Code = -32008
	CodeForbidden    jsonrpc2.Code = -32009
)

// ErrForbidden means the caller's scope does not allow the operation.
var ErrForbidden = errors.New("forbidden")

var codes = []struct {
	code jsonrpc2.Code
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
func ToWire(err error) error {
	if err == nil {
		return nil
	}
	var we *jsonrpc2.Error
	if errors.As(err, &we) {
		return we
	}
	for _, c := range codes {
		if errors.Is(err, c.err) {
			return jsonrpc2.NewError(c.code, err.Error())
		}
	}
	return jsonrpc2.NewError(jsonrpc2.InternalError, err.Error())
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
	var we *jsonrpc2.Error
	if !errors.As(err, &we) {
		return err
	}
	for _, c := range codes {
		if we.Code == c.code {
			return &remoteError{msg: we.Message, kind: c.err}
		}
	}
	return errors.New(we.Message)
}

// Codec is encoding/json, so json.RawMessage fields pass through verbatim.
type Codec struct{}

// Marshal implements jsonrpc2.Codec.
func (Codec) Marshal(v any) ([]byte, error) { return json.Marshal(v) }

// Unmarshal implements jsonrpc2.Codec.
func (Codec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }

// Client is a connection to a controller or session socket.
type Client struct {
	nc   net.Conn
	conn jsonrpc2.Conn
}

// Dial connects to a unix socket.
func Dial(ctx context.Context, path string) (*Client, error) {
	var d net.Dialer
	nc, err := d.DialContext(ctx, "unix", path)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w (is the controller running?)", path, err)
	}
	conn := jsonrpc2.NewConn(jsonrpc2.NewNDJSONStream(nc), jsonrpc2.WithCodec(Codec{}))
	conn.Go(ctx, jsonrpc2.MethodNotFoundHandler)
	return &Client{nc: nc, conn: conn}, nil
}

// Call invokes method and decodes the result into out (may be nil).
func (c *Client) Call(ctx context.Context, method string, params, out any) error {
	_, err := c.conn.Call(ctx, method, params, out)
	return FromWire(err)
}

// Close closes the connection.
func (c *Client) Close() error {
	err := c.conn.Close()
	if cerr := c.nc.Close(); err == nil && !errors.Is(cerr, net.ErrClosed) {
		err = cerr
	}
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
