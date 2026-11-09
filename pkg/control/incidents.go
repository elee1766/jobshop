package control

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// incidentCap bounds the in-memory WARN/ERROR log. A crash loop can log
// fast; old records drop first.
const incidentCap = 5000

// logRecord is one controller WARN/ERROR line kept for incident reports.
type logRecord struct {
	At    time.Time
	Level slog.Level
	Msg   string
	Attrs map[string]string
}

// incidentLog keeps recent WARN/ERROR records in memory. It is lost on
// restart, which is fine for a daily planner pass: the controller restarts
// on deploy and the store keeps sessions and steps.
type incidentLog struct {
	mu   sync.Mutex
	recs []logRecord
}

func (l *incidentLog) add(r logRecord) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.recs) >= incidentCap {
		l.recs = slices.Delete(l.recs, 0, len(l.recs)-incidentCap+1)
	}
	l.recs = append(l.recs, r)
}

func (l *incidentLog) since(t time.Time) []logRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []logRecord
	for _, r := range l.recs {
		if !r.At.Before(t) {
			out = append(out, r)
		}
	}
	return out
}

// incidentHandler tees WARN and above into an incidentLog.
type incidentHandler struct {
	slog.Handler
	log   *incidentLog
	attrs []slog.Attr
}

func (h incidentHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelWarn {
		rec := logRecord{At: r.Time, Level: r.Level, Msg: r.Message, Attrs: map[string]string{}}
		for _, a := range h.attrs {
			rec.Attrs[a.Key] = a.Value.String()
		}
		r.Attrs(func(a slog.Attr) bool {
			rec.Attrs[a.Key] = a.Value.String()
			return true
		})
		h.log.add(rec)
	}
	return h.Handler.Handle(ctx, r)
}

// Enabled always accepts WARN so the report sees it even if the base
// handler filters it.
func (h incidentHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= slog.LevelWarn || h.Handler.Enabled(ctx, l)
}

func (h incidentHandler) WithAttrs(as []slog.Attr) slog.Handler {
	return incidentHandler{h.Handler.WithAttrs(as), h.log, append(slices.Clip(h.attrs), as...)}
}

// WithGroup drops group names from captured keys; the report groups by
// message, so key prefixes do not matter.
func (h incidentHandler) WithGroup(n string) slog.Handler {
	return incidentHandler{h.Handler.WithGroup(n), h.log, h.attrs}
}

// groupLogs groups records by level and message, most frequent first.
// Sample keeps the attrs of the latest record of each group.
func groupLogs(recs []logRecord) []rpc.LogGroup {
	idx := map[string]int{}
	var out []rpc.LogGroup
	for _, r := range recs {
		k := r.Level.String() + "\x00" + r.Msg
		i, ok := idx[k]
		if !ok {
			i = len(out)
			idx[k] = i
			out = append(out, rpc.LogGroup{Level: r.Level.String(), Message: r.Msg, First: r.At})
		}
		g := &out[i]
		g.Count++
		if r.At.Before(g.First) {
			g.First = r.At
		}
		if !r.At.Before(g.Last) {
			g.Last = r.At
			g.Sample = r.Attrs
		}
	}
	slices.SortStableFunc(out, func(a, b rpc.LogGroup) int {
		if a.Count != b.Count {
			return b.Count - a.Count
		}
		return b.Last.Compare(a.Last)
	})
	return out
}

// incidentsLimit bounds the session and step scans.
const incidentsLimit = 2000

func (c *Controller) reportIncidents(ctx context.Context, sc scope, p rpc.IncidentsParams) (any, error) {
	if err := c.requireGrant(sc, "report-incidents", "report incidents"); err != nil {
		return nil, err
	}
	since := p.Since
	if since <= 0 {
		since = 24 * time.Hour
	}
	from := c.now().Add(-since)
	res := rpc.IncidentsResult{Since: from, Logs: groupLogs(c.incidents.since(from))}
	if res.Logs == nil {
		res.Logs = []rpc.LogGroup{}
	}
	res.Sessions, res.Steps = []rpc.SessionIncident{}, []rpc.StepIncident{}

	sessions, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusClosed, Order: store.OrderNewest, Limit: incidentsLimit})
	if err != nil {
		return nil, err
	}
	for _, s := range sessions {
		if s.ClosedAt.Before(from) || (s.Resolution != resolutionCrashed && s.Resolution != resolutionStuck) {
			continue
		}
		res.Sessions = append(res.Sessions, rpc.SessionIncident{
			ID: s.ID, Pool: s.Pool, Resolution: s.Resolution, ClosedAt: s.ClosedAt, Reason: c.lastComment(ctx, s.ID),
		})
	}

	steps, err := c.store.List(ctx, store.Filter{Kind: domain.KindStep, Order: store.OrderNewest, Limit: incidentsLimit})
	if err != nil {
		return nil, err
	}
	for _, s := range steps {
		if s.UpdatedAt.Before(from) {
			continue
		}
		f, err := domain.DecodeFields[domain.StepFields](s)
		if err != nil || f.Failures == 0 {
			continue
		}
		res.Steps = append(res.Steps, rpc.StepIncident{
			ID: s.ID, Run: f.Run, Step: f.Step, Title: s.Title, Status: s.Status, Resolution: s.Resolution,
			Failures: f.Failures, Retries: f.Retries, LastError: c.lastComment(ctx, s.ID),
		})
	}
	return res, nil
}

func (c *Controller) lastComment(ctx context.Context, id string) string {
	cs, err := store.Comments(ctx, c.store, id)
	if err != nil || len(cs) == 0 {
		return ""
	}
	last := cs[len(cs)-1]
	return fmt.Sprintf("%s: %s", last.Author, last.Body)
}
