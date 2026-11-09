// Package journal ships session journal frames to VictoriaLogs and
// queries them back (#530). It maps one ACP frame to one row and speaks
// VictoriaLogs' JSON-lines ingest and LogsQL query endpoints; the
// controller owns cursors, batching and access.
package journal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/elee1766/jobshop/pkg/redact"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Prompt _meta keys the controller sets on session/prompt. The shipper
// reads MetaTicket to tag later frames with the ticket the turn works on.
const (
	MetaTicket = "jobshop/ticket"
	MetaModel  = "jobshop/model"
	MetaEffort = "jobshop/effort"
)

// StreamFields are the VictoriaLogs stream fields: one stream per session.
const StreamFields = "project,pool,session"

// Source is the session a frame came from.
type Source struct {
	Project, Pool, Session string
}

// Work is what a ticket belongs to, for the row's ticket, run, workflow
// and step fields.
type Work struct {
	Ticket, Run, Workflow, Step string
}

// Mapper turns one session's frames into rows in seq order. It tracks the
// ticket of the current turn and tool titles by call id, so it must see
// every frame from the start of the journal, shipped or not.
type Mapper struct {
	Src    Source
	Redact *redact.Redactor
	// Lookup resolves a ticket id to its run, workflow and step. Nil or
	// an empty result leaves only the ticket.
	Lookup func(ticket string) Work

	work   Work
	titles map[string]string
}

// Row is one mapped frame.
type Row struct {
	Seq  int64
	T    int64  // unix ms
	Line []byte // JSON object, no newline
}

// FrameInfo is the part of an ACP frame rows are tagged from.
type FrameInfo struct {
	Method string `json:"method"`
	Params struct {
		Meta   map[string]any `json:"_meta"`
		Update struct {
			SessionUpdate string `json:"sessionUpdate"`
			ToolCallID    string `json:"toolCallId"`
			Title         string `json:"title"`
			Status        string `json:"status"`
		} `json:"update"`
	} `json:"params"`
}

// Observe updates the turn state from f without producing a row: frames
// already shipped still set the ticket and tool titles.
func (m *Mapper) Observe(f host.Frame) FrameInfo {
	var msg FrameInfo
	_ = json.Unmarshal(f.Msg, &msg)
	if m.titles == nil {
		m.titles = map[string]string{}
	}
	if f.From == "controller" && msg.Method == "session/prompt" {
		tk, _ := msg.Params.Meta[MetaTicket].(string)
		if tk != m.work.Ticket {
			m.work = Work{Ticket: tk}
			if tk != "" && m.Lookup != nil {
				m.work = m.Lookup(tk)
				m.work.Ticket = tk
			}
		}
		clear(m.titles)
	}
	u := msg.Params.Update
	if u.ToolCallID != "" && u.Title != "" {
		m.titles[u.ToolCallID] = u.Title
	}
	return msg
}

// Map observes f and returns its row.
func (m *Mapper) Map(f host.Frame) (Row, error) {
	msg := m.Observe(f)
	u := msg.Params.Update
	raw := f.Msg
	if m.Redact != nil {
		raw = m.Redact.JSON(raw)
	}
	r := map[string]any{
		"_time":   time.UnixMilli(f.T).UTC().Format(time.RFC3339Nano),
		"_msg":    string(raw),
		"project": m.Src.Project,
		"pool":    m.Src.Pool,
		"session": m.Src.Session,
		"seq":     f.Seq,
		"from":    f.From,
	}
	set := func(k, v string) {
		if v != "" {
			r[k] = v
		}
	}
	set("method", msg.Method)
	set("update", u.SessionUpdate)
	set("status", u.Status)
	if u.ToolCallID != "" {
		set("tool", m.titles[u.ToolCallID])
	}
	set("ticket", m.work.Ticket)
	set("run", m.work.Run)
	set("workflow", m.work.Workflow)
	set("step", m.work.Step)
	b, err := json.Marshal(r)
	if err != nil {
		return Row{}, err
	}
	return Row{Seq: f.Seq, T: f.T, Line: b}, nil
}

// Client talks to one VictoriaLogs.
type Client struct {
	URL     string // base, e.g. http://127.0.0.1:9428
	Headers map[string]string
	HTTP    *http.Client // default: 30s timeout
}

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return defaultHTTP
}

var defaultHTTP = telemetry.HTTPClient("journal", 30*time.Second)

func (c *Client) endpoint(path string, q url.Values) string {
	return strings.TrimRight(c.URL, "/") + path + "?" + q.Encode()
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		_ = resp.Body.Close()
		return nil, &StatusError{Code: resp.StatusCode, Body: strings.TrimSpace(string(b))}
	}
	return resp, nil
}

// StatusError is a non-2xx VictoriaLogs answer. A 400 on a query is a
// LogsQL syntax error the caller made.
type StatusError struct {
	Code int
	Body string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("victorialogs: HTTP %d: %s", e.Code, e.Body)
}

// Insert posts newline-separated rows to /insert/jsonline.
func (c *Client) Insert(ctx context.Context, body []byte) error {
	q := url.Values{"_stream_fields": {StreamFields}, "_time_field": {"_time"}, "_msg_field": {"_msg"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("/insert/jsonline", q), bytes.NewReader(body))
	if err != nil {
		return err
	}
	// VictoriaLogs ignores a form-encoded body: it must say JSON.
	req.Header.Set("Content-Type", "application/stream+json")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

// Query is one LogsQL request.
type Query struct {
	LogsQL     string
	Start, End time.Time // zero: unbounded
	Limit      int       // 0: no limit
	// Project, if set, restricts the query to that project's streams.
	// VictoriaLogs applies it outside the LogsQL, so the query text
	// cannot widen it.
	Project string
}

// Run sends q to /select/logsql/query and calls fn for each result row
// as it arrives.
func (c *Client) Run(ctx context.Context, q Query, fn func(json.RawMessage) error) error {
	v := url.Values{"query": {q.LogsQL}}
	if !q.Start.IsZero() {
		v.Set("start", q.Start.UTC().Format(time.RFC3339Nano))
	}
	if !q.End.IsZero() {
		v.Set("end", q.End.UTC().Format(time.RFC3339Nano))
	}
	if q.Limit > 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Project != "" {
		v.Set("extra_stream_filters", "{project="+strconv.Quote(q.Project)+"}")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.URL, "/")+"/select/logsql/query",
		strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64<<10), 64<<20)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		if !json.Valid(line) {
			return errors.New("victorialogs: query returned a non-JSON line")
		}
		if err := fn(append(json.RawMessage(nil), line...)); err != nil {
			return err
		}
	}
	return sc.Err()
}

// Flush asks VictoriaLogs to make recent rows searchable now. Tests and
// backfill use it; normal shipping waits for its own flush (about 1s).
func (c *Client) Flush(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.URL, "/")+"/internal/force_flush", nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}
