// Package http implements bounded HTTP capabilities for automation hooks.
package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/elee1766/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// RequestURL validates destinations before any external request is started.
func RequestURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u == nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return nil, errors.New("URL must be http(s) with a host and without credentials or fragment")
	}
	return u, nil
}

func origin(u *url.URL) string {
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return strings.ToLower(u.Scheme+"://"+u.Hostname()) + ":" + port
}

// ParseOrigin validates and normalizes a host origin grant.
func ParseOrigin(raw string) (string, error) {
	u, err := RequestURL(raw)
	if err != nil {
		return "", err
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" {
		return "", errors.New("origin must not include path or query")
	}
	return origin(u), nil
}

func allowedOrigin(u *url.URL, origins []string) bool {
	for _, raw := range origins {
		candidate, err := ParseOrigin(raw)
		if err == nil && candidate == origin(u) {
			return true
		}
	}
	return false
}

// Call executes http.get or http.post under the host's grants and limits.
func Call(host api.Host, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options := host.Options()
	if !options.AllowHTTP {
		return nil, errors.New("HTTP capability is disabled by host")
	}
	var raw, auth, timeout, body string
	var headers *starlark.Dict
	var jsonBody starlark.Value = starlark.None
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "url", &raw, "headers?", &headers, "auth?", &auth, "timeout?", &timeout, "body?", &body, "json?", &jsonBody); err != nil {
		return nil, err
	}
	u, err := RequestURL(raw)
	if err != nil {
		return nil, err
	}
	if len(options.HTTPOrigins) > 0 && !allowedOrigin(u, options.HTTPOrigins) {
		return nil, errors.New("HTTP destination is not granted by host")
	}
	method := http.MethodGet
	if fn.Name() == "http.post" {
		method = http.MethodPost
	}
	if body != "" && jsonBody != starlark.None {
		return nil, errors.New("http: body and json are mutually exclusive")
	}
	if jsonBody != starlark.None {
		data, convertErr := api.FromValue(jsonBody, options.Limits.DataBytes)
		if convertErr != nil {
			return nil, convertErr
		}
		encoded, marshalErr := json.Marshal(data)
		if marshalErr != nil {
			return nil, marshalErr
		}
		body = string(encoded)
	}
	if len(body) > options.Limits.DataBytes {
		return nil, errors.New("HTTP request body byte limit exceeded")
	}
	if method == http.MethodGet && body != "" {
		return nil, errors.New("http.get does not accept a body")
	}
	ctx, cancel, err := api.CallContext(host, timeout)
	if err != nil {
		return nil, err
	}
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, u.String(), strings.NewReader(body))
	if err != nil {
		return nil, errors.New("could not construct HTTP request")
	}
	if headers != nil {
		rawHeaders, convertErr := api.FromValue(headers, options.Limits.DataBytes)
		if convertErr != nil {
			return nil, convertErr
		}
		for key, value := range rawHeaders.(map[string]any) {
			text, ok := value.(string)
			if !ok {
				return nil, errors.New("HTTP header values must be strings")
			}
			req.Header.Set(key, text)
		}
	}
	if jsonBody != starlark.None {
		req.Header.Set("Content-Type", "application/json")
	}
	credentialed := auth != ""
	if credentialed {
		credential, ok := options.Auth[auth]
		if !ok {
			return nil, errors.New("unknown HTTP credential handle")
		}
		if !allowedOrigin(u, credential.Origins) {
			return nil, errors.New("HTTP credential is not granted for destination origin")
		}
		for key, value := range credential.Headers {
			req.Header.Set(key, value)
		}
	}
	client := *options.HTTP
	// A cookie jar could add host credentials outside our origin policy.
	client.Jar = nil
	client.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("HTTP redirect limit exceeded")
		}
		if _, err := RequestURL(next.URL.String()); err != nil {
			return err
		}
		if len(options.HTTPOrigins) > 0 && !allowedOrigin(next.URL, options.HTTPOrigins) {
			return errors.New("HTTP redirect destination is not granted by host")
		}
		if origin(next.URL) != origin(u) {
			if credentialed {
				return errors.New("cross-origin credential redirect refused")
			}
			// Custom headers may be credentials too, so retain no caller headers.
			next.Header = make(http.Header)
		}
		return nil
	}
	host.MarkStarted()
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("HTTP request failed or was cancelled")
	}
	defer func() { _ = response.Body.Close() }()
	data, err := api.ReadBounded(response.Body, options.Limits.OutputBytes)
	if err != nil {
		return nil, fmt.Errorf("HTTP response: %w", err)
	}
	host.MarkCompleted()
	responseHeaders := map[string]any{}
	for key, values := range response.Header {
		responseHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
	}
	return ResponseValue(response.StatusCode, data, responseHeaders, options)
}

// ResponseValue is the response struct http and github calls return.
func ResponseValue(status int, data []byte, responseHeaders map[string]any, options api.Options) (starlark.Value, error) {
	attrs := starlark.StringDict{"status_code": starlark.MakeInt(status), "body": starlark.String(data), "text": starlark.String(data)}
	headerValue, err := api.ToValue(responseHeaders, options.Limits.DataBytes)
	if err != nil {
		return nil, err
	}
	attrs["headers"] = headerValue
	attrs["json"] = starlark.NewBuiltin("response.json", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs("response.json", args, kwargs); err != nil {
			return nil, err
		}
		value, err := api.DecodeJSON(data)
		if err != nil {
			return nil, errors.New("HTTP response is not valid JSON")
		}
		return api.ToValue(value, options.Limits.DataBytes)
	})
	return starlarkstruct.FromStringDict(starlark.String("response"), attrs), nil
}
