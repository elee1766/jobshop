package storetest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
)

// BenchSize is the seed for Bench. The defaults are near gasserver
// (JOB-43): about 1000 sessions, with runs, steps and tasks around them.
type BenchSize struct {
	Sessions, Runs, Tasks, Comments int
}

// DefaultBenchSize reads JOBSHOP_BENCH_SESSIONS to scale every count.
func DefaultBenchSize() BenchSize {
	n := 1000
	if v, err := strconv.Atoi(os.Getenv("JOBSHOP_BENCH_SESSIONS")); err == nil && v > 0 {
		n = v
	}
	return BenchSize{Sessions: n, Runs: n, Tasks: 2 * n, Comments: 5}
}

const benchPool = "dev"

// Bench runs the hot-path benchmarks against one seeded store: List
// shapes, Get, ClaimNext with concurrent claimers and a mixed load of
// supervisor passes and Desk polls. Every benchmark leaves the store as
// it found it, so one seed serves them all.
func Bench(b *testing.B, s store.Store, size BenchSize) {
	ctx := b.Context()
	ids := seed(b, s, size)

	b.Run("ListSessions", func(b *testing.B) {
		for b.Loop() {
			must(b, list(ctx, s, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Limit: store.NoLimit}))
		}
	})
	b.Run("ListLabeledOpen", func(b *testing.B) {
		for b.Loop() {
			must(b, list(ctx, s, store.Filter{Label: "bug", Status: store.StatusOpen, Limit: store.NoLimit}))
		}
	})
	b.Run("ListRunState", func(b *testing.B) {
		for b.Loop() {
			must(b, list(ctx, s, store.Filter{Kind: store.KindRun, State: "running", Limit: store.NoLimit}))
		}
	})
	b.Run("ListReady", func(b *testing.B) {
		for b.Loop() {
			must(b, list(ctx, s, store.Filter{Ready: true, Pool: benchPool}))
		}
	})
	b.Run("Get", func(b *testing.B) {
		i := 0
		for b.Loop() {
			if _, err := s.Get(ctx, ids[i%len(ids)]); err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
	for _, n := range []int{1, 4, 16} {
		b.Run(fmt.Sprintf("ClaimNext/claimers=%d", n), func(b *testing.B) {
			benchClaim(b, s, n)
		})
	}
	b.Run("Mixed", func(b *testing.B) { benchMixed(b, s, ids) })
}

func list(ctx context.Context, s store.Store, f store.Filter) error {
	f.Project = proj
	_, err := s.List(ctx, f)
	return err
}

func must(b *testing.B, err error) {
	b.Helper()
	if err != nil {
		b.Fatal(err)
	}
}

// claimRelease claims the next ready ticket and puts it back open.
func claimRelease(ctx context.Context, s store.Store, holder string) error {
	_, c, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: benchPool, Holder: holder, Lease: lease})
	if err != nil {
		return err
	}
	_, err = s.Release(ctx, "", c.TicketID, &c, store.Outcome{Reason: "bench"})
	return err
}

// benchClaim runs b.N claim and release pairs split over n claimers.
func benchClaim(b *testing.B, s store.Store, n int) {
	ctx := b.Context()
	var left atomic.Int64
	left.Store(int64(b.N))
	var wg sync.WaitGroup
	errs := make(chan error, n)
	b.ResetTimer()
	for w := range n {
		wg.Go(func() {
			holder := fmt.Sprintf("bench-claimer-%d", w)
			for left.Add(-1) >= 0 {
				if err := claimRelease(ctx, s, holder); err != nil {
					errs <- err
					return
				}
			}
		})
	}
	wg.Wait()
	b.StopTimer()
	close(errs)
	for err := range errs {
		b.Fatal(err)
	}
}

// benchMixed models one supervisor pass (ready list, sessions, runs,
// a claim and release, a comment) against 4 Desk pollers reading
// sessions, labeled tickets and single tickets. One op is one pass.
func benchMixed(b *testing.B, s store.Store, ids []string) {
	ctx, cancel := context.WithCancel(b.Context())
	defer cancel()
	var wg sync.WaitGroup
	var polls atomic.Int64
	for w := range 4 {
		wg.Go(func() {
			for i := w; ctx.Err() == nil; i++ {
				_ = list(ctx, s, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Limit: store.NoLimit})
				_ = list(ctx, s, store.Filter{Label: "bug", Status: store.StatusOpen})
				_, _ = s.Get(ctx, ids[i%len(ids)])
				polls.Add(1)
			}
		})
	}
	i := 0
	for b.Loop() {
		must(b, list(ctx, s, store.Filter{Ready: true, Pool: benchPool}))
		must(b, list(ctx, s, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Limit: store.NoLimit}))
		must(b, list(ctx, s, store.Filter{Kind: store.KindRun, State: "running", Limit: store.NoLimit}))
		must(b, claimRelease(ctx, s, "bench-supervisor"))
		_, err := store.AddComment(ctx, s, "bench", ids[i%len(ids)], "supervisor pass note")
		must(b, err)
		i++
	}
	b.StopTimer()
	cancel()
	wg.Wait()
	b.ReportMetric(float64(polls.Load())/float64(b.N), "polls/op")
}

// seed writes size's tickets in batches and returns every id.
func seed(b *testing.B, s store.Store, size BenchSize) []string {
	b.Helper()
	ctx := b.Context()
	start := time.Now()
	var ids []string
	newID := func() string {
		id, err := s.NewID(proj)
		must(b, err)
		ids = append(ids, id)
		return id
	}
	apply := func(g store.Graph) {
		_, err := s.Apply(ctx, "bench", g)
		must(b, err)
	}
	var g store.Graph
	flush := func() {
		if len(g.Tickets) > 0 {
			apply(g)
		}
		g = store.Graph{}
	}
	for i := range size.Sessions {
		g.Tickets = append(g.Tickets, store.NewTicket{
			ID: newID(), Project: proj, Kind: store.KindSession, Pool: benchPool, Title: "session",
			Fields: json.RawMessage(fmt.Sprintf(`{"agent":"jcode","key":"bench:%d"}`, i)),
		})
		if len(g.Tickets) >= 200 {
			flush()
		}
	}
	flush()
	for i := range size.Runs {
		state := "done"
		if i%4 == 0 {
			state = "running"
		}
		run := newID()
		g.Tickets = append(g.Tickets, store.NewTicket{
			ID: run, Project: proj, Kind: store.KindRun, Title: "run",
			Fields: json.RawMessage(fmt.Sprintf(`{"workflow":"pr-work","state":%q,"key":"issue-%d"}`, state, i)),
			Meta:   json.RawMessage(fmt.Sprintf(`{"github":{"issue":"%d"}}`, i)),
		})
		for j, step := range []string{"implement", "review-merge", "post-merge"} {
			g.Tickets = append(g.Tickets, store.NewTicket{
				ID: newID(), Project: proj, Kind: store.KindStep, Pool: benchPool, ParentID: run, Title: step,
				Body:   "Step body with the workflow prompt, a few hundred bytes long in practice.",
				Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":%q,"index":%d,"retries":3}`, run, step, j)),
			})
		}
		if len(g.Tickets) >= 200 {
			flush()
		}
	}
	flush()
	for i := range size.Tasks {
		nt := store.NewTicket{ID: newID(), Project: proj, Kind: store.KindTask, Pool: benchPool, Title: fmt.Sprintf("task %d", i), Body: "task body"}
		if i%3 == 0 {
			nt.Labels = []string{"bug"}
		}
		g.Tickets = append(g.Tickets, nt)
		if len(g.Tickets) >= 200 {
			flush()
		}
	}
	flush()
	// Close most of what was seeded: on a real host the open set is
	// small next to the history.
	for i, id := range ids {
		if i%5 == 0 {
			continue
		}
		_, err := s.Release(ctx, "bench", id, nil, store.Outcome{Close: "done"})
		must(b, err)
	}
	for i, id := range ids {
		if i%2 != 0 {
			continue
		}
		for range size.Comments {
			_, err := store.AddComment(ctx, s, "bench", id, "a comment of ordinary length about progress, CI and a PR link")
			must(b, err)
		}
	}
	b.Logf("seeded %d tickets in %s", len(ids), time.Since(start).Round(time.Millisecond))
	return ids
}
