package control

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// nixosProfile is the part of NixOS's generated /etc/profile that reset
// PATH in a resumed session's tool shell (#253).
const nixosProfile = `if [ -z "$__NIXOS_SET_ENVIRONMENT_DONE" ]; then
  export PATH=/run/current-system/sw/bin
fi
`

// TestResumeShellEnv (#253): after a resume jcode runs tools as `bash
// -lc`, which reads /etc/profile. The host spec env must survive that
// with the session PATH (jobshop first) and JOBSHOP_* intact.
func TestResumeShellEnv(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("no bash")
	}
	c, _ := rulesController(t, rulesBase)
	exe := filepath.Join(t.TempDir(), "bin", "jobshop")
	c.opt.HostExe = exe
	spec, _, err := c.hostSpec(store.Ticket{ID: "js-s1", Pool: "dev"}, store.SessionFields{Agent: "fake"})
	require.NoError(t, err)
	assert.Contains(t, spec.Env, "__NIXOS_SET_ENVIRONMENT_DONE=1")

	profile := filepath.Join(t.TempDir(), "profile")
	require.NoError(t, os.WriteFile(profile, []byte(nixosProfile), 0o600))
	run := func(script string) string {
		cmd := exec.Command(bash, "--noprofile", "--norc", "-c", script)
		cmd.Env = append([]string{"HOME=" + t.TempDir()}, spec.Env...)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return strings.TrimSpace(string(out))
	}
	got := run(". " + profile + `; echo "$PATH|$JOBSHOP_SESSION"`)
	assert.True(t, strings.HasPrefix(got, filepath.Dir(exe)+":"), "session PATH kept: %s", got)
	assert.True(t, strings.HasSuffix(got, "|js-s1"), got)

	// On a NixOS host, the real login shell too.
	if b, err := os.ReadFile("/etc/profile"); err == nil && strings.Contains(string(b), "__NIXOS_SET_ENVIRONMENT_DONE") {
		cmd := exec.Command(bash, "-lc", `echo "$PATH"`)
		cmd.Env = append([]string{"HOME=" + t.TempDir()}, spec.Env...)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		assert.True(t, strings.HasPrefix(strings.TrimSpace(string(out)), filepath.Dir(exe)+":"), string(out))
	}
}
