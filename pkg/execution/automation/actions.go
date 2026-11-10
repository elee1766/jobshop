package automation

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Action is a validated request, not an effect. References must additionally be
// checked against the invoking pack by the controller before accepting a batch.
type Action struct {
	Kind     string `json:"kind"`
	Workflow string `json:"workflow,omitempty"`
	Key      string `json:"key,omitempty"`
	ID       string `json:"id,omitempty"`
	Pool     string `json:"pool,omitempty"`
	// Type is the start_step kind as written: "" or "agent" (prompt,
	// prompt_file or script) or "command".
	Type       string `json:"type,omitempty"`
	Prompt     string `json:"prompt,omitempty"`
	PromptFile string `json:"prompt_file,omitempty"`
	// Script is a step function, file.star:function in the invoking module's
	// directory. The controller runs it in-process as a durable step.
	Script string `json:"script,omitempty"`
	// Run is a command step's command: a string, which runs with sh -c, or
	// a []string argv. Same shape as a workflow command step's run.
	Run any `json:"run,omitempty"`
	// OnRestart is a command step's restart policy: "retry" or "fail".
	OnRestart    string         `json:"on_restart,omitempty"`
	Worktree     bool           `json:"worktree,omitempty"`
	Timeout      string         `json:"timeout,omitempty"`
	Title        string         `json:"title,omitempty"`
	Outcome      string         `json:"outcome,omitempty"`
	Reason       string         `json:"reason,omitempty"`
	Inputs       map[string]any `json:"inputs,omitempty"`
	Output       map[string]any `json:"output,omitempty"`
	Retries      int            `json:"retries,omitempty"`
	RetriesSet   bool           `json:"retries_set,omitempty"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
	RetryPool    string         `json:"retry_pool,omitempty"`
	Needs        []string       `json:"needs,omitempty"`
	// NeedsResources are resource keys the step holds while in progress,
	// like a workflow step's needs_resources (#400). Placement replaces
	// pool and needs_resources with weighted pool options (#409).
	NeedsResources []string         `json:"needs_resources,omitempty"`
	Placement      *ActionPlacement `json:"placement,omitempty"`
	// RetryAfter (Go duration) is the wait after a script step returns
	// result(not_yet=True) or a command step exits 75.
	RetryAfter string `json:"retry_after,omitempty"`
	// MaxReviewRounds caps the request_changes closes of the step (nil:
	// the default). OnRequestChanges is the step such a close reopens, one
	// of Needs.
	MaxReviewRounds  *int   `json:"max_review_rounds,omitempty"`
	OnRequestChanges string `json:"on_request_changes,omitempty"`
	// create_ticket and comment: Body is the ticket body or comment text,
	// Ticket the ticket commented on. Labels, Priority and Parent are for
	// create_ticket.
	Body     string   `json:"body,omitempty"`
	Ticket   string   `json:"ticket,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Priority *int     `json:"priority,omitempty"`
	Parent   string   `json:"parent,omitempty"`
}

type actionValue struct {
	*starlarkstruct.Struct
	action Action
}

func cloneAction(a Action) Action {
	if a.Inputs != nil {
		a.Inputs = cloneData(a.Inputs).(map[string]any)
	}
	if a.Output != nil {
		a.Output = cloneData(a.Output).(map[string]any)
	}
	if a.OutputSchema != nil {
		a.OutputSchema = cloneData(a.OutputSchema).(map[string]any)
	}
	if run, ok := a.Run.([]string); ok {
		a.Run = slices.Clone(run)
	}
	a.Needs = slices.Clone(a.Needs)
	a.Labels = slices.Clone(a.Labels)
	a.NeedsResources = slices.Clone(a.NeedsResources)
	a.Placement = clonePlacement(a.Placement)
	if a.MaxReviewRounds != nil {
		n := *a.MaxReviewRounds
		a.MaxReviewRounds = &n
	}
	if a.Priority != nil {
		p := *a.Priority
		a.Priority = &p
	}
	return a
}

func cloneData(value any) any {
	switch value := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for key, item := range value {
			out[key] = cloneData(item)
		}
		return out
	case []any:
		out := make([]any, len(value))
		for i, item := range value {
			out[i] = cloneData(item)
		}
		return out
	default:
		return value
	}
}

func constructAction(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("%s: use named arguments", fn.Name())
	}
	a := Action{Kind: fn.Name()}
	var inputs, output, outputSchema *starlark.Dict
	var needs *starlark.List
	var labels *starlark.List
	var needsResources *starlark.List
	var placement *starlark.Dict
	var script, run, maxReviewRounds starlark.Value
	var err error
	switch a.Kind {
	case "start_run":
		err = starlark.UnpackArgs(a.Kind, args, kwargs, "workflow", &a.Workflow, "key", &a.Key, "inputs?", &inputs, "title?", &a.Title, "reason?", &a.Reason)
	case "start_step":
		err = starlark.UnpackArgs(a.Kind, args, kwargs, "id", &a.ID, "type?", &a.Type, "pool?", &a.Pool, "run?", &run, "on_restart?", &a.OnRestart, "prompt?", &a.Prompt, "prompt_file?", &a.PromptFile, "script?", &script, "worktree?", &a.Worktree, "timeout?", &a.Timeout, "inputs?", &inputs, "title?", &a.Title, "retries?", &a.Retries, "retry_pool?", &a.RetryPool, "needs?", &needs, "output_schema?", &outputSchema, "needs_resources?", &needsResources, "retry_after?", &a.RetryAfter, "placement?", &placement, "max_review_rounds?", &maxReviewRounds, "on_request_changes?", &a.OnRequestChanges, "reason?", &a.Reason)
	case "finish_run":
		err = starlark.UnpackArgs(a.Kind, args, kwargs, "outcome", &a.Outcome, "output?", &output, "reason?", &a.Reason)
	case "create_ticket":
		priority := -1
		err = starlark.UnpackArgs(a.Kind, args, kwargs, "title", &a.Title, "key", &a.Key, "pool?", &a.Pool, "body?", &a.Body, "labels?", &labels, "priority?", &priority, "parent?", &a.Parent, "reason?", &a.Reason)
		if priority != -1 {
			a.Priority = &priority
		}
	case "comment":
		err = starlark.UnpackArgs(a.Kind, args, kwargs, "ticket", &a.Ticket, "body", &a.Body, "reason?", &a.Reason)
	}
	if err != nil {
		return nil, err
	}
	if a.Kind == "start_step" {
		for _, pair := range kwargs {
			if name, _ := starlark.AsString(pair[0]); name == "retries" {
				a.RetriesSet = true
			}
		}
	}
	required := map[string]string{}
	limit := 1 << 20
	if inv, _ := t.Local("invocation").(*invocation); inv != nil {
		limit = inv.engine.options.Limits.DataBytes
	}
	switch a.Kind {
	case "start_run":
		required["workflow"], required["key"] = a.Workflow, a.Key
	case "start_step":
		required["id"] = a.ID
		if placement == nil {
			required["pool"] = a.Pool
		}
		if err := checkStartStep(&a, t, script, run, outputSchema != nil); err != nil {
			return nil, err
		}
		if a.Retries < 0 || a.Retries > 100 {
			return nil, errors.New("start_step: retries must be between 0 and 100")
		}
		if err := a.checkStepFields(placement, needsResources, maxReviewRounds, limit); err != nil {
			return nil, err
		}
	case "finish_run":
		required["outcome"] = a.Outcome
	case "create_ticket":
		required["title"], required["key"] = a.Title, a.Key
		if a.Priority != nil && (*a.Priority < 0 || *a.Priority > 4) {
			return nil, errors.New("create_ticket: priority must be 0 to 4")
		}
		if labels != nil {
			for i := 0; i < labels.Len(); i++ {
				label, ok := starlark.AsString(labels.Index(i))
				if !ok || strings.TrimSpace(label) == "" {
					return nil, errors.New("create_ticket: labels must be nonempty strings")
				}
				a.Labels = append(a.Labels, label)
			}
		}
	case "comment":
		required["ticket"], required["body"] = a.Ticket, a.Body
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%s: %s must be nonempty", a.Kind, name)
		}
	}
	for _, pair := range []struct {
		dict   *starlark.Dict
		target *map[string]any
	}{{inputs, &a.Inputs}, {output, &a.Output}, {outputSchema, &a.OutputSchema}} {
		if pair.dict == nil {
			continue
		}
		value, convertErr := api.FromValue(pair.dict, limit)
		if convertErr != nil {
			return nil, fmt.Errorf("%s: %w", a.Kind, convertErr)
		}
		*pair.target = value.(map[string]any)
	}
	if needs != nil {
		for i := 0; i < needs.Len(); i++ {
			dependency, ok := starlark.AsString(needs.Index(i))
			if !ok || strings.TrimSpace(dependency) == "" || dependency == a.ID {
				return nil, errors.New("start_step: needs must contain nonempty other step ids")
			}
			a.Needs = append(a.Needs, dependency)
		}
	}
	if a.OnRequestChanges != "" && !slices.Contains(a.Needs, a.OnRequestChanges) {
		// A direct need: the reviewer reopens blocked behind it.
		return nil, fmt.Errorf("start_step: on_request_changes %q must be a step this agent step needs directly", a.OnRequestChanges)
	}
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, errors.New("action data byte limit exceeded")
	}
	value, err := api.ToValue(a, limit)
	if err != nil {
		return nil, err
	}
	result := &actionValue{Struct: value.(*starlarkstruct.Struct), action: a}
	result.Freeze()
	return result, nil
}

func (e *Engine) result(value starlark.Value) (Result, error) {
	var result Result
	if v, ok := value.(*resultValue); ok {
		result.Output = cloneData(v.output).(map[string]any)
		result.Outcome = v.outcome
		result.NotYet = v.notYet
		value = v.actions
	}
	actions, batch, err := actionBatch(value)
	if err != nil {
		return Result{}, err
	}
	if !batch {
		data, err := api.FromValue(value, e.options.Limits.DataBytes)
		if err != nil {
			return Result{}, err
		}
		var ok bool
		result.Output, ok = data.(map[string]any)
		if !ok {
			return Result{}, errors.New("callback output must be an object, an action, or an action list")
		}
	}
	if len(actions) > e.options.Limits.Actions {
		return Result{}, errors.New("action count limit exceeded")
	}
	result.Actions = actions
	data, err := json.Marshal(result)
	if err != nil {
		return Result{}, err
	}
	if len(data) > e.options.Limits.DataBytes {
		return Result{}, errors.New("callback result data byte limit exceeded")
	}
	return result, nil
}

func actionBatch(value starlark.Value) ([]Action, bool, error) {
	actions := []Action{}
	switch v := value.(type) {
	case *actionValue:
		actions = []Action{cloneAction(v.action)}
	case *starlark.List:
		for i := 0; i < v.Len(); i++ {
			a, ok := v.Index(i).(*actionValue)
			if !ok {
				return nil, false, errors.New("action batch may contain only action constructors")
			}
			actions = append(actions, cloneAction(a.action))
		}
	case starlark.Tuple:
		for _, item := range v {
			a, ok := item.(*actionValue)
			if !ok {
				return nil, false, errors.New("action batch may contain only action constructors")
			}
			actions = append(actions, cloneAction(a.action))
		}
	default:
		return nil, false, nil
	}
	return actions, true, nil
}

// checkStartStep validates the start_step fields that depend on its type and
// fills Script and Run. Fields that do not belong to the type are rejected
// rather than ignored.
func checkStartStep(a *Action, t *starlark.Thread, script, run starlark.Value, hasSchema bool) error {
	hasScript := script != nil && script != starlark.None
	hasRun := run != nil && run != starlark.None
	switch a.Type {
	case "command":
		if hasScript || a.Prompt != "" || a.PromptFile != "" || hasSchema {
			return errors.New("start_step: command excludes prompt, prompt_file, script and output_schema")
		}
		if a.RetryPool != "" {
			return errors.New("start_step: command excludes retry_pool")
		}
		if !hasRun {
			return errors.New("start_step: command needs run")
		}
		var err error
		if a.Run, err = commandRun(run); err != nil {
			return fmt.Errorf("start_step: %w", err)
		}
		switch a.OnRestart {
		case "", "retry", "fail":
		default:
			return fmt.Errorf("start_step: on_restart %q must be retry or fail", a.OnRestart)
		}
	case "", "agent":
		if hasRun || a.OnRestart != "" {
			return errors.New("start_step: run and on_restart need type=\"command\"")
		}
		if hasScript {
			engine, _ := t.Local("engine").(*Engine)
			if engine == nil {
				return errors.New("start_step: script is only available in a callback")
			}
			var err error
			if a.Script, err = engine.stepEntry(script); err != nil {
				return fmt.Errorf("start_step: %w", err)
			}
		}
		if a.Script != "" {
			if a.Prompt != "" || a.PromptFile != "" {
				return errors.New("start_step: script excludes prompt and prompt_file")
			}
		} else if (a.Prompt == "") == (a.PromptFile == "") {
			return errors.New("start_step: exactly one of script, prompt or prompt_file is required")
		}
		if a.Worktree && a.Script == "" {
			return errors.New("start_step: worktree needs script or type=\"command\"")
		}
		if a.Timeout != "" && a.Script == "" {
			return errors.New("start_step: timeout needs script or type=\"command\"")
		}
	default:
		return fmt.Errorf("start_step: type %q must be agent or command", a.Type)
	}
	if a.Timeout != "" {
		if d, err := time.ParseDuration(a.Timeout); err != nil || d <= 0 {
			return errors.New("start_step: timeout must be a positive duration")
		}
	}
	return nil
}

// errNULRun: a NUL byte cannot be passed to exec, so reject it here instead of
// failing when the step starts.
var errNULRun = errors.New("run must not contain NUL bytes")

// commandRun normalizes a command step's run to a string or a []string argv.
// Like a workflow step, the string is a shell command and the list is exec'd
// directly. Only the executable must be nonblank, later arguments may be empty.
func commandRun(v starlark.Value) (any, error) {
	if s, ok := starlark.AsString(v); ok {
		if strings.TrimSpace(s) == "" {
			return nil, errors.New("run must be nonempty")
		}
		if strings.ContainsRune(s, 0) {
			return nil, errNULRun
		}
		return s, nil
	}
	var items starlark.Indexable
	switch v := v.(type) {
	case *starlark.List:
		items = v
	case starlark.Tuple:
		items = v
	default:
		return nil, errors.New("run must be a string or a list of strings")
	}
	if items.Len() == 0 {
		return nil, errors.New("run must be nonempty")
	}
	argv := make([]string, items.Len())
	for i := range argv {
		s, ok := starlark.AsString(items.Index(i))
		if !ok {
			return nil, errors.New("run must be a string or a list of strings")
		}
		if strings.ContainsRune(s, 0) {
			return nil, errNULRun
		}
		argv[i] = s
	}
	if strings.TrimSpace(argv[0]) == "" {
		return nil, errors.New("run executable must be nonblank")
	}
	return argv, nil
}
