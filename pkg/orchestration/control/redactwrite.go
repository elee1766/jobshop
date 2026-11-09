package control

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/elee1766/jobshop/pkg/support/redact"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Redaction on write (#574): comments, reports, learnings and artifacts
// reach Desk, gateway digests and memory curation (committed to a public
// repo), so the controller masks secrets before storing them. The
// session's own env is not known here; the configured rules cover tokens.

// maxRedactFile is the largest artifact attachment scanned. Bigger files
// and binary ones are copied as they are.
const maxRedactFile = 1 << 20

// secretScan scans an outbound write's texts with the configured rules
// (#578), so the CLI's check (#574) uses the host's redact config.
// Open to every caller: it reveals rule ids and lines, not secrets.
func (c *Controller) secretScan(_ context.Context, _ scope, p rpc.SecretScanParams) (any, error) {
	sc := c.scanner()
	res := rpc.SecretScanResult{Findings: []rpc.SecretFinding{}}
	for _, t := range p.Texts {
		for _, f := range sc.Find(t.Text) {
			res.Findings = append(res.Findings, rpc.SecretFinding{From: t.From, Rule: f.Rule, Line: f.Line})
		}
	}
	return res, nil
}

// redactWrite masks secrets in each of texts, in place. When any changed
// it records one activity item naming what (never the secret).
func (c *Controller) redactWrite(sc scope, what, target string, texts ...*string) {
	changed := false
	red := c.redactor(nil)
	for _, s := range texts {
		if s == nil || *s == "" {
			continue
		}
		if r := red.Text(*s); r != *s {
			*s, changed = r, true
		}
	}
	if changed {
		c.recordRedacted(sc, what, target)
	}
}

// recordRedacted adds a redacted: true item to the activity feed.
func (c *Controller) recordRedacted(sc scope, what, target string) {
	actor := sc.session
	if actor == "" {
		actor = sc.actor
	}
	now := c.now()
	c.addActivity(rpc.ActivityItem{
		ID:     fmt.Sprintf("redact:%s:%d:%d", actor, now.UnixNano(), ghCallSeq.Add(1)),
		Source: "jobshop", Kind: "redact", Actor: actor, Bot: true, At: now,
		Title: strings.TrimSpace(fmt.Sprintf("secrets masked in %s %s", what, target)), Redacted: true,
	})
}

// redactFile masks secrets in a copied artifact attachment, if it is
// text and at most maxRedactFile. It reports whether anything changed.
func redactFile(red *redact.Redactor, path string) (bool, error) {
	st, err := os.Stat(path)
	if err != nil || st.Size() > maxRedactFile {
		return false, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if isBinary(b) {
		return false, nil
	}
	r := red.Text(string(b))
	if r == string(b) {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(r), 0o640)
}

// isBinary is a NUL byte in the first 8KB, as git decides.
func isBinary(b []byte) bool {
	if len(b) > 8192 {
		b = b[:8192]
	}
	for _, c := range b {
		if c == 0 {
			return true
		}
	}
	return false
}
