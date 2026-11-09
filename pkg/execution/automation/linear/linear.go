// Package linear implements the linear built-in: Linear GraphQL calls
// through the host's client for the identity covering a team. The client
// mints the token per request; it never becomes a script value.
package linear

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elee1766/jobshop/pkg/execution/automation/api"
	"github.com/elee1766/jobshop/pkg/integrations/apicall"
	"go.starlark.net/starlark"
)

// GraphQL runs linear.graphql(query, variables={}, team="", timeout="") and
// returns the data object. team (a key such as ENG) selects the identity. A
// transport error, a non-200 status or a response with errors raises, with
// Linear's error messages.
func GraphQL(host api.Host, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := host.Options()
	if options.Linear == nil {
		return nil, errors.New("linear capability is disabled by host")
	}
	var query, team, timeout string
	var variables *starlark.Dict
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "query", &query, "variables?", &variables, "team?", &team, "timeout?", &timeout); err != nil {
		return nil, err
	}
	vars := map[string]any{}
	if variables != nil {
		raw, err := api.FromValue(variables, options.Limits.DataBytes)
		if err != nil {
			return nil, err
		}
		vars = raw.(map[string]any)
	}
	ctx, cancel, err := api.CallContext(host, timeout)
	if err != nil {
		return nil, err
	}
	defer cancel()
	cl, err := options.Linear(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("linear: %w", err)
	}
	ctx = apicall.WithObserver(ctx, apicall.Observer{Token: host.AddSecret, Sent: host.MarkStarted, Done: host.MarkCompleted})
	var data json.RawMessage
	if err := cl.GraphQL(ctx, query, vars, &data); err != nil {
		return nil, err
	}
	if len(data) > options.Limits.OutputBytes {
		return nil, errors.New("linear.graphql: output byte limit exceeded")
	}
	if len(data) == 0 || string(data) == "null" {
		return nil, errors.New("linear.graphql: response has no data")
	}
	value, err := api.DecodeJSON(data)
	if err != nil {
		return nil, errors.New("linear.graphql: invalid data")
	}
	return api.ToValue(value, options.Limits.DataBytes)
}
