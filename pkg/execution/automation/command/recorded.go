package command

import (
	"errors"

	"github.com/gfx-labs/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Restore validates a recorded command result without executing the command.
// Fields must come from bounded JSON decoded with api.DecodeJSON.
func Restore(fields map[string]any, options api.Options) (starlark.Value, error) {
	if !options.AllowShell {
		return nil, errors.New("shell capability is disabled by host")
	}
	invalid := errors.New("recorded capability result does not match its capability contract")
	code, ok := api.RecordedInt(fields["exit_code"])
	if !ok || code < -1 {
		return nil, invalid
	}
	stdout, ok := fields["stdout"].(string)
	if !ok {
		return nil, invalid
	}
	stderr, ok := fields["stderr"].(string)
	if !ok || len(stdout)+len(stderr) > options.Limits.OutputBytes {
		return nil, invalid
	}
	return starlarkstruct.FromStringDict(starlark.String("shell_result"), starlark.StringDict{"stdout": starlark.String(stdout), "stderr": starlark.String(stderr), "exit_code": starlark.MakeInt(code)}), nil
}
