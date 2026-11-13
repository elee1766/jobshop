// Package scaleplan decides how many sessions a pool wants and which to
// cordon, lift or start (#281), split out of pkg/orchestration/control by the JOB-155
// audit. It is pure: values in, a plan out. The controller reads the
// sessions, holds the history under its lock and applies the plan.
package scaleplan

import (
	"cmp"
	"slices"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
)

// DefaultScaleDownAfter is the stabilization window when the pool sets
// none.
const DefaultScaleDownAfter = 10 * time.Minute

// Sample is one raw recommendation, kept for the stabilization window.
type Sample struct {
	At time.Time
	N  int
}

// Desired is clamp(min, max, demand) held at its highest value over
// the scale-down window, so a burst does not start and stop sessions in a
// loop. Scale up is immediate. hist is the pool's samples, pruned and
// appended in place.
func Desired(p config.Pool, demand int, now time.Time, hist []Sample) (int, []Sample) {
	rec := min(max(demand, p.MinSessions), p.MaxSessions)
	window := config.Duration(p.ScaleDownAfter, DefaultScaleDownAfter)
	hist = slices.DeleteFunc(hist, func(s Sample) bool { return now.Sub(s.At) >= window })
	hist = append(hist, Sample{At: now, N: rec})
	out := 0
	for _, s := range hist {
		out = max(out, s.N)
	}
	// A reload that lowers max_sessions applies at once, not after the window.
	return min(out, p.MaxSessions), hist
}

// Session is the reconciler's view of one open session.
type Session struct {
	ID        string
	Idle      bool // holds no ticket
	LastClaim time.Time
	Created   time.Time
	Cordon    string // store.Cordon reason, "" when in rotation
	// Repo is the pool repo key the session serves (JOB-207), "" for the
	// pool's default repo.
	Repo string
	// Exempt: a session resumed from a park (JOB-189). It holds a slot but
	// is never cordoned for being surplus, so a pool one over its
	// max_sessions does not cordon it.
	Exempt bool
}

// Plan is what one pool needs this tick.
type Plan struct {
	Cordon   []string // move to CordonScale
	Uncordon []string // scale cordons to lift
	Start    int      // new sessions wanted, before brakes and the host cap
}

// ForPool compares desired with the sessions in rotation. Surplus is
// cordoned idle first, then least recent claim, then oldest host. A
// shortfall first lifts scale cordons (the session is already warm) and
// starts the rest. Manual and drain cordons are never lifted here.
func ForPool(desired int, sessions []Session) Plan {
	var active, scaled []Session
	exempt := 0
	for _, s := range sessions {
		if s.Exempt && s.Cordon == "" {
			exempt++
			continue
		}
		switch s.Cordon {
		case "":
			active = append(active, s)
		case domain.CordonScale:
			scaled = append(scaled, s)
		}
	}
	var p Plan
	if n := len(active) - desired; n > 0 {
		slices.SortStableFunc(active, func(a, b Session) int {
			if a.Idle != b.Idle {
				if a.Idle {
					return -1
				}
				return 1
			}
			return cmp.Or(a.LastClaim.Compare(b.LastClaim), a.Created.Compare(b.Created), cmp.Compare(a.ID, b.ID))
		})
		for _, s := range active[:n] {
			p.Cordon = append(p.Cordon, s.ID)
		}
		return p
	}
	short := desired - len(active) - exempt
	// Lift the busiest first: a working session is the furthest from ending.
	slices.SortStableFunc(scaled, func(a, b Session) int {
		if a.Idle != b.Idle {
			if a.Idle {
				return 1
			}
			return -1
		}
		return b.LastClaim.Compare(a.LastClaim)
	})
	for _, s := range scaled {
		if short <= 0 {
			break
		}
		p.Uncordon = append(p.Uncordon, s.ID)
		short--
	}
	p.Start = max(short, 0)
	return p
}

// CapRequest is one pool's wish to start sessions under the host cap.
type CapRequest struct {
	Pool     string
	Priority int
	Start    int
}

// CapStarts grants starts in priority order (lower first, then name) until
// open plus granted reaches limit. limit 0 means no cap.
func CapStarts(reqs []CapRequest, open, limit int) map[string]int {
	out := map[string]int{}
	slices.SortStableFunc(reqs, func(a, b CapRequest) int {
		return cmp.Or(cmp.Compare(a.Priority, b.Priority), cmp.Compare(a.Pool, b.Pool))
	})
	room := limit - open
	for _, r := range reqs {
		n := r.Start
		if limit > 0 {
			n = max(min(n, room), 0)
			room -= n
		}
		if n > 0 {
			out[r.Pool] = n
		}
	}
	return out
}
