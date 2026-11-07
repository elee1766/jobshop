package automation

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Action is a validated request, not an effect. References must additionally be
// checked against the invoking pack by the controller before accepting a batch.
type Action struct {
	Kind         string         `json:"kind"`
	Workflow     string         `json:"workflow,omitempty"`
	Key          string         `json:"key,omitempty"`
	ID           string         `json:"id,omitempty"`
	Pool         string         `json:"pool,omitempty"`
	Prompt       string         `json:"prompt,omitempty"`
	PromptFile   string         `json:"prompt_file,omitempty"`
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
	a.Needs = slices.Clone(a.Needs)
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
	var err error
	switch a.Kind {
	case "start_run":
		err = starlark.UnpackArgs(a.Kind, args, kwargs, "workflow", &a.Workflow, "key", &a.Key, "inputs?", &inputs, "title?", &a.Title, "reason?", &a.Reason)
	case "start_step":
		err = starlark.UnpackArgs(a.Kind, args, kwargs, "id", &a.ID, "pool", &a.Pool, "prompt?", &a.Prompt, "prompt_file?", &a.PromptFile, "inputs?", &inputs, "title?", &a.Title, "retries?", &a.Retries, "retry_pool?", &a.RetryPool, "needs?", &needs, "output_schema?", &outputSchema, "reason?", &a.Reason)
	case "finish_run":
		err = starlark.UnpackArgs(a.Kind, args, kwargs, "outcome", &a.Outcome, "output?", &output, "reason?", &a.Reason)
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
	switch a.Kind {
	case "start_run":
		required["workflow"], required["key"] = a.Workflow, a.Key
	case "start_step":
		required["id"], required["pool"] = a.ID, a.Pool
		if (a.Prompt == "") == (a.PromptFile == "") {
			return nil, errors.New("start_step: exactly one of prompt or prompt_file is required")
		}
		if a.Retries < 0 || a.Retries > 100 {
			return nil, errors.New("start_step: retries must be between 0 and 100")
		}
	case "finish_run":
		required["outcome"] = a.Outcome
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%s: %s must be nonempty", a.Kind, name)
		}
	}
	limit := 1 << 20
	if inv, _ := t.Local("invocation").(*invocation); inv != nil {
		limit = inv.engine.options.Limits.DataBytes
	}
	for _, pair := range []struct {
		dict   *starlark.Dict
		target *map[string]any
	}{{inputs, &a.Inputs}, {output, &a.Output}, {outputSchema, &a.OutputSchema}} {
		if pair.dict == nil {
			continue
		}
		value, convertErr := fromValue(pair.dict, limit)
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
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, errors.New("action data byte limit exceeded")
	}
	value, err := toValue(a, limit)
	if err != nil {
		return nil, err
	}
	result := &actionValue{Struct: value.(*starlarkstruct.Struct), action: a}
	result.Freeze()
	return result, nil
}

func (e *Engine) result(value starlark.Value) (Result, error) {
	var actions []Action
	switch v := value.(type) {
	case *actionValue:
		actions = []Action{cloneAction(v.action)}
	case *starlark.List:
		for i := 0; i < v.Len(); i++ {
			if a, ok := v.Index(i).(*actionValue); ok {
				actions = append(actions, cloneAction(a.action))
			}
		}
		if len(actions) > 0 && len(actions) != v.Len() {
			return Result{}, errors.New("action batch may contain only action constructors")
		}
	case starlark.Tuple:
		for _, item := range v {
			if a, ok := item.(*actionValue); ok {
				actions = append(actions, cloneAction(a.action))
			}
		}
		if len(actions) > 0 && len(actions) != len(v) {
			return Result{}, errors.New("action batch may contain only action constructors")
		}
	}
	if len(actions) > e.options.Limits.Actions {
		return Result{}, errors.New("action count limit exceeded")
	}
	if actions != nil {
		data, err := json.Marshal(actions)
		if err != nil {
			return Result{}, err
		}
		if len(data) > e.options.Limits.DataBytes {
			return Result{}, errors.New("action batch data byte limit exceeded")
		}
		generic, err := decodeJSON(data)
		if err != nil {
			return Result{}, err
		}
		return Result{Value: generic, Actions: actions}, nil
	}
	data, err := fromValue(value, e.options.Limits.DataBytes)
	return Result{Value: data}, err
}
