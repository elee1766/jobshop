package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func requestURL(raw string) (*url.URL, error) {
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

func parseOrigin(raw string) (string, error) {
	u, err := requestURL(raw)
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
		candidate, err := parseOrigin(raw)
		if err == nil && candidate == origin(u) {
			return true
		}
	}
	return false
}

func callContext(inv *invocation, timeout string) (context.Context, context.CancelFunc, error) {
	duration := inv.engine.options.Limits.Duration
	if timeout != "" {
		parsed, err := time.ParseDuration(timeout)
		if err != nil || parsed <= 0 {
			return nil, nil, errors.New("timeout must be a positive duration string")
		}
		duration = min(duration, parsed)
	}
	ctx, cancel := context.WithTimeout(inv.ctx, duration)
	return ctx, cancel, nil
}

func httpCall(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	inv, err := active(t)
	if err != nil {
		return nil, err
	}
	if !inv.engine.options.AllowHTTP {
		return nil, errors.New("HTTP capability is disabled by host")
	}
	var raw, auth, timeout, body string
	var headers *starlark.Dict
	var jsonBody starlark.Value = starlark.None
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "url", &raw, "headers?", &headers, "auth?", &auth, "timeout?", &timeout, "body?", &body, "json?", &jsonBody); err != nil {
		return nil, err
	}
	u, err := requestURL(raw)
	if err != nil {
		return nil, err
	}
	options := inv.engine.options
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
		data, convertErr := fromValue(jsonBody, options.Limits.DataBytes)
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
	ctx, cancel, err := callContext(inv, timeout)
	if err != nil {
		return nil, err
	}
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, u.String(), strings.NewReader(body))
	if err != nil {
		return nil, errors.New("could not construct HTTP request")
	}
	if headers != nil {
		rawHeaders, convertErr := fromValue(headers, options.Limits.DataBytes)
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
		if _, err := requestURL(next.URL.String()); err != nil {
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
	inv.externalStarted = true
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("HTTP request failed or was cancelled")
	}
	defer func() { _ = response.Body.Close() }()
	data, err := readBounded(response.Body, options.Limits.OutputBytes)
	if err != nil {
		return nil, fmt.Errorf("HTTP response: %w", err)
	}
	inv.externalCompleted = true
	responseHeaders := map[string]any{}
	for key, values := range response.Header {
		responseHeaders[strings.ToLower(key)] = strings.Join(values, ", ")
	}
	return responseValue(response.StatusCode, data, responseHeaders, options)
}

func responseValue(status int, data []byte, responseHeaders map[string]any, options Options) (starlark.Value, error) {
	attrs := starlark.StringDict{"status_code": starlark.MakeInt(status), "body": starlark.String(data), "text": starlark.String(data)}
	headerValue, err := toValue(responseHeaders, options.Limits.DataBytes)
	if err != nil {
		return nil, err
	}
	attrs["headers"] = headerValue
	attrs["json"] = starlark.NewBuiltin("response.json", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs("response.json", args, kwargs); err != nil {
			return nil, err
		}
		value, err := decodeJSON(data)
		if err != nil {
			return nil, errors.New("HTTP response is not valid JSON")
		}
		return toValue(value, options.Limits.DataBytes)
	})
	return starlarkstruct.FromStringDict(starlark.String("response"), attrs), nil
}

func readBounded(reader io.Reader, limit int) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	if err != nil {
		return nil, errors.New("could not read bounded output")
	}
	if len(data) > limit {
		return nil, errors.New("output byte limit exceeded")
	}
	return data, nil
}

func decodeJSON(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected trailing JSON data")
	}
	return value, nil
}
