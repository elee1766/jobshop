package control

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// maxDeskExecOutput bounds a desk command's stdout kept in the result.
const maxDeskExecOutput = 1 << 20

// deskExec runs a desk_commands entry for a Desk plugin (#542): sh -c in
// the root dir, the input as JSON on stdin like hooks (#188). A non-zero
// exit or a timeout is an error. Every run is logged with the actor; with
// a ticket it also gets a desk_exec history row.
func (c *Controller) deskExec(ctx context.Context, sc scope, p rpc.DeskExecParams) (any, error) {
	if err := requireAdmin(sc, "desk exec"); err != nil {
		return nil, err
	}
	cfg := c.config()
	dc, ok := cfg.DeskCommands[p.Name]
	if !ok {
		return nil, fmt.Errorf("%w: desk command %q", domain.ErrNotFound, p.Name)
	}
	if p.Ticket != "" {
		if _, err := c.store.Get(ctx, p.Ticket); err != nil {
			return nil, err
		}
	}
	in := []byte(p.Input)
	if len(bytes.TrimSpace(in)) == 0 {
		in = []byte("null")
	}
	if !json.Valid(in) {
		return nil, fmt.Errorf("%w: input is not JSON", domain.ErrInvalid)
	}
	var res rpc.DeskExecResult
	var exit int
	var runErr error
	if dc.Handler != "" {
		res, exit, runErr = c.runDeskHandler(ctx, cfg, p.Name, dc, in)
	} else {
		res, exit, runErr = c.runDeskCommand(ctx, dc, in)
	}
	c.log.InfoContext(ctx, "desk exec", "name", p.Name, "actor", sc.actor, "ticket", p.Ticket, "exit", exit, "err", runErr)
	if p.Ticket != "" {
		d := map[string]any{"name": p.Name, "exit": exit}
		if runErr != nil {
			d["error"] = runErr.Error()
		}
		if err := c.store.Record(ctx, sc.actor, p.Ticket, "desk_exec", d); err != nil {
			return nil, err
		}
	}
	if runErr != nil {
		return nil, runErr
	}
	return res, nil
}

func (c *Controller) runDeskCommand(ctx context.Context, dc config.DeskCommand, in []byte) (rpc.DeskExecResult, int, error) {
	ctx, cancel := context.WithTimeout(ctx, config.Duration(dc.Timeout, 30*time.Second))
	defer cancel()
	// No pool applies to Desk exec: it runs unisolated in the root dir,
	// through proc.Start for the process-group kill (#810).
	var stdout bytes.Buffer
	var stderr strings.Builder
	ex, err := c.runIsolated(ctx, isolatedCmd{argv: shArgv(dc.Command), dir: c.root.Dir, stdin: bytes.NewReader(in), stdout: &stdout, stderr: &stderr, name: "desk-exec"})
	if ctx.Err() != nil {
		return rpc.DeskExecResult{}, -1, fmt.Errorf("desk command timed out after %s", config.Duration(dc.Timeout, 30*time.Second))
	}
	if err != nil {
		return rpc.DeskExecResult{}, -1, fmt.Errorf("desk command: %w", err)
	}
	if ex.ExitCode != 0 || ex.Signal != "" {
		return rpc.DeskExecResult{}, ex.ExitCode, fmt.Errorf("desk command exited %s: %s", exitText(ex), truncate(strings.TrimSpace(stderr.String())))
	}
	out := stdout.Bytes()
	if len(out) > maxDeskExecOutput {
		return rpc.DeskExecResult{}, 0, fmt.Errorf("desk command output over %d bytes", maxDeskExecOutput)
	}
	if t := bytes.TrimSpace(out); len(t) > 0 && json.Valid(t) {
		return rpc.DeskExecResult{Output: t}, 0, nil
	}
	return rpc.DeskExecResult{Text: string(out)}, 0, nil
}
