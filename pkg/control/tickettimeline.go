package control

import (
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// The timeline (#1103) draws a trace as bars on one time axis. It is built
// from history the store already keeps: each ticket's rows are replayed to
// find the state it was in over time. Every transition has a row: create
// (with its labels), claim, release, lease_expired, handoff, close, reopen,
// label_add, label_remove, and update (fields.wait, fields.state,
// defer_until). A ticket with no history, or whose history does not start
// at its creation, gets an "unknown" span and its row is marked partial.

// maxCommentMarkers bounds the comment ticks on one bar: the newest.
const maxCommentMarkers = 40

// tlDiff is the part of a history diff the timeline reads.
type tlDiff struct {
	Reason     string    `json:"reason"`
	DeferUntil time.Time `json:"defer_until"`
	Resolution string    `json:"resolution"`
	Label      string    `json:"label"`
	Labels     []string  `json:"labels"`
	Failures   *int      `json:"failures"` // set when the event counted a failed attempt
	Fields     *struct {
		State string       `json:"state"`
		Wait  *domain.Wait `json:"wait"`
	} `json:"fields"`
}

// humanLabels put a ticket on hold for a person.
var humanLabels = []string{needsHumanLabel, questionLabel, proposedLabel}

// ticketTimeline returns the bars, nesting and arrows of the graph around
// a ticket.
func (c *Controller) ticketTimeline(ctx context.Context, sc scope, p rpc.GraphParams) (any, error) {
	g, threads, err := c.graphWalk(ctx, sc, p)
	if err != nil {
		return nil, err
	}
	return buildTimeline(c.config(), g, threads, c.now()), nil
}

func buildTimeline(cfg *config.Config, g rpc.GraphResult, threads map[string]store.Thread, now time.Time) rpc.TimelineResult {
	out := rpc.TimelineResult{Root: g.Root, Rows: []rpc.TimelineRow{}, Arrows: []rpc.TimelineArrow{}, Truncated: g.Truncated}
	tickets := map[string]domain.Ticket{}
	rows := map[string]*rpc.TimelineRow{}
	for _, n := range g.Nodes {
		if n.Kind == domain.KindSession || n.Kind == domain.KindMessage {
			continue
		}
		tickets[n.ID] = n.Ticket
		row := timelineRow(cfg, n.Ticket, threads[n.ID], now)
		rows[n.ID] = &row
	}
	// Nesting: parent_id first, else the ticket whose holder created it.
	parent := map[string]string{}
	for _, kind := range []string{"parent", "created"} {
		for _, r := range g.Relations {
			if r.Kind != kind || rows[r.From] == nil || rows[r.To] == nil {
				continue
			}
			if _, ok := parent[r.To]; !ok {
				parent[r.To] = r.From
			}
		}
	}
	kids := map[string][]string{}
	var roots []string
	for _, n := range g.Nodes {
		if rows[n.ID] == nil {
			continue
		}
		if p, ok := parent[n.ID]; ok {
			kids[p] = append(kids[p], n.ID)
		} else {
			roots = append(roots, n.ID)
		}
	}
	byAge := func(a, b string) int {
		if d := rows[a].CreatedAt.Compare(rows[b].CreatedAt); d != 0 {
			return d
		}
		return strings.Compare(a, b)
	}
	slices.SortFunc(roots, byAge)
	seen := map[string]bool{}
	var walk func(id string, depth int)
	walk = func(id string, depth int) {
		if seen[id] {
			return
		}
		seen[id] = true
		row := rows[id]
		row.Depth, row.Parent = depth, parent[id]
		out.Rows = append(out.Rows, *row)
		slices.SortFunc(kids[id], byAge)
		for _, k := range kids[id] {
			walk(k, depth+1)
		}
	}
	for _, id := range roots {
		walk(id, 0)
	}
	// A cycle of created-by links has no root: list those flat.
	var rest []string
	for id := range rows {
		if !seen[id] {
			rest = append(rest, id)
		}
	}
	slices.SortFunc(rest, byAge)
	for _, id := range rest {
		walk(id, 0)
	}

	seenArrow := map[rpc.TimelineArrow]bool{}
	add := func(a rpc.TimelineArrow) {
		if !seenArrow[a] {
			seenArrow[a] = true
			out.Arrows = append(out.Arrows, a)
		}
	}
	for _, r := range g.Relations {
		blocker, blocked := tickets[r.To], tickets[r.From]
		if rows[r.To] == nil || rows[r.From] == nil {
			continue
		}
		a := rpc.TimelineArrow{From: blocker.ID, To: blocked.ID, Kind: r.Kind}
		switch r.Kind {
		case "blocks", "waits_for":
			a.State = arrowState(blocker, rows[blocked.ID].Started)
		case "wait":
			a.State = r.State
		case "tracks":
		default:
			continue
		}
		add(a)
	}
	return out
}

// arrowState is how a blocking edge stands. The blocked ticket started
// at start (zero: not yet). A blocker step that closed without finishing
// still blocks.
func arrowState(blocker domain.Ticket, start time.Time) string {
	open := blocker.Status != domain.StatusClosed || blocker.Kind == domain.KindStep && !domain.StepFinished(blocker)
	switch {
	case open && start.IsZero():
		return rpc.ArrowBlocking
	case open, !start.IsZero() && start.Before(blocker.ClosedAt):
		return rpc.ArrowConflict
	}
	return rpc.ArrowMet
}

// tlState is what the replay knows about a ticket after an event.
type tlState struct {
	closed     bool
	held       bool
	holder     string
	wait       *domain.Wait
	labels     []string
	deferUntil time.Time
	runState   string
}

// tlBuilder turns state changes into segments.
type tlBuilder struct {
	cfg      *config.Config
	t        domain.Ticket
	comments []domain.Comment
	st       tlState
	segs     []rpc.TimelineSegment
	open     bool
	key      string       // identifies the open segment
	segWait  *domain.Wait // the wait the open segment shows
	marks    []rpc.TimelineMarker
}

func timelineRow(cfg *config.Config, t domain.Ticket, th store.Thread, now time.Time) rpc.TimelineRow {
	row := rpc.TimelineRow{
		ID: t.ID, Kind: t.Kind, Title: t.Title, Pool: t.Pool, Status: t.Status, Resolution: t.Resolution,
		CreatedAt: t.CreatedAt, ClosedAt: t.ClosedAt, Segments: []rpc.TimelineSegment{}, Markers: []rpc.TimelineMarker{},
	}
	hist := th.History
	closed := t.Status == domain.StatusClosed
	end := now
	if closed && !t.ClosedAt.IsZero() {
		end = t.ClosedAt
	}
	b := &tlBuilder{cfg: cfg, t: t, comments: th.Comments}
	b.st.runState = domain.RunRunning
	if len(hist) == 0 || hist[0].Op != "create" {
		row.Partial = true
		first := end
		if len(hist) > 0 && hist[0].At.Before(end) {
			first = hist[0].At
		}
		if first.After(t.CreatedAt) {
			b.segs = append(b.segs, rpc.TimelineSegment{State: rpc.TimelineUnknown, Start: t.CreatedAt, End: first})
		}
	}
	lastClose := -1
	for i, h := range hist {
		if h.Op == "close" {
			lastClose = i
		}
	}
	var lastActor string
	for i, h := range hist {
		b.tick(h.At)
		var d tlDiff
		_ = json.Unmarshal(h.Diff, &d)
		fail, why := false, ""
		switch h.Op {
		case "create":
			b.st.labels = slices.Clone(d.Labels)
		case "claim":
			b.st.held, b.st.holder, b.st.deferUntil = true, h.Actor, time.Time{}
			if row.Started.IsZero() {
				row.Started = h.At
			}
			if lastActor != "" && lastActor != h.Actor {
				row.Markers = append(row.Markers, rpc.TimelineMarker{At: h.At, Kind: rpc.MarkerSession, Actor: h.Actor, Label: "taken over by " + h.Actor})
			}
			lastActor = h.Actor
		case "release", "lease_expired":
			b.st.held, b.st.holder = false, ""
			b.st.deferUntil = d.DeferUntil
			switch {
			case h.Op == "lease_expired":
				fail, why = true, "lease expired"
			case d.Failures != nil:
				fail, why = true, d.Reason
			}
		case "handoff":
			b.st.held, b.st.holder = false, ""
		case "close":
			b.st.held, b.st.holder = false, ""
			fail, why = d.Failures != nil, "closed failed"
			// A failed, rebase or request_changes close of a step reopens
			// it unless it is the last close and the step stayed closed.
			reopens := t.Kind == domain.KindStep && slices.Contains([]string{"failed", domain.ResolutionRebase, domain.ResolutionRequestChanges}, d.Resolution)
			b.st.closed = !reopens || i == lastClose && closed
		case "reopen":
			b.st.closed = false
		case "update":
			if !d.DeferUntil.IsZero() {
				b.st.deferUntil = d.DeferUntil
			}
			if d.Fields != nil {
				b.st.wait = d.Fields.Wait
				if d.Fields.State != "" {
					b.st.runState = d.Fields.State
				}
			}
		case "label_add":
			if !slices.Contains(b.st.labels, d.Label) {
				b.st.labels = append(b.st.labels, d.Label)
			}
		case "label_remove":
			b.st.labels = slices.DeleteFunc(b.st.labels, func(l string) bool { return l == d.Label })
		}
		b.apply(h.At, fail, why)
	}
	b.tick(end)
	// The ticket itself wins over the replay (a compacted history): a
	// closed ticket's bar ends at its close.
	if closed {
		b.st.closed = true
	}
	b.apply(end, false, "")
	row.Segments = b.segs
	row.Markers = append(row.Markers, b.marks...)
	row.Markers = append(row.Markers, commentMarkers(th.Comments)...)
	slices.SortStableFunc(row.Markers, func(a, c rpc.TimelineMarker) int { return a.At.Compare(c.At) })
	return row
}

// commentMarkers are the ticks for comments (the newest maxCommentMarkers)
// and for the PR a step announced.
func commentMarkers(comments []domain.Comment) []rpc.TimelineMarker {
	var out []rpc.TimelineMarker
	for _, cm := range comments {
		if m := stepPRComment.FindStringSubmatch(cm.Body); m != nil {
			out = append(out, rpc.TimelineMarker{At: cm.CreatedAt, Kind: rpc.MarkerPR, Label: "PR opened", Actor: cm.Author, URL: m[1]})
			continue
		}
		first, _, _ := strings.Cut(strings.TrimSpace(cm.Body), "\n")
		if r := []rune(first); len(r) > 80 {
			first = string(r[:79]) + "…"
		}
		out = append(out, rpc.TimelineMarker{At: cm.CreatedAt, Kind: rpc.MarkerComment, Label: first, Actor: cm.Author})
	}
	n := 0
	for i := len(out) - 1; i >= 0; i-- {
		if out[i].Kind != rpc.MarkerComment {
			continue
		}
		if n++; n > maxCommentMarkers {
			out = slices.Delete(out, i, i+1)
		}
	}
	return out
}

// tick applies a deferral that ended before to.
func (b *tlBuilder) tick(to time.Time) {
	if u := b.st.deferUntil; b.open && !u.IsZero() && !u.After(to) {
		b.st.deferUntil = time.Time{}
		b.apply(u, false, "")
	}
}

// current is the segment the state calls for at at, nil for none (closed).
func (b *tlBuilder) current(at time.Time) (*rpc.TimelineSegment, *domain.Wait) {
	s := &b.st
	switch {
	case s.closed:
		return nil, nil
	case s.held:
		return &rpc.TimelineSegment{State: rpc.TimelineWorking, Actor: s.holder}, nil
	case b.t.Kind == domain.KindRun && s.runState == domain.RunNeedsHuman:
		return &rpc.TimelineSegment{State: rpc.TimelineHuman, Detail: "run needs a human"}, nil
	case b.t.Kind == domain.KindRun || b.t.Kind == domain.KindBatch:
		return &rpc.TimelineSegment{State: rpc.TimelineWorking}, nil
	}
	for _, l := range humanLabels {
		if slices.Contains(s.labels, l) {
			return &rpc.TimelineSegment{State: rpc.TimelineHuman, Detail: l}, nil
		}
	}
	if w := s.wait; w != nil {
		seg := &rpc.TimelineSegment{State: rpc.TimelineWaiting, Detail: w.Reason, Wait: b.waitConds(*w, "")}
		// A wait on a time alone is a deferral until that time.
		timed := len(w.Conds) > 0
		for _, cd := range w.Conds {
			timed = timed && cd.Type == domain.WaitTime
			if cd.Type != domain.WaitTime {
				continue
			}
			// The wait ends at the earliest time with mode any, else the latest.
			if earlier := cd.At.Before(seg.Until); seg.Until.IsZero() || earlier == (w.Mode == domain.WaitAny) {
				seg.Until = cd.At
			}
		}
		if timed {
			seg.State = rpc.TimelineDeferred
		} else {
			seg.Until = time.Time{}
		}
		return seg, w
	}
	if s.deferUntil.After(at) {
		return &rpc.TimelineSegment{State: rpc.TimelineDeferred, Until: s.deferUntil}, nil
	}
	return &rpc.TimelineSegment{State: rpc.TimelineQueued}, nil
}

// apply moves the open segment to the one the state calls for at at. fail
// marks the claim that ends here as a failed attempt.
func (b *tlBuilder) apply(at time.Time, fail bool, why string) {
	next, w := b.current(at)
	key := ""
	if next != nil {
		next.Start = at
		key = next.State + "|" + next.Actor + "|" + next.Detail + "|" + next.Until.String()
		if w != nil {
			key += "|" + w.Since.String()
		}
	}
	if b.open && next != nil && key == b.key && !fail {
		// The same span: keep the freshest condition states.
		b.segs[len(b.segs)-1].Wait = next.Wait
		return
	}
	b.close(at, fail, why)
	if next != nil {
		b.segs = append(b.segs, *next)
		b.key, b.open, b.segWait = key, true, w
	}
}

// close ends the open segment at at. A claim that ends in a failed attempt
// becomes a retry span. A wait that is gone records how it ended.
func (b *tlBuilder) close(at time.Time, fail bool, why string) {
	if !b.open {
		return
	}
	b.open = false
	i := len(b.segs) - 1
	seg := &b.segs[i]
	seg.End = at
	switch {
	case seg.State == rpc.TimelineWorking && fail:
		seg.State, seg.Detail = rpc.TimelineRetry, why
	case b.segWait != nil && (b.st.wait == nil || !b.st.wait.Since.Equal(b.segWait.Since)):
		// The wait is over (a different cause, such as a human label, leaves
		// it set and the span without an end reason).
		reason := endReason(waitEpisode{wait: *b.segWait, end: at}, b.comments)
		seg.Ended = endedState(*b.segWait, domain.WaitCond{Met: true}, reason)
		seg.Wait = b.waitConds(*b.segWait, reason)
		b.merged(*b.segWait, seg, at)
	}
	b.segWait = nil
	if !seg.End.After(seg.Start) {
		b.segs = slices.Delete(b.segs, i, i+1)
	}
}

// waitConds lists a wait's conditions with their state: waiting or met
// while live, and once over, how it ended (see endedState).
func (b *tlBuilder) waitConds(w domain.Wait, why string) []rpc.TimelineWait {
	out := make([]rpc.TimelineWait, 0, len(w.Conds))
	for _, cd := range w.Conds {
		tw := rpc.TimelineWait{Type: cd.Type, State: waitStateWaiting}
		if cd.Met {
			tw.State = waitStateMet
		}
		if why != "" {
			tw.State = endedState(w, cd, why)
		}
		target, ext := waitTarget(b.cfg, b.t.Project, cd)
		switch {
		case ext != nil:
			tw.Label, tw.URL = ext.Label, ext.URL
		case target != "":
			tw.Label, tw.Target = cd.Type+" "+target, target
		case cd.Type == domain.WaitTime:
			tw.Label = "time " + cd.At.UTC().Format(time.RFC3339)
		default:
			tw.Label = cd.Type
		}
		out = append(out, tw)
	}
	return out
}

// merged marks the merge of a PR a finished wait was waiting for.
func (b *tlBuilder) merged(w domain.Wait, seg *rpc.TimelineSegment, at time.Time) {
	if seg.Ended != waitStateMet {
		return
	}
	for i, cd := range w.Conds {
		if cd.Type == domain.WaitPR && cd.State == "merged" && seg.Wait[i].State == waitStateMet {
			pr := strconv.Itoa(cd.PR)
			b.marks = append(b.marks, rpc.TimelineMarker{At: at, Kind: rpc.MarkerMerged, Label: cd.Repo + " PR #" + pr + " merged", URL: "https://github.com/" + cd.Repo + "/pull/" + pr})
		}
	}
}
