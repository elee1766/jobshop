// Package desk serves Jobshop Desk: a read-only dashboard over the admin
// API (#202). It is an admin-socket client like the gateway; it never
// opens the database. The UI is built into dist/ and embedded.
package desk

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/gzhttp"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// dist holds the built UI. Only dist/.keep is tracked, so the package
// builds and its tests run before the UI is built.
//
//go:embed all:dist
var dist embed.FS

// unbuilt is the page served when the binary was built without the UI.
const unbuilt = "<!doctype html>\n<html>\n  <head>\n    <title>jobshop desk</title>\n  </head>\n  <body>\n    <p>Desk UI is not built. Run <code>npm --prefix ui ci &amp;&amp; npm --prefix ui run build</code>, then rebuild jobshop.</p>\n  </body>\n</html>\n"

// Built reports whether the binary embeds the built UI.
func Built() bool {
	_, err := fs.Stat(dist, "dist/index.html")
	return err == nil
}

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
	// ChatPool is the type: chat pool Desk chats with (escalation keys and
	// new desk:<id> sessions). Empty: chat is read-only.
	ChatPool string
	// Push sends Web Push notifications for new attention items. Nil: off.
	Push *Push
	// UserHeader names the header an authenticating proxy sets to the
	// signed-in user (#515). Empty: no user; calls act as "admin".
	UserHeader string
	// BasePath is the URL prefix Desk is served under, "/desk/" (#516).
	// Empty or "/": the root, unless a trusted proxy sends
	// X-Forwarded-Prefix. Requests with the prefix have it stripped, so a
	// proxy may keep or strip it.
	BasePath string
	// TrustedProxies are the TCP peers whose UserHeader and
	// X-Forwarded-Prefix are believed. Required when UserHeader is set.
	TrustedProxies []netip.Prefix
	// Root holds Desk's notification, acknowledgement and focus state.
	Root string
	// Favicon replaces the built-in icon.svg as the page icon. Nil: built-in.
	Favicon []byte
}

type userKey struct{}

// userFrom is the proxy user of a request, empty when not configured.
func userFrom(ctx context.Context) string {
	u, _ := ctx.Value(userKey{}).(string)
	return u
}

// proxyUser returns the user for r. ok is false when the request must be
// refused: an untrusted peer, or a missing or invalid header.
func (s *Server) proxyUser(r *http.Request) (user string, ok bool) {
	if !s.trustedPeer(r) {
		return "", false
	}
	vals := r.Header.Values(s.cfg.UserHeader)
	if len(vals) != 1 {
		return "", false
	}
	u := strings.TrimSpace(vals[0])
	if rpc.ValidActor("desk:"+u) != nil || u == "" {
		return "", false
	}
	return u, true
}

// ForwardedPrefixHeader carries the prefix from a proxy (#516).
const ForwardedPrefixHeader = "X-Forwarded-Prefix"

// Server is the Desk HTTP handler.
type Server struct {
	cfg Config
	ctl Controller
	log *slog.Logger
	ui  fs.FS
	// index is dist/index.html; serveIndex adds the <base> tag.
	index []byte
	// attn is <root>/desk/attention.json (#535).
	attn attentionFile
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
	index, err := fs.ReadFile(ui, "index.html")
	if err != nil {
		index = []byte(unbuilt)
	}
	// The CLI rejects an invalid base; here it falls back to the root.
	if cfg.BasePath = CleanBasePath(cfg.BasePath); cfg.BasePath == "" {
		cfg.BasePath = "/"
	}
	s := &Server{cfg: cfg, ctl: ctl, log: log, ui: ui, index: index}
	if cfg.Root != "" {
		s.attn.path = filepath.Join(cfg.Root, "desk", "attention.json")
	}
	return s
}

// basePathRe is what a base path may contain. A leading "//" would make
// <base href> another origin.
var basePathRe = regexp.MustCompile(`^/([A-Za-z0-9._~-]+/)*$`)

// CleanBasePath returns p with one leading and one trailing slash, or ""
// when p is not a valid base path. "" and "/" give "/".
func CleanBasePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	if !basePathRe.MatchString(p) {
		return ""
	}
	return p
}

// ParseTrustedProxies parses addresses and CIDRs; a bare address is a
// single-host range.
func ParseTrustedProxies(vals []string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for _, v := range vals {
		if p, err := netip.ParsePrefix(v); err == nil {
			out = append(out, p.Masked())
			continue
		}
		a, err := netip.ParseAddr(v)
		if err != nil {
			return nil, fmt.Errorf("desk: trusted proxy %q: want an address or CIDR", v)
		}
		out = append(out, netip.PrefixFrom(a.Unmap(), a.Unmap().BitLen()))
	}
	return out, nil
}

// trustedPeer reports whether the request's TCP peer is a trusted proxy.
func (s *Server) trustedPeer(r *http.Request) bool {
	ap, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	addr := ap.Addr().Unmap()
	for _, p := range s.cfg.TrustedProxies {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// basePath is the prefix this request is served under: the flag, else a
// valid X-Forwarded-Prefix from a trusted proxy, else "/".
func (s *Server) basePath(r *http.Request) string {
	if s.cfg.BasePath != "/" {
		return s.cfg.BasePath
	}
	if v := r.Header.Get(ForwardedPrefixHeader); v != "" && s.trustedPeer(r) {
		if p := CleanBasePath(v); p != "" {
			return p
		}
	}
	return "/"
}

// stripBase removes the base from the request path when present (prefix
// kept by the proxy) and records the base for the index page.
func (s *Server) stripBase(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := s.basePath(r)
		if base != "/" {
			if r.URL.Path == strings.TrimSuffix(base, "/") {
				u := *r.URL
				u.Path = base
				http.Redirect(w, r, u.RequestURI(), http.StatusMovedPermanently)
				return
			}
			if rest, ok := strings.CutPrefix(r.URL.Path, base); ok {
				r = r.Clone(r.Context())
				r.URL.Path = "/" + rest
				r.URL.RawPath = ""
			}
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), baseKey{}, base)))
	})
}

type baseKey struct{}

// serveIndex writes index.html with a <base> tag for the request's base,
// so the relative asset URLs and the UI's API URLs resolve under it.
func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	base, _ := r.Context().Value(baseKey{}).(string)
	if base == "" {
		base = "/"
	}
	tag := `<base href="` + html.EscapeString(base) + `">`
	page := strings.Replace(string(s.index), "<head>", "<head>\n    "+tag, 1)
	if s.cfg.Favicon != nil {
		page = strings.Replace(page, `<link rel="icon" href="icon.svg" type="image/svg+xml" />`,
			`<link rel="icon" href="favicon" type="`+html.EscapeString(faviconType(s.cfg.Favicon))+`" />`, 1)
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Cache-Control", "no-cache")
	if r.Method == http.MethodHead {
		return
	}
	page = s.indexEmbeds(w, r, page)
	_, _ = io.WriteString(w, page)
}

// Start opens push state and runs the notifier until ctx ends. No-op
// without Push.
func (s *Server) Start(ctx context.Context) error {
	if s.cfg.Push == nil {
		return nil
	}
	if s.cfg.Push.Interval <= 0 {
		s.cfg.Push.Interval = time.Minute
	}
	if err := s.cfg.Push.open(); err != nil {
		return fmt.Errorf("desk: push: %w", err)
	}
	go s.runPush(ctx)
	go s.watchEvents(ctx)
	return nil
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
	mux.HandleFunc("GET /api/whoami", s.json(s.whoami))
	mux.HandleFunc("GET /api/embeds", s.json(s.embedList))
	mux.HandleFunc("GET /api/attention", s.json(s.attention))
	mux.HandleFunc("GET /api/work", s.json(s.work))
	mux.HandleFunc("GET /api/sessions", s.json(s.sessions))
	mux.HandleFunc("GET /api/pools", s.json(s.pools))
	mux.HandleFunc("GET /api/activity", s.json(s.activity))
	mux.HandleFunc("GET /api/ticket/{id}", s.json(s.ticket))
	mux.HandleFunc("GET /api/events", s.events)
	mux.HandleFunc("GET /api/session/{id}/stream", s.sessionStream)
	mux.HandleFunc("GET /api/session/{id}/journal", s.json(s.sessionJournal))
	mux.HandleFunc("GET /api/journal/search", s.json(s.journalSearch))
	mux.HandleFunc("GET /api/session/{id}/access", s.json(s.chatAccess))
	mux.HandleFunc("GET /api/chat", s.json(s.chatList))
	mux.HandleFunc("GET /api/escalations", s.json(s.escalations))
	mux.HandleFunc("POST /api/chat/send", s.json(s.chatSend))
	mux.HandleFunc("POST /api/chat/new", s.json(s.chatNew))
	mux.HandleFunc("POST /api/chat/cancel", s.json(s.chatCancel))
	mux.HandleFunc("POST /api/attention/chat", s.json(s.attentionChat))
	mux.HandleFunc("POST /api/attention/ack", s.json(s.attentionAck))
	mux.HandleFunc("GET /api/session/{id}/events", s.json(s.sessionEvents))
	mux.HandleFunc("POST /api/pool/scale", s.json(s.poolScale))
	mux.HandleFunc("POST /api/session/cordon", s.json(s.sessionCordon))
	mux.HandleFunc("GET /api/push/key", s.json(s.pushKey))
	if s.cfg.Push != nil {
		mux.HandleFunc("POST /api/push/subscribe", s.json(s.pushSubscribe))
		mux.HandleFunc("POST /api/push/unsubscribe", s.json(s.pushUnsubscribe))
		mux.HandleFunc("POST /api/push/test", s.json(s.pushTest))
	}
	mux.HandleFunc("GET /api/artifacts", s.json(s.artifactList))
	mux.HandleFunc("GET /api/artifacts/{id}", s.json(s.artifactShow))
	mux.HandleFunc("GET /api/artifacts/{id}/files/{name}", s.artifactFile)
	mux.HandleFunc("POST /api/artifacts/{id}/ack", s.json(s.artifactAck))
	mux.HandleFunc("GET /api/overview", s.json(s.overview))
	mux.HandleFunc("GET /api/live", s.live)
	mux.HandleFunc("GET /api/trace/{ref}", s.json(s.traceView))
	mux.HandleFunc("GET /api/system", s.json(s.systemView))
	mux.HandleFunc("GET /api/improvements", s.json(s.improvements))
	mux.HandleFunc("POST /api/attention/answer", s.json(s.attentionAnswer))
	mux.HandleFunc("POST /api/chat/manage", s.json(s.chatManage))
	mux.HandleFunc("GET /api/focus", s.json(s.focusGet))
	mux.HandleFunc("POST /api/focus", s.json(s.focusSet))
	mux.HandleFunc("GET /api/", http.NotFound)
	mux.HandleFunc("GET /plugins/", http.NotFound)
	mux.HandleFunc("GET /api/rpc", s.json(s.rpcList))
	mux.HandleFunc("POST /api/rpc/{method}", s.json(s.rpcCall))
	// Only the static files are compressed: SSE and the JSON API stay
	// untouched so streams are never buffered (#1088).
	mux.Handle("GET /", gzhttp.GzipHandler(s.static()))
	return s.guard(s.stripBase(mux))
}

func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", baseCSP)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		if !s.hostAllowed(r.Host) {
			http.Error(w, "host not allowed", http.StatusMisdirectedRequest)
			return
		}
		if s.cfg.UserHeader != "" {
			// A request that bypasses the proxy must not claim a user.
			u, ok := s.proxyUser(r)
			if !ok {
				http.Error(w, "no trusted user", http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), userKey{}, u)
			r = r.WithContext(rpc.WithActor(ctx, "desk:"+u))
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
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "favicon" && s.cfg.Favicon != nil {
			w.Header().Set("Content-Type", faviconType(s.cfg.Favicon))
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(s.cfg.Favicon)
			return
		}
		if _, err := fs.Stat(s.ui, name); err != nil || name == "" || name == "index.html" {
			s.serveIndex(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if r.URL.Path == "/sw.js" || r.URL.Path == "/manifest.webmanifest" {
			// Browsers check the worker for updates; keep it and the manifest fresh.
			w.Header().Set("Cache-Control", "no-cache")
			if r.URL.Path == "/manifest.webmanifest" {
				w.Header().Set("Content-Type", "application/manifest+json")
			}
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
			case errors.Is(err, domain.ErrNotFound):
				code = http.StatusNotFound
			case errors.Is(err, errBadRequest), errors.Is(err, domain.ErrInvalid), errors.Is(err, domain.ErrClosed):
				code = http.StatusBadRequest
			case errors.Is(err, domain.ErrConflict):
				code = http.StatusConflict
			case errors.Is(err, errForbidden):
				code = http.StatusForbidden
			}
			s.log.WarnContext(ctx, "desk: api", "path", r.URL.Path, "err", err)
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "code": rpcErrCode(err)})
			return
		}
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Whoami is GET /api/whoami.
type Whoami struct {
	User string `json:"user"`
}

func (s *Server) whoami(r *http.Request) (any, error) {
	return Whoami{User: userFrom(r.Context())}, nil
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
	Kind  string    `json:"kind"` // rpc.Attention* kinds
	ID    string    `json:"id,omitempty"`
	Title string    `json:"title"`
	Why   string    `json:"why"`
	Since time.Time `json:"since,omitzero"`
	// Ref is the linked issue or PR (#12, PR #3, JOB-5), from the
	// controller. URL links to it.
	Ref string `json:"ref,omitempty"`
	URL string `json:"url,omitempty"`
	// Summary is one or two lines for the row (#607).
	Summary string `json:"summary,omitempty"`
	// Context (markdown) and Options are a question's (JOB-101): the
	// row expands the context and shows each option as a button.
	Context string                  `json:"context,omitempty"`
	Options []domain.QuestionOption `json:"options,omitempty"`
}

// attention is the controller's attention list (#577): Desk adds no
// items of its own, so a tool that reports through a needs-human ticket
// shows here, in the CLI and in digests alike.
func (s *Server) attention(r *http.Request) (any, error) {
	var res rpc.AttentionResult
	if err := s.ctl.Call(r.Context(), rpc.MethodAttention, struct{}{}, &res); err != nil {
		return nil, err
	}
	out := make([]AttentionItem, 0, len(res.Items))
	for _, it := range res.Items {
		var why []string
		if it.Repo != "" || it.Ref != "" {
			why = append(why, strings.TrimSpace(it.Repo+" "+it.Ref))
		}
		if it.Owner != "" {
			why = append(why, "owner "+it.Owner)
		}
		out = append(out, AttentionItem{
			Kind: it.Kind, ID: it.ID, Title: it.Title, Why: strings.Join(why, ", "), Since: it.Since, Ref: it.Ref, URL: it.URL,
			Summary: it.Summary, Context: it.Context, Options: it.Options,
		})
	}
	// Items the operator acknowledged in Desk stay hidden (#535).
	return s.dropAcked(out), nil
}

// Work is runs and batches. With ?detail=1 (Attention's In progress,
// JOB-54) Detail holds each run's steps and last comment, and Sessions
// the holder of every in-progress step, so a card needs no request of
// its own.
type Work struct {
	Runs     []domain.Ticket           `json:"runs"`
	Batches  []domain.Ticket           `json:"batches"`
	Detail   map[string]RunDetail      `json:"detail,omitempty"`
	Sessions map[string]rpc.HolderInfo `json:"sessions,omitempty"`
}

// RunDetail is a run's steps and its last comment.
type RunDetail struct {
	Steps   []domain.Ticket `json:"steps"`
	Comment *domain.Comment `json:"comment,omitempty"`
}

// maxDetailRuns bounds the per-run calls one detail request makes.
const maxDetailRuns = 20

func (s *Server) work(r *http.Request) (any, error) {
	ctx := r.Context()
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	w := Work{Runs: []domain.Ticket{}, Batches: []domain.Ticket{}}
	p := rpc.RunListParams{State: r.URL.Query().Get("state"), Limit: limit}
	if err := s.ctl.Call(ctx, rpc.MethodRunList, p, &w.Runs); err != nil {
		return nil, err
	}
	if err := s.ctl.Call(ctx, rpc.MethodBatchList, rpc.BatchListParams{Limit: limit}, &w.Batches); err != nil {
		return nil, err
	}
	if r.URL.Query().Get("detail") == "" {
		return w, nil
	}
	w.Detail = map[string]RunDetail{}
	var holders []string
	for i, run := range w.Runs {
		if i >= maxDetailRuns {
			break
		}
		var d RunDetail
		var rr rpc.RunResult
		if err := s.ctl.Call(ctx, rpc.MethodRunShow, rpc.IDParams{ID: run.ID}, &rr); err == nil {
			d.Steps = rr.Steps
		}
		var sh rpc.ShowResult
		if err := s.ctl.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: run.ID}, &sh); err == nil && len(sh.Comments) > 0 {
			d.Comment = &sh.Comments[len(sh.Comments)-1]
		}
		for _, st := range d.Steps {
			if st.Status == domain.StatusInProgress && st.Holder != "" {
				holders = append(holders, st.Holder)
			}
		}
		w.Detail[run.ID] = d
	}
	w.Sessions = s.holders(ctx, holders)
	return w, nil
}

// holders looks up the given session ids with one session list call.
// An id that is not a live session gets only its id.
func (s *Server) holders(ctx context.Context, ids []string) map[string]rpc.HolderInfo {
	out := map[string]rpc.HolderInfo{}
	if len(ids) == 0 {
		return out
	}
	for _, id := range ids {
		out[id] = rpc.HolderInfo{ID: id}
	}
	var list []rpc.SessionInfo
	if err := s.ctl.Call(ctx, rpc.MethodSessionList, rpc.SessionListParams{}, &list); err != nil {
		return out
	}
	for _, si := range list {
		if _, ok := out[si.ID]; !ok {
			continue
		}
		h := rpc.HolderInfo{
			ID: si.ID, DisplayName: si.DisplayName, Model: si.Model, Alive: si.Alive, Busy: si.Busy, Tool: si.Tool,
			ToolTarget: si.ToolTarget, ToolIntent: si.ToolIntent, Agent: si.Agent,
		}
		if si.Ticket != nil {
			h.Pool = si.Ticket.Pool
		}
		// A controller-driven session whose host died. An external session
		// never had a host (no pid, no exit file): nothing is known of it.
		h.HostLost = !si.Alive && (si.HostPid != 0 || si.Exit != "")
		out[si.ID] = h
	}
	return out
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

// TicketView is the drawer: the ticket, its run when it is one, and its
// trace. Sessions describes every in-progress holder in it (JOB-54); trace
// nodes carry theirs in Session.
type TicketView struct {
	rpc.ShowResult
	Inspected  *domain.Ticket            `json:"inspected,omitempty"`
	Inspection *rpc.ShowResult           `json:"inspection,omitempty"`
	Graph      *rpc.GraphResult          `json:"graph,omitempty"`
	Timeline   *rpc.TimelineResult       `json:"timeline,omitempty"`
	Run        *rpc.RunResult            `json:"run,omitempty"`
	Trace      *rpc.TraceNode            `json:"trace,omitempty"`
	Sessions   map[string]rpc.HolderInfo `json:"sessions,omitempty"`
}

func (s *Server) ticket(r *http.Request) (any, error) {
	return s.ticketView(r, nil)
}

// ticketView is ticket with extra session ids to describe in the same
// session.list call as the holders it finds (the trace timeline adds the
// holders of its open working spans, JOB-184).
func (s *Server) ticketView(r *http.Request, extra []string) (any, error) {
	ctx := r.Context()
	id := r.PathValue("id")
	var v TicketView
	if err := s.ctl.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: id}, &v.ShowResult); err != nil {
		return nil, err
	}
	if inspect := r.URL.Query().Get("inspect"); inspect != "" && inspect != id {
		var shown rpc.ShowResult
		if err := s.ctl.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: inspect}, &shown); err != nil {
			return nil, err
		}
		v.Inspected = &shown.Ticket
		v.Inspection = &shown
	}
	if v.Ticket.Kind == domain.KindRun {
		var run rpc.RunResult
		if err := s.ctl.Call(ctx, rpc.MethodRunShow, rpc.IDParams{ID: id}, &run); err == nil {
			v.Run = &run
		}
	}
	var tr rpc.TraceNode
	if err := s.ctl.Call(ctx, rpc.MethodTicketTrace, rpc.TraceParams{ID: id, Depth: 3}, &tr); err == nil {
		v.Trace = &tr
	}
	ids := slices.Clone(extra)
	if v.Ticket.Status == domain.StatusInProgress && v.Ticket.Holder != "" {
		ids = append(ids, v.Ticket.Holder)
	}
	if v.Run != nil {
		for _, st := range v.Run.Steps {
			if st.Status == domain.StatusInProgress && st.Holder != "" {
				ids = append(ids, st.Holder)
			}
		}
	}
	if v.Trace != nil {
		ids = traceHolders(*v.Trace, ids)
	}
	if len(ids) > 0 {
		v.Sessions = s.holders(ctx, ids)
		if v.Trace != nil {
			fillTrace(v.Trace, v.Sessions)
		}
	}
	return v, nil
}

func traceHolders(n rpc.TraceNode, ids []string) []string {
	if n.Holder != "" {
		ids = append(ids, n.Holder)
	}
	for _, c := range n.Children {
		ids = traceHolders(c, ids)
	}
	return ids
}

func fillTrace(n *rpc.TraceNode, ss map[string]rpc.HolderInfo) {
	if h, ok := ss[n.Holder]; ok && n.Holder != "" {
		n.Session = &h
	}
	for i := range n.Children {
		fillTrace(&n.Children[i], ss)
	}
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
	h.Set("X-Accel-Buffering", "no")
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

// events proxies events.subscribe. Ordered history resumes from Last-Event-ID
// or since. Dashboard state uses live snapshots instead of invalidations.
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
	// Headers go out first so the browser stops showing "connecting" (#520).
	e, ok := newSSE(w)
	if !ok {
		return
	}
	since := lastID(r)
	tail, _ := strconv.ParseInt(r.URL.Query().Get("tail"), 10, 64)
	if tail <= 0 {
		tail = 200
	}
	// The controller resolves tail itself; listing every session to find
	// this one's seq took seconds (#614).
	res, err := s.ctl.Subscribe(r.Context(), rpc.SessionSubscribeParams{ID: id, SinceSeq: since, Tail: tail}, func(ev rpc.SessionEvent) {
		e.send("session", ev.Seq, ev)
	})
	switch {
	case err != nil && r.Context().Err() == nil:
		e.send("error", 0, map[string]string{"error": err.Error()})
	case err == nil:
		e.send("end", 0, res)
	}
}

// faviconType is the favicon's media type. Sniffing does not know SVG.
func faviconType(b []byte) string {
	if t := http.DetectContentType(b); !strings.HasPrefix(t, "text/") {
		return t
	}
	if bytes.Contains(b[:min(len(b), 512)], []byte("<svg")) {
		return "image/svg+xml"
	}
	return "image/x-icon"
}
