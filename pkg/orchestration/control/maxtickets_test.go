package control_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// singleUse is a pool config with max_tickets; hostMax 0 is no host limit.
type singleUse struct{ sessions, tickets, minSessions, hostMax int }

func writeSingleUseConfig(t *testing.T, root config.Root, c singleUse) {
	t.Helper()
	limits := ""
	if c.hostMax > 0 {
		limits = fmt.Sprintf("limits:\n  max_sessions: %d\n", c.hostMax)
	}
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\n%spools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: %d\n    max_tickets: %d\n    min_sessions: %d\n    idle_timeout: 1m\n    dir: %q\n",
		os.Args[0], limits, c.sessions, c.tickets, c.minSessions, root.Dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
}

func sessionFields(t *testing.T, tk domain.Ticket) domain.SessionFields {
	t.Helper()
	f, err := domain.DecodeFields[domain.SessionFields](tk)
	require.NoError(t, err)
	return f
}

// harness is a started single-use controller and helpers to drive it.
type harness struct {
	t     *testing.T
	root  config.Root
	admin *rpc.Client
}

func newHarness(t *testing.T, c singleUse) *harness {
	t.Helper()
	root, start := poolRoot(t, "1m")
	writeSingleUseConfig(t, root, c)
	stop := start()
	t.Cleanup(stop)
	return &harness{t: t, root: root, admin: dial(t, root.AdminSocket())}
}

func (h *harness) file(name string, body []byte) {
	h.t.Helper()
	require.NoError(h.t, os.WriteFile(filepath.Join(h.root.Dir, name), body, 0o600))
}

func (h *harness) exists(name string) bool {
	_, err := os.Stat(filepath.Join(h.root.Dir, name))
	return err == nil
}

func (h *harness) create(title string) domain.Ticket {
	h.t.Helper()
	var tk domain.Ticket
	require.NoError(h.t, h.admin.Call(h.t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: title, Pool: "dev"}, &tk))
	return tk
}

func (h *harness) show(id string) domain.Ticket {
	h.t.Helper()
	var got rpc.ShowResult
	require.NoError(h.t, h.admin.Call(h.t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &got))
	return got.Ticket
}

func (h *harness) sessions(status domain.Status) []domain.Ticket {
	h.t.Helper()
	var out []domain.Ticket
	require.NoError(h.t, h.admin.Call(h.t.Context(), rpc.MethodTicketList,
		rpc.ListParams{Kind: domain.KindSession, Status: status}, &out))
	return out
}

func (h *harness) until(cond func() bool, msg string) {
	h.t.Helper()
	passEventually(h.t, h.root.Dir, cond, msg)
}

// parkNext makes the next turn park its ticket with a resume wait.
func (h *harness) parkNext() {
	h.t.Helper()
	park, err := json.Marshal(rpc.WaitAddParams{
		Reason: "CI on PR #5", Resume: true,
		Conds: []domain.WaitCond{{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"}},
	})
	require.NoError(h.t, err)
	h.file("park", park)
}

// parkedSession waits until a session is suspended on ticket and returns it.
func (h *harness) parkedSession(ticket string) string {
	h.t.Helper()
	var id string
	h.until(func() bool {
		for _, s := range h.sessions(domain.StatusOpen) {
			if f := sessionFields(h.t, s); f.Suspended && f.Parked == ticket {
				id = s.ID
				return true
			}
		}
		return false
	}, "a session parks "+ticket)
	return id
}

func (h *harness) clearWait(ticket string) {
	h.t.Helper()
	require.NoError(h.t, h.admin.Call(h.t.Context(), rpc.MethodWaitClear, rpc.WaitClearParams{ID: ticket, Reason: "checks green"}, nil))
}

func (h *harness) pools() rpc.PoolScale {
	h.t.Helper()
	var res rpc.PoolStatusResult
	require.NoError(h.t, h.admin.Call(h.t.Context(), rpc.MethodPoolList, struct{}{}, &res))
	require.Len(h.t, res.Pools, 1)
	return res.Pools[0]
}

// holders is the sessions that held each ticket at some point, from the
// ticket history is not needed: a session closes one ticket, so the
// closed session count and the tickets' closers are enough.
func countSessions(ss []domain.Ticket, resolution string) int {
	n := 0
	for _, s := range ss {
		if s.Resolution == resolution {
			n++
		}
	}
	return n
}

// TestMaxTicketsEndsSession (JOB-189): max_tickets 1, max_sessions 1 and
// two ready tickets. The first session closes one and ends ticket_limit,
// a second takes the other. No session works two tickets.
func TestMaxTicketsEndsSession(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1, tickets: 1})
	a, b := h.create("a"), h.create("b")
	waitClosed(t, h.root, h.admin, a.ID)
	waitClosed(t, h.root, h.admin, b.ID)
	h.until(func() bool { return countSessions(h.sessions(domain.StatusClosed), "ticket_limit") >= 2 }, "both single-use sessions end")
	for _, s := range h.sessions(domain.StatusClosed) {
		assert.Equal(t, 1, sessionFields(t, s).Tickets, "session %s closed one ticket", s.ID)
	}
	assert.NotEqual(t, h.show(a.ID).Holder, h.show(b.ID).Holder, "two sessions, one ticket each")
}

// TestMaxTicketsUnsetKeepsSession (JOB-189): without max_tickets one
// session claims the next ticket in the same conversation.
func TestMaxTicketsUnsetKeepsSession(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1})
	a, b := h.create("a"), h.create("b")
	waitClosed(t, h.root, h.admin, a.ID)
	waitClosed(t, h.root, h.admin, b.ID)
	assert.Equal(t, h.show(a.ID).Holder, h.show(b.ID).Holder, "one session worked both")
	assert.Len(t, h.sessions(domain.StatusOpen), 1)
	assert.Empty(t, h.sessions(domain.StatusClosed))
}

// TestMaxTicketsReleaseDoesNotCount (JOB-189): a ticket the session drops
// with a release does not use up its limit, so it can still close one.
func TestMaxTicketsReleaseDoesNotCount(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1, tickets: 1})
	h.file("release", nil)
	a := h.create("a")
	// The first turn releases the ticket; the same session claims it again
	// and closes it, then ends at its one closed ticket.
	waitClosed(t, h.root, h.admin, a.ID)
	h.until(func() bool { return countSessions(h.sessions(domain.StatusClosed), "ticket_limit") == 1 }, "the session ends ticket_limit")
	closed := h.sessions(domain.StatusClosed)
	require.Len(t, closed, 1)
	assert.Equal(t, 1, sessionFields(t, closed[0]).Tickets)
}

// TestMaxTicketsParkResume (JOB-189): a session whose ticket waits with
// resume is suspended and takes no slot, so another session works another
// ticket. When the wait ends the ticket goes to the parking session, which
// resumes the same conversation and closes it.
func TestMaxTicketsParkResume(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1, tickets: 1})
	h.parkNext()
	x := h.create("x")
	a := h.parkedSession(x.ID)
	acp := sessionFields(t, h.show(a)).ACPSession
	require.NotEmpty(t, acp)

	h.file("block", nil)
	y := h.create("y")
	h.until(func() bool {
		got := h.show(y.ID)
		return got.Status == domain.StatusInProgress && got.Holder != "" && got.Holder != a
	}, "another session works Y while A is parked")
	p := h.pools()
	assert.Equal(t, 1, p.Parked)
	assert.Equal(t, 1, p.Open, "the parked session is in no slot count")
	info := sessionInfo(t, h.admin, a)
	assert.True(t, info.Suspended)
	assert.Equal(t, x.ID, info.Parked)

	// Let Y's session finish, so there is room, then end X's wait.
	h.file("go", nil)
	waitClosed(t, h.root, h.admin, y.ID)
	h.clearWait(x.ID)
	assert.Equal(t, a, h.show(x.ID).Holder, "the wait hands X back to A")
	waitClosed(t, h.root, h.admin, x.ID)
	h.until(func() bool { return h.show(a).Status == domain.StatusClosed }, "A ends")
	got := h.show(a)
	assert.Equal(t, "ticket_limit", got.Resolution)
	f := sessionFields(t, got)
	assert.Equal(t, 1, f.Tickets, "the parked ticket counted once, on close")
	assert.Equal(t, acp, f.ACPSession, "the conversation resumed, not replaced")
}

// TestMaxTicketsResumeOverCap (JOB-189): a wait ends while the pool is
// full. The session resumes at once and the pool runs max_sessions + 1.
func TestMaxTicketsResumeOverCap(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1, tickets: 1})
	h.parkNext()
	x := h.create("x")
	a := h.parkedSession(x.ID)

	h.file("block", nil)
	y := h.create("y")
	h.until(func() bool { return h.show(y.ID).Status == domain.StatusInProgress }, "B works Y")
	h.clearWait(x.ID)
	h.until(func() bool { return !sessionFields(t, h.show(a)).Suspended }, "A resumes over the cap")
	h.until(func() bool { return h.exists("started") && h.show(x.ID).Status == domain.StatusInProgress }, "A works X")
	assert.Len(t, h.sessions(domain.StatusOpen), 2, "max_sessions + 1")
	assert.Equal(t, a, h.show(x.ID).Holder)

	// The scaler does not cordon the resumed session as surplus, and
	// starts no third session while over the cap.
	pass(t, h.root.Dir, 3)
	assert.Empty(t, sessionFields(t, h.show(a)).Cordoned, "not cordoned as surplus")
	h.create("z")
	pass(t, h.root.Dir, 3)
	assert.Len(t, h.sessions(domain.StatusOpen), 2, "no new session while over the cap")

	h.file("go", nil)
	waitClosed(t, h.root, h.admin, x.ID)
	waitClosed(t, h.root, h.admin, y.ID)
}

// TestMaxTicketsTwoResumesOneOverCap (JOB-189): two waits end while the
// pool is full. One session resumes, the other stays suspended until a
// slot frees, ahead of a new session for other ready work. Never more
// than max_sessions + 1.
func TestMaxTicketsTwoResumesOneOverCap(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1, tickets: 1})
	h.parkNext()
	x1 := h.create("x1")
	a1 := h.parkedSession(x1.ID)
	h.parkNext()
	x2 := h.create("x2")
	a2 := h.parkedSession(x2.ID)

	h.file("block", nil)
	y := h.create("y")
	h.until(func() bool { return h.show(y.ID).Status == domain.StatusInProgress }, "B works Y")
	h.create("z") // other ready work: waits behind the resumes

	h.clearWait(x1.ID)
	h.clearWait(x2.ID)
	h.until(func() bool { return !sessionFields(t, h.show(a1)).Suspended }, "the older wait resumes")
	pass(t, h.root.Dir, 3)
	assert.True(t, sessionFields(t, h.show(a2)).Suspended, "the second stays suspended")
	assert.Len(t, h.sessions(domain.StatusOpen), 3, "B, the resumed A1 and the suspended A2, and nothing new")
	p := h.pools()
	assert.Equal(t, 1, p.Parked)
	assert.Equal(t, 2, p.Open, "max_sessions + 1")

	// A1 finishes; the next slot goes to A2, before a new session for Z.
	h.file("go", nil)
	waitClosed(t, h.root, h.admin, x1.ID)
	waitClosed(t, h.root, h.admin, x2.ID)
	assert.Equal(t, a2, h.show(x2.ID).Holder)
}

// TestMaxTicketsResumeAtHostLimit (JOB-189): at limits.max_sessions the
// session stays suspended until a host slot frees.
func TestMaxTicketsResumeAtHostLimit(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 2, tickets: 1, hostMax: 1})
	h.parkNext()
	x := h.create("x")
	a := h.parkedSession(x.ID)

	h.file("block", nil)
	y := h.create("y")
	h.until(func() bool { return h.show(y.ID).Status == domain.StatusInProgress }, "B works Y at the host limit")
	h.clearWait(x.ID)
	pass(t, h.root.Dir, 3)
	assert.True(t, sessionFields(t, h.show(a)).Suspended, "no host slot, so A stays suspended")

	h.file("go", nil)
	waitClosed(t, h.root, h.admin, y.ID)
	waitClosed(t, h.root, h.admin, x.ID)
}

// TestMaxTicketsParkedGone (JOB-189): a parked session whose ticket was
// closed by someone else ends instead of staying suspended forever.
func TestMaxTicketsParkedGone(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1, tickets: 2})
	h.parkNext()
	x := h.create("x")
	a := h.parkedSession(x.ID)
	require.NoError(t, h.admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: x.ID, Comment: "done by hand"}, nil))
	h.until(func() bool { return h.show(a).Status == domain.StatusClosed }, "A ends")
	assert.Equal(t, "parked_gone", h.show(a).Resolution)
}

// TestMaxTicketsParkedNotIdle (JOB-189): a suspended session does not end
// on idle_timeout while its ticket waits.
func TestMaxTicketsParkedNotIdle(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1, tickets: 1})
	h.parkNext()
	x := h.create("x")
	a := h.parkedSession(x.ID)
	// idle_timeout is 1m; run the fake clock well past it.
	clk := fakeClock(h.root.Dir)
	require.NotNil(t, clk)
	clk.Advance(10 * time.Minute)
	pass(t, h.root.Dir, 3)
	got := h.show(a)
	assert.Equal(t, domain.StatusOpen, got.Status, "still parked after idle_timeout")
	assert.True(t, sessionFields(t, got).Suspended)
}

// TestMaxTicketsTimedDeferParks (JOB-189): a ticket the session defers to
// a later time parks it too.
func TestMaxTicketsTimedDeferParks(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 1, tickets: 1})
	h.file("defer", nil)
	x := h.create("x")
	a := h.parkedSession(x.ID)
	assert.Equal(t, a, h.show(x.ID).Holder, "the deferred ticket is assigned back to the session")
	// The scaler may start a warm spare while it holds its recent peak
	// (scale_down_after); the parked session is not one of the open count.
	assert.Equal(t, 1, h.pools().Parked)
}

// TestMaxTicketsWarmReplaced (JOB-189): with min_sessions a warm session
// that used its ticket ends, and the warm count is rebuilt.
func TestMaxTicketsWarmReplaced(t *testing.T) {
	h := newHarness(t, singleUse{sessions: 2, tickets: 1, minSessions: 1})
	h.until(func() bool { return len(h.sessions(domain.StatusOpen)) == 1 }, "one warm session")
	warm := h.sessions(domain.StatusOpen)[0].ID
	a := h.create("a")
	waitClosed(t, h.root, h.admin, a.ID)
	h.until(func() bool { return h.show(warm).Status == domain.StatusClosed }, "the used session ends")
	h.until(func() bool {
		open := h.sessions(domain.StatusOpen)
		return len(open) == 1 && open[0].ID != warm
	}, "a new warm session replaces it")
	assert.Equal(t, "ticket_limit", h.show(warm).Resolution)
}

func sessionInfo(t *testing.T, c *rpc.Client, id string) rpc.SessionInfo {
	t.Helper()
	var res []rpc.SessionInfo
	require.NoError(t, c.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &res))
	for _, s := range res {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("session %s not listed", id)
	return rpc.SessionInfo{}
}
