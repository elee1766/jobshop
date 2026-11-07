package hooks

import (
	"crypto/hmac"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// A url guardrail hook (#188) answers like a command hook, signed like an
// admission url hook. 503 and 429 are retried with backoff or the
// server's Retry-After; 4xx is final; out of tries is deny.
func TestHookURLRetry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	secret := []byte("topsecret")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sec"), secret, 0o600))
	var mu sync.Mutex
	codes := []int{} // status per request, popped in order; empty is 200 allow
	tries := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !hmac.Equal([]byte(r.Header.Get("X-Jobshop-Signature-256")), []byte(hookSignature(secret, body))) {
			http.Error(w, "bad signature", http.StatusUnauthorized)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		tries++
		if len(codes) > 0 {
			code := codes[0]
			codes = codes[1:]
			if code == http.StatusTooManyRequests {
				w.Header().Set("Retry-After", "7")
			}
			http.Error(w, "busy", code)
			return
		}
		_, _ = io.WriteString(w, `{"decision":"deny","reason":"remote says no"}`)
	}))
	t.Cleanup(srv.Close)

	var waits []time.Duration
	after := func(d time.Duration) <-chan time.Time {
		waits = append(waits, d)
		ch := make(chan time.Time, 1)
		ch <- time.Time{}
		return ch
	}
	h := config.Hook{
		On: config.HookAction, URL: srv.URL, SecretFile: config.SecretPath(filepath.Join(dir, "sec")),
		Retry: config.HTTPRetry{Attempts: 3, Backoff: "2s", MaxBackoff: "5s"},
	}
	run := func(set ...int) Answer {
		mu.Lock()
		codes, tries, waits = set, 0, nil
		mu.Unlock()
		p := hookPost{url: h.URL, secret: h.SecretFile, timeout: time.Second, retry: h.Retry, after: after}
		out, err := p.do(t.Context(), []byte(`{"on":"action"}`))
		if err != nil {
			return Answer{rpc.GuardDeny, err.Error()}
		}
		return parseHookAnswer(out)
	}

	// The answer comes from the response body.
	a := run()
	assert.Equal(t, Answer{rpc.GuardDeny, "remote says no"}, a)
	assert.Equal(t, 1, tries)

	// 503 then 429 (Retry-After 7s, capped to max_backoff 5s), then the answer.
	a = run(http.StatusServiceUnavailable, http.StatusTooManyRequests)
	assert.Equal(t, "remote says no", a.Reason)
	assert.Equal(t, 3, tries)
	assert.Equal(t, []time.Duration{2 * time.Second, 5 * time.Second}, waits)

	// Out of tries: deny with the last error.
	a = run(http.StatusBadGateway, http.StatusBadGateway, http.StatusBadGateway)
	assert.Equal(t, rpc.GuardDeny, a.Decision)
	assert.Contains(t, a.Reason, "HTTP 502")
	assert.Contains(t, a.Reason, "after 3 tries")
	assert.Equal(t, 3, tries)

	// A 4xx is final.
	a = run(http.StatusForbidden)
	assert.Contains(t, a.Reason, "HTTP 403")
	assert.Equal(t, 1, tries)
}
