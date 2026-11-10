package control_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// A deliver -> review -> merge workflow like packs/agent's, with a review
// round cap of 2 (JOB-98).
const requestChangesYAML = `projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
pools:
  dev: {project: js, agent: fake, max_sessions: 1, idle_timeout: 1m, dir: %q}
  humans: {project: js, agent: fake, max_sessions: 0, dir: %q}
workflows:
  pr-work:
    vars: [task]
    escalate_pool: humans
    steps:
      - {id: deliver, pool: dev, title: "Deliver {{.task}}", prompt: x}
      - {id: review, pool: dev, title: "Review {{.task}}", prompt: x, max_review_rounds: 2}
      - {id: merge, pool: dev, title: "Merge {{.task}}", prompt: x}
`

// TestRequestChanges (JOB-98): with the in-process agent, review closes
// request_changes, deliver reopens with the findings in its prompt, the
// second review approves and the run finishes, with no failure counted.
// A reviewer that keeps requesting changes past its cap sends the run to
// needs_human.
func TestRequestChanges(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(root.Config(), fmt.Appendf(nil, requestChangesYAML, os.Args[0], root.Dir, root.Dir), 0o600))
	_, err := config.LoadFile(root.Config())
	require.NoError(t, err)
	rounds := filepath.Join(root.Dir, "request_changes")
	require.NoError(t, os.WriteFile(rounds, []byte("1"), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"task": "one round"}, Key: "one"})
	passEventually(t, root.Dir, func() bool { return showRun(t, admin, r.Run.ID).Run.Status == domain.StatusClosed }, "run did not finish after one review round")
	got := showRun(t, admin, r.Run.ID)
	assert.Equal(t, "done", got.Run.Resolution)
	s := stepByName(t, got)
	for n, st := range s {
		assert.Equal(t, "done", st.Resolution, n)
		assert.Zero(t, failures(t, st), n)
	}
	rf, err := domain.DecodeFields[domain.StepFields](s["review"])
	require.NoError(t, err)
	assert.Equal(t, 1, rf.ReviewRounds)
	assert.Equal(t, "deliver", rf.OnRequestChanges, "default: the previous step")
	df, err := domain.DecodeFields[domain.StepFields](s["deliver"])
	require.NoError(t, err)
	assert.Nil(t, df.Changes, "cleared when deliver closed done")

	// The reopened deliver turn carried the findings in its prompt.
	b, err := os.ReadFile(filepath.Join(root.Dir, "prompts"))
	require.NoError(t, err)
	var deliverTurns []string
	for line := range strings.Lines(string(b)) {
		if strings.Contains(line, `# Deliver one round`) {
			deliverTurns = append(deliverTurns, line)
		}
	}
	require.Len(t, deliverTurns, 2, "deliver ran twice")
	assert.NotContains(t, deliverTurns[0], "Changes requested")
	assert.Contains(t, deliverTurns[1], "# Changes requested")
	assert.Contains(t, deliverTurns[1], "review round 1 of 2")
	assert.Contains(t, deliverTurns[1], "finding 1: handle the empty case")

	// Over the cap: rounds 1 and 2 reopen deliver, round 3 escalates.
	require.NoError(t, os.WriteFile(rounds, []byte("10"), 0o600))
	r = startRun(t, admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"task": "endless"}, Key: "endless"})
	// The controller sets needs_human, then files the escalation ticket in
	// a later write, so wait for both (#923).
	passEventually(t, root.Dir, func() bool {
		return runState(t, showRun(t, admin, r.Run.ID)) == domain.RunNeedsHuman && len(escalations(t, admin, r.Run.ID)) > 0
	}, "run did not reach needs_human with an escalation")
	s = stepByName(t, showRun(t, admin, r.Run.ID))
	assert.Equal(t, domain.ResolutionRequestChanges, s["review"].Resolution)
	rf, err = domain.DecodeFields[domain.StepFields](s["review"])
	require.NoError(t, err)
	assert.Equal(t, 3, rf.ReviewRounds)
	assert.Zero(t, rf.Failures)
	assert.Equal(t, "done", s["deliver"].Resolution, "deliver is not reopened past the cap")
	assert.Equal(t, domain.StatusOpen, s["merge"].Status)
	esc := escalations(t, admin, r.Run.ID)
	require.Len(t, esc, 1)
	assert.Contains(t, esc[0].Body, "past its review round limit")
	assert.Contains(t, esc[0].Body, "review rounds 3/2")
}

// TestRequestChangesRejected: request_changes needs findings, a step to
// reopen and a closed target; a refused close changes nothing.
func TestRequestChangesRejected(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "7"}})
	impl, review := r.Steps[0].ID, r.Steps[1].ID
	call := func(id, comment string) error {
		return admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: id, Resolution: domain.ResolutionRequestChanges, Comment: comment}, nil)
	}
	require.ErrorIs(t, call(impl, "x"), domain.ErrInvalid, "the first step has no step to send changes to")
	require.ErrorIs(t, call(review, "x"), domain.ErrInvalid, "implement is not finished")
	closeStep(t, admin, impl, "done")
	require.ErrorIs(t, call(review, "  "), domain.ErrInvalid, "findings are required")
	err := admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: review, Resolution: "failed", Outcome: domain.ResolutionRequestChanges}, nil)
	require.ErrorIs(t, err, domain.ErrInvalid, "outcome request_changes with another resolution")
	assert.Equal(t, domain.StatusOpen, getTicket(t, admin, review).Ticket.Status)
	assert.Equal(t, domain.StatusClosed, getTicket(t, admin, impl).Ticket.Status)

	// --outcome request_changes is the same close.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: review, Outcome: domain.ResolutionRequestChanges, Comment: "rename x"}, nil))
	sh := getTicket(t, admin, impl)
	assert.Equal(t, domain.StatusOpen, sh.Ticket.Status)
	f, err := domain.DecodeFields[domain.StepFields](sh.Ticket)
	require.NoError(t, err)
	require.NotNil(t, f.Changes)
	assert.Equal(t, "rename x", f.Changes.Findings)
	assert.Equal(t, 3, f.Changes.Max, "default cap")
	rs := getTicket(t, admin, review)
	assert.Equal(t, domain.StatusOpen, rs.Ticket.Status)
	require.NotEmpty(t, rs.Comments)
	assert.Contains(t, rs.Comments[0].Body, "Changes requested:\n\nrename x")
	assert.Equal(t, domain.RunRunning, runState(t, showRun(t, admin, r.Run.ID)))
}
