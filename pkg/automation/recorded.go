package automation

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// restore accepts only completed, unredacted results chosen by the host. The
// host must match invocation/source identity, capability, index and InputDigest,
// and return an error for unknown outcomes. This package stores no call history.
func (inv *invocation) restore(name string, cached any, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := inv.engine.options
	limit := options.Limits.DataBytes
	if name == "http.get" || name == "http.post" {
		// JSON can escape each raw body byte to six bytes. Metadata has its own bound.
		maxInt := int(^uint(0) >> 1)
		if options.Limits.OutputBytes > (maxInt-limit)/6 {
			limit = maxInt
		} else {
			limit += 6 * options.Limits.OutputBytes
		}
	}
	encoded, err := json.Marshal(cached)
	if err != nil || len(encoded) > limit {
		return nil, errors.New("recorded capability result byte limit exceeded")
	}
	value, err := decodeJSON(encoded)
	if err != nil {
		return nil, errors.New("invalid recorded capability result JSON")
	}
	fields, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("recorded capability result must be an object")
	}
	invalid := errors.New("recorded capability result does not match its capability contract")
	switch name {
	case "http.get", "http.post":
		if !options.AllowHTTP {
			return nil, errors.New("HTTP capability is disabled by host")
		}
		status, ok := recordedInt(fields["status_code"])
		if !ok || status < 100 || status > 999 {
			return nil, invalid
		}
		body, ok := fields["body"].(string)
		if !ok || len(body) > options.Limits.OutputBytes {
			return nil, invalid
		}
		if text, exists := fields["text"]; exists {
			if text != body {
				return nil, invalid
			}
		}
		headers, ok := fields["headers"].(map[string]any)
		if !ok {
			return nil, invalid
		}
		for _, value := range headers {
			if _, ok := value.(string); !ok {
				return nil, invalid
			}
		}
		metadata := map[string]any{"status_code": fields["status_code"], "headers": headers}
		if _, err := toValue(metadata, options.Limits.DataBytes); err != nil {
			return nil, invalid
		}
		encodedMetadata, err := json.Marshal(metadata)
		if err != nil || len(encodedMetadata) > options.Limits.DataBytes {
			return nil, invalid
		}
		return responseValue(status, []byte(body), headers, options)
	case "shell":
		if !options.AllowShell {
			return nil, errors.New("shell capability is disabled by host")
		}
		code, ok := recordedInt(fields["exit_code"])
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
	case "jev":
		if !options.Jev.Enabled {
			return nil, errors.New("jev capability is disabled by host")
		}
		choice, ok := fields["choice"].(string)
		if !ok || choice == "" {
			return nil, invalid
		}
		choices, _ := callArgument(args, kwargs, "options", 1).(*starlark.Dict)
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
			defaultChoice, _ := starlark.AsString(callArgument(args, kwargs, "default", 3))
			if defaultChoice != choice {
				return nil, invalid
			}
		}
		decision := Decision{Choice: choice, Reason: reason, Defaulted: fallback}
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
		inv.decisions = append(inv.decisions, decision)
		return toValue(decision, options.Limits.DataBytes)
	default:
		return nil, errors.New("unknown recorded capability")
	}
}

func callArgument(args starlark.Tuple, kwargs []starlark.Tuple, name string, index int) starlark.Value {
	for _, pair := range kwargs {
		if key, _ := starlark.AsString(pair[0]); key == name {
			return pair[1]
		}
	}
	if index < len(args) {
		return args[index]
	}
	return starlark.None
}

func recordedInt(value any) (int, bool) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	integer, err := strconv.Atoi(string(number))
	return integer, err == nil
}
