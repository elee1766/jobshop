package ingest

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

var secrets = map[string]string{
	"gh-a": "secret-a", "gh-b": "secret-b", "gh-only": "secret-only",
	"lin": "lin-secret", "linear-app1": "app1-secret", "linear-app2": "app2-secret",
}

var testEndpoints = []config.IngestEndpoint{
	{Name: "gh-a", Source: config.IngestGitHub},
	{Name: "gh-b", Source: config.IngestGitHub},
	{Name: "gh-only", Source: config.IngestGitHub, Repos: []string{"acme/app"}},
	{Name: "lin", Source: config.IngestLinear},
	// Two implicit identity endpoints share /linear/webhook.
	{Name: "linear-app1", Source: config.IngestLinear, Identity: "app1", Path: config.LinearWebhookPath, Implicit: true},
	{Name: "linear-app2", Source: config.IngestLinear, Identity: "app2", Path: config.LinearWebhookPath, Implicit: true},
}

func mac(secret string, b []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(b)
	return hex.EncodeToString(m.Sum(nil))
}

func openSpool(t *testing.T, path string) *Spool {
	t.Helper()
	sp, err := OpenSpool(t.Context(), path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sp.Close() })
	return sp
}

type seen struct{ signers []string }

func testHandler(t *testing.T, now time.Time) (*Handler, *Spool, *seen) {
	t.Helper()
	sp := openSpool(t, filepath.Join(t.TempDir(), "ingest", "spool.db"))
	s := &seen{}
	h := &Handler{
		Spool:     sp,
		Endpoints: testEndpoints,
		Secret: func(e config.IngestEndpoint) (Secret, error) {
			k, ok := secrets[e.Name]
			if !ok {
				return Secret{}, errors.New("no secret")
			}
			return Secret{Identity: e.Identity, Key: []byte(k)}, nil
		},
		Now: func() time.Time { return now },
		OnLinear: func(signer string, ev LinearEvent) {
			s.signers = append(s.signers, signer+":"+ev.IssueKey())
		},
	}
	return h, sp, s
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

func ghHeaders(secret, delivery string, body []byte) map[string]string {
	return map[string]string{
		"X-GitHub-Event": "issues", "X-GitHub-Delivery": delivery,
		"X-Hub-Signature-256": "sha256=" + mac(secret, body),
	}
}

func rows(t *testing.T, sp *Spool) []Delivery {
	t.Helper()
	ds, err := sp.List(t.Context(), 0)
	require.NoError(t, err)
	return ds
}

// TestEndpointSecrets: two GitHub endpoints with different secrets each
// accept only their own signature.
func TestEndpointSecrets(t *testing.T) {
	h, sp, _ := testHandler(t, time.Now())
	body := []byte(`{"action":"opened","repository":{"full_name":"acme/x"}}`)
	assert.Equal(t, http.StatusUnauthorized, post(h, "/github/gh-a", body, ghHeaders("secret-b", "d-1", body)), "b's secret on a")
	assert.Equal(t, http.StatusUnauthorized, post(h, "/github/gh-b", body, ghHeaders("secret-a", "d-1", body)), "a's secret on b")
	assert.Equal(t, http.StatusUnauthorized, post(h, "/github/gh-a", body, map[string]string{"X-GitHub-Event": "issues"}), "no signature")
	assert.Empty(t, rows(t, sp))
	assert.Equal(t, http.StatusAccepted, post(h, "/github/gh-a", body, ghHeaders("secret-a", "d-1", body)))
	assert.Equal(t, http.StatusAccepted, post(h, "/github/gh-b", body, ghHeaders("secret-b", "d-2", body)))
	got := rows(t, sp)
	require.Len(t, got, 2)
	assert.Equal(t, "gh-a", got[0].Endpoint)
	assert.Equal(t, "gh-b", got[1].Endpoint)

	ping := []byte(`{"zen":"x"}`)
	hdr := ghHeaders("secret-a", "d-3", ping)
	hdr["X-GitHub-Event"] = "ping"
	assert.Equal(t, http.StatusNoContent, post(h, "/github/gh-a", ping, hdr), "ping filtered")
	assert.Equal(t, http.StatusNotFound, post(h, "/github/webhook", body, ghHeaders("secret-a", "d-4", body)), "no default path")
	assert.Len(t, rows(t, sp), 2)
}

func TestRepoAllowlist(t *testing.T) {
	h, sp, _ := testHandler(t, time.Now())
	other := []byte(`{"repository":{"full_name":"acme/other"}}`)
	assert.Equal(t, http.StatusForbidden, post(h, "/github/gh-only", other, ghHeaders("secret-only", "o-1", other)))
	org := []byte(`{"organization":{"login":"acme"}}`)
	assert.Equal(t, http.StatusForbidden, post(h, "/github/gh-only", org, ghHeaders("secret-only", "o-2", org)), "no repository with an allowlist")
	assert.Empty(t, rows(t, sp))
	listed := []byte(`{"repository":{"full_name":"Acme/App"}}`)
	assert.Equal(t, http.StatusAccepted, post(h, "/github/gh-only", listed, ghHeaders("secret-only", "o-3", listed)))
	// No allowlist: org-level deliveries pass.
	assert.Equal(t, http.StatusAccepted, post(h, "/github/gh-a", org, ghHeaders("secret-a", "o-4", org)))
	assert.Len(t, rows(t, sp), 2)
}

func TestLinearSignatureAndFilters(t *testing.T) {
	now := time.Date(2026, 11, 4, 9, 0, 0, 0, time.UTC)
	h, sp, s := testHandler(t, now)
	good := linearBody(t, now, "Issue", "user")
	sig := func(secret string, b []byte) map[string]string {
		return map[string]string{"Linear-Signature": mac(secret, b), "Linear-Delivery": "ld-" + mac("id", b)[:12]}
	}
	assert.Equal(t, http.StatusUnauthorized, post(h, "/linear/lin", good, sig("wrong", good)), "bad signature")
	assert.Equal(t, http.StatusUnauthorized, post(h, "/linear/lin", good, sig("app1-secret", good)), "another endpoint's secret")
	stale := linearBody(t, now.Add(-2*time.Minute), "Issue", "user")
	assert.Equal(t, http.StatusUnauthorized, post(h, "/linear/lin", stale, sig("lin-secret", stale)), "stale")
	future := linearBody(t, now.Add(2*time.Minute), "Issue", "user")
	assert.Equal(t, http.StatusUnauthorized, post(h, "/linear/lin", future, sig("lin-secret", future)), "future")
	bot := linearBody(t, now, "Issue", "app")
	assert.Equal(t, http.StatusNoContent, post(h, "/linear/lin", bot, sig("lin-secret", bot)), "bot filtered")
	other := linearBody(t, now, "Project", "user")
	assert.Equal(t, http.StatusNoContent, post(h, "/linear/lin", other, sig("lin-secret", other)), "type filtered")
	assert.Empty(t, rows(t, sp))
	assert.Empty(t, s.signers)

	assert.Equal(t, http.StatusAccepted, post(h, "/linear/lin", good, sig("lin-secret", good)))
	assert.Equal(t, []string{":ENG-5"}, s.signers, "plain secret: no identity")
}

// TestImplicitLinearEndpoints: identity endpoints share /linear/webhook
// and the secret that signed names the endpoint and identity (#300).
func TestImplicitLinearEndpoints(t *testing.T) {
	now := time.Now()
	h, sp, s := testHandler(t, now)
	b := linearBody(t, now, "Issue", "user")
	require.Equal(t, http.StatusAccepted, post(h, config.LinearWebhookPath, b, map[string]string{"Linear-Signature": mac("app2-secret", b)}))
	got := rows(t, sp)
	require.Len(t, got, 1)
	assert.Equal(t, "linear-app2", got[0].Endpoint)
	assert.Equal(t, []string{"app2:ENG-5"}, s.signers)
	assert.Equal(t, http.StatusUnauthorized, post(h, config.LinearWebhookPath, b, map[string]string{"Linear-Signature": mac("lin-secret", b)}), "only the path's endpoints")
}

func TestDuplicateWritesOnce(t *testing.T) {
	now := time.Now()
	h, sp, s := testHandler(t, now)
	b := linearBody(t, now, "Comment", "user")
	hdr := map[string]string{"Linear-Signature": mac("lin-secret", b), "Linear-Delivery": "same"}
	require.Equal(t, http.StatusAccepted, post(h, "/linear/lin", b, hdr))
	require.Equal(t, http.StatusAccepted, post(h, "/linear/lin", b, hdr), "redelivery is accepted")
	assert.Len(t, rows(t, sp), 1)
	assert.Len(t, s.signers, 1, "consumer runs once")
	// Without a delivery header, the body hash dedupes.
	nohdr := map[string]string{"Linear-Signature": mac("lin-secret", b)}
	require.Equal(t, http.StatusAccepted, post(h, "/linear/lin", b, nohdr))
	require.Equal(t, http.StatusAccepted, post(h, "/linear/lin", b, nohdr))
	assert.Len(t, rows(t, sp), 2)
}

func TestSpoolWriteError(t *testing.T) {
	h, sp, _ := testHandler(t, time.Now())
	require.NoError(t, sp.Close())
	body := []byte(`{}`)
	assert.Equal(t, http.StatusInternalServerError, post(h, "/github/gh-a", body, ghHeaders("secret-a", "d", body)))
}

// TestDurableAfterAccept: a row acknowledged with 202 is there when the
// database is opened fresh, as after a kill right after the response.
func TestDurableAfterAccept(t *testing.T) {
	h, _, _ := testHandler(t, time.Now())
	path := filepath.Join(t.TempDir(), "spool.db")
	sp2, err := OpenSpool(t.Context(), path)
	require.NoError(t, err)
	h.Spool = sp2
	srv := httptest.NewServer(h.Mux())
	body := []byte(`{"action":"opened","number":7,"repository":{"full_name":"acme/app"}}`)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL+"/github/gh-a", bytes.NewReader(body))
	require.NoError(t, err)
	for k, v := range ghHeaders("secret-a", "abc-123", body) {
		req.Header.Set(k, v)
	}
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("Authorization", "not kept")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	srv.Close()
	require.NoError(t, sp2.Close())

	fresh := openSpool(t, path)
	got := rows(t, fresh)
	require.Len(t, got, 1)
	d := got[0]
	assert.Equal(t, "gh-a", d.Endpoint)
	assert.Equal(t, SourceGitHub, d.Source)
	assert.Equal(t, "abc-123", d.DeliveryID)
	assert.Equal(t, "pull_request", d.Kind)
	assert.False(t, d.ReceivedAt.IsZero())
	assert.Equal(t, "pull_request", d.Headers["X-GitHub-Event"])
	assert.NotContains(t, d.Headers, "Authorization")
	assert.JSONEq(t, string(body), string(d.Payload))

	require.NoError(t, fresh.Delete(t.Context(), []int64{d.ID, 999}))
	assert.Empty(t, rows(t, fresh))
}

func TestSpoolBasics(t *testing.T) {
	ctx := t.Context()
	sp := openSpool(t, filepath.Join(t.TempDir(), "s.db"))
	_, err := sp.Write(ctx, Delivery{Source: "jira", DeliveryID: "a", Endpoint: "e"})
	require.ErrorIs(t, err, ErrBadDelivery)
	_, err = sp.Write(ctx, Delivery{Source: SourceGitHub, Endpoint: "e"})
	require.ErrorIs(t, err, ErrBadDelivery)
	for i := range 3 {
		ok, err := sp.Write(ctx, Delivery{Source: SourceGitHub, DeliveryID: string(rune('a' + i)), Endpoint: "e"})
		require.NoError(t, err)
		assert.True(t, ok)
	}
	got, err := sp.List(ctx, 2)
	require.NoError(t, err)
	require.Len(t, got, 2, "limit")
	assert.Equal(t, "a", got[0].DeliveryID, "oldest first")
	require.NoError(t, sp.Delete(ctx, nil))
}

func TestSpoolNewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.db")
	sp := openSpool(t, path)
	_, err := sp.db.ExecContext(t.Context(), "PRAGMA user_version = 99")
	require.NoError(t, err)
	require.NoError(t, sp.Close())
	_, err = OpenSpool(t.Context(), path)
	require.ErrorContains(t, err, "newer than this binary")
}

func TestReadSecret(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "s"), []byte("  k\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "empty"), []byte("\n"), 0o600))
	s, err := ReadSecret(context.Background(), "id", filepath.Join(dir, "s"))
	require.NoError(t, err)
	assert.Equal(t, "k", string(s.Key))
	_, err = ReadSecret(context.Background(), "", filepath.Join(dir, "empty"))
	require.ErrorIs(t, err, errNoSecret)
}
