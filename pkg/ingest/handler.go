package ingest

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
)

// MaxBody is the largest delivery body accepted.
const MaxBody = 1 << 20

// Secret is one webhook signing secret. Identity names the Linear
// identity it belongs to, "" for a plain secret_file.
type Secret struct {
	Identity string
	Key      []byte
}

// Handler verifies, filters and spools webhook deliveries, one route per
// endpoint (#339). Status codes: 202 spooled (or already spooled), 204
// filtered, 401 bad signature or stale, 403 repo not on the endpoint's
// allowlist, 400 bad body, 500 spool write error so the sender retries.
type Handler struct {
	Spool     *Spool
	Endpoints []config.IngestEndpoint
	// Secret returns an endpoint's current secret. It runs per request,
	// so a rotated secret file needs no restart.
	Secret func(config.IngestEndpoint) (Secret, error)
	Now    func() time.Time
	Log    *slog.Logger
	// OnSpool, if set, runs after any new delivery is spooled (not for
	// duplicates), whatever its source (#679). The controller's listener
	// wakes its drain with it; `ingest serve` notifies the controller.
	// It must not block.
	OnSpool func(d Delivery)
}

// Mux serves every endpoint at its path. Implicit identity endpoints
// share /linear/webhook: a request there is checked against each of
// their secrets and the one that signed it names the endpoint.
func (h *Handler) Mux() *http.ServeMux {
	byPath := map[string][]config.IngestEndpoint{}
	var order []string
	for _, e := range h.Endpoints {
		p := e.EndpointPath()
		if _, ok := byPath[p]; !ok {
			order = append(order, p)
		}
		byPath[p] = append(byPath[p], e)
	}
	mux := http.NewServeMux()
	for _, p := range order {
		eps := byPath[p]
		mux.HandleFunc("POST "+p, func(w http.ResponseWriter, r *http.Request) { h.serve(w, r, eps) })
	}
	return mux
}

func (h *Handler) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

func (h *Handler) log() *slog.Logger {
	if h.Log != nil {
		return h.Log
	}
	return slog.New(slog.DiscardHandler)
}

func (h *Handler) serve(w http.ResponseWriter, r *http.Request, eps []config.IngestEndpoint) {
	body, err := io.ReadAll(io.LimitReader(r.Body, MaxBody+1))
	if err != nil || len(body) > MaxBody {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	sigHeader := "Linear-Signature"
	switch eps[0].Source {
	case config.IngestGitHub:
		sigHeader = "X-Hub-Signature-256"
	case config.IngestGeneric:
		sigHeader = eps[0].SignatureHeader
	}
	prefix := "sha256="
	if eps[0].Source == config.IngestGeneric {
		prefix = eps[0].SignaturePrefix
	}
	sig, _ := strings.CutPrefix(r.Header.Get(sigHeader), prefix)
	ep, _, ok := h.verify(eps, sig, body)
	if !ok {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	switch ep.Source {
	case config.IngestGitHub:
		h.serveGitHub(w, r, ep, body)
	case config.IngestGeneric:
		h.serveGeneric(w, r, ep, body)
	default:
		h.serveLinear(w, r, ep, body)
	}
}

// verify returns the endpoint whose secret's HMAC-SHA256 of body is the
// hex sig. Only the endpoints on this path are tried.
func (h *Handler) verify(eps []config.IngestEndpoint, sig string, body []byte) (config.IngestEndpoint, Secret, bool) {
	got, err := hex.DecodeString(strings.TrimSpace(sig))
	if err != nil || len(got) == 0 {
		return config.IngestEndpoint{}, Secret{}, false
	}
	for _, e := range eps {
		s, err := h.Secret(e)
		if err != nil {
			h.log().Warn("ingest: secret", "endpoint", e.Name, "err", err)
			continue
		}
		mac := hmac.New(sha256.New, s.Key)
		mac.Write(body)
		if hmac.Equal(mac.Sum(nil), got) {
			return e, s, true
		}
	}
	return config.IngestEndpoint{}, Secret{}, false
}

var validID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// deliveryID is the sender's id, or a hash of the body when it sent
// none, so a redelivery of the same body still dedupes.
func deliveryID(header string, body []byte) string {
	if validID.MatchString(header) {
		return header
	}
	sum := sha256.Sum256(body)
	return "sha256-" + hex.EncodeToString(sum[:16])
}

func pick(r *http.Request, names ...string) map[string]string {
	out := map[string]string{}
	for _, n := range names {
		if v := r.Header.Get(n); v != "" {
			out[n] = v
		}
	}
	return out
}

// write spools d, answers, and runs OnSpool when d was new.
func (h *Handler) write(w http.ResponseWriter, r *http.Request, d Delivery) {
	fresh, err := h.Spool.Write(r.Context(), d)
	if err != nil {
		h.log().ErrorContext(r.Context(), "ingest: spool write", "endpoint", d.Endpoint, "delivery", d.DeliveryID, "err", err)
		http.Error(w, "spool write failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	if fresh && h.OnSpool != nil {
		h.OnSpool(d)
	}
}

// GitHubWanted reports whether a GitHub event kind is worth spooling.
// ping is the hook-creation check and carries no work.
func GitHubWanted(kind string) bool { return kind != "" && kind != "ping" }

func (h *Handler) serveGitHub(w http.ResponseWriter, r *http.Request, ep config.IngestEndpoint, body []byte) {
	kind := r.Header.Get("X-GitHub-Event")
	if !GitHubWanted(kind) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var p struct {
		Repository *struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}
	// With an allowlist, org-level and App deliveries that name no
	// repository are refused too: the endpoint is for those repos only.
	if len(ep.Repos) > 0 && (p.Repository == nil || !slices.ContainsFunc(ep.Repos, func(s string) bool {
		return strings.EqualFold(s, p.Repository.FullName)
	})) {
		http.Error(w, "repository not allowed on this endpoint", http.StatusForbidden)
		return
	}
	h.write(w, r, Delivery{
		Endpoint: ep.Name, Source: SourceGitHub,
		DeliveryID: deliveryID(r.Header.Get("X-GitHub-Delivery"), body),
		Kind:       kind, ReceivedAt: h.now().UTC(),
		Headers: pick(r, "X-GitHub-Event", "X-GitHub-Delivery", "X-Hub-Signature-256", "X-GitHub-Hook-ID", "X-GitHub-Hook-Installation-Target-ID"),
		Payload: body,
	})
}

func (h *Handler) serveLinear(w http.ResponseWriter, r *http.Request, ep config.IngestEndpoint, body []byte) {
	var ev LinearEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}
	if !ev.Fresh(h.now()) {
		http.Error(w, "stale", http.StatusUnauthorized)
		return
	}
	if !ev.Wanted() {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	h.write(w, r, Delivery{
		Endpoint: ep.Name, Source: SourceLinear,
		DeliveryID: deliveryID(r.Header.Get("Linear-Delivery"), body),
		Kind:       ev.Type, ReceivedAt: h.now().UTC(),
		Headers: pick(r, "Linear-Delivery", "Linear-Event", "Linear-Signature"),
		Payload: body,
	})
}

// errNoSecret is returned by ReadSecret for an empty file.
var errNoSecret = errors.New("empty secret")

// ReadSecret reads a secret file or runs a secret command (config.ReadSecret),
// trimming surrounding space.
func ReadSecret(ctx context.Context, identity string, src config.Secret) (Secret, error) {
	b, err := config.ReadSecret(ctx, src)
	if err != nil {
		return Secret{}, err
	}
	k := strings.TrimSpace(string(b))
	if k == "" {
		// src.String() is the path, or "secret command": a command line
		// may hold a token, so it stays out of logs.
		return Secret{}, fmt.Errorf("%s: %w", src, errNoSecret)
	}
	return Secret{Identity: identity, Key: []byte(k)}, nil
}
