// Package automation loads trusted Starlark hooks and invokes them with bounded
// host capabilities. Actions are data only. The caller owns persistence, scope
// checks, deduplication and applying actions outside this package.
package automation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/automation/api"
	"github.com/elee1766/jobshop/pkg/automation/command"
	automationhttp "github.com/elee1766/jobshop/pkg/automation/http"
	"github.com/elee1766/jobshop/pkg/automation/jev"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

// Limits are host ceilings. Zero values select bounded defaults.
type Limits = api.Limits

// Credential is host-owned and never made available as a script value.
type Credential = api.Credential

// JevOptions names a full Decisions endpoint, not a service base URL.
type JevOptions = api.JevOptions

// Options grants host capabilities. Shell runs as the controller OS user and is
// not an agent sandbox. HTTP permits private networks unless HTTPOrigins limits it.
type Options = api.Options

// Decision is local audit evidence, not authorization for privileged work.
type Decision = api.Decision

// Result separates host-owned metadata and effects from arbitrary callback output.
type Result struct {
	Metadata  ResultMetadata `json:"metadata"`
	Output    map[string]any `json:"output"`
	Actions   []Action       `json:"actions"`
	Logs      string         `json:"logs"`
	Decisions []Decision     `json:"decisions"`
}

// ResultMetadata identifies the loaded code and observed capability count.
// Scripts cannot supply or override these fields.
type ResultMetadata struct {
	Entry  string `json:"entry"`
	Digest string `json:"digest"`
	Calls  int    `json:"calls"`
}

// Engine is immutable after Load and safe for concurrent invocations.
type Engine struct {
	options   Options
	functions map[string]*starlark.Function
	files     []string
	digest    string
}

type invocation struct {
	ctx               context.Context
	engine            *Engine
	output            bytes.Buffer
	decisions         []Decision
	calls             int
	externalStarted   bool
	externalCompleted bool
	callSecrets       []string
}

func (inv *invocation) Context() context.Context { return inv.ctx }
func (inv *invocation) Options() api.Options     { return inv.engine.options }
func (inv *invocation) MarkStarted()             { inv.externalStarted = true }
func (inv *invocation) MarkCompleted()           { inv.externalCompleted = true }
func (inv *invocation) RecordDecision(decision api.Decision) {
	inv.decisions = append(inv.decisions, decision)
}

func defaults(o Options) (Options, error) {
	l := &o.Limits
	if l.Duration == 0 {
		l.Duration = 30 * time.Second
	}
	if l.Steps == 0 {
		l.Steps = 1_000_000
	}
	if l.OutputBytes == 0 {
		l.OutputBytes = 1 << 20
	}
	if l.SourceBytes == 0 {
		l.SourceBytes = 1 << 20
	}
	if l.Modules == 0 {
		l.Modules = 64
	}
	if l.DataBytes == 0 {
		l.DataBytes = 1 << 20
	}
	if l.Actions == 0 {
		l.Actions = 1000
	}
	if l.Calls == 0 {
		l.Calls = 100
	}
	if l.Calls < 0 || l.Duration < 0 || l.OutputBytes < 0 || l.SourceBytes < 0 || l.Modules < 0 || l.DataBytes < 0 || l.Actions < 0 {
		return o, errors.New("automation limits must be positive")
	}
	o.HTTPOrigins = slices.Clone(o.HTTPOrigins)
	o.ShellEnv = slices.Clone(o.ShellEnv)
	auth := make(map[string]Credential, len(o.Auth))
	for name, cred := range o.Auth {
		headers := make(map[string]string, len(cred.Headers))
		for key, value := range cred.Headers {
			headers[key] = value
		}
		cred.Secrets = slices.Clone(cred.Secrets)
		cred.Headers, cred.Origins = headers, slices.Clone(cred.Origins)
		for _, origin := range cred.Origins {
			if _, err := automationhttp.ParseOrigin(origin); err != nil {
				return o, fmt.Errorf("credential %q: invalid origin", name)
			}
		}
		auth[name] = cred
	}
	o.Auth = auth
	for _, origin := range o.HTTPOrigins {
		if _, err := automationhttp.ParseOrigin(origin); err != nil {
			return o, errors.New("invalid HTTP origin")
		}
	}
	if o.HTTP == nil {
		o.HTTP = &http.Client{}
	} else {
		client := *o.HTTP
		o.HTTP = &client
	}
	if o.Jev.Enabled {
		if _, err := automationhttp.RequestURL(o.Jev.Endpoint); err != nil || strings.TrimSpace(o.Jev.APIKey) == "" {
			return o, errors.New("enabled Jev requires a valid endpoint and API key")
		}
		if o.Jev.Model == "" {
			o.Jev.Model = "typesafe/jev-1.13"
		}
	}
	return o, nil
}

// Load evaluates local modules without permitting capability calls. Paths are
// relative to root and cannot escape it, including through symlinks.
func Load(ctx context.Context, root string, entries []string, options Options) (*Engine, error) {
	options, err := defaults(options)
	if err != nil {
		return nil, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	fs, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("hook root: %w", err)
	}
	defer func() { _ = fs.Close() }()
	e := &Engine{options: options, functions: make(map[string]*starlark.Function)}
	ctx, cancel := context.WithTimeout(ctx, options.Limits.Duration)
	defer cancel()
	thread := e.thread(ctx, nil)
	stop := context.AfterFunc(ctx, func() { thread.Cancel("hook load cancelled") })
	defer stop()
	cache := make(map[string]starlark.StringDict)
	loading := make(map[string]bool)
	current := "."
	total := 0
	sources := make(map[string][]byte)
	load := func(name string) (starlark.StringDict, error) {
		if globals, ok := cache[name]; ok {
			return globals, nil
		}
		if loading[name] {
			return nil, fmt.Errorf("cyclic module load: %s", name)
		}
		if len(loading) >= options.Limits.Modules {
			return nil, errors.New("hook module limit exceeded")
		}
		loading[name] = true
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		info, statErr := fs.Stat(name)
		if statErr != nil {
			return nil, statErr
		}
		if !info.Mode().IsRegular() {
			return nil, errors.New("hook module must be a regular file")
		}
		f, openErr := fs.Open(name)
		if openErr != nil {
			return nil, fmt.Errorf("hook module %s: %w", name, openErr)
		}
		info, statErr = f.Stat()
		if statErr != nil {
			_ = f.Close()
			return nil, statErr
		}
		if !info.Mode().IsRegular() {
			_ = f.Close()
			return nil, errors.New("hook module must be a regular file")
		}
		data, readErr := io.ReadAll(io.LimitReader(f, int64(options.Limits.SourceBytes-total)+1))
		_ = f.Close()
		if readErr != nil {
			return nil, readErr
		}
		total += len(data)
		if total > options.Limits.SourceBytes {
			return nil, errors.New("hook source byte limit exceeded")
		}
		sources[name] = data
		previous := current
		current = filepath.Dir(name)
		globals, execErr := starlark.ExecFileOptions(&syntax.FileOptions{}, thread, filepath.Join(root, name), data, predeclared())
		current = previous
		if execErr != nil {
			return nil, (&invocation{engine: e}).safeError(execErr)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		globals.Freeze()
		cache[name] = globals
		loading[name] = false
		e.files = append(e.files, filepath.Join(root, name))
		return globals, nil
	}
	thread.Load = func(_ *starlark.Thread, module string) (starlark.StringDict, error) {
		name, pathErr := localPath(current, module)
		if pathErr != nil {
			return nil, pathErr
		}
		return load(name)
	}
	for _, entry := range entries {
		file, function, splitErr := splitEntry(entry)
		if splitErr != nil {
			return nil, splitErr
		}
		globals, loadErr := load(file)
		if loadErr != nil {
			return nil, fmt.Errorf("hook %s: %w", entry, loadErr)
		}
		fn, ok := globals[function].(*starlark.Function)
		if !ok {
			return nil, fmt.Errorf("hook %s: exported function not found", entry)
		}
		e.functions[entry] = fn
	}
	slices.Sort(e.files)
	hasher := sha256.New()
	for _, name := range slices.Sorted(maps.Keys(sources)) {
		for _, data := range [][]byte{[]byte(name), sources[name]} {
			var size [8]byte
			binary.BigEndian.PutUint64(size[:], uint64(len(data)))
			_, _ = hasher.Write(size[:])
			_, _ = hasher.Write(data)
		}
	}
	e.digest = hex.EncodeToString(hasher.Sum(nil))
	return e, nil
}

// Files returns loaded absolute paths for reload watching.
func (e *Engine) Files() []string { return slices.Clone(e.files) }

// Digest identifies exactly the loaded relative module names and source bytes.
func (e *Engine) Digest() string { return e.digest }

// CheckArity validates a host's positional callback contract without calling it.
func (e *Engine) CheckArity(entry string, positional int) error {
	fn := e.functions[entry]
	if fn == nil {
		return fmt.Errorf("hook %s: entry is not loaded", entry)
	}
	if positional < 0 {
		return errors.New("callback positional argument count must be nonnegative")
	}
	ordinary := fn.NumParams()
	if fn.HasVarargs() {
		ordinary--
	}
	if fn.HasKwargs() {
		ordinary--
	}
	positionals := ordinary - fn.NumKwonlyParams()
	for i := 0; i < ordinary; i++ {
		if fn.ParamDefault(i) != nil {
			continue
		}
		if i >= positionals || i >= positional {
			name, _ := fn.Param(i)
			return fmt.Errorf("hook %s: %d positional arguments do not supply required parameter %s", entry, positional, name)
		}
	}
	if positional > positionals && !fn.HasVarargs() {
		return fmt.Errorf("hook %s: accepts at most %d positional arguments, got %d", entry, positionals, positional)
	}
	return nil
}

// Invoke calls a previously validated hook with maps as dictionaries and Go
// structs as attribute-accessible values. Every call has its own thread locals.
func (e *Engine) Invoke(ctx context.Context, entry string, args ...any) (result Result, err error) {
	var inv *invocation
	defer func() {
		result.Metadata = ResultMetadata{Entry: entry, Digest: e.digest}
		if inv != nil {
			result.Metadata.Calls = inv.calls
		}
		if result.Output == nil {
			result.Output = map[string]any{}
		}
		if result.Actions == nil {
			result.Actions = []Action{}
		}
		if result.Decisions == nil {
			result.Decisions = []Decision{}
		}
	}()
	fn, ok := e.functions[entry]
	if !ok {
		return Result{}, fmt.Errorf("hook %s is not loaded", entry)
	}
	ctx, cancel := context.WithTimeout(ctx, e.options.Limits.Duration)
	defer cancel()
	inv = &invocation{ctx: ctx, engine: e}
	thread := e.thread(ctx, inv)
	stop := context.AfterFunc(ctx, func() { thread.Cancel("hook invocation cancelled") })
	defer stop()
	values, err := api.Arguments(args, e.options.Limits.DataBytes)
	if err != nil {
		return Result{}, fmt.Errorf("hook %s %w", entry, err)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	value, err := starlark.Call(thread, fn, values, nil)
	output := inv.redact(inv.output.String()).(string)
	if len(output) > e.options.Limits.OutputBytes {
		return Result{Decisions: inv.decisions}, fmt.Errorf("hook %s: redacted output byte limit exceeded", entry)
	}
	if err != nil {
		return Result{Logs: output, Decisions: inv.decisions}, fmt.Errorf("hook %s: %w", entry, inv.safeError(err))
	}
	result, err = e.result(value)
	result.Logs = output
	result.Decisions = inv.decisions
	if err != nil {
		return Result{Logs: output, Decisions: inv.decisions}, fmt.Errorf("hook %s result: %w", entry, inv.safeError(err))
	}
	return result, nil
}

func (e *Engine) thread(ctx context.Context, inv *invocation) *starlark.Thread {
	t := &starlark.Thread{Name: "automation"}
	t.SetMaxExecutionSteps(e.options.Limits.Steps)
	t.SetLocal("invocation", inv)
	t.Print = func(t *starlark.Thread, msg string) {
		if inv == nil {
			t.Cancel("print is unavailable during module loading")
			return
		}
		if inv.output.Len()+len(msg)+1 > e.options.Limits.OutputBytes {
			t.Cancel("hook output byte limit exceeded")
			return
		}
		inv.output.WriteString(msg)
		inv.output.WriteByte('\n')
	}
	if ctx.Err() != nil {
		t.Cancel("hook context cancelled")
	}
	return t
}

func localPath(base, name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || strings.Contains(name, "\\") {
		return "", errors.New("hook module must be a local relative path")
	}
	path := filepath.Clean(filepath.Join(base, name))
	if path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) || filepath.Ext(path) != ".star" {
		return "", errors.New("hook module must remain under root and end in .star")
	}
	return path, nil
}

func splitEntry(entry string) (string, string, error) {
	file, fn, ok := strings.Cut(entry, ":")
	if !ok || fn == "" || strings.HasPrefix(fn, "_") || strings.ContainsAny(fn, ":/\\") {
		return "", "", fmt.Errorf("hook %q must be file.star:exported_function", entry)
	}
	file, err := localPath(".", file)
	return file, fn, err
}

func diagnostic(err error) error {
	var eval *starlark.EvalError
	if errors.As(err, &eval) {
		return fmt.Errorf("%s: %w", eval.Backtrace(), err)
	}
	return err
}

func active(t *starlark.Thread) (*invocation, error) {
	inv, _ := t.Local("invocation").(*invocation)
	if inv == nil {
		return nil, errors.New("capability calls are unavailable during module loading")
	}
	if err := inv.ctx.Err(); err != nil {
		return nil, err
	}
	return inv, nil
}

func predeclared() starlark.StringDict {
	return starlark.StringDict{
		"result":     starlark.NewBuiltin("result", constructResult),
		"start_run":  starlark.NewBuiltin("start_run", constructAction),
		"start_step": starlark.NewBuiltin("start_step", constructAction),
		"finish_run": starlark.NewBuiltin("finish_run", constructAction),
		"http": starlarkstruct.FromStringDict(starlark.String("http"), starlark.StringDict{
			"get":  capability("http.get", automationhttp.Call),
			"post": capability("http.post", automationhttp.Call),
		}),
		"shell": capability("shell", command.Call),
		"jev":   capability("jev", jev.Call),
	}
}
