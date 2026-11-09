package control

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHostInheritedEnvKeepsTools: a respawned host gets the controller's
// PATH and GH_TOKEN; only per-start systemd variables are dropped (#244).
func TestHostInheritedEnvKeepsTools(t *testing.T) {
	t.Setenv("PATH", "/nix/store/x-gh/bin")
	t.Setenv("GH_TOKEN", "tok")
	t.Setenv("INVOCATION_ID", "abc")
	env := hostInheritedEnv()
	assert.True(t, slices.Contains(env, "PATH=/nix/store/x-gh/bin"))
	assert.True(t, slices.Contains(env, "GH_TOKEN=tok"))
	assert.False(t, slices.Contains(env, "INVOCATION_ID=abc"))
}
