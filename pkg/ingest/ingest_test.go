package ingest

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ghSecret  = "gh-secret"
	linSecret = "lin-secret"
)

func mac(secret string, b []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(b)
	return hex.EncodeToString(m.Sum(nil))
}

func testHandler(t *testing.T, now time.Time) (*Handler, Spool, *[]string) {
	t.Helper()
	sp := Spool{Dir: filepath.Join(t.TempDir(), "spool")}
	var seen []string
	h := &Handler{
		Spool:  sp,
		GitHub: func() []Secret { return []Secret{{Key: []byte(ghSecret)}} },
		Linear: func() []Secret {
			return []Secret{{Name: "other", Key: []byte("x")}, {Name: "lin", Key: []byte(linSecret)}}
		},
		Now: func() time.Time { return now },
		OnLinear: func(signer string, ev LinearEvent) {
			seen = append(seen, signer+":"+ev.IssueKey())
		},
	}
	return h, sp, &seen
}

func linearBody(t *testing.T, at time.Time, typ, actorType string) []byte {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"action": "create", "type": typ, "organizationId": "org-1", "webhookTimestamp": at.UnixMilli(),
		"actor": map[string]any{"id": "u-1", "type": actorType},
		"data":  map[string]any{"identifier": "ENG-5"},
	})
	require.NoError(t, err)
	return b
}

func post(h *Handler, path string, body []byte, hdr map[string]string) int {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.Mux().ServeHTTP(rec, req)
	return rec.Code
}

func TestGitHubSignature(t *testing.T) {
	h, sp, _ := testHandler(t, time.Now())
	body := []byte(`{"action":"opened"}`)
	hdr := func(sig string) map[string]string {
		return map[string]string{"X-GitHub-Event": "issues", "X-GitHub-Delivery": "d-1", "X-Hub-Signature-256": sig}
	}
	assert.Equal(t, http.StatusUnauthorized, post(h, GitHubPath, body, hdr("sha256="+mac("wrong", body))))
	assert.Equal(t, http.StatusUnauthorized, post(h, GitHubPath, body, hdr("")))
	names, err := sp.List()
	require.NoError(t, err)
	assert.Empty(t, names)
	assert.Equal(t, http.StatusAccepted, post(h, GitHubPath, body, hdr("sha256="+mac(ghSecret, body))))
	names, err = sp.List()
	require.NoError(t, err)
	assert.Equal(t, []string{"github-d-1.json"}, names)

	ping := []byte(`{"zen":"x"}`)
	assert.Equal(t, http.StatusNoContent, post(h, GitHubPath, ping, map[string]string{
		"X-GitHub-Event": "ping", "X-GitHub-Delivery": "d-2", "X-Hub-Signature-256": "sha256=" + mac(ghSecret, ping),
	}), "ping filtered")
	names, err = sp.List()
	require.NoError(t, err)
	assert.Len(t, names, 1)
}

func TestLinearSignatureAndFilters(t *testing.T) {
	now := time.Date(2026, 11, 4, 9, 0, 0, 0, time.UTC)
	h, sp, seen := testHandler(t, now)
	good := linearBody(t, now, "Issue", "user")
	sig := func(b []byte) map[string]string {
		return map[string]string{"Linear-Signature": mac(linSecret, b), "Linear-Delivery": "ld-" + mac("id", b)[:12]}
	}
	assert.Equal(t, http.StatusUnauthorized, post(h, LinearPath, good, map[string]string{"Linear-Signature": mac("wrong", good)}), "bad signature")
	stale := linearBody(t, now.Add(-2*time.Minute), "Issue", "user")
	assert.Equal(t, http.StatusUnauthorized, post(h, LinearPath, stale, sig(stale)), "stale")
	future := linearBody(t, now.Add(2*time.Minute), "Issue", "user")
	assert.Equal(t, http.StatusUnauthorized, post(h, LinearPath, future, sig(future)), "future")
	bot := linearBody(t, now, "Issue", "app")
	assert.Equal(t, http.StatusNoContent, post(h, LinearPath, bot, sig(bot)), "bot filtered")
	other := linearBody(t, now, "Project", "user")
	assert.Equal(t, http.StatusNoContent, post(h, LinearPath, other, sig(other)), "type filtered")
	names, err := sp.List()
	require.NoError(t, err)
	assert.Empty(t, names)
	assert.Empty(t, *seen)

	assert.Equal(t, http.StatusAccepted, post(h, LinearPath, good, sig(good)))
	assert.Equal(t, []string{"lin:ENG-5"}, *seen, "signer named")
}

func TestDuplicateWritesOnce(t *testing.T) {
	now := time.Now()
	h, sp, seen := testHandler(t, now)
	b := linearBody(t, now, "Comment", "user")
	hdr := map[string]string{"Linear-Signature": mac(linSecret, b), "Linear-Delivery": "same"}
	require.Equal(t, http.StatusAccepted, post(h, LinearPath, b, hdr))
	info1, err := os.Stat(filepath.Join(sp.Dir, "linear-same.json"))
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, post(h, LinearPath, b, hdr), "redelivery is accepted")
	info2, err := os.Stat(filepath.Join(sp.Dir, "linear-same.json"))
	require.NoError(t, err)
	assert.True(t, os.SameFile(info1, info2), "file not rewritten")
	assert.Len(t, *seen, 1, "consumer runs once")
	names, err := sp.List()
	require.NoError(t, err)
	assert.Len(t, names, 1)

	ok, err := sp.Write(Delivery{Source: SourceLinear, DeliveryID: "same", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestPartialWriteInvisible: a crash between the temp write and the
// link leaves a dot file that List skips; the next Write succeeds.
func TestPartialWriteInvisible(t *testing.T) {
	sp := Spool{Dir: t.TempDir()}
	require.NoError(t, os.WriteFile(filepath.Join(sp.Dir, ".tmp-123"), []byte(`{"source":"git`), 0o600))
	names, err := sp.List()
	require.NoError(t, err)
	assert.Empty(t, names)
	ok, err := sp.Write(Delivery{Source: SourceGitHub, DeliveryID: "a", Kind: "push", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)
	assert.True(t, ok)
	names, err = sp.List()
	require.NoError(t, err)
	assert.Equal(t, []string{"github-a.json"}, names)
	ents, err := os.ReadDir(sp.Dir)
	require.NoError(t, err)
	assert.Len(t, ents, 2, "no temp file left by a finished write")
}

func TestSpoolRejectsBadIDs(t *testing.T) {
	sp := Spool{Dir: t.TempDir()}
	for _, id := range []string{"", "../x", ".hidden", "a/b"} {
		_, err := sp.Write(Delivery{Source: SourceGitHub, DeliveryID: id})
		require.ErrorIs(t, err, ErrBadDelivery, id)
	}
	_, err := sp.Write(Delivery{Source: "other", DeliveryID: "a"})
	require.ErrorIs(t, err, ErrBadDelivery)
	_, err = sp.Read("../etc")
	require.ErrorIs(t, err, ErrBadDelivery)
}

func TestSpoolWriteError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "notadir")
	require.NoError(t, os.WriteFile(file, nil, 0o600))
	h, _, _ := testHandler(t, time.Now())
	h.Spool = Spool{Dir: file}
	body := []byte(`{}`)
	assert.Equal(t, http.StatusInternalServerError, post(h, GitHubPath, body, map[string]string{
		"X-GitHub-Event": "push", "X-GitHub-Delivery": "d", "X-Hub-Signature-256": "sha256=" + mac(ghSecret, body),
	}))
}

// TestEndToEnd: a real server, POST, then the spooled file decodes.
func TestEndToEnd(t *testing.T) {
	h, sp, _ := testHandler(t, time.Now())
	srv := httptest.NewServer(h.Mux())
	defer srv.Close()
	body := []byte(`{"action":"opened","number":7}`)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL+GitHubPath, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-GitHub-Delivery", "abc-123")
	req.Header.Set("X-Hub-Signature-256", "sha256="+mac(ghSecret, body))
	req.Header.Set("Authorization", "not kept")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	names, err := sp.List()
	require.NoError(t, err)
	require.Equal(t, []string{"github-abc-123.json"}, names)
	d, err := sp.Read(names[0])
	require.NoError(t, err)
	assert.Equal(t, SourceGitHub, d.Source)
	assert.Equal(t, "abc-123", d.DeliveryID)
	assert.Equal(t, "pull_request", d.Kind)
	assert.False(t, d.ReceivedAt.IsZero())
	assert.Equal(t, "pull_request", d.Headers["X-GitHub-Event"])
	assert.NotContains(t, d.Headers, "Authorization")
	assert.JSONEq(t, string(body), string(d.Payload))
	require.NoError(t, sp.Remove(names[0]))
	names, err = sp.List()
	require.NoError(t, err)
	assert.Empty(t, names)
	require.NoError(t, sp.Remove("github-abc-123.json"), "missing is fine")
}

func TestSourceOff(t *testing.T) {
	h := &Handler{Spool: Spool{Dir: t.TempDir()}, Linear: func() []Secret { return nil }}
	assert.Equal(t, http.StatusNotFound, post(h, GitHubPath, []byte(`{}`), nil))
}
