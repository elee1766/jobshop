package identity

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/config"
)

// HTTPError is a token endpoint answering outside 2xx. It carries the
// status so Outage can tell a broken host from a bad request.
type HTTPError struct {
	Op     string // what was asked, e.g. "installation 7 token"
	Code   int
	Status string // the response status line, e.g. "401 Unauthorized"
	Body   string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return e.Op + ": " + e.Status
	}
	return e.Op + ": " + e.Status + ": " + e.Body
}

// Outage reports whether err says the host cannot get credentials
// (JOB-194), as opposed to a request that is wrong: a credential command
// that failed or timed out, a secret file that is missing or unreadable,
// an auth refusal (401, 403 that is not a rate limit), a provider 5xx, or
// an unreachable endpoint. Retrying the work against it only spends
// attempts, so the controller pauses the pools that use the identity
// instead. A caller's own cancellation is not an outage.
func Outage(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	var se *config.SecretError
	if errors.As(err, &se) {
		return true
	}
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission) {
		return true
	}
	var he *HTTPError
	if errors.As(err, &he) {
		switch {
		case he.Code == http.StatusUnauthorized, he.Code >= 500:
			return true
		case he.Code == http.StatusForbidden:
			return !strings.Contains(strings.ToLower(he.Body), "rate limit")
		}
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return true
	}
	return errors.Is(err, io.ErrUnexpectedEOF)
}

// Probe asks whether name can get credentials again: it drops the cached
// token, mints a fresh one, and for a GitHub identity reads GET
// /rate_limit with it, which costs no quota and proves the token is
// accepted. A nil result lifts the pool pause.
func (m *Minter) Probe(ctx context.Context, name string, id config.Identity, target string) error {
	m.Drop(name)
	t, err := m.Token(ctx, name, id, target)
	if err != nil {
		return err
	}
	if !config.IsGitHub(id.Kind) {
		return nil
	}
	base := id.APIURL
	if base == "" {
		base = DefaultAPIURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/rate_limit", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+t.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("identity %s: GET /rate_limit: %w", name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return &HTTPError{Op: "identity " + name + ": GET /rate_limit", Code: resp.StatusCode, Status: resp.Status, Body: strings.TrimSpace(string(body))}
	}
	return nil
}
