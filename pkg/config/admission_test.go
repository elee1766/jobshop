package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdmissionConfig(t *testing.T) {
	base := "projects:\n  js: {}\nagents:\n  a:\n    command: x\ningest:\n  admission:\n"
	c, err := Parse([]byte(base + `    - name: redact
      type: mutating
      command: ./redact
      match: {sources: [github], repos: ["acme/*"], bots: false}
    - name: remote
      type: validating
      url: https://hooks.example/check
      secret_file: secrets/hook
      timeout: 5s
      retry: {attempts: 3, backoff: 500ms, max_backoff: 10s}
      failure_policy: ignore
      match: {projects: [js], teams: [ENG]}
`))
	require.NoError(t, err)
	require.Len(t, c.Ingest.Admission, 2)
	assert.Equal(t, FailurePolicyFail, c.Ingest.Admission[0].Policy())
	assert.Equal(t, FailurePolicyIgnore, c.Ingest.Admission[1].Policy())
	assert.Equal(t, HTTPRetry{Attempts: 3, Backoff: "500ms", MaxBackoff: "10s"}, c.Ingest.Admission[1].Retry)

	for _, tt := range []struct{ name, hook, want string }{
		{"bad name", "    - {name: Bad, type: mutating, command: x}\n", "name must match"},
		{"bad type", "    - {name: a, type: sometimes, command: x}\n", "type"},
		{"no transport", "    - {name: a, type: mutating}\n", "exactly one of command and url"},
		{"both transports", "    - {name: a, type: mutating, command: x, url: http://h}\n", "exactly one"},
		{"url no secret", "    - {name: a, type: mutating, url: http://h}\n", "secret_file is required"},
		{"bad url", "    - {name: a, type: mutating, url: ftp://h, secret_file: s}\n", "http or https"},
		{"secret on command", "    - {name: a, type: mutating, command: x, secret_file: s}\n", "secret_file is for url"},
		{"bad timeout", "    - {name: a, type: mutating, command: x, timeout: soon}\n", "timeout"},
		{"retry on command", "    - {name: a, type: mutating, command: x, retry: {attempts: 3}}\n", "retry is for url"},
		{"bad retry", "    - {name: a, type: mutating, url: http://h, secret_file: s, retry: {attempts: 50}}\n", "retry.attempts"},
		{"bad backoff", "    - {name: a, type: mutating, url: http://h, secret_file: s, retry: {backoff: soon}}\n", "retry.backoff"},
		{"bad policy", "    - {name: a, type: mutating, command: x, failure_policy: maybe}\n", "failure_policy"},
		{"bad source", "    - {name: a, type: mutating, command: x, match: {sources: [jira]}}\n", "match.sources"},
		{"bad glob", "    - {name: a, type: mutating, command: x, match: {repos: [\"[\"]}}\n", "glob"},
		{"bad project", "    - {name: a, type: mutating, command: x, match: {projects: [zz]}}\n", "unknown project"},
		{"duplicate", "    - {name: a, type: mutating, command: x}\n    - {name: a, type: mutating, command: x}\n", "duplicate"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(base + tt.hook))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "ingest.admission")
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

// Backoff doubles to max_backoff; Budget, which sizes the admission
// lease, counts every try and the longest wait between them.
func TestHTTPRetry(t *testing.T) {
	r := HTTPRetry{Attempts: 4, Backoff: "1s", MaxBackoff: "3s"}
	assert.Equal(t, []time.Duration{time.Second, 2 * time.Second, 3 * time.Second}, []time.Duration{r.Wait(2), r.Wait(3), r.Wait(4)})
	assert.Equal(t, 4*10*time.Second+3*3*time.Second, r.Budget(10*time.Second))
	assert.Equal(t, 30*time.Second, HTTPRetry{}.Budget(30*time.Second))
}

func TestHookURLConfig(t *testing.T) {
	base := "projects:\n  js: {}\nhooks:\n"
	c, err := Parse([]byte(base + "  - {on: action, url: https://h/guard, secret_file: s, retry: {attempts: 2}}\n"))
	require.NoError(t, err)
	assert.Equal(t, "https://h/guard", c.Hooks[0].URL)
	for _, tt := range []struct{ hook, want string }{
		{"  - {on: action}\n", "exactly one of command and url"},
		{"  - {on: action, url: https://h}\n", "secret_file is required"},
		{"  - {on: action, command: x, retry: {attempts: 2}}\n", "retry is for url"},
	} {
		_, err := Parse([]byte(base + tt.hook))
		require.Error(t, err)
		assert.Contains(t, err.Error(), tt.want)
	}
}
