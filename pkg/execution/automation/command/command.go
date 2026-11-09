// Package command implements bounded argv execution for automation hooks.
package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type processOutput struct {
	mu             sync.Mutex
	remaining      int
	exceeded       bool
	cancel         func()
	stdout, stderr bytes.Buffer
}
type processWriter struct {
	output *processOutput
	stderr bool
}

func (w processWriter) Write(data []byte) (int, error) {
	o := w.output
	o.mu.Lock()
	defer o.mu.Unlock()
	length := min(len(data), o.remaining)
	if w.stderr {
		o.stderr.Write(data[:length])
	} else {
		o.stdout.Write(data[:length])
	}
	o.remaining -= length
	if length < len(data) {
		o.exceeded = true
		o.cancel()
	}
	return len(data), nil
}

// Call executes shell's explicit argv, environment and working directory.
func Call(host api.Host, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := host.Options()
	if !options.AllowShell {
		return nil, errors.New("shell capability is disabled by host")
	}
	var argv starlark.Value
	var cwd, timeout string
	var env *starlark.Dict
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "argv", &argv, "cwd?", &cwd, "env?", &env, "timeout?", &timeout); err != nil {
		return nil, err
	}
	value, err := api.FromValue(argv, options.Limits.DataBytes)
	if err != nil {
		return nil, err
	}
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return nil, errors.New("shell: argv must be a nonempty list of strings")
	}
	command := make([]string, len(items))
	for i, item := range items {
		text, ok := item.(string)
		if !ok || strings.ContainsRune(text, 0) {
			return nil, errors.New("shell: argv items must be strings without NUL")
		}
		command[i] = text
	}
	if strings.TrimSpace(command[0]) == "" {
		return nil, errors.New("shell: executable must be nonempty")
	}
	if options.Exec == nil {
		if cwd == "" {
			cwd = options.ShellDir
		} else if !filepath.IsAbs(cwd) && options.ShellDir != "" {
			cwd = filepath.Join(options.ShellDir, cwd)
		}
	}
	environment := make(map[string]string)
	for _, item := range options.ShellEnv {
		key, text, ok := strings.Cut(item, "=")
		if ok {
			environment[key] = text
		}
	}
	if env != nil {
		data, convertErr := api.FromValue(env, options.Limits.DataBytes)
		if convertErr != nil {
			return nil, convertErr
		}
		for key, item := range data.(map[string]any) {
			text, ok := item.(string)
			if !ok || key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(text, 0) {
				return nil, errors.New("shell: env must map valid names to strings")
			}
			environment[key] = text
		}
	}
	ctx, cancel, err := api.CallContext(host, timeout)
	if err != nil {
		return nil, err
	}
	defer cancel()
	output := &processOutput{remaining: options.Limits.OutputBytes, cancel: cancel}
	if options.Exec != nil {
		req := api.ExecRequest{
			Argv: command, Dir: cwd, Env: environment,
			Stdout: processWriter{output: output}, Stderr: processWriter{output: output, stderr: true},
		}
		return hostExec(ctx, host, options.Exec, req, output)
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = cwd
	cmd.Env = make([]string, 0, len(environment))
	for key, text := range environment {
		cmd.Env = append(cmd.Env, key+"="+text)
	}
	sort.Strings(cmd.Env)
	cmd.WaitDelay = time.Second
	configureProcess(cmd)
	cmd.Stdout = processWriter{output: output}
	cmd.Stderr = processWriter{output: output, stderr: true}
	if err := cmd.Start(); err != nil {
		return nil, errors.New("shell could not execute command")
	}
	host.MarkStarted()
	err = cmd.Wait()
	if output.exceeded {
		return nil, errors.New("shell output byte limit exceeded")
	}
	if ctx.Err() != nil {
		return nil, fmt.Errorf("shell cancelled: %w", ctx.Err())
	}
	host.MarkCompleted()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return nil, errors.New("shell could not execute command")
		}
		code = exit.ExitCode()
	}
	return shellResult(output, code), nil
}

// hostExec runs a shell call through the host. The host owns process start, so
// a start failure and a lost process look the same here: the outcome is unknown.
func hostExec(ctx context.Context, host api.Host, run api.Exec, req api.ExecRequest, output *processOutput) (starlark.Value, error) {
	host.MarkStarted()
	code, err := run(ctx, req)
	if output.exceeded {
		return nil, errors.New("shell output byte limit exceeded")
	}
	if ctx.Err() != nil {
		return nil, fmt.Errorf("shell cancelled: %w", ctx.Err())
	}
	if err != nil {
		return nil, fmt.Errorf("shell could not execute command: %w", err)
	}
	host.MarkCompleted()
	return shellResult(output, code), nil
}

func shellResult(output *processOutput, code int) starlark.Value {
	output.mu.Lock()
	defer output.mu.Unlock()
	return starlarkstruct.FromStringDict(starlark.String("shell_result"), starlark.StringDict{
		"stdout": starlark.String(output.stdout.String()), "stderr": starlark.String(output.stderr.String()), "exit_code": starlark.MakeInt(code),
	})
}
