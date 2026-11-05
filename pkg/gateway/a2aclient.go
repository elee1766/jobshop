package gateway

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
)

// A2A client (#547): the outbound half of the A2A v1.0 JSON-RPC binding,
// stdlib only. It sends messages and reads tasks; it serves nothing, so
// Jobshop needs no public endpoint. Field names follow v1.0; replies in
// the v0.3 shape (kind-tagged results, lowercase states) are read too.

// A2A task states, normalized to the v0.3 spelling: v1.0's
// TASK_STATE_INPUT_REQUIRED reads as input-required.
const (
	A2ASubmitted     = "submitted"
	A2AWorking       = "working"
	A2AInputRequired = "input-required"
	A2AAuthRequired  = "auth-required"
	A2ACompleted     = "completed"
	A2AFailed        = "failed"
	A2ARejected      = "rejected"
	A2ACanceled      = "canceled"
)

// a2aState normalizes a task state from either protocol version.
func a2aState(s string) string {
	s = strings.ToLower(strings.TrimPrefix(s, "TASK_STATE_"))
	s = strings.ReplaceAll(s, "_", "-")
	if s == "cancelled" {
		s = A2ACanceled
	}
	return s
}

// a2aTerminal reports a state the task does not leave.
func a2aTerminal(s string) bool {
	switch s {
	case A2ACompleted, A2AFailed, A2ARejected, A2ACanceled:
		return true
	}
	return false
}

// a2aPart is one message or artifact part. Only text is used.
type a2aPart struct {
	Kind string `json:"kind,omitempty"` // v0.3
	Text string `json:"text,omitempty"`
}

type a2aMessage struct {
	Kind      string    `json:"kind,omitempty"`
	MessageID string    `json:"messageId"`
	Role      string    `json:"role"`
	Parts     []a2aPart `json:"parts"`
	ContextID string    `json:"contextId,omitempty"`
	TaskID    string    `json:"taskId,omitempty"`
}

type a2aTask struct {
	ID        string `json:"id"`
	ContextID string `json:"contextId"`
	Status    struct {
		State   string      `json:"state"`
		Message *a2aMessage `json:"message,omitempty"`
	} `json:"status"`
	Artifacts []struct {
		Parts []a2aPart `json:"parts"`
	} `json:"artifacts,omitempty"`
}

// a2aReply is a SendMessage or GetTask result: a message or a task.
type a2aReply struct {
	Message *a2aMessage
	Task    *a2aTask
}

// text joins the text parts of the reply: a message's parts, or a task's
// status message and artifacts.
func (r a2aReply) text() string {
	var out []string
	add := func(ps []a2aPart) {
		for _, p := range ps {
			if t := strings.TrimSpace(p.Text); t != "" {
				out = append(out, t)
			}
		}
	}
	if r.Message != nil {
		add(r.Message.Parts)
	}
	if r.Task != nil {
		if r.Task.Status.Message != nil {
			add(r.Task.Status.Message.Parts)
		}
		for _, a := range r.Task.Artifacts {
			add(a.Parts)
		}
	}
	return strings.Join(out, "\n\n")
}

// parseA2AReply reads v1.0 ({"task":...} or {"message":...}) and v0.3
// (a Task or Message tagged by kind) results.
func parseA2AReply(raw json.RawMessage) (a2aReply, error) {
	var wrap struct {
		Task    *a2aTask    `json:"task"`
		Message *a2aMessage `json:"message"`
		Kind    string      `json:"kind"`
		ID      string      `json:"id"`
		Status  *struct{}   `json:"status"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return a2aReply{}, err
	}
	switch {
	case wrap.Task != nil:
		return a2aReply{Task: wrap.Task}, nil
	case wrap.Message != nil:
		return a2aReply{Message: wrap.Message}, nil
	case wrap.Kind == "message":
		var m a2aMessage
		err := json.Unmarshal(raw, &m)
		return a2aReply{Message: &m}, err
	case wrap.Kind == "task" || wrap.Status != nil:
		var t a2aTask
		err := json.Unmarshal(raw, &t)
		return a2aReply{Task: &t}, err
	}
	return a2aReply{}, errors.New("a2a: result is neither a task nor a message")
}

// a2aCard is the part of an agent card the client reads.
type a2aCard struct {
	Name string `json:"name"`
	// v1.0: interfaces in preference order.
	SupportedInterfaces []struct {
		URL             string `json:"url"`
		ProtocolBinding string `json:"protocolBinding"`
		ProtocolVersion string `json:"protocolVersion"`
	} `json:"supportedInterfaces"`
	// v0.3: one main url and its transport.
	URL                string `json:"url"`
	PreferredTransport string `json:"preferredTransport"`
	ProtocolVersion    string `json:"protocolVersion"`
}

// endpoint picks the JSON-RPC interface: the first v1.x JSONRPC entry,
// then any JSONRPC entry, then the v0.3 url. Empty: none usable.
func (c a2aCard) endpoint() string {
	for _, i := range c.SupportedInterfaces {
		if strings.EqualFold(i.ProtocolBinding, "JSONRPC") && strings.HasPrefix(i.ProtocolVersion, "1.") {
			return i.URL
		}
	}
	for _, i := range c.SupportedInterfaces {
		if strings.EqualFold(i.ProtocolBinding, "JSONRPC") {
			return i.URL
		}
	}
	if c.URL != "" && (c.PreferredTransport == "" || strings.EqualFold(c.PreferredTransport, "JSONRPC")) {
		return c.URL
	}
	return ""
}

// a2aClient calls one peer. The token is only ever put in the
// Authorization header: errors carry the method and HTTP status, never
// the request or response body.
type a2aClient struct {
	base     string // configured url; the card is under it
	token    string
	http     *http.Client
	endpoint string // from the card
	ids      atomic.Int64
}

// errA2AAuth is a 401 or 403 from the peer.
var errA2AAuth = errors.New("a2a: peer refused the token")

func (c *a2aClient) do(ctx context.Context, method, url string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("A2A-Version", "1.0")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		// net/http errors name the URL, which holds no secret.
		return fmt.Errorf("a2a: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w (HTTP %d)", errA2AAuth, resp.StatusCode)
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("a2a: %s %s: HTTP %d", method, url, resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(out)
}

// card fetches the agent card and picks the endpoint. A card without a
// JSON-RPC interface falls back to the configured url.
func (c *a2aClient) card(ctx context.Context) error {
	var card a2aCard
	if err := c.do(ctx, http.MethodGet, strings.TrimRight(c.base, "/")+"/.well-known/agent-card.json", nil, &card); err != nil {
		return fmt.Errorf("agent card: %w", err)
	}
	c.endpoint = card.endpoint()
	if c.endpoint == "" {
		c.endpoint = c.base
	}
	return nil
}

func (c *a2aClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if c.endpoint == "" {
		if err := c.card(ctx); err != nil {
			return nil, err
		}
	}
	body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": c.ids.Add(1), "method": method, "params": params})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := c.do(ctx, http.MethodPost, c.endpoint, body, &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("a2a: %s: error %d: %s", method, resp.Error.Code, resp.Error.Message)
	}
	return resp.Result, nil
}

// send is SendMessage with text in contextID, continuing taskID when the
// peer asked for input on it.
func (c *a2aClient) send(ctx context.Context, contextID, taskID, text string) (a2aReply, error) {
	msg := a2aMessage{MessageID: newMessageID(), Role: "ROLE_USER", Parts: []a2aPart{{Text: text}}, ContextID: contextID, TaskID: taskID}
	raw, err := c.call(ctx, "SendMessage", map[string]any{"message": msg})
	if err != nil {
		return a2aReply{}, err
	}
	return parseA2AReply(raw)
}

// task is GetTask.
func (c *a2aClient) task(ctx context.Context, id string) (a2aTask, error) {
	raw, err := c.call(ctx, "GetTask", map[string]any{"id": id})
	if err != nil {
		return a2aTask{}, err
	}
	r, err := parseA2AReply(raw)
	if err != nil {
		return a2aTask{}, err
	}
	if r.Task == nil {
		return a2aTask{}, errors.New("a2a: GetTask returned no task")
	}
	return *r.Task, nil
}

func newMessageID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
