package config

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// Hook points. HookAction covers a session's outside actions, which the
// session reports with `jobshop guard <action>` before doing them.
const (
	HookClaim  = "claim"  // before a session claims a ticket
	HookAction = "action" // before push, pr-create, merge, comment, message
	HookMerge  = "merge"  // before a merge (also an action)
	// HookIntake runs on untrusted Linear issue text before it reaches a
	// prompt: run start for a Linear issue, subscription delivery (#288).
	HookIntake = "intake"
)

// HookPoints lists the valid Hook.On values.
var HookPoints = []string{HookClaim, HookAction, HookMerge, HookIntake}

// ActionLinearAPIMutation is the action name `jobshop linear api` guards
// a mutation with (#546); hooks list it in actions, next to the gh
// action names (pr-create, merge, comment, ...) in pkg/transport/cli/gh.go.
const ActionLinearAPIMutation = "linear-api-mutation"

// Hook is a guardrail command run at a fixed point (#188). It gets a JSON
// event on stdin and answers on stdout with one JSON object
// {"decision": "allow"|"deny"|"ask_human", "reason": "..."}. Empty
// stdout with exit 0 is allow; a non-zero exit, a timeout or bad output
// is deny (fail closed).
type Hook struct {
	On string `yaml:"on"`
	// Exactly one of Command (sh -c in the root dir, JSON on stdin, the
	// answer on stdout), URL (HTTP POST of the JSON, the answer in the
	// response body, signed with SecretFile like admission url hooks) and
	// Handler (a Starlark file.star:function called as fn(event, args)).
	Command    string         `yaml:"command"`
	URL        string         `yaml:"url"`
	Handler    string         `yaml:"handler,omitempty"`
	Args       map[string]any `yaml:"args,omitempty"`
	SecretFile Secret         `yaml:"secret_file"`
	Retry      HTTPRetry      `yaml:"retry"` // url only
	// Actions limits an action hook to these action names; empty is all.
	Actions []string `yaml:"actions"`
	// Pools limits the hook to sessions of these pools; empty is all.
	Pools   []string `yaml:"pools"`
	Timeout string   `yaml:"timeout"` // Go duration per try, default 30s
}

// Applies reports whether h runs for point on, the action and pool.
func (h Hook) Applies(on, action, pool string) bool {
	match := h.On == on || (h.On == HookAction && on == HookMerge)
	if !match {
		return false
	}
	if len(h.Actions) > 0 && !slices.Contains(h.Actions, action) {
		return false
	}
	return len(h.Pools) == 0 || slices.Contains(h.Pools, pool)
}

func (c *Config) checkHook(h Hook) error {
	if !slices.Contains(HookPoints, h.On) {
		return fmt.Errorf("on %q must be one of %v", h.On, HookPoints)
	}
	if h.Handler != "" {
		if h.Command != "" || h.URL != "" {
			return errors.New("set exactly one of command, url and handler")
		}
		if h.On == HookIntake {
			return errors.New("handler is not supported for on: intake")
		}
		if !h.SecretFile.IsZero() || h.Retry != (HTTPRetry{}) {
			return errors.New("secret_file and retry are for url hooks")
		}
	} else if err := checkHookURL(h.Command, h.URL, h.SecretFile, h.Retry); err != nil {
		return err
	} else if len(h.Args) > 0 {
		return errors.New("args needs handler")
	}
	if len(h.Actions) > 0 && (h.On == HookClaim || h.On == HookIntake) {
		return errors.New("actions needs on: action or merge")
	}
	for _, p := range h.Pools {
		if _, ok := c.Pools[p]; !ok {
			return fmt.Errorf("unknown pool %q", p)
		}
	}
	if h.Timeout != "" {
		if d, err := time.ParseDuration(h.Timeout); err != nil || d <= 0 {
			return fmt.Errorf("timeout %q must be a positive duration", h.Timeout)
		}
	}
	return nil
}
