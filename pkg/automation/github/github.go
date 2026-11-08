// Package github implements the github built-in: GitHub REST and GraphQL
// calls authenticated by the host's identities. The token is minted per call
// by the host and never becomes a script value.
package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/elee1766/jobshop/pkg/automation/api"
	automationhttp "github.com/elee1766/jobshop/pkg/automation/http"
	"go.starlark.net/starlark"
)

// maxPages bounds pages=N: each page is one request.
const maxPages = 10

var repoPath = regexp.MustCompile(`^/repos/([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)(?:/|$)`)

// linkNext finds rel="next" in a Link header.
var linkNext = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

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
	auth, err := options.GitHub(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("github: %w", err)
	}
	host.AddSecret(auth.Token)
	base, err := url.Parse(strings.TrimRight(auth.APIURL, "/"))
	if err != nil || base.Host == "" {
		return nil, errors.New("github: host API URL is invalid")
	}
	next := base.String() + path
	if len(query) > 0 {
		next += "?" + query.Encode()
	}
	client := *options.HTTP
	client.Jar = nil
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 || req.URL.Host != base.Host || req.URL.Scheme != base.Scheme {
			return errors.New("github: redirect away from the API refused")
		}
		return nil
	}
	var status int
	var data []byte
	var respHeaders map[string]any
	var merged []any
	total := 0
	host.MarkStarted()
	for page := 0; page < pages && next != ""; page++ {
		req, err := http.NewRequestWithContext(ctx, method, next, bytes.NewReader(body))
		if err != nil {
			return nil, errors.New("github: could not construct request")
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		for key, value := range extra {
			req.Header.Set(key, value)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Authorization", "Bearer "+auth.Token)
		resp, err := client.Do(req)
		if err != nil {
			return nil, errors.New("github: request failed or was cancelled")
		}
		chunk, err := api.ReadBounded(resp.Body, options.Limits.OutputBytes-total)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("github response: %w", err)
		}
		total += len(chunk)
		status, data = resp.StatusCode, chunk
		respHeaders = map[string]any{}
		for key, values := range resp.Header {
			respHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
		}
		next = ""
		if pages == 1 || status != http.StatusOK {
			break
		}
		decoded, err := api.DecodeJSON(chunk)
		items, ok := decoded.([]any)
		if err != nil || !ok {
			return nil, errors.New("github: pages needs a JSON array response")
		}
		merged = append(merged, items...)
		if m := linkNext.FindStringSubmatch(resp.Header.Get("Link")); m != nil {
			u, err := url.Parse(m[1])
			if err != nil || u.Host != base.Host || u.Scheme != base.Scheme {
				return nil, errors.New("github: next page link leaves the API")
			}
			next = u.String()
		}
	}
	host.MarkCompleted()
	if merged != nil {
		if data, err = json.Marshal(merged); err != nil {
			return nil, err
		}
		if len(data) > options.Limits.OutputBytes {
			return nil, errors.New("github: output byte limit exceeded")
		}
		delete(respHeaders, "link")
	}
	return automationhttp.ResponseValue(status, data, respHeaders, options)
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
	auth, err := options.GitHub(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("github: %w", err)
	}
	host.AddSecret(auth.Token)
	// GraphQL lives at /graphql on api.github.com and /api/graphql on GHES (/api/v3).
	endpoint := strings.TrimRight(auth.APIURL, "/")
	if strings.HasSuffix(endpoint, "/api/v3") {
		endpoint = strings.TrimSuffix(endpoint, "/v3")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("github: could not construct request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+auth.Token)
	client := *options.HTTP
	client.Jar = nil
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	host.MarkStarted()
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("github: request failed or was cancelled")
	}
	data, err := api.ReadBounded(resp.Body, options.Limits.OutputBytes)
	_ = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("github response: %w", err)
	}
	host.MarkCompleted()
	var decoded struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, fmt.Errorf("github.graphql: HTTP %d: invalid JSON response", resp.StatusCode)
	}
	if len(decoded.Errors) > 0 {
		msgs := make([]string, 0, len(decoded.Errors))
		for _, e := range decoded.Errors {
			msgs = append(msgs, e.Message)
		}
		return nil, fmt.Errorf("github.graphql: %s", strings.Join(msgs, "; "))
	}
	if resp.StatusCode != http.StatusOK || len(decoded.Data) == 0 {
		return nil, fmt.Errorf("github.graphql: HTTP %d", resp.StatusCode)
	}
	value, err := api.DecodeJSON(decoded.Data)
	if err != nil {
		return nil, errors.New("github.graphql: invalid data")
	}
	return api.ToValue(value, options.Limits.DataBytes)
}
