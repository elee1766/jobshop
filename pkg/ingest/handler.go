package ingest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Paths the handler serves.
const (
	GitHubPath = "/github/webhook"
	LinearPath = "/linear/webhook"
)

const (
	// LinearMaxAge is how far webhookTimestamp may be from now.
	LinearMaxAge = 60 * time.Second
	maxBody      = 1 << 20
)

// Secret is one webhook signing secret. Name tells the consumer which
// identity signed (a linear-app identity, or "" for a plain file).
type Secret struct {
	Name string
	Key  []byte
}

// Handler verifies, filters and spools webhook deliveries. Status codes:
// 202 spooled (or already spooled), 204 filtered, 401 bad signature or
// stale, 400 bad body, 500 spool write error so the sender retries.
type Handler struct {
	Spool Spool
	// GitHub and Linear return the current secrets; nil turns the
	// source off (404). They are called per request so a reload or a
	// rotated secret file needs no restart.
	GitHub func() []Secret
	Linear func() []Secret
	Now    func() time.Time
	Log    *slog.Logger
	// OnLinear, if set, runs after a new Linear delivery is spooled
	// (not for duplicates). The controller uses it until the drain lands.
	OnLinear func(signer string, ev LinearEvent)
}

// Mux serves the enabled sources.
func (h *Handler) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	if h.GitHub != nil {
		mux.HandleFunc("POST "+GitHubPath, h.serveGitHub)
	}
	if h.Linear != nil {
		mux.HandleFunc("POST "+LinearPath, h.serveLinear)
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
	return slog.Default()
}

func readBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	if err != nil || len(body) > maxBody {
		http.Error(w, "bad body", http.StatusBadRequest)
		return nil, false
	}
	return body, true
}

// verify returns the secret whose HMAC-SHA256 of body is the hex sig.
func verify(secrets []Secret, sig string, body []byte) (Secret, bool) {
	got, err := hex.DecodeString(strings.TrimSpace(sig))
	if err != nil || len(got) == 0 {
		return Secret{}, false
	}
	for _, s := range secrets {
		if len(s.Key) == 0 {
			continue
		}
		mac := hmac.New(sha256.New, s.Key)
		mac.Write(body)
		if hmac.Equal(mac.Sum(nil), got) {
			return s, true
		}
	}
	return Secret{}, false
}

// deliveryID is the sender's id, or a hash of the body when it sent
// none, so a redelivery of the same body still dedupes.
func deliveryID(header string, body []byte) string {
	if validID.MatchString(header) && !strings.HasPrefix(header, ".") {
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

// write spools d and answers. It reports whether d was new.
func (h *Handler) write(w http.ResponseWriter, d Delivery) bool {
	fresh, err := h.Spool.Write(d)
	if err != nil {
		h.log().Error("ingest: spool write", "source", d.Source, "delivery", d.DeliveryID, "err", err)
		http.Error(w, "spool write failed", http.StatusInternalServerError)
		return false
	}
	w.WriteHeader(http.StatusAccepted)
	return fresh
}

func (h *Handler) serveGitHub(w http.ResponseWriter, r *http.Request) {
	body, ok := readBody(w, r)
	if !ok {
		return
	}
	sig, _ := strings.CutPrefix(r.Header.Get("X-Hub-Signature-256"), "sha256=")
	if _, ok := verify(h.GitHub(), sig, body); !ok {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	kind := r.Header.Get("X-GitHub-Event")
	if !GitHubWanted(kind) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if !json.Valid(body) {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}
	h.write(w, Delivery{
		Source: SourceGitHub, DeliveryID: deliveryID(r.Header.Get("X-GitHub-Delivery"), body),
		Kind: kind, ReceivedAt: h.now().UTC(),
		Headers: pick(r, "X-GitHub-Event", "X-GitHub-Delivery", "X-Hub-Signature-256", "X-GitHub-Hook-ID", "X-GitHub-Hook-Installation-Target-ID"),
		Payload: body,
	})
}

// GitHubWanted reports whether a GitHub event kind is worth spooling.
// ping is the hook-creation check and carries no work.
func GitHubWanted(kind string) bool { return kind != "" && kind != "ping" }

func (h *Handler) serveLinear(w http.ResponseWriter, r *http.Request) {
	body, ok := readBody(w, r)
	if !ok {
		return
	}
	signer, ok := verify(h.Linear(), r.Header.Get("Linear-Signature"), body)
	if !ok {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
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
	fresh := h.write(w, Delivery{
		Source: SourceLinear, DeliveryID: deliveryID(r.Header.Get("Linear-Delivery"), body),
		Kind: ev.Type, ReceivedAt: h.now().UTC(),
		Headers: pick(r, "Linear-Delivery", "Linear-Event", "Linear-Signature"),
		Payload: body,
	})
	if fresh && h.OnLinear != nil {
		h.OnLinear(signer.Name, ev)
	}
}

// LinearEvent is the part of a Linear webhook payload jobshop reads.
type LinearEvent struct {
	Action         string `json:"action"`
	Type           string `json:"type"`
	OrganizationID string `json:"organizationId"`
	// WebhookTimestamp is Unix milliseconds.
	WebhookTimestamp int64 `json:"webhookTimestamp"`
	Actor            *struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"actor"`
	Data struct {
		Identifier string    `json:"identifier"`
		Number     int       `json:"number"`
		Team       *teamRef  `json:"team"`
		UserID     string    `json:"userId"` // Comment events
		BotActor   any       `json:"botActor"`
		Issue      *issueRef `json:"issue"`
	} `json:"data"`
}

type teamRef struct {
	Key string `json:"key"`
}

type issueRef struct {
	Identifier string   `json:"identifier"`
	Number     int      `json:"number"`
	Team       *teamRef `json:"team"`
}

// Fresh reports whether the event's timestamp is within LinearMaxAge
// of now, which bounds replay of a captured delivery.
func (e LinearEvent) Fresh(now time.Time) bool {
	age := now.Sub(time.UnixMilli(e.WebhookTimestamp))
	return age <= LinearMaxAge && age >= -LinearMaxAge
}

// Wanted is the filter that needs only the payload (#300): Issue and
// Comment events by a human. Apps, bots and integrations are skipped:
// their writes need no push. The organization and self checks need a
// Linear token and stay with the consumer.
func (e LinearEvent) Wanted() bool {
	if e.Type != "Issue" && e.Type != "Comment" {
		return false
	}
	return e.Actor != nil && strings.EqualFold(e.Actor.Type, "user") && e.Data.BotActor == nil
}

// IssueKey is the issue the event is about, "" when the payload does
// not name it.
func (e LinearEvent) IssueKey() string {
	key := func(ident string, n int, team *teamRef) string {
		if ident != "" {
			return ident
		}
		if team != nil && team.Key != "" && n > 0 {
			return fmt.Sprintf("%s-%d", team.Key, n)
		}
		return ""
	}
	switch e.Type {
	case "Issue":
		return key(e.Data.Identifier, e.Data.Number, e.Data.Team)
	case "Comment":
		if e.Data.Issue != nil {
			return key(e.Data.Issue.Identifier, e.Data.Issue.Number, e.Data.Issue.Team)
		}
	}
	return ""
}

// errNoSecret is returned by ReadSecret for an empty file.
var errNoSecret = errors.New("empty secret")

// ReadSecret reads a secret file, trimming surrounding space.
func ReadSecret(name, path string) (Secret, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Secret{}, err
	}
	k := strings.TrimSpace(string(b))
	if k == "" {
		return Secret{}, fmt.Errorf("%s: %w", path, errNoSecret)
	}
	return Secret{Name: name, Key: []byte(k)}, nil
}
