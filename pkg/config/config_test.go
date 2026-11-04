package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const poolsYAML = `
projects:
  js: {}
agents:
  claude:
    command: claude-acp
    args: [--fast, -v]
    env:
      FOO: bar
pools:
  main:
    agent: claude
    max_sessions: 3
    repo: /src/jobshop
    ref: origin/master
    idle_timeout: 10m
    turn_timeout: 1h
  local:
    project: js
    agent: claude
    dir: /work
`

func TestParsePools(t *testing.T) {
	c, err := Parse([]byte(poolsYAML))
	require.NoError(t, err)

	assert.Equal(t, Agent{
		Command: "claude-acp",
		Args:    []string{"--fast", "-v"},
		Env:     map[string]string{"FOO": "bar"},
	}, c.Agents["claude"])
	assert.Equal(t, Pool{
		Project:     "js",
		Agent:       "claude",
		MaxSessions: 3,
		Repo:        "/src/jobshop",
		Ref:         "origin/master",
		IdleTimeout: "10m",
		TurnTimeout: "1h",
	}, c.Pools["main"], "project defaults to default_project")
	assert.Equal(t, Pool{Project: "js", Agent: "claude", Dir: "/work"}, c.Pools["local"])
}

// TestUnknownAdminMethod: a method this binary does not know is dropped
// with a warning, so a newer pack or last good config still loads (#325).
func TestUnknownAdminMethod(t *testing.T) {
	c, err := Parse([]byte("projects:\n  js: {}\nagents:\n  a:\n    command: x\npools:\n  p:\n    agent: a\n    dir: /w\n    admin: [shutdown, session-list]\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{"session-list"}, c.Pools["p"].Admin)
	require.Len(t, c.Warnings, 1)
	assert.Contains(t, c.Warnings[0], `admin method "shutdown" is unknown`)
	assert.False(t, c.Pools["p"].Allows("shutdown"))
}

func TestParseErrors(t *testing.T) {
	base := "projects:\n  js: {}\nagents:\n  a:\n    command: x\n"
	tests := []struct {
		name, yaml, want string
	}{
		{"unknown agent", base + "pools:\n  p:\n    agent: nope\n    dir: /w\n", `unknown agent "nope"`},
		{"unknown project", base + "pools:\n  p:\n    project: zz\n    agent: a\n    dir: /w\n", `unknown project "zz"`},
		{"no dir or repo", base + "pools:\n  p:\n    agent: a\n", "set dir or repo"},
		{"unknown top key", base + "bogus: 1\n", "bogus"},
		{"unknown pool key", base + "pools:\n  p:\n    agent: a\n    dir: /w\n    extra: 1\n", "extra"},
		{"unknown agent key", "agents:\n  a:\n    command: x\n    cmd: y\n", "cmd"},
		{"bad timeout", base + "pools:\n  p:\n    agent: a\n    dir: /w\n    idle_timeout: soon\n", "idle_timeout"},
		{"unknown telemetry key", base + "telemetry:\n  metrics: x\n", "metrics"},
		{"export without logs endpoint", base + "pools:\n  p:\n    agent: a\n    dir: /w\n    conversation_export: true\n", "conversation_export needs"},
		{"bad inbox retention", base + "ingest:\n  retention: 2w\n", "ingest.retention"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestParseConversationExport(t *testing.T) {
	c, err := Parse([]byte("projects:\n  js: {}\nagents:\n  a:\n    command: x\ntelemetry:\n  otlp_endpoint: http://o:4318\npools:\n  p:\n    agent: a\n    dir: /w\n    conversation_export: true\n  q:\n    agent: a\n    dir: /w\n"))
	require.NoError(t, err)
	assert.True(t, c.Pools["p"].ConversationExport)
	assert.False(t, c.Pools["q"].ConversationExport, "default off")
}

// TestSessionOptions (#304): the pool overrides the agent per option id,
// and model and effort are shorthand for their ACP ids.
func TestSessionOptions(t *testing.T) {
	c, err := Parse([]byte("projects:\n  js: {}\nagents:\n  a:\n    command: x\n    model: claude-oauth:opus\n    effort: low\n    config_options: {mode: fast}\npools:\n  p:\n    agent: a\n    dir: /w\n    model: openai-oauth:gpt\n  q:\n    agent: a\n    dir: /w\n    config_options: {reasoning_effort: high}\n"))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"model": "openai-oauth:gpt", "reasoning_effort": "low", "mode": "fast"}, c.SessionOptions("p", "a"))
	assert.Equal(t, map[string]string{"model": "claude-oauth:opus", "reasoning_effort": "high", "mode": "fast"}, c.SessionOptions("q", "a"))
	assert.Equal(t, map[string]string{"model": "claude-oauth:opus", "reasoning_effort": "low", "mode": "fast"}, c.SessionOptions("", "a"), "manual session")
}

func TestParsePoolAdmin(t *testing.T) {
	c, err := Parse([]byte("projects:\n  js: {}\nagents:\n  a:\n    command: x\npools:\n  p:\n    agent: a\n    dir: /w\n    admin: [config-reload, session-list]\n"))
	require.NoError(t, err)
	assert.True(t, c.Pools["p"].Allows("session-list"))
	assert.False(t, c.Pools["p"].Allows("drain"))
}

func TestParseTriggers(t *testing.T) {
	base := "projects:\n  js: {}\nagents:\n  a:\n    command: x\npools:\n  dev:\n    agent: a\n    dir: /w\n"
	c, err := Parse([]byte(base + `triggers:
  replan:
    when: {pool: dev, ready_below: 1}
    ticket: {pool: dev, title: Plan, priority: 1, labels: [plan]}
    cooldown: 10m
  nightly:
    when: {cron: "0 6 * * *", timezone: Europe/Berlin}
    ticket: {title: Review}
`))
	require.NoError(t, err)
	r := c.Triggers["replan"]
	require.NotNil(t, r.When.ReadyBelow)
	assert.Equal(t, 1, *r.When.ReadyBelow)
	assert.Equal(t, "js", r.Ticket.Project, "project from the ticket's pool")
	assert.Equal(t, "js", c.Triggers["nightly"].Ticket.Project, "project defaults to default_project")

	trig := func(s string) string { return base + "triggers:\n  t:\n" + s }
	tests := []struct{ name, yaml, want string }{
		{"empty when", trig("    when: {}\n    ticket: {title: x}\n"), "a check, learnings or edits"},
		{"learnings min", trig("    when: {learnings: {min: 0}}\n    ticket: {title: x}\n"), "learnings.min must be >= 1"},
		{"learnings every", trig("    when: {learnings: {min: 1, every: soon}}\n    ticket: {title: x}\n"), "learnings.every"},
		{"timeout without check", trig("    when: {ready_below: 1, check_timeout: 1s}\n    ticket: {title: x}\n"), "check_timeout needs check"},
		{"bad check timeout", trig("    when: {check: 'true', check_timeout: 0s}\n    ticket: {title: x}\n"), "positive duration"},
		{"unknown pool", trig("    when: {pool: nope, ready_below: 1}\n    ticket: {title: x}\n"), `unknown pool "nope"`},
		{"unknown ticket pool", trig("    when: {ready_below: 1}\n    ticket: {pool: nope, title: x}\n"), `ticket: unknown pool "nope"`},
		{"bad cron", trig("    when: {cron: \"61 * * * *\"}\n    ticket: {title: x}\n"), "minute"},
		{"negative cooldown", trig("    when: {ready_below: 1}\n    ticket: {title: x}\n    cooldown: -1m\n"), "cooldown"},
		{"negative threshold", trig("    when: {open_above: -1}\n    ticket: {title: x}\n"), ">= 0"},
		{"no title", trig("    when: {ready_below: 1}\n    ticket: {pool: dev}\n"), "title is required"},
		{"bad timezone", trig("    when: {cron: \"0 6 * * *\", timezone: Mars/Base}\n    ticket: {title: x}\n"), "timezone"},
		{"unknown key", trig("    when: {ready_below: 1, nope: 1}\n    ticket: {title: x}\n"), "nope"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestParseTelemetry(t *testing.T) {
	c, err := Parse([]byte("telemetry:\n  metrics_listen: 127.0.0.1:9464\n  otlp_endpoint: http://127.0.0.1:4318\n"))
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:9464", c.Telemetry.MetricsListen)
	assert.Equal(t, "http://127.0.0.1:4318", c.Telemetry.OTLPEndpoint)

	c, err = Parse([]byte("projects:\n  js: {}\n"))
	require.NoError(t, err)
	assert.Empty(t, c.Telemetry.MetricsListen, "off by default")
	assert.Empty(t, c.Telemetry.TracesURL())
	assert.Empty(t, c.Telemetry.LogsURL())
	assert.Nil(t, c.Telemetry.Headers(func(string) string { return "" }))
}

func TestTelemetrySignals(t *testing.T) {
	c, err := Parse([]byte(`
telemetry:
  otlp_endpoint: http://127.0.0.1:4318/
  otlp_logs_endpoint: http://vl:9428/insert/opentelemetry/v1/logs
  otlp_headers:
    Authorization: Bearer ${TOKEN}
    X-Static: plain
`))
	require.NoError(t, err)
	tel := c.Telemetry
	assert.Equal(t, "http://127.0.0.1:4318/v1/traces", tel.TracesURL(), "base path when no override")
	assert.Equal(t, "http://vl:9428/insert/opentelemetry/v1/logs", tel.LogsURL())
	assert.Equal(t, "Bearer ${TOKEN}", tel.OTLPHeaders["Authorization"], "stored unexpanded")
	env := map[string]string{"TOKEN": "s3cret"}
	assert.Equal(t, map[string]string{"Authorization": "Bearer s3cret", "X-Static": "plain"},
		tel.Headers(func(k string) string { return env[k] }))

	c, err = Parse([]byte("telemetry:\n  otlp_traces_endpoint: http://vt/x\n"))
	require.NoError(t, err)
	assert.Equal(t, "http://vt/x", c.Telemetry.TracesURL(), "override alone enables traces")
	assert.Empty(t, c.Telemetry.LogsURL())
}

func TestDuration(t *testing.T) {
	def := 5 * time.Minute
	tests := []struct {
		in   string
		want time.Duration
	}{
		{"90s", 90 * time.Second},
		{"", def},
		{"garbage", def},
		{"-1m", def},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Duration(tt.in, def), "%q", tt.in)
	}
}

func TestParseGateway(t *testing.T) {
	base := poolsYAML + `  chat:
    agent: claude
    type: chat
    dir: /work
gateway:
  ops:
    network: matrix
    homeserver: http://127.0.0.1:6167
    token_file: matrix-token
    pool: chat
    rooms: ["!r:x"]
    senders: ["@op:x"]
`
	c, err := Parse([]byte(base))
	require.NoError(t, err)
	assert.Equal(t, "chat", c.Gateway["ops"].Pool)

	// discord needs no homeserver; the escalation room is a forum channel id.
	discord := strings.Replace(base, "network: matrix\n    homeserver: http://127.0.0.1:6167\n", "network: discord\n", 1)
	discord = strings.Replace(discord, `rooms: ["!r:x"]`, `rooms: ["111", "222"]`+"\n    escalation_room: \"222\"", 1)
	c, err = Parse([]byte(discord))
	require.NoError(t, err)
	assert.Equal(t, "discord", c.Gateway["ops"].Network)
	assert.Equal(t, "222", c.Gateway["ops"].EscalationRoom)
	_, err = Parse([]byte(strings.Replace(base, "    homeserver: http://127.0.0.1:6167\n", "", 1)))
	assert.Error(t, err, "matrix without homeserver")

	for name, bad := range map[string][2]string{
		"ticket pool":                  {"pool: chat", "pool: main"},
		"no senders":                   {`senders: ["@op:x"]`, "senders: []"},
		"no token":                     {"token_file: matrix-token", `token_file: ""`},
		"network":                      {"network: matrix", "network: irc"},
		"escalation room not in rooms": {`senders: ["@op:x"]`, "senders: [\"@op:x\"]\n    escalation_room: \"!other:x\""},
		"digest without room":          {`senders: ["@op:x"]`, "senders: [\"@op:x\"]\n    attention_digest: \"0 9 * * *\""},
		"bad digest cron":              {`senders: ["@op:x"]`, "senders: [\"@op:x\"]\n    escalation_room: \"!r:x\"\n    attention_digest: \"daily\""},
	} {
		_, err := Parse([]byte(strings.Replace(base, bad[0], bad[1], 1)))
		assert.Error(t, err, name)
	}
}

func TestIngestRetention(t *testing.T) {
	assert.Equal(t, DefaultInboxRetention, Ingest{}.RetentionDuration())
	assert.Equal(t, 3*24*time.Hour, Ingest{Retention: "3d"}.RetentionDuration())
	assert.Equal(t, 36*time.Hour, Ingest{Retention: "36h"}.RetentionDuration())
}

const ingestApp = "identities:\n  lapp: {kind: linear-app, client_id_file: /i, client_secret_file: /s, webhook_secret_file: /w, workspace: w, teams: [ENG]}\n"

// TestIngestEndpoints (#339): endpoints parse with default paths, and a
// linear-app identity with a webhook secret and no endpoint gets an
// implicit one at /linear/webhook.
func TestIngestEndpoints(t *testing.T) {
	c, err := Parse([]byte(ingestApp + `ingest:
  spool: data/spool.db
  endpoints:
    - {name: gh-app, source: github, secret_file: secrets/gh, repos: [acme/app]}
    - {name: gh-org, source: github, secret_file: /abs/gh, path: /hooks/org}
`))
	require.NoError(t, err)
	eps := c.IngestEndpoints()
	require.Len(t, eps, 3)
	assert.Equal(t, "/github/gh-app", eps[0].EndpointPath())
	assert.Equal(t, "/hooks/org", eps[1].EndpointPath())
	assert.Equal(t, IngestEndpoint{Name: "linear-lapp", Source: IngestLinear, Identity: "lapp", Path: LinearWebhookPath, Implicit: true}, eps[2])
	assert.Equal(t, "/r/data/spool.db", c.Ingest.SpoolPath(Root{Dir: "/r"}))
	assert.Equal(t, "/r/ingest/spool.db", Ingest{}.SpoolPath(Root{Dir: "/r"}))

	// Naming the identity in an endpoint replaces the implicit one.
	c, err = Parse([]byte(ingestApp + "ingest:\n  endpoints:\n    - {name: lin, source: linear, identity: lapp}\n"))
	require.NoError(t, err)
	eps = c.IngestEndpoints()
	require.Len(t, eps, 1)
	assert.Equal(t, "/linear/lin", eps[0].EndpointPath())

	// Identity only, no ingest block: /linear/webhook still served.
	c, err = Parse([]byte(ingestApp))
	require.NoError(t, err)
	require.Len(t, c.IngestEndpoints(), 1)
	assert.Equal(t, LinearWebhookPath, c.IngestEndpoints()[0].EndpointPath())
}

// TestIngestPollers (#341): defaults, and every validation error.
func TestIngestPollers(t *testing.T) {
	const pat = "identities:\n  me: {kind: github-token, token_file: /pat, repos: [\"me/*\"]}\n  lin: {kind: linear, token_file: /l, workspace: w, teams: [ENG]}\n"
	c, err := Parse([]byte(pat + "ingest:\n  pollers:\n    - {name: p, source: github, repos: [me/app], identity: me}\n    - {name: q, source: github, repos: [x/y], events: [issues], interval: 30s}\n"))
	require.NoError(t, err)
	ps := c.Ingest.Pollers
	require.Len(t, ps, 2)
	assert.Equal(t, time.Minute, ps[0].Every())
	assert.Equal(t, PollerEvents, ps[0].EventSet())
	assert.Equal(t, 30*time.Second, ps[1].Every())
	assert.Equal(t, []string{"issues"}, ps[1].EventSet())

	for name, tc := range map[string]struct{ yaml, want string }{
		"bad name":       {"ingest:\n  pollers: [{name: A_b, source: github, repos: [a/b]}]\n", "[a-z0-9-]+"},
		"dup name":       {"ingest:\n  pollers: [{name: a, source: github, repos: [a/b]}, {name: a, source: github, repos: [a/b]}]\n", "already used"},
		"endpoint name":  {"ingest:\n  endpoints: [{name: a, source: github, secret_file: s}]\n  pollers: [{name: a, source: github, repos: [a/b]}]\n", "already used"},
		"source":         {"ingest:\n  pollers: [{name: a, source: linear, repos: [a/b]}]\n", "source must be github"},
		"no repos":       {"ingest:\n  pollers: [{name: a, source: github}]\n", "needs repos"},
		"bad repo":       {"ingest:\n  pollers: [{name: a, source: github, repos: [nope]}]\n", "owner/name"},
		"bad event":      {"ingest:\n  pollers: [{name: a, source: github, repos: [a/b], events: [push]}]\n", "must be one of"},
		"short interval": {"ingest:\n  pollers: [{name: a, source: github, repos: [a/b], interval: 1s}]\n", "at least 10s"},
		"linear id":      {pat + "ingest:\n  pollers: [{name: a, source: github, repos: [a/b], identity: lin}]\n", "github-app or github-token"},
		"not covered":    {pat + "ingest:\n  pollers: [{name: a, source: github, repos: [a/b], identity: me}]\n", "does not cover a/b"},
	} {
		_, err := Parse([]byte(tc.yaml))
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), tc.want, name)
	}
}

func TestIngestEndpointErrors(t *testing.T) {
	for name, tc := range map[string]struct{ yaml, want string }{
		"old github secret":  {"ingest:\n  github_secret_file: x\n", "ingest.endpoints"},
		"old linear secret":  {"ingest:\n  linear_secret_file: x\n", "ingest.endpoints"},
		"old spool dir":      {"ingest:\n  spool_dir: x\n", "ingest.spool"},
		"bad name":           {"ingest:\n  endpoints: [{name: Bad_Name, source: github, secret_file: s}]\n", "[a-z0-9-]+"},
		"duplicate name":     {"ingest:\n  endpoints: [{name: a, source: github, secret_file: s}, {name: a, source: github, secret_file: t, path: /x}]\n", "duplicate name"},
		"duplicate path":     {"ingest:\n  endpoints: [{name: a, source: github, secret_file: s, path: /x}, {name: b, source: github, secret_file: t, path: /x}]\n", "already used"},
		"bad source":         {"ingest:\n  endpoints: [{name: a, source: jira, secret_file: s}]\n", "linear or generic"},
		"generic no header":  {"ingest:\n  endpoints: [{name: a, source: generic, secret_file: s}]\n", "needs signature_header"},
		"header on github":   {"ingest:\n  endpoints: [{name: a, source: github, secret_file: s, signature_header: X}]\n", "for generic endpoints"},
		"no secret":          {"ingest:\n  endpoints: [{name: a, source: github}]\n", "exactly one"},
		"both secrets":       {ingestApp + "ingest:\n  endpoints: [{name: a, source: linear, secret_file: s, identity: lapp}]\n", "exactly one"},
		"unknown identity":   {"ingest:\n  endpoints: [{name: a, source: linear, identity: nope}]\n", "Linear identity"},
		"repos on linear":    {"ingest:\n  endpoints: [{name: a, source: linear, secret_file: s, repos: [a/b]}]\n", "repos is for github"},
		"bad repo":           {"ingest:\n  endpoints: [{name: a, source: github, secret_file: s, repos: [nope]}]\n", "owner/name"},
		"relative path":      {"ingest:\n  endpoints: [{name: a, source: github, secret_file: s, path: x}]\n", "must start with /"},
		"shadows identities": {ingestApp + "ingest:\n  endpoints: [{name: a, source: linear, secret_file: s, path: /linear/webhook}]\n", "name the identity"},
	} {
		_, err := Parse([]byte(tc.yaml))
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), tc.want, name)
	}
}
