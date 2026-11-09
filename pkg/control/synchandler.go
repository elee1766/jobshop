package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// syncPruneEvery spaces sync journal prunes; syncRetention is how long a
// guard or desk handler's journal stays for automation show.
const (
	syncPruneEvery = time.Hour
	syncRetention  = 72 * time.Hour
	syncPruneBatch = 500
)

// callSync runs a handler loaded with cfg: a guard hook, a desk command
// or a signal script. Its arguments go through JSON so the script sees
// plain dicts and lists; a nil map is an empty dict.
func (c *Controller) callSync(ctx context.Context, cfg *config.Config, name, entry string, timeout time.Duration, args ...any) (automation.Result, error) {
	runtime := hookHost{c}.CurrentHooks()
	if runtime.Config != cfg {
		// A reload landed since cfg was read: its programs may differ.
		return automation.Result{}, errors.New("configuration reloaded during the call; retry")
	}
	program := runtime.Programs[entry]
	if program == nil {
		return automation.Result{}, fmt.Errorf("handler %s is not loaded", entry)
	}
	values := make([]any, len(args))
	for i, arg := range args {
		if m, ok := arg.(map[string]any); ok && m == nil {
			arg = map[string]any{}
		}
		data, err := json.Marshal(arg)
		if err != nil {
			return automation.Result{}, err
		}
		if err := hooks.DecodeJSON(data, &values[i]); err != nil {
			return automation.Result{}, err
		}
	}
	return c.hooks().Sync(ctx, name, program, timeout, values...)
}

// runHookHandler answers a guard hook from its Starlark handler. It fails
// closed like a command hook: an error or a bad answer is deny, and an
// empty dict is allow.
func (c *Controller) runHookHandler(ctx context.Context, cfg *config.Config, h config.Hook, ev hooks.GuardEvent) hooks.Answer {
	result, err := c.callSync(ctx, cfg, "hooks."+ev.On, h.Handler, config.Duration(h.Timeout, 30*time.Second), ev, h.Args)
	if err != nil {
		return hooks.Answer{Decision: domain.GuardDeny, Reason: "hook " + h.Handler + ": " + err.Error()}
	}
	if len(result.Output) == 0 {
		return hooks.Answer{Decision: domain.GuardAllow}
	}
	decision, _ := result.Output["decision"].(string)
	reason, _ := result.Output["reason"].(string)
	switch decision {
	case domain.GuardAllow, domain.GuardDeny, domain.GuardAskHuman:
		return hooks.Answer{Decision: decision, Reason: reason}
	}
	return hooks.Answer{Decision: domain.GuardDeny, Reason: fmt.Sprintf("hook decision %q is not allow, deny or ask_human", decision)}
}

// runDeskHandler is runDeskCommand for a Starlark handler: the returned
// dict is the result.
func (c *Controller) runDeskHandler(ctx context.Context, cfg *config.Config, name string, dc config.DeskCommand, in []byte) (rpc.DeskExecResult, int, error) {
	var input any
	if err := hooks.DecodeJSON(in, &input); err != nil {
		return rpc.DeskExecResult{}, -1, err
	}
	result, err := c.callSync(ctx, cfg, "desk_commands."+name, dc.Handler, config.Duration(dc.Timeout, 30*time.Second), input, dc.Args)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return rpc.DeskExecResult{}, -1, fmt.Errorf("desk command timed out after %s", config.Duration(dc.Timeout, 30*time.Second))
		}
		return rpc.DeskExecResult{}, 1, fmt.Errorf("desk command: %w", err)
	}
	out, err := json.Marshal(result.Output)
	if err != nil {
		return rpc.DeskExecResult{}, 1, err
	}
	if len(out) > maxDeskExecOutput {
		return rpc.DeskExecResult{}, 0, fmt.Errorf("desk command output over %d bytes", maxDeskExecOutput)
	}
	return rpc.DeskExecResult{Output: out}, 0, nil
}

// pruneSync deletes old guard and desk handler journals.
func (c *Controller) pruneSync(ctx context.Context, _ *config.Config) {
	now := c.now()
	if now.Sub(c.syncPrunedAt) < syncPruneEvery {
		return
	}
	c.syncPrunedAt = now
	n, err := c.hooks().PruneSync(ctx, now.Add(-syncRetention), syncPruneBatch)
	if err != nil {
		if ctx.Err() == nil {
			c.log.ErrorContext(ctx, "sync journal prune", "err", err)
		}
		return
	}
	if n > 0 {
		c.log.InfoContext(ctx, "sync journal pruned", "invocations", n)
	}
}
