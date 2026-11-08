// Package linear implements the linear built-in: Linear GraphQL calls
// authenticated by the host's Linear identities. The token is minted per call
// by the host and never becomes a script value.
package linear

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/elee1766/jobshop/pkg/automation/api"
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
	vars := any(map[string]any{})
	if variables != nil {
		var err error
		if vars, err = api.FromValue(variables, options.Limits.DataBytes); err != nil {
			return nil, err
		}
	}
	body, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return nil, err
	}
	ctx, cancel, err := api.CallContext(host, timeout)
	if err != nil {
		return nil, err
	}
	defer cancel()
	auth, err := options.Linear(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("linear: %w", err)
	}
	host.AddSecret(auth.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, auth.APIURL, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("linear: could not construct request")
	}
	req.Header.Set("Content-Type", "application/json")
	// Personal API keys go in Authorization as they are; OAuth tokens need Bearer.
	authz := auth.Token
	if !strings.HasPrefix(authz, "lin_api_") {
		authz = "Bearer " + authz
	}
	req.Header.Set("Authorization", authz)
	client := *options.HTTP
	client.Jar = nil
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	host.MarkStarted()
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("linear: request failed or was cancelled")
	}
	data, err := api.ReadBounded(resp.Body, options.Limits.OutputBytes)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("linear response: %w", err)
	}
	host.MarkCompleted()
	var decoded struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, fmt.Errorf("linear.graphql: HTTP %d: invalid JSON response", resp.StatusCode)
	}
	if len(decoded.Errors) > 0 {
		msgs := make([]string, 0, len(decoded.Errors))
		for _, e := range decoded.Errors {
			msgs = append(msgs, e.Message)
		}
		return nil, fmt.Errorf("linear.graphql: %s", strings.Join(msgs, "; "))
	}
	if resp.StatusCode != http.StatusOK || len(decoded.Data) == 0 || string(decoded.Data) == "null" {
		return nil, fmt.Errorf("linear.graphql: HTTP %d", resp.StatusCode)
	}
	value, err := api.DecodeJSON(decoded.Data)
	if err != nil {
		return nil, errors.New("linear.graphql: invalid data")
	}
	return api.ToValue(value, options.Limits.DataBytes)
}
