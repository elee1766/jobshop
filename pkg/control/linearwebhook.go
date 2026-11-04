package control

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
)

// Linear webhook intake (#300). Linear pushes Issue and Comment events
// to POST /linear/webhook on linear.webhook_listen. The payload is only
// a hint: the handler checks the signature, age, organization and actor,
// then re-reads the issue and runs the polling code early. Polling stays
// the fallback, so a lost or refused event only costs latency.

const (
	linearWebhookPath   = "/linear/webhook"
	linearWebhookMaxAge = 60 * time.Second
	linearWebhookMaxLen = 1 << 20
)

// linearWebhook is the part of a Linear webhook payload jobshop reads.
type linearWebhook struct {
	Action         string `json:"action"`
	Type           string `json:"type"`
	OrganizationID string `json:"organizationId"`
	// WebhookTimestamp is Unix milliseconds.
	WebhookTimestamp int64 `json:"webhookTimestamp"`
	Actor            *struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"actor"`
	Data struct {
		Identifier string `json:"identifier"`
		Number     int    `json:"number"`
		Team       *struct {
			Key string `json:"key"`
		} `json:"team"`
		// Comment events.
		UserID   string `json:"userId"`
		BotActor any    `json:"botActor"`
		Issue    *struct {
			Identifier string `json:"identifier"`
			Number     int    `json:"number"`
			Team       *struct {
				Key string `json:"key"`
			} `json:"team"`
		} `json:"issue"`
	} `json:"data"`
}

// issueKey is the issue the event is about, "" when the payload does not
// name it.
func (w linearWebhook) issueKey() string {
	key := func(ident string, n int, team *struct {
		Key string `json:"key"`
	},
	) string {
		if ident != "" {
			return ident
		}
		if team != nil && team.Key != "" && n > 0 {
			return fmt.Sprintf("%s-%d", team.Key, n)
		}
		return ""
	}
	switch w.Type {
	case "Issue":
		return key(w.Data.Identifier, w.Data.Number, w.Data.Team)
	case "Comment":
		if w.Data.Issue != nil {
			return key(w.Data.Issue.Identifier, w.Data.Issue.Number, w.Data.Issue.Team)
		}
	}
	return ""
}

// serveLinearWebhook serves POST /linear/webhook on linear.webhook_listen
// until the controller stops. Empty listen address: off.
func (c *Controller) serveLinearWebhook() error {
	addr := c.config().Linear.WebhookListen
	if addr == "" {
		return nil
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("linear webhook listen: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST "+linearWebhookPath, c.handleLinearWebhook)
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	c.wg.Go(func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.log.Error("linear webhook server", "err", err)
		}
	})
	c.wg.Go(func() {
		<-c.ctx.Done()
		_ = srv.Close()
	})
	c.log.Info("serving linear webhook", "addr", ln.Addr().String())
	return nil
}

// handleLinearWebhook answers fast; the work runs on the controller's
// context after the response.
func (c *Controller) handleLinearWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, linearWebhookMaxLen+1))
	if err != nil || len(body) > linearWebhookMaxLen {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	name, ok := c.linearWebhookIdentity(r.Header.Get("Linear-Signature"), body)
	if !ok {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	var ev linearWebhook
	if err := json.Unmarshal(body, &ev); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}
	age := c.now().Sub(time.UnixMilli(ev.WebhookTimestamp))
	if age > linearWebhookMaxAge || age < -linearWebhookMaxAge {
		http.Error(w, "stale", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
	if ev.Type != "Issue" && ev.Type != "Comment" {
		return
	}
	c.wg.Go(func() {
		if err := c.linearWebhookEvent(c.ctx, name, ev); err != nil {
			c.log.Warn("linear webhook", "type", ev.Type, "issue", ev.issueKey(), "err", err)
		}
	})
}

// linearWebhookIdentity returns the linear-app identity whose webhook
// secret signed body. sig is hex HMAC-SHA256 of the raw body.
func (c *Controller) linearWebhookIdentity(sig string, body []byte) (string, bool) {
	got, err := hex.DecodeString(strings.TrimSpace(sig))
	if err != nil || len(got) == 0 {
		return "", false
	}
	cfg := c.config()
	for _, name := range cfg.IdentityNames() {
		id := cfg.Identities[name]
		if id.Kind != config.IdentityLinearApp || id.WebhookSecretFile == "" {
			continue
		}
		secret, err := os.ReadFile(id.WebhookSecretFile)
		if err != nil {
			c.log.Warn("linear webhook: secret", "identity", name, "err", err)
			continue
		}
		mac := hmac.New(sha256.New, []byte(strings.TrimSpace(string(secret))))
		mac.Write(body)
		if hmac.Equal(mac.Sum(nil), got) {
			return name, true
		}
	}
	return "", false
}

// linearWebhookEvent filters a verified event and runs the polling path
// for its issue.
func (c *Controller) linearWebhookEvent(ctx context.Context, name string, ev linearWebhook) error {
	if ev.Actor == nil || !strings.EqualFold(ev.Actor.Type, "user") || ev.Data.BotActor != nil {
		return nil // apps, bots and integrations: their writes need no push
	}
	key := ev.issueKey()
	team, ok := linear.Team(key)
	if !ok {
		return nil // polling picks it up
	}
	cfg := c.config()
	id := cfg.Identities[name]
	if !id.CoversTeam(team) {
		return nil
	}
	tok, err := c.minter.Token(ctx, name, id, team)
	if err != nil {
		return err
	}
	lctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cl := linear.New(c.linearURL, tok.Token)
	me, org, err := cl.Me(lctx)
	if err != nil {
		return err
	}
	if ev.OrganizationID != org || ev.Actor.ID == me || ev.Data.UserID == me {
		return nil
	}
	// Never trust the payload: the issue must exist as the identity
	// reads it.
	is, err := cl.GetIssue(lctx, key)
	if err != nil {
		return err
	}
	switch ev.Type {
	case "Issue":
		c.linearIntakeInvalidate(is.Team.Key)
		// Same pass the tick runs, so the stall fires now.
		c.tickMu.Lock()
		if !c.draining() {
			c.runTriggers()
		}
		c.tickMu.Unlock()
	case "Comment":
		c.pollSubscriptionThread(ctx, thread{Linear: is.Identifier})
	}
	return nil
}

// linearIntakeInvalidate drops cached linear_issue_no_run reads of team,
// so the next tick reads Linear again.
func (c *Controller) linearIntakeInvalidate(team string) {
	c.linIntake.mu.Lock()
	defer c.linIntake.mu.Unlock()
	for k := range c.linIntake.cache {
		if t, _, _ := strings.Cut(k, "|"); strings.EqualFold(t, team) {
			delete(c.linIntake.cache, k)
		}
	}
}
