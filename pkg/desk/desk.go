// Package desk serves Jobshop Desk: a read-only dashboard over the admin
// API (#202). It is an admin-socket client like the gateway; it never
// opens the database. The Svelte UI is built into dist/ and embedded.
package desk

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

//go:embed all:dist
var dist embed.FS

// Controller is the part of the admin API Desk reads.
type Controller interface {
	Call(ctx context.Context, method string, params, out any) error
	Events(ctx context.Context, p rpc.EventsSubscribeParams, fn func(rpc.Event)) (rpc.EventsSubscribeResult, error)
	Subscribe(ctx context.Context, p rpc.SessionSubscribeParams, fn func(rpc.SessionEvent)) (rpc.SessionSubscribeResult, error)
}

// Config is one Desk server.
type Config struct {
	// Hosts are the Host header values accepted besides loopback names.
	Hosts []string
	// DeployState is the self-deploy state dir (deployed, failed files).
	// Empty: the Deploys view shows the controller version only.
	DeployState string
	// ChatPool is the type: chat pool Desk chats with (escalation keys and
	// new desk:<id> sessions). Empty: chat is read-only.
	ChatPool string
}

// Server is the Desk HTTP handler.
type Server struct {
	cfg Config
	ctl Controller
	log *slog.Logger
	ui  fs.FS
}

// NeedsHuman is the escalation label.
const NeedsHuman = "needs-human"

// CSRFHeader must be set on every non-GET request. Browsers do not send a
// custom header cross-origin without a CORS preflight, which Desk refuses.
const CSRFHeader = "X-Jobshop-Desk"

// New returns a server. log may be nil.
func New(cfg Config, ctl Controller, log *slog.Logger) *Server {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	ui, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return &Server{cfg: cfg, ctl: ctl, log: log, ui: ui}
}

// Loopback reports whether addr (host:port) binds only loopback.
func Loopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Handler returns the routes behind the Host, CSRF and header checks.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/attention", s.json(s.attention))
	mux.HandleFunc("GET /api/work", s.json(s.work))
	mux.HandleFunc("GET /api/sessions", s.json(s.sessions))
	mux.HandleFunc("GET /api/pools", s.json(s.pools))
	mux.HandleFunc("GET /api/deploys", s.json(s.deploys))
	mux.HandleFunc("GET /api/activity", s.json(s.activity))
	mux.HandleFunc("GET /api/ticket/{id}", s.json(s.ticket))
	mux.HandleFunc("GET /api/events", s.events)
	mux.HandleFunc("GET /api/session/{id}/stream", s.sessionStream)
	mux.HandleFunc("GET /api/session/{id}/access", s.json(s.chatAccess))
	mux.HandleFunc("GET /api/chat", s.json(s.chatList))
	mux.HandleFunc("POST /api/chat/send", s.json(s.chatSend))
	mux.HandleFunc("POST /api/chat/new", s.json(s.chatNew))
	mux.HandleFunc("POST /api/pool/scale", s.json(s.poolScale))
	mux.HandleFunc("POST /api/session/cordon", s.json(s.sessionCordon))
	mux.Handle("GET /", s.static())
	return s.guard(mux)
}

func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		if !s.hostAllowed(r.Host) {
			http.Error(w, "host not allowed", http.StatusMisdirectedRequest)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && (r.Header.Get(CSRFHeader) == "" || !s.sameOrigin(r)) {
			http.Error(w, "missing "+CSRFHeader+" header", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// hostAllowed blocks DNS rebinding: only loopback names and configured hosts.
func (s *Server) hostAllowed(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" || slices.Contains(s.cfg.Hosts, host) {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (s *Server) static() http.Handler {
	files := http.FileServerFS(s.ui)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Views live in the query string, so every unknown path is the app.
		if _, err := fs.Stat(s.ui, strings.TrimPrefix(r.URL.Path, "/")); err != nil || r.URL.Path == "/" {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
			w.Header().Set("Cache-Control", "no-cache")
		} else if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}

func (s *Server) json(fn func(*http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		v, err := fn(r.WithContext(ctx))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if err != nil {
			code := http.StatusBadGateway
			switch {
			case errors.Is(err, store.ErrNotFound):
				code = http.StatusNotFound
			case errors.Is(err, errBadRequest), errors.Is(err, store.ErrInvalid), errors.Is(err, store.ErrClosed):
				code = http.StatusBadRequest
			case errors.Is(err, store.ErrConflict):
				code = http.StatusConflict
			}
			s.log.Warn("desk: api", "path", r.URL.Path, "err", err)
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(v)
	}
}

// activity proxies activity.list (#255). Query: since (Go duration),
// source, repo, kind, actor, bots, ticket, limit.
func (s *Server) activity(r *http.Request) (any, error) {
	q := r.URL.Query()
	p := rpc.ActivityListParams{
		Source: q.Get("source"), Repo: q.Get("repo"), Kind: q.Get("kind"),
		Actor: q.Get("actor"), Bots: q.Get("bots"), Ticket: q.Get("ticket"),
	}
	if v := q.Get("since"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("%w: since: %w", errBadRequest, err)
		}
		p.Since = d
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("%w: limit: %w", errBadRequest, err)
		}
		p.Limit = n
	}
	var res rpc.ActivityListResult
	if err := s.ctl.Call(r.Context(), rpc.MethodActivityList, p, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// AttentionItem is one thing waiting on a human.
type AttentionItem struct {
	Kind  string    `json:"kind"` // run, ticket, pool, deploy
	ID    string    `json:"id,omitempty"`
	Title string    `json:"title"`
	Why   string    `json:"why"`
	Since time.Time `json:"since,omitzero"`
	// Chat is the key of the escalation chat the controller opened for
	// the ticket (meta.escalation.key), when there is one (#379).
	Chat string `json:"chat,omitempty"`
}

// escalationKey is meta.escalation.key of t, or "".
func escalationKey(t store.Ticket) string {
	var m struct {
		Escalation struct {
			Key string `json:"key"`
		} `json:"escalation"`
	}
	if json.Unmarshal(t.Meta, &m) != nil {
		return ""
	}
	return m.Escalation.Key
}

// attention lists what waits on a human, oldest first; empty when healthy.
func (s *Server) attention(r *http.Request) (any, error) {
	ctx := r.Context()
	out := []AttentionItem{}
	var runs []store.Ticket
	if err := s.ctl.Call(ctx, rpc.MethodRunList, rpc.RunListParams{State: store.RunNeedsHuman}, &runs); err != nil {
		return nil, err
	}
	for _, t := range runs {
		out = append(out, AttentionItem{Kind: "run", ID: t.ID, Title: t.Title, Why: "run needs a human", Since: t.UpdatedAt})
	}
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		var ts []store.Ticket
		if err := s.ctl.Call(ctx, rpc.MethodTicketList, rpc.ListParams{Label: NeedsHuman, Status: st}, &ts); err != nil {
			return nil, err
		}
		for _, t := range ts {
			out = append(out, AttentionItem{Kind: "ticket", ID: t.ID, Title: t.Title, Why: "escalation", Since: t.CreatedAt, Chat: escalationKey(t)})
		}
	}
	var pools rpc.PoolStatusResult
	if err := s.ctl.Call(ctx, rpc.MethodPoolList, struct{}{}, &pools); err != nil {
		return nil, err
	}
	for name, why := range pools.Quarantined {
		out = append(out, AttentionItem{Kind: "pool", ID: name, Title: "pool " + name + " quarantined", Why: why})
	}
	if d := s.readDeploy(); d.Failed != "" && d.Failed != d.Deployed {
		out = append(out, AttentionItem{Kind: "deploy", ID: d.Failed, Title: "deploy of " + short(d.Failed) + " rolled back", Why: "self-deploy failed; waiting for a newer green commit", Since: d.FailedAt})
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].Since, out[j].Since
		if a.IsZero() != b.IsZero() {
			return b.IsZero()
		}
		return a.Before(b)
	})
	return out, nil
}

// Work is runs and batches.
type Work struct {
	Runs    []store.Ticket `json:"runs"`
	Batches []store.Ticket `json:"batches"`
}

func (s *Server) work(r *http.Request) (any, error) {
	ctx := r.Context()
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	w := Work{Runs: []store.Ticket{}, Batches: []store.Ticket{}}
	p := rpc.RunListParams{State: r.URL.Query().Get("state"), Limit: limit}
	if err := s.ctl.Call(ctx, rpc.MethodRunList, p, &w.Runs); err != nil {
		return nil, err
	}
	if err := s.ctl.Call(ctx, rpc.MethodBatchList, rpc.BatchListParams{Limit: limit}, &w.Batches); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Server) sessions(r *http.Request) (any, error) {
	var out []rpc.SessionInfo
	p := rpc.SessionListParams{}
	if d, err := time.ParseDuration(r.URL.Query().Get("since")); err == nil && d > 0 {
		p.Since = d
	}
	if err := s.ctl.Call(r.Context(), rpc.MethodSessionList, p, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []rpc.SessionInfo{}
	}
	return out, nil
}

// Pools is pool status and triggers.
type Pools struct {
	rpc.PoolStatusResult
	Triggers []rpc.TriggerInfo `json:"triggers"`
}

func (s *Server) pools(r *http.Request) (any, error) {
	var p Pools
	if err := s.ctl.Call(r.Context(), rpc.MethodPoolList, struct{}{}, &p.PoolStatusResult); err != nil {
		return nil, err
	}
	if err := s.ctl.Call(r.Context(), rpc.MethodTriggerList, struct{}{}, &p.Triggers); err != nil {
		return nil, err
	}
	sort.Slice(p.Triggers, func(i, j int) bool { return p.Triggers[i].Name < p.Triggers[j].Name })
	return p, nil
}

// Deploys is the running controller and the self-deploy state.
type Deploys struct {
	Server     rpc.VersionResult `json:"server"`
	Deployed   string            `json:"deployed,omitempty"`
	DeployedAt time.Time         `json:"deployed_at,omitzero"`
	Failed     string            `json:"failed,omitempty"`
	FailedAt   time.Time         `json:"failed_at,omitzero"`
}

func (s *Server) deploys(r *http.Request) (any, error) {
	d := s.readDeploy()
	if err := s.ctl.Call(r.Context(), rpc.MethodVersion, struct{}{}, &d.Server); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Server) readDeploy() Deploys {
	var d Deploys
	if s.cfg.DeployState == "" {
		return d
	}
	read := func(name string) (string, time.Time) {
		p := filepath.Join(s.cfg.DeployState, name)
		b, err := os.ReadFile(p)
		if err != nil {
			return "", time.Time{}
		}
		var mt time.Time
		if fi, err := os.Stat(p); err == nil {
			mt = fi.ModTime()
		}
		return strings.TrimSpace(string(b)), mt
	}
	d.Deployed, d.DeployedAt = read("deployed")
	d.Failed, d.FailedAt = read("failed")
	return d
}

// TicketView is the drawer: the ticket, its run when it is one, and its trace.
type TicketView struct {
	rpc.ShowResult
	Run   *rpc.RunResult `json:"run,omitempty"`
	Trace *rpc.TraceNode `json:"trace,omitempty"`
}

func (s *Server) ticket(r *http.Request) (any, error) {
	ctx := r.Context()
	id := r.PathValue("id")
	var v TicketView
	if err := s.ctl.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: id}, &v.ShowResult); err != nil {
		return nil, err
	}
	if v.Ticket.Kind == store.KindRun {
		var run rpc.RunResult
		if err := s.ctl.Call(ctx, rpc.MethodRunShow, rpc.IDParams{ID: id}, &run); err == nil {
			v.Run = &run
		}
	}
	var tr rpc.TraceNode
	if err := s.ctl.Call(ctx, rpc.MethodTicketTrace, rpc.TraceParams{ID: id, Depth: 3}, &tr); err == nil {
		v.Trace = &tr
	}
	return v, nil
}

// sse writes server-sent events; the browser's EventSource reconnects.
type sse struct {
	w http.ResponseWriter
	f http.Flusher
}

func newSSE(w http.ResponseWriter) (*sse, bool) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return nil, false
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	f.Flush()
	return &sse{w: w, f: f}, true
}

func (e *sse) send(event string, id int64, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	if id > 0 {
		_, _ = fmt.Fprintf(e.w, "id: %d\n", id)
	}
	_, _ = fmt.Fprintf(e.w, "event: %s\ndata: %s\n\n", event, b)
	e.f.Flush()
}

// lastID is the EventSource resume point.
func lastID(r *http.Request) int64 {
	v := r.Header.Get("Last-Event-ID")
	if v == "" {
		v = r.URL.Query().Get("since")
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return max(n, 0)
}

// events proxies events.subscribe. Live only by default: the UI refetches
// views on change, so a replay of old history is not needed.
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	e, ok := newSSE(w)
	if !ok {
		return
	}
	res, err := s.ctl.Events(r.Context(), rpc.EventsSubscribeParams{SinceSeq: lastID(r)}, func(ev rpc.Event) {
		e.send(ev.Kind, ev.Seq, ev)
	})
	if err != nil && r.Context().Err() == nil {
		e.send("error", 0, map[string]string{"error": err.Error()})
	} else if res.Lagged {
		e.send(rpc.EventKindLagged, 0, res)
	}
}

// sessionStream proxies session.subscribe from tail seqs before the end,
// or from Last-Event-ID on reconnect.
func (s *Server) sessionStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	since := lastID(r)
	if since == 0 {
		tail, _ := strconv.ParseInt(r.URL.Query().Get("tail"), 10, 64)
		if tail <= 0 {
			tail = 200
		}
		var list []rpc.SessionInfo
		if err := s.ctl.Call(r.Context(), rpc.MethodSessionList, rpc.SessionListParams{All: true}, &list); err == nil {
			for _, si := range list {
				if si.ID == id {
					since = max(si.Seq-tail, 0)
				}
			}
		}
	}
	e, ok := newSSE(w)
	if !ok {
		return
	}
	res, err := s.ctl.Subscribe(r.Context(), rpc.SessionSubscribeParams{ID: id, SinceSeq: since}, func(ev rpc.SessionEvent) {
		e.send("session", ev.Seq, ev)
	})
	switch {
	case err != nil && r.Context().Err() == nil:
		e.send("error", 0, map[string]string{"error": err.Error()})
	case err == nil:
		e.send("end", 0, res)
	}
}

func short(rev string) string {
	if len(rev) > 7 {
		return rev[:7]
	}
	return rev
}
