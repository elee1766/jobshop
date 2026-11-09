package identity

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/telemetry"
)

// hostTokenTTL is how long HostToken keeps a token before reading it again.
const hostTokenTTL = 10 * time.Minute

// HostToken is the host's own GitHub credential for repos no configured
// identity covers: GH_TOKEN, else `gh auth token`, cached ten minutes. It
// satisfies apicall.Source. The zero value is ready.
type HostToken struct {
	mu  sync.Mutex
	tok string
	at  time.Time
}

// Token returns the cached token, reading it again when it is old.
func (h *HostToken) Token(ctx context.Context) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.tok != "" && time.Since(h.at) < hostTokenTTL {
		return h.tok, nil
	}
	tok := os.Getenv("GH_TOKEN")
	if tok == "" {
		cmd, done := telemetry.ExecContext(ctx, "gh-auth-token", "gh", "auth", "token")
		out, err := cmd.Output()
		done(err)
		if err != nil {
			return "", fmt.Errorf("no identity covers the repo and gh auth token failed: %w", err)
		}
		tok = strings.TrimSpace(string(out))
	}
	h.tok, h.at = tok, time.Now()
	return tok, nil
}

// Renew forgets a refused token and reads it again.
func (h *HostToken) Renew(ctx context.Context, stale string) (string, error) {
	h.mu.Lock()
	if h.tok == stale {
		h.tok = ""
	}
	h.mu.Unlock()
	return h.Token(ctx)
}
