package control

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"go.opentelemetry.io/otel/metric"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
)

// Injection dedup (#397): a session that keeps its conversation already
// holds the rule text it was sent. Each runner records a hash per rule
// it injected; a turn sends only rules that are new or changed since,
// and tells the session once about a rule that was removed from config.
// The record is kept in the session's fields, so it survives a controller
// restart and a respawn whose session/load restores the conversation. It
// resets with a new conversation (session/new, a failed load, a replaced
// conversation), so it can only resend more. The pool prompt is deduped
// the same way under poolKey.

// poolKey records the pool prompt in the injected map. It is not a rule
// name, so it never gets a removal notice.
const poolKey = "@pool"

// Part statuses (PromptPart.Status) for rule parts.
const (
	injectNew       = "new"
	injectChanged   = "changed"
	injectUnchanged = "unchanged" // skipped
	injectAlways    = "always"
	injectRemoved   = "removed"
)

func hashText(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}

// dedupe marks each rule part in parts against sent (rule name -> hash of
// the text last sent) and returns the parts to send and the record after
// sending them. sent is not modified. Skipped parts are kept in all with
// status unchanged, for `session prompt`.
func dedupe(cfg *config.Config, sent map[string]string, parts []PromptPart) (send, all []PromptPart, next map[string]string) {
	mode := map[string]string{}
	for _, r := range cfg.PromptRules {
		mode[r.Name] = r.Inject
	}
	next = make(map[string]string, len(sent))
	for k, v := range sent {
		next[k] = v
	}
	var removed []PromptPart
	for _, name := range sortedNames(sent) {
		if _, ok := mode[name]; !ok && name != poolKey {
			removed = append(removed, PromptPart{
				Source: "removed:" + name, Status: injectRemoved,
				Text: fmt.Sprintf("Prompt rule %q no longer applies.", name),
			})
			delete(next, name)
		}
	}
	// Removal notices go just before the body.
	for i, p := range parts {
		if i == len(parts)-1 {
			send = append(send, removed...)
			all = append(all, removed...)
		}
		name, ok := strings.CutPrefix(p.Source, "rule:")
		if p.Source == "pool" {
			name, ok = poolKey, true
		}
		if !ok {
			send = append(send, p)
			all = append(all, p)
			continue
		}
		h := hashText(p.Text)
		prev, had := sent[name]
		switch {
		case mode[name] == config.InjectAlways:
			p.Status = injectAlways
		case !had:
			p.Status = injectNew
		case prev == h || mode[name] == config.InjectOnce:
			p.Status = injectUnchanged
		default:
			p.Status = injectChanged
			if name == poolKey {
				p.Text = "Updated pool instructions:\n\n" + p.Text
			} else {
				p.Text = fmt.Sprintf("Updated prompt rule %q:\n\n%s", name, p.Text)
			}
		}
		all = append(all, p)
		if p.Status == injectUnchanged {
			continue
		}
		if mode[name] != config.InjectOnce || !had {
			next[name] = h
		}
		send = append(send, p)
	}
	return send, all, next
}

func sortedNames(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// recordInjected adds a turn's rule bytes to the injected-bytes metric,
// by status (new, changed, always, unchanged for skipped, removed).
func (c *Controller) recordInjected(pool string, all []PromptPart) {
	if c.metrics == nil {
		return
	}
	for _, p := range all {
		if p.Status == "" {
			continue
		}
		c.metrics.InjectedBytes.Add(c.ctx, int64(len(p.Text)), metric.WithAttributes(
			telemetry.AttrPool.String(pool), telemetry.AttrReason.String(p.Status)))
	}
}
