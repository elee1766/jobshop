package config

import (
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
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
