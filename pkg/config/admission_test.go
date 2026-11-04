package config

import (
	"testing"

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
      failure_policy: ignore
      match: {projects: [js], teams: [ENG]}
`))
	require.NoError(t, err)
	require.Len(t, c.Ingest.Admission, 2)
	assert.Equal(t, FailurePolicyFail, c.Ingest.Admission[0].Policy())
	assert.Equal(t, FailurePolicyIgnore, c.Ingest.Admission[1].Policy())

	for _, tt := range []struct{ name, hook, want string }{
		{"bad name", "    - {name: Bad, type: mutating, command: x}\n", "name must match"},
		{"bad type", "    - {name: a, type: sometimes, command: x}\n", "type"},
		{"no transport", "    - {name: a, type: mutating}\n", "exactly one of command and url"},
		{"both transports", "    - {name: a, type: mutating, command: x, url: http://h}\n", "exactly one"},
		{"url no secret", "    - {name: a, type: mutating, url: http://h}\n", "secret_file is required"},
		{"bad url", "    - {name: a, type: mutating, url: ftp://h, secret_file: s}\n", "http or https"},
		{"secret on command", "    - {name: a, type: mutating, command: x, secret_file: s}\n", "secret_file is for url"},
		{"bad timeout", "    - {name: a, type: mutating, command: x, timeout: soon}\n", "timeout"},
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
