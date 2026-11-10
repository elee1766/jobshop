package stepexec

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/orchestration/hooks"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
)

// httpStep runs one attempt of an http step.
func (e Executor) httpStep(ctx context.Context, cfg *config.Config, s config.Step, tpl Templates, attemptKey string) Result {
	fill := tpl.Fill
	url, err := fill(s.URL)
	if err != nil {
		return failed("", err)
	}
	res := Result{Command: s.Method + " " + url}
	body, err := fill(s.Body)
	if err != nil {
		res.Result, res.Err = ResultFailed, err.Error()
		return res
	}
	req, err := http.NewRequestWithContext(telemetry.WithRoute(ctx, "/step"), s.Method, url, strings.NewReader(body))
	if err != nil {
		res.Result, res.Err = ResultFailed, err.Error()
		return res
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range s.Headers {
		if v, err = fill(v); err != nil {
			res.Result, res.Err = ResultFailed, err.Error()
			return res
		}
		req.Header.Set(k, v)
	}
	var secrets []string
	for k, f := range s.HeaderFiles {
		b, err := os.ReadFile(f)
		if err != nil {
			res.Result, res.Err = ResultFailed, fmt.Sprintf("header %s: %v", k, err)
			return res
		}
		v := strings.TrimSpace(string(b))
		req.Header.Set(k, v)
		secrets = append(secrets, "HEADER_SECRET="+v)
	}
	req.Header.Set("Idempotency-Key", attemptKey)
	resp, err := e.HTTP.Do(req)
	if err != nil {
		res.Result, res.Err = ResultFailed, err.Error()
		if ctx.Err() != nil {
			res.Err = "timed out"
		}
		return res
	}
	defer func() { _ = resp.Body.Close() }()
	var out LimitedBuffer
	_, _ = io.Copy(&out, resp.Body)
	res.Output = cfg.Redact.Scanner().Redactor(secrets).Text(out.String())
	res.Outcome = OutcomeOf(res.Output)
	switch code := resp.StatusCode; {
	case code >= 200 && code < 300:
		res.Result = ResultOK
	case code == http.StatusConflict || code == http.StatusTooEarly || code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable:
		res.Result = ResultNotYet
		res.RetryAfter = hooks.RetryAfter(resp.Header.Get("Retry-After"), e.Now())
	default:
		res.Result, res.Err = ResultFailed, resp.Status
	}
	return res
}
