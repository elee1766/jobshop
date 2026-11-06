package control

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// hookPost is one url hook call (#348 admission, #188 guardrail): a
// signed JSON POST, retried per its retry settings.
type hookPost struct {
	url     string
	secret  config.Secret // already resolved against the root dir
	headers map[string]string
	timeout time.Duration // per try
	retry   config.HTTPRetry
	// after waits between tries; nil is time.After. Tests pass a clock.
	after func(time.Duration) <-chan time.Time
}

// errHookTimeout marks a try that ran out of its timeout.
var errHookTimeout = errors.New("timed out")

// do posts body and returns the response body of the first 2xx answer.
// A transport error, a timeout, 429 or 5xx is retried; any other status
// is final. The error names the last failure and how many tries ran.
func (p hookPost) do(ctx context.Context, body []byte) ([]byte, error) {
	secret, err := config.ReadSecret(ctx, p.secret)
	if err != nil {
		return nil, fmt.Errorf("secret: %w", err)
	}
	secret = bytes.TrimSpace(secret)
	after := p.after
	if after == nil {
		after = time.After
	}
	tries := p.retry.Tries()
	var last error
	for n := 1; n <= tries; n++ {
		out, wait, retry, err := p.try(ctx, secret, body)
		if err == nil {
			return out, nil
		}
		last = err
		if !retry || n == tries || ctx.Err() != nil {
			break
		}
		if wait <= 0 {
			wait = p.retry.Wait(n + 1)
		}
		select {
		case <-after(min(wait, p.retry.Limit())):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if tries > 1 {
		return nil, fmt.Errorf("%w (after %d tries)", last, tries)
	}
	return nil, last
}

// try is one POST. retry reports whether the failure may succeed later;
// wait is the server's Retry-After, or 0.
func (p hookPost) try(ctx context.Context, secret, body []byte) (out []byte, wait time.Duration, retry bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(telemetry.WithRoute(ctx, "/hook"), http.MethodPost, p.url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, false, err
	}
	reqID := make([]byte, 16)
	_, _ = rand.Read(reqID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jobshop-Request-Id", hex.EncodeToString(reqID))
	req.Header.Set("X-Jobshop-Signature-256", hookSignature(secret, body))
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}
	resp, err := hookHTTP.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, 0, true, fmt.Errorf("%w after %s", errHookTimeout, p.timeout)
		}
		return nil, 0, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	out, err = io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, 0, true, err
	}
	if resp.StatusCode/100 == 2 {
		return out, 0, false, nil
	}
	err = fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(out[:min(len(out), 200)])))
	retry = resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
	return nil, retryAfter(resp.Header.Get("Retry-After"), time.Now()), retry, err
}

// hookSignature is the X-Jobshop-Signature-256 value for body.
func hookSignature(secret, body []byte) string {
	m := hmac.New(sha256.New, secret)
	m.Write(body)
	return "sha256=" + hex.EncodeToString(m.Sum(nil))
}
