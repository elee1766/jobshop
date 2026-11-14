package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLinearIntakeSkipsTargets (JOB-4): a fixture pack shaped like the
// github pack's intake. The default skip labels reach the intake signal
// with target and proposed in them, and an instance can still add its
// own on top.
func TestLinearIntakeSkipsTargets(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"pk/pack.yaml": `defaults:
  linear_team: ""
  linear_skip_labels: Blocked,target,proposed
workflows:
  check:
    vars: [signal, subject, detail]
    steps:
      - {id: d, pool: plan, title: x, prompt: x}
pools:
  plan: {agent: a, dir: /w}
triggers:
  scan:
    when:
      scan:
        - {signal: linear_issue_no_run, team: "{{.linear_team}}", skip_labels: "{{.linear_skip_labels}}"}
    run:
      workflow: check
      vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}
`,
		"jobshop.yaml": hostBase + "packs:\n  - path: pk\n    vars: {linear_team: ENG}\n",
	})
	skip := func() []string {
		t.Helper()
		c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
		require.NoError(t, err)
		sigs := c.Triggers["scan"].When.Scan
		require.Len(t, sigs, 1)
		assert.Equal(t, "ENG", sigs[0].Team)
		return strings.Split(sigs[0].SkipLabels, ",")
	}
	assert.Subset(t, skip(), []string{"target", "proposed", "Blocked"})

	writeFiles(t, dir, map[string]string{
		"jobshop.yaml": hostBase + "packs:\n  - path: pk\n    vars: {linear_team: ENG, linear_skip_labels: \"Blocked,target,proposed,wontfix\"}\n",
	})
	assert.Subset(t, skip(), []string{"target", "proposed", "wontfix"})
}

// TestLinearIntakeStart (JOB-211): start names the workflow the scan
// starts itself. A pack prefix renames it with the workflows, and a
// workflow that is not a Linear issue workflow fails the load.
func TestLinearIntakeStart(t *testing.T) {
	dir := t.TempDir()
	pack := func(start string) string {
		return `workflows:
  check:
    vars: [signal, subject, detail]
    steps:
      - {id: d, pool: plan, title: x, prompt: x}
  work:
    vars: [issue]
    issue: issue
    tracker: linear
    linear_workspace: acme
    repo: acme/app
    steps:
      - {id: i, pool: plan, title: x, prompt: x}
pools:
  plan: {agent: a, dir: /w}
triggers:
  scan:
    when:
      scan:
        - {signal: linear_issue_no_run, team: ENG, start: ` + start + `}
    run:
      workflow: check
      vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}
`
	}
	load := func(start string) (*Config, error) {
		writeFiles(t, dir, map[string]string{
			"pk/pack.yaml": pack(start),
			"jobshop.yaml": hostBase + "packs:\n  - path: pk\n    prefix: app-\n",
		})
		return LoadFile(filepath.Join(dir, "jobshop.yaml"))
	}
	c, err := load("work")
	require.NoError(t, err)
	tr := c.Triggers["app-scan"]
	assert.Equal(t, "app-work", tr.ScanStart("linear_issue_no_run"))
	assert.Empty(t, tr.ScanStart("issue_no_run"))

	_, err = load("check")
	require.ErrorContains(t, err, "must have tracker linear")
	_, err = load("nope")
	require.ErrorContains(t, err, "unknown workflow")
}
