package control

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProbeShell: the spec PATH overrides the inherited one, a missing
// command or JOBSHOP_SESSION is a permanent error naming it, and a
// sandboxed spec is not probed on the host (JOB-67).
func TestProbeShell(t *testing.T) {
	sh, err := exec.LookPath("sh")
	require.NoError(t, err)
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "jobshop"), []byte("#!/bin/sh\n"), 0o700))
	want := []string{"jobshop", "sh"}
	inherited := []string{"PATH=" + filepath.Dir(sh), "HOME=/x"}

	ok := host.Spec{Env: []string{"JOBSHOP_SESSION=js-s1", "PATH=" + bin + ":" + filepath.Dir(sh)}}
	require.NoError(t, probeShell(ok, inherited, want))

	noPath := host.Spec{Env: []string{"JOBSHOP_SESSION=js-s1", "PATH=" + bin}}
	err = probeShell(noPath, inherited, want)
	require.Error(t, err)
	assert.True(t, host.Permanent(err))
	assert.Contains(t, err.Error(), "missing sh on")

	noSession := host.Spec{Env: []string{"PATH=" + bin + ":" + filepath.Dir(sh)}}
	err = probeShell(noSession, inherited, want)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing $JOBSHOP_SESSION on")

	assert.NoError(t, probeShell(host.Spec{Sandbox: &sandbox.Sandbox{}}, nil, want))
	assert.NoError(t, probeShell(noPath, inherited, nil))
}
