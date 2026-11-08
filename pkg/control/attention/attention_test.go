package attention_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control/attention"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

var t0 = time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)

// fake is an attention.Source over fixed tickets.
type fake struct {
	byLabel map[string][]store.Ticket
	runs    []store.Ticket
	waits   []attention.Waiting
	sub     []rpc.AttentionItem
	stale   []rpc.AttentionItem
	health  attention.Health
	tickets map[string]store.Ticket
	steps   map[string][]store.Ticket
}

func (f fake) Now() time.Time { return t0 }
func (f fake) Ticket(_ context.Context, id string) (store.Ticket, error) {
	if t, ok := f.tickets[id]; ok {
		return t, nil
	}
	return store.Ticket{}, store.ErrNotFound
}

func (f fake) Steps(_ context.Context, run string) ([]store.Ticket, error) {
	return f.steps[run], nil
}

func (f fake) LabeledOpen(_ context.Context, label string) ([]store.Ticket, error) {
	return f.byLabel[label], nil
}
func (f fake) NeedsHumanRuns(context.Context) ([]store.Ticket, error)  { return f.runs, nil }
func (f fake) TimedOut(context.Context) ([]attention.Waiting, error)   { return f.waits, nil }
func (f fake) Subsystems(context.Context) ([]rpc.AttentionItem, error) { return f.sub, nil }
func (f fake) Stale(context.Context) ([]rpc.AttentionItem, error)      { return f.stale, nil }
func (f fake) Health() attention.Health                                { return f.health }

func ticket(id string, at time.Duration, labels []string, meta string) store.Ticket {
	return store.Ticket{ID: id, Title: id, Labels: labels, Meta: json.RawMessage(meta), CreatedAt: t0.Add(at), UpdatedAt: t0.Add(at)}
}

func kinds(items []rpc.AttentionItem) []string {
	out := []string{}
	for _, it := range items {
		out = append(out, it.Kind+" "+it.ID)
	}
	return out
}

func TestHealthyIsEmpty(t *testing.T) {
	t.Parallel()
	items, err := attention.Build(t.Context(), fake{})
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.Empty(t, items)
}

// An escalation stands for its run and pool once: the run row and the
// pool row are not listed beside it, and a ticket with two labels is one row.
func TestEscalationDedup(t *testing.T) {
	t.Parallel()
	esc := ticket("esc", time.Minute, []string{attention.NeedsHumanLabel}, `{"run":{"id":"r1"},"pool":{"name":"dev"}}`)
	both := ticket("both", 2*time.Minute, []string{attention.NeedsHumanLabel, attention.QuestionLabel}, `{}`)
	f := fake{
		byLabel: map[string][]store.Ticket{attention.NeedsHumanLabel: {esc, both}, attention.QuestionLabel: {both}},
		runs: []store.Ticket{
			{ID: "r1", Kind: store.KindRun, UpdatedAt: t0},
			{ID: "r2", Kind: store.KindRun, UpdatedAt: t0.Add(time.Hour)},
		},
		health: attention.Health{Started: t0, Quarantined: map[string]string{"dev": "bad", "plan": "bad too"}},
	}
	items, err := attention.Build(t.Context(), f)
	require.NoError(t, err)
	assert.Equal(t, []string{"pool plan", "ticket esc", "ticket both", "run r2"}, kinds(items))
}

func TestKindOfLabel(t *testing.T) {
	t.Parallel()
	f := fake{byLabel: map[string][]store.Ticket{
		attention.QuestionLabel: {ticket("q", 0, []string{attention.QuestionLabel}, `{}`)},
		attention.ProposedLabel: {
			ticket("p", time.Minute, []string{attention.ProposedLabel}, `{}`),
			ticket("held", 2*time.Minute, []string{attention.ProposedLabel, attention.NeedsHumanLabel}, `{}`),
		},
	}}
	f.byLabel[attention.NeedsHumanLabel] = []store.Ticket{f.byLabel[attention.ProposedLabel][1]}
	items, err := attention.Build(t.Context(), f)
	require.NoError(t, err)
	assert.Equal(t, []string{"question q", "decision p", "ticket held"}, kinds(items))
}

// A wait past its timeout replaces the ticket row it already has, with
// the owner and the timeout as the time it has waited since.
func TestTimedOutWaitReplacesTicketRow(t *testing.T) {
	t.Parallel()
	tk := ticket("w", 0, []string{attention.NeedsHumanLabel}, `{}`)
	f := fake{
		byLabel: map[string][]store.Ticket{attention.NeedsHumanLabel: {tk}},
		waits:   []attention.Waiting{{Ticket: tk, Wait: &store.Wait{Reason: "ci", Owner: "ann", Timeout: t0.Add(time.Hour)}}},
	}
	items, err := attention.Build(t.Context(), f)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, rpc.AttentionWait, items[0].Kind)
	assert.Equal(t, "w: wait past timeout: ci", items[0].Title)
	assert.Equal(t, "ann", items[0].Owner)
	assert.Equal(t, t0.Add(time.Hour), items[0].Since)
	assert.Equal(t, "ci", items[0].Summary)
	var titles []string
	for _, o := range items[0].Options {
		titles = append(titles, o.Title)
	}
	assert.Equal(t, []string{"Clear wait", "Extend one hour"}, titles)
}

func TestStaleListedOnce(t *testing.T) {
	t.Parallel()
	tk := ticket("s", 0, []string{attention.NeedsHumanLabel}, `{}`)
	f := fake{
		byLabel: map[string][]store.Ticket{attention.NeedsHumanLabel: {tk}},
		stale:   []rpc.AttentionItem{{Kind: rpc.AttentionStale, ID: "s"}, {Kind: rpc.AttentionStale, ID: "o", Since: t0.Add(time.Minute)}},
	}
	items, err := attention.Build(t.Context(), f)
	require.NoError(t, err)
	assert.Equal(t, []string{"ticket s", "stale o"}, kinds(items))
}

func TestHealthRows(t *testing.T) {
	t.Parallel()
	f := fake{health: attention.Health{
		Started: t0, ConfigErr: "bad pack", ConfigErrAt: t0.Add(time.Minute), WatchErr: "bad yaml", WatchErrAt: t0.Add(2 * time.Minute),
		Quarantined: map[string]string{"dev": "no dir"},
		Paused:      []attention.Pause{{Pool: "dev", Until: t0.Add(time.Hour), Why: "x"}, {Pool: "plan", Until: t0.Add(time.Hour), Why: "crash\nsecond line"}},
		PausedFor:   15 * time.Minute,
	}}
	items, err := attention.Build(t.Context(), f)
	require.NoError(t, err)
	assert.Equal(t, []string{"pool dev", "config config", "config-invalid config-invalid", "pool plan"}, kinds(items))
	plan := items[3]
	assert.Equal(t, t0.Add(45*time.Minute), plan.Since)
	assert.Equal(t, "pool plan paused until 2026-11-08T13:00:00Z, sessions crash-looping: crash", plan.Title)
}

func TestTicketItem(t *testing.T) {
	t.Parallel()
	q := ticket("q", 0, []string{attention.QuestionLabel}, `{"linear":{"workspace":"acme","issue":"ENG-5"},"github":{"repo":"acme/w"},"question":{"ticket":"t1","context":"why","options":[{"id":1,"title":"A"}]}}`)
	it := attention.TicketItem(rpc.AttentionQuestion, q)
	assert.Equal(t, "ENG-5", it.Ref)
	assert.Equal(t, "https://linear.app/acme/issue/ENG-5", it.URL)
	assert.Equal(t, "acme/w", it.Repo)
	assert.Equal(t, "why", it.Context)
	assert.Len(t, it.Options, 1)

	desk := ticket("d", 0, nil, `{"github":{"repo":"a/b","pr":7},"desk":{"attention_summary":"short","attention_context":"full text"}}`)
	desk.Body = "body"
	it = attention.TicketItem(rpc.AttentionTicket, desk)
	assert.Equal(t, "PR #7", it.Ref)
	assert.Equal(t, "https://github.com/a/b/pull/7", it.URL)
	assert.Equal(t, "short", it.Summary)
	assert.Equal(t, "full text", it.Context)

	run := ticket("r", time.Minute, nil, `{}`)
	run.UpdatedAt = t0.Add(time.Hour)
	assert.Equal(t, t0.Add(time.Hour), attention.TicketItem(rpc.AttentionRun, run).Since)
}

func TestSummaryOf(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "one two", attention.SummaryOf("\n one \n\n two \nthree"))
	long := attention.SummaryOf(string(make([]byte, 0)) + stringOf('a', 400))
	assert.Len(t, long, attention.MaxSummary+3)
	assert.Equal(t, "", attention.SummaryOf(" \n "))
}

func stringOf(b byte, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return string(out)
}

func titles(as []attention.Action) []string {
	var out []string
	for _, a := range as {
		out = append(out, a.Option.Title)
	}
	return out
}

func TestActions(t *testing.T) {
	t.Parallel()
	done := store.Ticket{Kind: store.KindStep, Status: store.StatusClosed, Resolution: "done", Fields: json.RawMessage(`{"step":"build"}`)}
	f := fake{
		tickets: map[string]store.Ticket{
			"esc":   ticket("esc", 0, nil, `{"run":{"id":"r1"}}`),
			"guard": ticket("guard", 0, nil, `{"hook":{"key":"k"}}`),
			"w":     ticket("w", 0, nil, `{"run":{"id":"r1"}}`),
			"rec":   ticket("rec", 0, nil, `{"desk":{"attention_recommend":"2"}}`),
		},
		steps: map[string][]store.Ticket{"r1": {done, {Kind: store.KindStep, Status: store.StatusOpen}}},
	}
	ctx := t.Context()
	run := attention.Actions(ctx, f, rpc.AttentionItem{Kind: rpc.AttentionTicket, ID: "esc"})
	assert.Equal(t, []string{"Resume", "Resume at build", "Cancel run"}, titles(run))
	assert.Equal(t, attention.VerbResume, run[1].Verb)
	assert.Equal(t, "r1", run[1].Target)
	assert.Equal(t, "build", run[1].Step)

	assert.Equal(t, []string{"Allow", "Deny"}, titles(attention.Actions(ctx, f, rpc.AttentionItem{Kind: rpc.AttentionTicket, ID: "guard"})))
	assert.Equal(t, []string{"Clear wait", "Extend one hour", "Cancel run"}, titles(attention.Actions(ctx, f, rpc.AttentionItem{Kind: rpc.AttentionWait, ID: "w"})))
	assert.Equal(t, []string{"Approve", "Reject", "Dismiss"}, titles(attention.Actions(ctx, f, rpc.AttentionItem{Kind: rpc.AttentionDecision, ID: "x"})))
	assert.Equal(t, []string{"Unquarantine after fix", "Keep quarantined"}, titles(attention.Actions(ctx, f, rpc.AttentionItem{Kind: rpc.AttentionPool, ID: "dev"})))
	assert.Equal(t, []string{"Discuss", "Leave waiting"}, titles(attention.Actions(ctx, f, rpc.AttentionItem{Kind: rpc.AttentionConfig, ID: "config"})))

	q := attention.Actions(ctx, f, rpc.AttentionItem{Kind: rpc.AttentionQuestion, ID: "q", Options: []rpc.QuestionOption{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}}})
	assert.Equal(t, attention.VerbQuestion, q[1].Verb)
	assert.Equal(t, 2, q[1].Option.ID)

	rec := attention.Actions(ctx, f, rpc.AttentionItem{Kind: rpc.AttentionTicket, ID: "rec"})
	assert.False(t, rec[0].Option.Recommended)
	assert.True(t, rec[1].Option.Recommended)
}

func TestKey(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "attention:question:js-1", attention.Key("question", "js-1"))
	kind, id, ok := attention.ParseKey("attention:run:js-1")
	assert.True(t, ok)
	assert.Equal(t, "run", kind)
	assert.Equal(t, "js-1", id)
	_, _, ok = attention.ParseKey("desk:abc")
	assert.False(t, ok)
}

type failing struct{ fake }

func (failing) NeedsHumanRuns(context.Context) ([]store.Ticket, error) {
	return nil, errors.New("boom")
}

func TestBuildPropagatesReadError(t *testing.T) {
	t.Parallel()
	_, err := attention.Build(t.Context(), failing{})
	require.Error(t, err)
}
