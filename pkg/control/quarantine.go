package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/host"
	"github.com/elee1766/jobshop/pkg/store"
)

const (
	// crashReleaseReason marks a release by a crashed session in history,
	// so crashes per ticket can be counted.
	crashReleaseReason = "session crashed"
	quarantineLabel    = "quarantined"
	quarantineCrashes  = 3
	quarantineFor      = time.Hour
	poolCrashes        = 3
	poolCrashWindow    = 10 * time.Minute
	poolPause          = 15 * time.Minute
)

// crashRelease releases what a crashed session held with crashReleaseReason
// and quarantines tickets that have now crashed quarantineCrashes sessions.
func (c *Controller) crashRelease(ctx context.Context, sessionID string) {
	held, err := c.store.List(ctx, store.Filter{Holder: sessionID, Status: store.StatusInProgress})
	if err != nil {
		return
	}
	for _, t := range held {
		cl := store.Claim{TicketID: t.ID, Holder: sessionID, Fence: t.Fence}
		if _, err := c.store.Release(ctx, cl, crashReleaseReason); err != nil {
			c.log.Warn("crash release", "ticket", t.ID, "err", err)
			continue
		}
		c.maybeQuarantine(ctx, t.ID)
	}
}

// maybeQuarantine defers a ticket for quarantineFor and labels it once
// quarantineCrashes crashed sessions have released it since it was last
// quarantined (removing the label resets the count).
func (c *Controller) maybeQuarantine(ctx context.Context, id string) {
	hist, err := c.store.History(ctx, id)
	if err != nil {
		return
	}
	var sessions []string
	for _, h := range hist {
		var d struct {
			Reason string `json:"reason"`
			Label  string `json:"label"`
		}
		_ = json.Unmarshal(h.Diff, &d)
		switch {
		case (h.Op == "label_add" || h.Op == "label_remove") && d.Label == quarantineLabel:
			sessions = nil
		case h.Op == "release" && d.Reason == crashReleaseReason:
			sessions = append(sessions, h.Actor)
		}
	}
	if len(sessions) < quarantineCrashes {
		return
	}
	until := c.now().Add(quarantineFor)
	if _, err := c.store.Update(ctx, "controller", id, store.Update{DeferUntil: &until}); err != nil {
		c.log.Error("quarantine", "ticket", id, "err", err)
		return
	}
	if err := c.store.AddLabel(ctx, "controller", id, quarantineLabel); err != nil {
		c.log.Error("quarantine label", "ticket", id, "err", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Quarantined until %s: released by %d crashed sessions.\n", until.UTC().Format(time.RFC3339), len(sessions))
	for _, s := range sessions {
		reason := "unknown"
		if ex, ok := host.ReadExit(c.root.SessionDir(s)); ok {
			reason = fmt.Sprintf("%s (code %d %s)", ex.Reason, ex.ExitCode, ex.Signal)
		}
		fmt.Fprintf(&b, "- %s: %s\n", s, reason)
	}
	b.WriteString("Undo with `ticket update " + id + " --defer 0` and `label rm " + id + " quarantined`.")
	_, _ = c.store.AddComment(ctx, "controller", id, b.String())
	c.log.Warn("quarantined ticket", "ticket", id, "sessions", sessions)
}

// poolCrashed records a crashed session of pool and pauses the pool when
// poolCrashes happened within poolCrashWindow.
func (c *Controller) poolCrashed(pool string) {
	if pool == "" {
		return
	}
	now := c.now()
	c.mu.Lock()
	defer c.mu.Unlock()
	keep := c.poolCrashLog[pool][:0]
	for _, t := range c.poolCrashLog[pool] {
		if now.Sub(t) < poolCrashWindow {
			keep = append(keep, t)
		}
	}
	keep = append(keep, now)
	c.poolCrashLog[pool] = keep
	if len(keep) >= poolCrashes {
		c.poolPausedUntil[pool] = now.Add(poolPause)
		c.poolCrashLog[pool] = nil
		c.log.Warn("pool quarantined: sessions crash-looping", "pool", pool, "crashes", len(keep), "until", now.Add(poolPause))
	}
}

// poolPaused reports whether the supervisor should start no sessions for pool.
func (c *Controller) poolPaused(pool string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now().Before(c.poolPausedUntil[pool])
}
