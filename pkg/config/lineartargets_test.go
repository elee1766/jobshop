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
  stall:
    when:
      stall:
        - {signal: linear_issue_no_run, team: "{{.linear_team}}", skip_labels: "{{.linear_skip_labels}}"}
    run:
      workflow: check
      vars: {signal: "{{.stall_signal}}", subject: "{{.stall_subject}}", detail: "{{.stall_detail}}"}
`,
		"jobshop.yaml": hostBase + "packs:\n  - path: pk\n    vars: {linear_team: ENG}\n",
	})
	skip := func() []string {
		t.Helper()
		c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
		require.NoError(t, err)
		sigs := c.Triggers["stall"].When.Stall
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
