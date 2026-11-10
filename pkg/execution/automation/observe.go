package automation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Observation records an ordered capability attempt, not proof of an external
// effect. Index is invocation-local. An unknown outcome must be reconciled before
// retrying. Inputs are redacted, while InputDigest identifies their original data.
type Observation = api.Observation

func capability(name string, call api.Call) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		inv, err := active(t)
		if err != nil {
			return nil, err
		}
		inv.calls++
		if inv.calls > inv.engine.options.Limits.Calls {
			return nil, errors.New("hook capability call limit exceeded")
		}
		inv.externalStarted, inv.externalCompleted = false, false
		inv.captureHeaders(name, args, kwargs)
		options := inv.engine.options
		invoke := func() (starlark.Value, error) {
			// An observer may cancel the invocation before external work starts.
			if err := inv.ctx.Err(); err != nil {
				return nil, err
			}
			return call(inv, fn, args, kwargs)
		}
		if options.Observe == nil && options.Replay == nil {
			return invoke()
		}
		inputs, err := api.FromValue(args, options.Limits.DataBytes)
		if err != nil {
			return nil, err
		}
		named := starlark.NewDict(len(kwargs))
		for _, pair := range kwargs {
			if err := named.SetKey(pair[0], pair[1]); err != nil {
				return nil, err
			}
		}
		keywords, err := api.FromValue(named, options.Limits.DataBytes)
		if err != nil {
			return nil, err
		}
		raw := map[string]any{"args": inputs, "kwargs": keywords}
		encoded, err := json.Marshal(raw)
		if err != nil || len(encoded) > options.Limits.DataBytes {
			return nil, errors.New("capability input byte limit exceeded")
		}
		digest := sha256.Sum256(encoded)
		started := time.Now()
		observation := Observation{Index: inv.calls, Phase: "started", Capability: name, InputDigest: hex.EncodeToString(digest[:]), Inputs: inv.redact(raw).(map[string]any), StartedAt: started.UTC()}
		if options.Replay != nil {
			cached, found, err := options.Replay(inv.ctx, observation)
			if err != nil {
				return nil, errors.New("capability recorded-result lookup failed: reconcile the recorded outcome before retrying")
			}
			if found {
				return inv.restore(name, cached, args, kwargs)
			}
		}
		if options.Observe != nil {
			if err := options.Observe(inv.ctx, observation); err != nil {
				return nil, errors.New("capability observation failed before execution")
			}
		}
		result, callErr := invoke()
		finished := time.Now()
		observation.FinishedAt, observation.Duration = finished.UTC(), finished.Sub(started)
		observation.Phase = "completed"
		if callErr != nil {
			observation.Phase = "failed"
			if inv.externalStarted && !inv.externalCompleted {
				observation.Phase = "unknown"
			}
			observation.Error = inv.redact(callErr.Error()).(string)
		} else {
			data, err := observedResult(name, result, options.Limits)
			if err != nil {
				return nil, errors.New("capability result observation exceeded structured data limit: external outcome may be ambiguous")
			}
			observation.Result = inv.redact(data)
			before, _ := json.Marshal(data)
			after, _ := json.Marshal(observation.Result)
			observation.Redacted = string(before) != string(after)
		}
		if options.Observe != nil {
			if err := options.Observe(inv.ctx, observation); err != nil {
				return nil, errors.New("capability observation failed after execution: external outcome may be ambiguous")
			}
		}
		return result, callErr
	})
}

func observedResult(capability string, value starlark.Value, limits Limits) (any, error) {
	if structure, ok := value.(*starlarkstruct.Struct); ok {
		isHTTP := responseCapability(capability)
		var body string
		dict := starlark.NewDict(0)
		for _, name := range structure.AttrNames() {
			attr, err := structure.Attr(name)
			if err != nil {
				return nil, err
			}
			if _, callable := attr.(starlark.Callable); callable {
				continue
			}
			if isHTTP && name == "text" {
				continue
			}
			if isHTTP && name == "body" {
				var ok bool
				body, ok = starlark.AsString(attr)
				if !ok || len(body) > limits.OutputBytes {
					return nil, errors.New("HTTP observation body byte limit exceeded")
				}
				continue
			}
			if err := dict.SetKey(starlark.String(name), attr); err != nil {
				return nil, err
			}
		}
		data, err := api.FromValue(dict, limits.DataBytes)
		if err != nil {
			return nil, err
		}
		if isHTTP {
			// Body is charged once against OutputBytes, not again as text or JSON escapes.
			data.(map[string]any)["body"] = body
		}
		return data, nil
	}
	return api.FromValue(value, limits.DataBytes)
}

func sensitiveHeader(name string) bool {
	switch strings.ToLower(name) {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key":
		return true
	default:
		return false
	}
}

func (inv *invocation) captureHeaders(name string, args starlark.Tuple, kwargs []starlark.Tuple) {
	if name != "http.get" && name != "http.post" {
		return
	}
	headers, ok := api.Argument(args, kwargs, "headers", 1).(*starlark.Dict)
	if !ok {
		return
	}
	for _, header := range headers.Items() {
		key, _ := starlark.AsString(header[0])
		value, ok := starlark.AsString(header[1])
		if !ok || !sensitiveHeader(key) {
			continue
		}
		inv.callSecrets = append(inv.callSecrets, headerSecrets(key, value)...)
	}
}

func headerSecrets(name, value string) []string {
	if value == "" {
		return nil
	}
	out := []string{value}
	if strings.EqualFold(name, "cookie") || strings.EqualFold(name, "set-cookie") {
		for _, part := range strings.Split(value, ";") {
			if _, secret, ok := strings.Cut(strings.TrimSpace(part), "="); ok && secret != "" {
				out = append(out, secret)
			}
		}
	} else if _, secret, ok := strings.Cut(value, " "); ok && secret != "" {
		out = append(out, secret)
	}
	return out
}

func (inv *invocation) redact(value any) any {
	secrets := append([]string{}, inv.callSecrets...)
	if key := inv.engine.options.Jev.APIKey; key != "" {
		secrets = append(secrets, key)
	}
	for _, credential := range inv.engine.options.Auth {
		secrets = append(secrets, credential.Secrets...)
		for name, secret := range credential.Headers {
			secrets = append(secrets, headerSecrets(name, secret)...)
		}
	}
	var walk func(any) any
	walk = func(value any) any {
		switch value := value.(type) {
		case string:
			for _, secret := range secrets {
				if secret != "" {
					value = strings.ReplaceAll(value, secret, "[redacted]")
				}
			}
			return value
		case map[string]any:
			out := make(map[string]any, len(value))
			for key, item := range value {
				if sensitiveHeader(key) {
					out[key] = "[redacted]"
				} else {
					out[walk(key).(string)] = walk(item)
				}
			}
			return out
		case []any:
			out := make([]any, len(value))
			for i, item := range value {
				out[i] = walk(item)
			}
			return out
		default:
			return value
		}
	}
	return walk(value)
}

func (inv *invocation) safeError(err error) error {
	var eval *starlark.EvalError
	if errors.As(err, &eval) {
		return diagnostic(&starlark.EvalError{Msg: inv.redact(eval.Msg).(string), CallStack: eval.CallStack})
	}
	message := inv.redact(err.Error()).(string)
	if message == err.Error() {
		return err
	}
	return errors.New(message)
}

// Observer is called synchronously outside any controller lock. It may reject a
// call before execution. Failure after execution does not undo external effects.
type Observer = api.Observer

// responseCapability: the call returns an http response struct, whose body
// is observed once and whose json() method is rebuilt on replay.
func responseCapability(name string) bool {
	switch name {
	case "http.get", "http.post", "github.get", "github.post", "github.patch", "github.put", "github.delete":
		return true
	}
	return false
}
