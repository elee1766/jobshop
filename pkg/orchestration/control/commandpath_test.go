package control_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	packpkg "github.com/gfx-labs/jobshop/pkg/pack"
	"github.com/gfx-labs/jobshop/pkg/support/clock"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestCommandStepFindsJobshop: a command step runs `jobshop version` and
// "$JOBSHOP_BIN" version with a controller PATH that has no jobshop on
// it (#669). The controller's binary is a symlink named jobshop to the
// test binary, which answers `version` (TestMain).
func TestCommandStepFindsJobshop(t *testing.T) {
	sh, err := exec.LookPath("sh")
	require.NoError(t, err)
	// PATH for the controller: only sh's dir, with no jobshop in it.
	t.Setenv("PATH", filepath.Dir(sh))

	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	bin := filepath.Join(dir, "bin")
	require.NoError(t, os.MkdirAll(bin, 0o750))
	exe := filepath.Join(bin, "jobshop")
	require.NoError(t, os.Symlink(os.Args[0], exe))
	pack := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pack, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, packpkg.FileName), []byte(`workflows:
  ver:
    vars: [issue]
    steps:
      - id: v
        type: command
        pool: dev
        run: 'jobshop version && "$JOBSHOP_BIN" version'
`), 0o600))
	cfg := fmt.Sprintf("packs:\n  - path: pk\nprojects:\n  js: {}\nagents:\n  fake:\n    command: %q\npools:\n  dev: {project: js, agent: fake, max_sessions: 0, dir: %q}\n", os.Args[0], dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	clk := clock.NewFake(time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC))
	go func() {
		done <- control.Run(ctx, root, withPass(dir, control.Options{Tick: time.Hour, Clock: clk, HostExe: exe}))
	}()
	t.Cleanup(func() {
		cancel()
		if err := <-done; !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	})
	admin := dial(t, root.AdminSocket())
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "ver", Vars: map[string]string{"issue": "1"}})
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r.Run.ID).Steps[0].Status == domain.StatusClosed
	}, "step done")
	f := stepFields(t, showRun(t, admin, r.Run.ID).Steps[0])
	require.Equal(t, "ok", f.Result, f.Output)
	assert.Equal(t, 2, strings.Count(f.Output, "jobshop test-version"), f.Output)
}
