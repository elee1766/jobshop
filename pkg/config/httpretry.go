package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// HTTPRetry retries a url hook call. A transport error, HTTP 5xx or 429
// is retried; any other status is final. Attempts counts the first try,
// so 0 or 1 is no retry. Backoff (default 1s) doubles after each try up
// to MaxBackoff (default 30s); a Retry-After header replaces it, capped
// the same way. The hook's timeout bounds each try.
type HTTPRetry struct {
	Attempts   int    `yaml:"attempts"`
	Backoff    string `yaml:"backoff"`
	MaxBackoff string `yaml:"max_backoff"`
}

// Tries is the number of tries, at least one.
func (r HTTPRetry) Tries() int { return max(r.Attempts, 1) }

// Wait is the backoff before try n (1-based, n >= 2).
func (r HTTPRetry) Wait(n int) time.Duration {
	d := Duration(r.Backoff, time.Second)
	limit := r.Limit()
	for i := 2; i < n && d < limit; i++ {
		d *= 2
	}
	return min(d, limit)
}

// Limit is the longest wait between tries.
func (r HTTPRetry) Limit() time.Duration { return Duration(r.MaxBackoff, 30*time.Second) }

// Budget is the most time all tries can take with per-try timeout t.
func (r HTTPRetry) Budget(t time.Duration) time.Duration {
	total := time.Duration(r.Tries()) * t
	for n := 2; n <= r.Tries(); n++ {
		total += r.Limit() // a Retry-After may wait up to the limit
	}
	return total
}

func (r HTTPRetry) check() error {
	if r.Attempts < 0 || r.Attempts > 20 {
		return fmt.Errorf("retry.attempts %d must be 0 to 20", r.Attempts)
	}
	for _, f := range []struct{ key, v string }{{"retry.backoff", r.Backoff}, {"retry.max_backoff", r.MaxBackoff}} {
		if f.v == "" {
			continue
		}
		if d, err := time.ParseDuration(f.v); err != nil || d <= 0 {
			return fmt.Errorf("%s %q must be a positive duration", f.key, f.v)
		}
	}
	return nil
}

// checkHookURL checks a hook's url, secret_file and retry: exactly one of
// command and url, and secret_file and retry only with url.
func checkHookURL(command, u string, secret Secret, retry HTTPRetry) error {
	if (command == "") == (u == "") {
		return errors.New("set exactly one of command and url")
	}
	if u == "" {
		if !secret.IsZero() {
			return errors.New("secret_file is for url hooks")
		}
		if retry != (HTTPRetry{}) {
			return errors.New("retry is for url hooks")
		}
		return nil
	}
	p, err := url.Parse(u)
	if err != nil || (p.Scheme != "http" && p.Scheme != "https") || p.Host == "" {
		return fmt.Errorf("url %q must be an http or https URL", u)
	}
	if secret.IsZero() {
		return errors.New("secret_file is required with url")
	}
	return retry.check()
}
