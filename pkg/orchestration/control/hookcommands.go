package control

import (
	"bytes"
	"context"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/orchestration/hooks"
)

// hookCommands keeps sandbox and process-group mechanics behind the hook executor.
type hookCommands struct {
	c   *Controller
	cfg *config.Config
}

func (h hookCommands) ExecuteHook(ctx context.Context, command, pool string, body []byte) (hooks.CommandResult, error) {
	var stdout bytes.Buffer
	var stderr strings.Builder
	ic := isolatedCmd{argv: shArgv(command), pool: pool, stdin: bytes.NewReader(body), name: "hook", cfg: h.cfg, stdout: &stdout, stderr: &stderr}
	if !h.c.isolates(h.cfg, pool) {
		ic.pool, ic.dir = "", h.c.root.Dir
	}
	ex, err := h.c.runIsolated(ctx, ic)
	return hooks.CommandResult{Output: stdout.Bytes(), Stderr: stderr.String(), ExitCode: ex.ExitCode, Signal: ex.Signal}, err
}
