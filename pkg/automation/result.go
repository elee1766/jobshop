package automation

import (
	"errors"

	"github.com/elee1766/jobshop/pkg/automation/api"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type resultValue struct {
	*starlarkstruct.Struct
	actions starlark.Value
	output  map[string]any
	outcome string
}

func constructResult(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 0 {
		return nil, errors.New("result: use named arguments")
	}
	var actions starlark.Value = starlark.Tuple{}
	output := starlark.NewDict(0)
	var outcome string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "actions?", &actions, "output?", &output, "outcome?", &outcome); err != nil {
		return nil, err
	}
	if _, ok, err := actionBatch(actions); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("result: actions must be an action or action list")
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
		Struct:  starlarkstruct.FromStringDict(starlark.String("result"), starlark.StringDict{"actions": actions, "output": output, "outcome": starlark.String(outcome)}),
		actions: actions,
		output:  data.(map[string]any),
		outcome: outcome,
	}
	value.Freeze()
	return value, nil
}
