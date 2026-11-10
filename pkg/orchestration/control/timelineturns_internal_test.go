package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

var turnBase = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func tm(mins int) time.Time { return turnBase.Add(time.Duration(mins) * time.Minute) }

type hist []store.HistoryEntry

func (h hist) add(mins int, actor, op string, diff any) hist {
	d, _ := json.Marshal(diff)
	return append(h, store.HistoryEntry{Seq: int64(len(h) + 1), At: tm(mins), Actor: actor, Op: op, Diff: d})
}

func (h hist) turnEnd(mins int, actor, stop string) hist {
	return h.add(mins, actor, domain.OpTurnEnd, domain.TurnDiff{Stop: stop})
}

// turnRow replays a task created at minute 0 with the given history.
func turnRow(t *testing.T, h hist, closedAt int, now int, busy bool) rpc.TimelineRow {
	t.Helper()
	tk := domain.Ticket{ID: "js-t", Kind: domain.KindTask, Status: domain.StatusOpen, CreatedAt: tm(0)}
	if closedAt > 0 {
		tk.Status, tk.ClosedAt = domain.StatusClosed, tm(closedAt)
	}
	full := hist{}.add(0, "admin", "create", map[string]any{})
	full = append(full, h...)
	for i := range full {
		full[i].Seq = int64(i + 1)
	}
	return timelineRow(&config.Config{}, tk, store.Thread{History: full}, tm(now), func(string) bool { return busy })
}

func workingSeg(t *testing.T, r rpc.TimelineRow) rpc.TimelineSegment {
	t.Helper()
	for _, s := range r.Segments {
		if s.State == rpc.TimelineWorking || s.State == rpc.TimelineRetry {
			return s
		}
	}
	require.Fail(t, "no working segment", "%v", segStates(r))
	return rpc.TimelineSegment{}
}

// TestTimelineTurnsPairs (JOB-184): turns inside a claim are listed with
// their stop reason, in order.
func TestTimelineTurnsPairs(t *testing.T) {
	t.Parallel()
	h := hist{}.add(1, "s1", "claim", nil).
		add(2, "s1", domain.OpTurnStart, nil).turnEnd(5, "s1", "end_turn").
		add(10, "s1", domain.OpTurnStart, nil).turnEnd(12, "s1", "cancelled").
		add(13, "s1", "close", map[string]any{"resolution": "done"})
	s := workingSeg(t, turnRow(t, h, 13, 20, false))
	assert.Equal(t, []rpc.TimelineTurn{
		{Start: tm(2), End: tm(5), Stop: "end_turn"},
		{Start: tm(10), End: tm(12), Stop: "cancelled"},
	}, s.Turns)
}

// TestTimelineTurnsNone (JOB-184): a claim without turn rows has no list, so
// it is drawn as held and not as idle.
func TestTimelineTurnsNone(t *testing.T) {
	t.Parallel()
	h := hist{}.add(1, "s1", "claim", nil).add(9, "s1", "close", map[string]any{"resolution": "done"})
	s := workingSeg(t, turnRow(t, h, 9, 20, false))
	assert.Nil(t, s.Turns)
}

// TestTimelineTurnsCutByClose (JOB-184): a turn that closed its ticket writes
// turn_end after the close. It is clipped to the claim and keeps its stop.
func TestTimelineTurnsCutByClose(t *testing.T) {
	t.Parallel()
	h := hist{}.add(1, "s1", "claim", nil).
		add(2, "s1", domain.OpTurnStart, nil).
		add(6, "s1", "close", map[string]any{"resolution": "done"}).
		turnEnd(8, "s1", "end_turn")
	row := turnRow(t, h, 6, 20, false)
	s := workingSeg(t, row)
	assert.Equal(t, []rpc.TimelineTurn{{Start: tm(2), End: tm(6), Stop: "end_turn"}}, s.Turns)
	assert.Equal(t, tm(6), s.End)
}

// TestTimelineTurnsMissingEnd (JOB-184): a turn with no turn_end (the
// controller stopped) ends at the next turn_start of the session, or at the
// claim's end. It runs only as the last turn of an open claim whose session
// is in a turn now.
func TestTimelineTurnsMissingEnd(t *testing.T) {
	t.Parallel()
	h := hist{}.add(1, "s1", "claim", nil).
		add(2, "s1", domain.OpTurnStart, nil).
		add(7, "s1", domain.OpTurnStart, nil)

	// Open claim, session in a turn: the last turn runs, the first is cut.
	s := workingSeg(t, turnRow(t, h, 0, 20, true))
	require.Len(t, s.Turns, 2)
	assert.Equal(t, rpc.TimelineTurn{Start: tm(2), End: tm(7)}, s.Turns[0])
	assert.Equal(t, rpc.TimelineTurn{Start: tm(7)}, s.Turns[1])

	// Open claim, session not in a turn: nothing is shown as running.
	s = workingSeg(t, turnRow(t, h, 0, 20, false))
	require.Len(t, s.Turns, 2)
	assert.Equal(t, tm(20), s.Turns[1].End, "ends at the claim's end, now")

	// Closed claim: the last turn ends at the close even if the session is busy.
	h2 := append(hist{}, h...).add(15, "s1", "close", map[string]any{"resolution": "done"})
	s = workingSeg(t, turnRow(t, h2, 15, 20, true))
	assert.Equal(t, tm(15), s.Turns[1].End)
}

// TestTimelineTurnsOutsideClaim (JOB-184): turn rows that fall in no claim
// of their session are dropped, and a turn belongs to its own session's claim.
func TestTimelineTurnsOutsideClaim(t *testing.T) {
	t.Parallel()
	h := hist{}.
		add(1, "s1", domain.OpTurnStart, nil).turnEnd(2, "s1", "end_turn"). // before any claim
		add(3, "s1", "claim", nil).
		add(4, "s1", domain.OpTurnStart, nil).turnEnd(5, "s1", "end_turn").
		add(6, "s1", "release", map[string]any{"reason": "later"}).
		add(7, "s1", domain.OpTurnStart, nil).turnEnd(8, "s1", "end_turn"). // after the release
		add(9, "s2", "claim", nil).
		add(10, "s1", domain.OpTurnStart, nil).turnEnd(11, "s1", "end_turn"). // s1 does not hold it
		add(12, "s2", domain.OpTurnStart, nil).turnEnd(13, "s2", "end_turn").
		add(14, "s2", "close", map[string]any{"resolution": "done"})
	row := turnRow(t, h, 14, 20, false)
	var work []rpc.TimelineSegment
	for _, s := range row.Segments {
		if s.State == rpc.TimelineWorking {
			work = append(work, s)
		}
	}
	require.Len(t, work, 2)
	assert.Equal(t, []rpc.TimelineTurn{{Start: tm(4), End: tm(5), Stop: "end_turn"}}, work[0].Turns)
	assert.Equal(t, []rpc.TimelineTurn{{Start: tm(12), End: tm(13), Stop: "end_turn"}}, work[1].Turns)
}

// TestTimelineTurnsRetry (JOB-184): a failed attempt keeps the turns it ran.
func TestTimelineTurnsRetry(t *testing.T) {
	t.Parallel()
	one := 1
	h := hist{}.add(1, "s1", "claim", nil).
		add(2, "s1", domain.OpTurnStart, nil).turnEnd(4, "s1", "error").
		add(5, "s1", "release", map[string]any{"reason": "tests failed", "failures": one})
	s := workingSeg(t, turnRow(t, h, 0, 20, false))
	assert.Equal(t, rpc.TimelineRetry, s.State)
	assert.Equal(t, []rpc.TimelineTurn{{Start: tm(2), End: tm(4), Stop: "error"}}, s.Turns)
}
