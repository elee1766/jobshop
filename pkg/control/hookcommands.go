package control

import (
	"bytes"
	"context"
	"strings"

	"github.com/elee1766/acphost/pkg/proc"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
)

// hookCommands keeps sandbox and process-group mechanics behind the hook executor.
type hookCommands struct {
	c   *Controller
	cfg *config.Config
	dir string
}

func (h hookCommands) ExecuteHook(ctx context.Context, command, pool string, body []byte) (hooks.CommandResult, error) {
	var stdout bytes.Buffer
	var stderr strings.Builder
	var ex proc.Exit
	var err error
	if h.c != nil {
		ic := isolatedCmd{argv: shArgv(command), pool: pool, stdin: bytes.NewReader(body), name: "hook", cfg: h.cfg, stdout: &stdout, stderr: &stderr}
		if !h.c.isolates(h.cfg, pool) {
			ic.pool, ic.dir = "", h.c.root.Dir
		}
		ex, err = h.c.runIsolated(ctx, ic)
	} else {
		// Admission has no pool. Its process group still dies on timeout.
		spec := proc.Spec{Command: "sh", Args: []string{"-c", command}, Dir: h.dir}
		p, startErr := proc.Start(ctx, spec, proc.Stdin(bytes.NewReader(body)), proc.Stdout(&stdout), proc.Stderr(&stderr))
		if startErr != nil {
			return hooks.CommandResult{}, startErr
		}
		ex, err = p.Wait()
	}
	return hooks.CommandResult{Output: stdout.Bytes(), Stderr: stderr.String(), ExitCode: ex.ExitCode, Signal: ex.Signal}, err
}
