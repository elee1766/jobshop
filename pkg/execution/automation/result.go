package automation

import (
	"errors"

	"github.com/elee1766/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type resultValue struct {
	*starlarkstruct.Struct
	actions starlark.Value
	output  map[string]any
	outcome string
	notYet  bool
}

func constructResult(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 0 {
		return nil, errors.New("result: use named arguments")
	}
	var actions starlark.Value = starlark.Tuple{}
	output := starlark.NewDict(0)
	var outcome string
	var notYet bool
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "actions?", &actions, "output?", &output, "outcome?", &outcome, "not_yet?", &notYet); err != nil {
		return nil, err
	}
	if _, ok, err := actionBatch(actions); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("result: actions must be an action or action list")
	}
	if notYet && outcome != "" {
		return nil, errors.New("result: not_yet excludes outcome")
	}
	limit := 1 << 20
	if inv, _ := t.Local("invocation").(*invocation); inv != nil {
		limit = inv.engine.options.Limits.DataBytes
	}
	data, err := api.FromValue(output, limit)
	if err != nil {
		return nil, err
	}
	value := &resultValue{
		Struct:  starlarkstruct.FromStringDict(starlark.String("result"), starlark.StringDict{"actions": actions, "output": output, "outcome": starlark.String(outcome), "not_yet": starlark.Bool(notYet)}),
		actions: actions,
		output:  data.(map[string]any),
		outcome: outcome,
		notYet:  notYet,
	}
	value.Freeze()
	return value, nil
}
