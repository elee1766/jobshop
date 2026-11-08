// Package github implements the github built-in: GitHub REST and GraphQL
// calls authenticated by the host's identities. The token is minted per call
// by the host and never becomes a script value.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"go.starlark.net/starlark"

	"github.com/elee1766/jobshop/pkg/apicall"
	"github.com/elee1766/jobshop/pkg/automation/api"
	automationhttp "github.com/elee1766/jobshop/pkg/automation/http"
	gh "github.com/elee1766/jobshop/pkg/github"
)

// maxPages bounds pages=N: each page is one request.
const maxPages = 10

var repoPath = regexp.MustCompile(`^/repos/([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)(?:/|$)`)

// Call runs github.get, post, patch, put or delete:
//
//	github.get(path, repo="", params={}, headers={}, json=None, pages=1, timeout="")
//
// path is an API path such as /repos/o/r/pulls. repo selects the identity;
// it defaults to the owner/repo in a /repos/ path. pages>1 follows
// rel="next" links and concatenates JSON array pages into one body.
func Call(host api.Host, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := host.Options()
	if options.GitHub == nil {
		return nil, errors.New("github capability is disabled by host")
	}
	var path, repo, timeout string
	var params, headers *starlark.Dict
	var jsonBody starlark.Value = starlark.None
	pages := 1
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "path", &path, "repo?", &repo, "params?", &params, "headers?", &headers, "json?", &jsonBody, "pages?", &pages, "timeout?", &timeout); err != nil {
		return nil, err
	}
	method := strings.ToUpper(strings.TrimPrefix(fn.Name(), "github."))
	if pages < 1 || pages > maxPages {
		return nil, fmt.Errorf("%s: pages must be 1 to %d", fn.Name(), maxPages)
	}
	if pages > 1 && method != http.MethodGet {
		return nil, fmt.Errorf("%s: pages is only for github.get", fn.Name())
	}
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") || strings.ContainsAny(path, "?#") {
		return nil, fmt.Errorf("%s: path must be an absolute API path without query or fragment; use params", fn.Name())
	}
	if repo == "" {
		if m := repoPath.FindStringSubmatch(path); m != nil {
			repo = m[1]
		}
	}
	var body []byte
	if jsonBody != starlark.None {
		if method == http.MethodGet {
			return nil, errors.New("github.get does not accept a body")
		}
		data, err := api.FromValue(jsonBody, options.Limits.DataBytes)
		if err != nil {
			return nil, err
		}
		if body, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	query := url.Values{}
	if params != nil {
		raw, err := api.FromValue(params, options.Limits.DataBytes)
		if err != nil {
			return nil, err
		}
		for key, value := range raw.(map[string]any) {
			switch v := value.(type) {
			case string:
				query.Set(key, v)
			case int64, uint64, bool:
				query.Set(key, fmt.Sprint(v))
			default:
				return nil, fmt.Errorf("%s: params values must be strings, integers or booleans", fn.Name())
			}
		}
	}
	extra := map[string]string{}
	if headers != nil {
		raw, err := api.FromValue(headers, options.Limits.DataBytes)
		if err != nil {
			return nil, err
		}
		for key, value := range raw.(map[string]any) {
			text, ok := value.(string)
			if !ok {
				return nil, errors.New("github: header values must be strings")
			}
			if strings.EqualFold(key, "authorization") {
				return nil, errors.New("github: the host sets Authorization")
			}
			extra[key] = text
		}
	}
	ctx, cancel, err := api.CallContext(host, timeout)
	if err != nil {
		return nil, err
	}
	defer cancel()
	cl, err := options.GitHub(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("github: %w", err)
	}
	ctx = observe(ctx, host)
	req := gh.Request{Method: method, Path: path, Query: query, Header: extra}
	if body != nil {
		req.Body = body
	}
	// Each page is one request; a non-200 page ends the walk and is the
	// response, for the script to inspect. JSON array pages concatenate.
	var resp *gh.Response
	var merged []json.RawMessage
	total := 0
	for page := 0; page < pages && req.Path != ""; page++ {
		if req.MaxBody = options.Limits.OutputBytes - total; req.MaxBody <= 0 {
			return nil, errors.New("github: output byte limit exceeded")
		}
		if resp, err = cl.Raw(ctx, req); err != nil {
			return nil, err
		}
		total += len(resp.Body)
		if pages == 1 || resp.Status != http.StatusOK {
			merged = nil
			break
		}
		var items []json.RawMessage
		if err := json.Unmarshal(resp.Body, &items); err != nil {
			return nil, errors.New("github: pages needs a JSON array response")
		}
		merged = append(merged, items...)
		req = gh.Request{Path: resp.Next(), Header: extra}
	}
	if merged != nil {
		data, err := json.Marshal(merged)
		if err != nil {
			return nil, err
		}
		if len(data) > options.Limits.OutputBytes {
			return nil, errors.New("github: output byte limit exceeded")
		}
		resp.Body = data
		resp.Header.Del("Link")
	}
	respHeaders := map[string]any{}
	for key, values := range resp.Header {
		respHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
	}
	return automationhttp.ResponseValue(resp.Status, resp.Body, respHeaders, options)
}

// observe redacts the tokens a call sends and marks when it reaches
// GitHub, for the journal's started/completed outcome.
func observe(ctx context.Context, host api.Host) context.Context {
	return apicall.WithObserver(ctx, apicall.Observer{Token: host.AddSecret, Sent: host.MarkStarted, Done: host.MarkCompleted})
}

// GraphQL runs github.graphql(query, variables={}, repo="", timeout="") and
// returns the data object. A transport error, a non-200 status or a
// response with errors raises, with GitHub's error messages.
func GraphQL(host api.Host, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := host.Options()
	if options.GitHub == nil {
		return nil, errors.New("github capability is disabled by host")
	}
	var query, repo, timeout string
	var variables *starlark.Dict
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "query", &query, "variables?", &variables, "repo?", &repo, "timeout?", &timeout); err != nil {
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
	cl, err := options.GitHub(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("github: %w", err)
	}
	var data json.RawMessage
	if err := cl.GraphQL(observe(ctx, host), query, vars, &data); err != nil {
		return nil, err
	}
	if len(data) > options.Limits.OutputBytes {
		return nil, errors.New("github.graphql: output byte limit exceeded")
	}
	value, err := api.DecodeJSON(data)
	if err != nil {
		return nil, errors.New("github.graphql: invalid data")
	}
	return api.ToValue(value, options.Limits.DataBytes)
}
