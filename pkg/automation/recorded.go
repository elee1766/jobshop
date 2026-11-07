package automation

import (
	"encoding/json"
	"errors"

	"github.com/elee1766/jobshop/pkg/automation/api"
	"github.com/elee1766/jobshop/pkg/automation/command"
	automationhttp "github.com/elee1766/jobshop/pkg/automation/http"
	"github.com/elee1766/jobshop/pkg/automation/jev"
	"go.starlark.net/starlark"
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
	if name == "jev" {
		return jev.Restore(inv, encoded, args, kwargs)
	}
	value, err := api.DecodeJSON(encoded)
	if err != nil {
		return nil, errors.New("invalid recorded capability result JSON")
	}
	fields, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("recorded capability result must be an object")
	}
	switch name {
	case "http.get", "http.post":
		return automationhttp.Restore(fields, options)
	case "shell":
		return command.Restore(fields, options)
	default:
		return nil, errors.New("unknown recorded capability")
	}
}
