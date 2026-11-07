package jev

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/elee1766/jobshop/pkg/automation/api"
	"go.starlark.net/starlark"
)

// Restore validates a recorded decision against the original offered choices.
// Fields must come from bounded JSON decoded with api.DecodeJSON.
func Restore(host api.Host, fields map[string]any, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := host.Options()
	if !options.Jev.Enabled {
		return nil, errors.New("jev capability is disabled by host")
	}
	invalid := errors.New("recorded capability result does not match its capability contract")
	choice, ok := fields["choice"].(string)
	if !ok || choice == "" {
		return nil, invalid
	}
	choices, _ := api.Argument(args, kwargs, "options", 1).(*starlark.Dict)
	if choices == nil {
		return nil, invalid
	}
	if _, found, err := choices.Get(starlark.String(choice)); err != nil || !found {
		return nil, invalid
	}
	reason, ok := fields["reason"].(string)
	if !ok || reason == "" {
		return nil, invalid
	}
	fallback, ok := fields["defaulted"].(bool)
	if !ok {
		return nil, invalid
	}
	if fallback {
		defaultChoice, _ := starlark.AsString(api.Argument(args, kwargs, "default", 3))
		if defaultChoice != choice {
			return nil, invalid
		}
	}
	decision := api.Decision{Choice: choice, Reason: reason, Defaulted: fallback}
	if confidence := fields["confidence"]; confidence != nil {
		number, ok := confidence.(json.Number)
		if !ok {
			return nil, invalid
		}
		f, err := number.Float64()
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f > 1 {
			return nil, invalid
		}
		decision.Confidence = &f
	}
	host.RecordDecision(decision)
	return api.ToValue(decision, options.Limits.DataBytes)
}
