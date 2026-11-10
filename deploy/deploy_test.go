// Package deploy_test runs the deploy/ shell scripts against saved input.
// It has no non-test code, so the jobshop binary does not depend on it.
package deploy_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeGH is a gh stand-in: `gh pr view <url> ...` prints the state named for
// that URL in $FAKE_GH_STATES ("url=STATE ..."), else OPEN.
const fakeGH = `#!/bin/sh
url=$3
for kv in $FAKE_GH_STATES; do
	if [ "${kv%%=*}" = "$url" ]; then echo "${kv#*=}"; exit 0; fi
done
echo OPEN
`

func runClosedLoop(t *testing.T, trace []byte, states string) (string, int) {
	t.Helper()
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not installed")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	require.NoError(t, os.WriteFile(gh, []byte(fakeGH), 0o700))
	in := filepath.Join(dir, "trace.json")
	require.NoError(t, os.WriteFile(in, trace, 0o600))
	cmd := exec.Command("sh", "verify-closed-loop.sh", "--trace", in)
	cmd.Env = append(os.Environ(), "GH="+gh, "FAKE_GH_STATES="+states)
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(out), exit.ExitCode()
	}
	require.NoError(t, err)
	return string(out), 0
}

const merged = "https://github.com/gfx-labs/jobshop/pull/74=MERGED"

func TestClosedLoopPass(t *testing.T) {
	b, err := os.ReadFile("testdata/closed-loop-pass.json")
	require.NoError(t, err)
	out, code := runClosedLoop(t, b, merged)
	assert.Equal(t, 0, code, out)
	assert.Contains(t, out, "ok    created by trigger replan")
	assert.Contains(t, out, "ok    dev tickets from the planner: js-dev1 js-dev2")
	assert.Contains(t, out, "ok    reached refinery: js-dev1")
	assert.Contains(t, out, "ok    merged PR: js-dev1 https://github.com/gfx-labs/jobshop/pull/74")
	assert.Contains(t, out, "closed-loop check passed")
}

// TestClosedLoopFail breaks one link of the saved trace at a time and checks
// the script names it.
func TestClosedLoopFail(t *testing.T) {
	b, err := os.ReadFile("testdata/closed-loop-pass.json")
	require.NoError(t, err)
	cases := []struct {
		name   string
		edit   func(m map[string]any)
		states string
		want   string
	}{
		{"no trigger", func(m map[string]any) { delete(m, "trigger") }, merged, "plan ticket js-plan1 was not created by a trigger"},
		{"plan open", func(m map[string]any) { m["status"] = "in_progress" }, merged, "plan ticket js-plan1 is in_progress, not closed"},
		{"no dev ticket", func(m map[string]any) { m["holders"] = []any{"js-other"} }, merged, "no dev ticket created by the planner session (js-other)"},
		{"no refinery", func(m map[string]any) { child(m, 0)["pool"] = "dev" }, merged, "no dev ticket reached the refinery pool"},
		{"pr not merged", func(map[string]any) {}, "", "no merged PR referenced by a refinery ticket (js-dev1)"},
		{"no pr link", func(m map[string]any) { child(m, 0)["links"] = []any{"https://github.com/gfx-labs/jobshop/issues/70"} }, merged, "no merged PR"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m map[string]any
			require.NoError(t, json.Unmarshal(b, &m))
			tc.edit(m)
			in, err := json.Marshal(m)
			require.NoError(t, err)
			out, code := runClosedLoop(t, in, tc.states)
			assert.Equal(t, 1, code, out)
			assert.Contains(t, out, "FAIL  "+tc.want)
			assert.True(t, strings.HasSuffix(strings.TrimSpace(out), "closed-loop check FAILED"), out)
		})
	}
}

func TestClosedLoopUsage(t *testing.T) {
	out, code := runClosedLoop(t, []byte(`[1,2]`), "")
	assert.Equal(t, 2, code, out)
	assert.Contains(t, out, "not a ticket trace")
}

func child(m map[string]any, i int) map[string]any {
	return m["children"].([]any)[i].(map[string]any)
}
