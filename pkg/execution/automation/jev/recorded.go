package jev

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/gfx-labs/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
)

// Restore validates a recorded decision against the original offered choices.
// The engine bounds the encoded JSON before dispatching it here.
func Restore(host api.Host, data []byte, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := host.Options()
	if !options.Jev.Enabled {
		return nil, errors.New("jev capability is disabled by host")
	}
	invalid := errors.New("recorded capability result does not match its capability contract")
	var recorded struct {
		Choice     string   `json:"choice"`
		Reason     string   `json:"reason"`
		Defaulted  *bool    `json:"defaulted"`
		Confidence *float64 `json:"confidence"`
	}
	if err := json.Unmarshal(data, &recorded); err != nil || recorded.Choice == "" || recorded.Reason == "" || recorded.Defaulted == nil {
		return nil, invalid
	}
	choices, _ := api.Argument(args, kwargs, "options", 1).(*starlark.Dict)
	if choices == nil {
		return nil, invalid
	}
	if _, found, err := choices.Get(starlark.String(recorded.Choice)); err != nil || !found {
		return nil, invalid
	}
	if *recorded.Defaulted {
		defaultChoice, _ := starlark.AsString(api.Argument(args, kwargs, "default", 3))
		if defaultChoice != recorded.Choice {
			return nil, invalid
		}
	}
	decision := api.Decision{Choice: recorded.Choice, Reason: recorded.Reason, Defaulted: *recorded.Defaulted, Confidence: recorded.Confidence}
	if confidence := recorded.Confidence; confidence != nil {
		f := *confidence
		if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 || f > 1 {
			return nil, invalid
		}
	}
	host.RecordDecision(decision)
	return api.ToValue(decision, options.Limits.DataBytes)
}
