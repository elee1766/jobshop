package control

import (
	"crypto/hmac"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

// admissionController is a subController with the given admission hooks
// and on: intake hooks.
func admissionController(t *testing.T, hooks []config.AdmissionHook, intake ...string) (*Controller, *stallClock) {
	t.Helper()
	c, clk, _ := subController(t)
	cfg := *c.config()
	cfg.Ingest.Admission = hooks
	cfg.Hooks = nil
	for _, cmd := range intake {
		cfg.Hooks = append(cfg.Hooks, config.Hook{On: config.HookIntake, Command: cmd})
	}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	return c, clk
}

func putInbox(t *testing.T, c *Controller, it store.InboxItem) store.InboxItem {
	t.Helper()
	if it.Source == "" {
		it.Source = "github"
	}
	if it.DeliveryID == "" {
		it.DeliveryID = "d1"
	}
	if it.Payload == nil {
		it.Payload = json.RawMessage(`{"body":"hello","token":"s3cret"}`)
	}
	it.ReceivedAt = c.now()
	out, _, err := c.store.InboxPut(t.Context(), it)
	require.NoError(t, err)
	return out
}

func admitOne(t *testing.T, c *Controller, id int64) (store.InboxItem, []admissionResult) {
	t.Helper()
	ok, err := c.admitNext(t.Context())
	require.NoError(t, err)
	require.True(t, ok)
	it, err := c.store.InboxGet(t.Context(), id)
	require.NoError(t, err)
	var rs []admissionResult
	require.NoError(t, json.Unmarshal(it.HookResults, &rs))
	return it, rs
}

func mut(name, cmd string) config.AdmissionHook {
	return config.AdmissionHook{Name: name, Type: config.AdmissionMutating, Command: cmd}
}

func val(name, cmd string) config.AdmissionHook {
	return config.AdmissionHook{Name: name, Type: config.AdmissionValidating, Command: cmd}
}

func TestAdmissionNoHooksAdmits(t *testing.T) {
	t.Parallel()
	c, _ := admissionController(t, nil)
	it := putInbox(t, c, store.InboxItem{})
	got, rs := admitOne(t, c, it.ID)
	assert.Equal(t, store.InboxAdmitted, got.State)
	assert.JSONEq(t, string(it.Payload), string(got.Admitted))
	assert.Empty(t, rs)
	ok, err := c.admitNext(t.Context())
	require.NoError(t, err)
	assert.False(t, ok, "nothing pending")
}

func TestAdmissionOrderAndMutation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	log := filepath.Join(dir, "log")
	// Validating is listed first but runs after both mutating hooks, and
	// sees their result.
	c, _ := admissionController(t, []config.AdmissionHook{
		val("check", "cat > "+filepath.Join(dir, "seen.json")+"; echo check >> "+log),
		mut("redact", `echo redact >> `+log+`; echo '{"patch":[{"op":"replace","path":"/payload/token","value":"[redacted]"}],"reason":"secret"}'`),
		mut("full", `echo full >> `+log+`; echo '{"event":{"kind":"issue_comment","payload":{"body":"HELLO","token":"[redacted]"}}}'`),
	})
	it := putInbox(t, c, store.InboxItem{})
	got, rs := admitOne(t, c, it.ID)
	b, err := os.ReadFile(log)
	require.NoError(t, err)
	assert.Equal(t, "redact\nfull\ncheck\n", string(b))
	assert.Equal(t, store.InboxAdmitted, got.State)
	assert.JSONEq(t, `{"body":"HELLO","token":"[redacted]"}`, string(got.Admitted))
	assert.JSONEq(t, string(it.Payload), string(got.Payload), "the raw payload stays")
	require.Len(t, rs, 3)
	assert.Equal(t, []string{"redact", "full", "check"}, []string{rs[0].Name, rs[1].Name, rs[2].Name})
	assert.Equal(t, admitMutated, rs[0].Decision)
	assert.Equal(t, "secret", rs[0].Reason)
	assert.Contains(t, string(rs[0].Diff), "[redacted]")
	assert.Equal(t, admitMutated, rs[1].Decision)
	assert.Contains(t, string(rs[1].Diff), "HELLO")
	assert.Equal(t, "allow", rs[2].Decision)
	seen, err := os.ReadFile(filepath.Join(dir, "seen.json"))
	require.NoError(t, err)
	var ev admissionEvent
	require.NoError(t, json.Unmarshal(seen, &ev))
	assert.Equal(t, "d1", ev.DeliveryID)
	assert.JSONEq(t, `{"body":"HELLO","token":"[redacted]"}`, string(ev.Payload))
}

func TestAdmissionDeny(t *testing.T) {
	t.Parallel()
	c, _ := admissionController(t, []config.AdmissionHook{
		val("ok", ""+`echo '{"decision":"allow"}'`),
		val("no", `echo '{"decision":"deny","reason":"spam"}'`),
	})
	it := putInbox(t, c, store.InboxItem{})
	got, rs := admitOne(t, c, it.ID)
	assert.Equal(t, store.InboxDenied, got.State)
	assert.Nil(t, got.Admitted)
	assert.Equal(t, "spam", rs[1].Reason)
}

func TestAdmissionAskHumanHolds(t *testing.T) {
	t.Parallel()
	c, _ := admissionController(t, []config.AdmissionHook{val("ask", `echo '{"decision":"ask_human","reason":"unsure"}'`)})
	it := putInbox(t, c, store.InboxItem{RepoOrTeam: "acme/app", Number: 7, Actor: "mallory"})
	got, _ := admitOne(t, c, it.ID)
	assert.Equal(t, store.InboxHeld, got.State)
	ts, err := c.store.List(t.Context(), store.Filter{MetaPack: hookMetaPack, MetaKey: "key", MetaValue: "admission/d1"})
	require.NoError(t, err)
	require.Len(t, ts, 1)
	assert.Contains(t, ts[0].Labels, needsHumanLabel)
	assert.Contains(t, ts[0].Body, "unsure")
	assert.Contains(t, ts[0].Body, "acme/app#7")
}

// TestAdmissionHeldReleased (#353): closing the needs-human ticket moves
// the held item back to pending, and the next admission applies the
// answer without re-running the validating hook.
func TestAdmissionHeldReleased(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		resolution, want string
	}{{"done", store.InboxAdmitted}, {"wontfix", store.InboxDenied}} {
		t.Run(tc.resolution, func(t *testing.T) {
			c, _ := admissionController(t, []config.AdmissionHook{val("ask", `echo '{"decision":"ask_human","reason":"unsure"}'`)})
			it := putInbox(t, c, store.InboxItem{})
			got, _ := admitOne(t, c, it.ID)
			require.Equal(t, store.InboxHeld, got.State)

			// Unanswered: stays held.
			require.NoError(t, c.releaseAnswered(t.Context()))
			got, err := c.store.InboxGet(t.Context(), it.ID)
			require.NoError(t, err)
			assert.Equal(t, store.InboxHeld, got.State)

			ts, err := c.store.List(t.Context(), store.Filter{MetaPack: hookMetaPack, MetaKey: "key", MetaValue: "admission/d1"})
			require.NoError(t, err)
			require.Len(t, ts, 1)
			_, err = c.store.CloseTicket(t.Context(), "human", ts[0].ID, nil, tc.resolution)
			require.NoError(t, err)

			require.NoError(t, c.releaseAnswered(t.Context()))
			got, err = c.store.InboxGet(t.Context(), it.ID)
			require.NoError(t, err)
			assert.Equal(t, store.InboxPending, got.State)

			got, rs := admitOne(t, c, it.ID)
			assert.Equal(t, tc.want, got.State)
			require.Len(t, rs, 1, "the human's answer, not the hook again")
			assert.Equal(t, "human", rs[0].Name)
		})
	}
}

func TestAdmissionTimeout(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		policy, typ, state, decision string
	}{
		{"", config.AdmissionValidating, store.InboxDenied, "deny"},
		{"ignore", config.AdmissionValidating, store.InboxAdmitted, admitIgnored},
		{"", config.AdmissionMutating, store.InboxAdmitted, admitFailed},
		{"ignore", config.AdmissionMutating, store.InboxAdmitted, admitIgnored},
	} {
		t.Run(tc.typ+"/"+tc.policy, func(t *testing.T) {
			h := config.AdmissionHook{Name: "slow", Type: tc.typ, Command: "sleep 5", Timeout: "100ms", FailurePolicy: tc.policy}
			c, _ := admissionController(t, []config.AdmissionHook{h})
			it := putInbox(t, c, store.InboxItem{})
			got, rs := admitOne(t, c, it.ID)
			assert.Equal(t, tc.state, got.State)
			require.Len(t, rs, 1)
			assert.Equal(t, tc.decision, rs[0].Decision)
			assert.Contains(t, rs[0].Error, "timed out")
			if tc.state == store.InboxAdmitted {
				assert.JSONEq(t, string(it.Payload), string(got.Admitted), "a failed mutation leaves the event unchanged")
			}
		})
	}
}

func TestAdmissionBadJSONDenies(t *testing.T) {
	t.Parallel()
	c, _ := admissionController(t, []config.AdmissionHook{val("bad", "echo nope")})
	it := putInbox(t, c, store.InboxItem{})
	got, rs := admitOne(t, c, it.ID)
	assert.Equal(t, store.InboxDenied, got.State)
	assert.Contains(t, rs[0].Error, "not JSON")
}

func TestAdmissionMatch(t *testing.T) {
	t.Parallel()
	yes, no := true, false
	base := store.InboxItem{Source: "github", Kind: "issue_comment", RepoOrTeam: "acme/app", Actor: "bob"}
	for name, tc := range map[string]struct {
		m    config.AdmissionMatch
		it   store.InboxItem
		runs bool
	}{
		"empty":         {config.AdmissionMatch{}, base, true},
		"source":        {config.AdmissionMatch{Sources: []string{"linear"}}, base, false},
		"kind":          {config.AdmissionMatch{Kinds: []string{"issue_comment"}}, base, true},
		"kind miss":     {config.AdmissionMatch{Kinds: []string{"issues"}}, base, false},
		"repo glob":     {config.AdmissionMatch{Repos: []string{"acme/*"}}, base, true},
		"repo miss":     {config.AdmissionMatch{Repos: []string{"other/*"}}, base, false},
		"team glob":     {config.AdmissionMatch{Teams: []string{"EN*"}}, store.InboxItem{Source: "linear", RepoOrTeam: "ENG"}, true},
		"team miss":     {config.AdmissionMatch{Teams: []string{"OPS"}}, store.InboxItem{Source: "linear", RepoOrTeam: "ENG"}, false},
		"actor":         {config.AdmissionMatch{Actors: []string{"alice"}}, base, false},
		"bots only":     {config.AdmissionMatch{Bots: &yes}, base, false},
		"humans only":   {config.AdmissionMatch{Bots: &no}, base, true},
		"bot actor":     {config.AdmissionMatch{Bots: &yes}, store.InboxItem{Source: "github", Actor: "dep[bot]"}, true},
		"project":       {config.AdmissionMatch{Projects: []string{"nope"}}, base, false},
		"project match": {config.AdmissionMatch{Projects: []string{"PROJECT"}}, base, true},
	} {
		t.Run(name, func(t *testing.T) {
			h := val("v", `echo '{"decision":"deny"}'`)
			h.Match = tc.m
			c, _ := admissionController(t, []config.AdmissionHook{h})
			for i, p := range h.Match.Projects {
				if p == "PROJECT" {
					h.Match.Projects[i] = c.config().DefaultProject
				}
			}
			it := putInbox(t, c, tc.it)
			got, rs := admitOne(t, c, it.ID)
			assert.Equal(t, tc.runs, len(rs) == 1)
			assert.Equal(t, tc.runs, got.State == store.InboxDenied)
		})
	}
}

func TestAdmissionIntakeHooksCoverLinear(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c, _ := admissionController(t, nil, "cat > "+filepath.Join(dir, "ev.json")+`; echo '{"decision":"deny","reason":"inj"}'`)
	gh := putInbox(t, c, store.InboxItem{Source: "github", DeliveryID: "g"})
	got, _ := admitOne(t, c, gh.ID)
	assert.Equal(t, store.InboxAdmitted, got.State, "intake hooks match linear only")
	ln := putInbox(t, c, store.InboxItem{Source: "linear", DeliveryID: "l"})
	got, _ = admitOne(t, c, ln.ID)
	assert.Equal(t, store.InboxDenied, got.State)
	b, err := os.ReadFile(filepath.Join(dir, "ev.json"))
	require.NoError(t, err)
	assert.Contains(t, string(b), `"on":"intake"`)
}

func TestAdmissionURLSignature(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secret := []byte("topsecret")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sec"), append(secret, '\n'), 0o600))
	var deliveries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !hmac.Equal([]byte(r.Header.Get("X-Jobshop-Signature-256")), []byte(admissionSignature(secret, body))) {
			http.Error(w, "bad signature", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Jobshop-Request-Id") == "" {
			http.Error(w, "no request id", http.StatusBadRequest)
			return
		}
		deliveries = append(deliveries, r.Header.Get("X-Jobshop-Delivery"))
		_, _ = io.WriteString(w, `{"decision":"allow"}`)
	}))
	t.Cleanup(srv.Close)

	// The test server rejects unsigned and wrongly signed requests.
	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(`{}`))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(`{}`))
	require.NoError(t, err)
	req.Header.Set("X-Jobshop-Signature-256", admissionSignature([]byte("wrong"), []byte(`{}`)))
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	h := config.AdmissionHook{Name: "remote", Type: config.AdmissionValidating, URL: srv.URL, SecretFile: config.SecretPath(filepath.Join(dir, "sec"))}
	c, _ := admissionController(t, []config.AdmissionHook{h})
	it := putInbox(t, c, store.InboxItem{})
	got, _ := admitOne(t, c, it.ID)
	assert.Equal(t, store.InboxAdmitted, got.State)
	assert.Equal(t, []string{"d1"}, deliveries)

	// A wrong secret is refused by the server, which denies under fail.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sec"), []byte("other"), 0o600))
	it2 := putInbox(t, c, store.InboxItem{DeliveryID: "d2"})
	got, rs := admitOne(t, c, it2.ID)
	assert.Equal(t, store.InboxDenied, got.State)
	assert.Contains(t, rs[0].Error, "HTTP 401")
}

func TestAdmissionReclaimAfterLeaseExpiry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	log := filepath.Join(dir, "log")
	c, clk := admissionController(t, []config.AdmissionHook{val("v", `grep -o '"delivery_id":"[^"]*"' >> `+log)})
	it := putInbox(t, c, store.InboxItem{DeliveryID: "dx"})
	// A first admitter claims the item, runs the hooks and dies before
	// storing the result.
	claimed, err := c.store.InboxClaim(t.Context(), []string{store.InboxPending}, admissionHolder, time.Minute)
	require.NoError(t, err)
	c.admissionRunner().run(t.Context(), "", admissionHooks(c.config()), claimed, false)
	ok, err := c.admitNext(t.Context())
	require.NoError(t, err)
	assert.False(t, ok, "leased item is not claimable")
	clk.advance(2 * time.Minute)
	got, _ := admitOne(t, c, it.ID)
	assert.Equal(t, store.InboxAdmitted, got.State)
	assert.EqualValues(t, 2, got.Attempts)
	b, err := os.ReadFile(log)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat(`"delivery_id":"dx"`+"\n", 2), string(b))
}

func TestApplyJSONPatch(t *testing.T) {
	t.Parallel()
	out, err := applyJSONPatch([]byte(`{"a":[1,2],"b":{"c":"x"}}`), []byte(`[
		{"op":"add","path":"/a/1","value":9},
		{"op":"remove","path":"/b/c"},
		{"op":"copy","from":"/a/0","path":"/b/d"},
		{"op":"move","from":"/a/2","path":"/e"},
		{"op":"test","path":"/e","value":2}
	]`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"a":[1,9],"b":{"d":1},"e":2}`, string(out))
	_, err = applyJSONPatch([]byte(`{"a":1}`), []byte(`[{"op":"test","path":"/a","value":2}]`))
	require.Error(t, err)
	_, err = applyJSONPatch([]byte(`{"a":1}`), []byte(`[{"op":"replace","path":"/nope","value":2}]`))
	require.Error(t, err)
}
