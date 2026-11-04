package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// runDigest posts the attention list at each Digest time (#185).
func (g *Gateway) runDigest(ctx context.Context) {
	for {
		now := time.Now().UTC()
		t := time.NewTimer(g.cfg.Digest.Next(now).Sub(now))
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
		g.digest(ctx)
	}
}

// digest posts the attention list as a new thread in the escalation
// room. An empty list posts nothing: a healthy day needs no reading.
func (g *Gateway) digest(ctx context.Context) {
	res, err := g.ctl.Attention(ctx)
	if err != nil {
		g.log.Warn("gateway: attention", "err", err)
		return
	}
	if len(res.Items) == 0 {
		return
	}
	if _, err := g.ad.Open(ctx, g.settings().EscalationRoom, digestText(res.Items, time.Now())); err != nil {
		g.log.Warn("gateway: digest", "err", err)
	}
}

func digestText(items []rpc.AttentionItem, now time.Time) string {
	var b strings.Builder
	fmt.Fprintf(&b, "jobshop: %d waiting on a human (oldest first, `jobshop attention`):\n", len(items))
	for _, it := range items {
		ref := ""
		if it.Repo != "" || it.Ref != "" {
			ref = " (" + strings.TrimSpace(it.Repo+" "+it.Ref) + ")"
		}
		fmt.Fprintf(&b, "- %s %s: %s%s, %s\n", it.Kind, it.ID, it.Title, ref, waited(now.Sub(it.Since)))
	}
	return b.String()
}

func waited(d time.Duration) string {
	switch {
	case d >= 48*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
