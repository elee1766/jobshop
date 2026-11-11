// Package workflow decides what a wait (fields.wait, #381) is waiting
// for. It evaluates a wait's conditions against observations, plans which
// waits an event can change and when the next time condition or timeout
// falls due, and nothing else.
//
// The engine reads and decides; it never writes. The controller owns the
// persisted transition: authorization, the waitMu pass lock, the
// revision-checked update that clears or saves a wait, labels, comments
// and dispatch eligibility. A Decision is advice computed from a
// snapshot, and the controller applies it only if the ticket is still at
// the revision the snapshot came from, so a stale observation cannot
// overwrite a wait that changed meanwhile (JOB-144).
//
// Provider clients stay where they are (the gh CLI adapter in
// pkg/orchestration/control, pkg/integrations/linear); the engine sees them through the small
// interfaces below.
package workflow

import (
	"context"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/github"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
)

// Source reads PR, check, issue, tag and commit state from GitHub.
type Source interface {
	PR(ctx context.Context, repo string, n int) (github.PRInfo, error)
	Reviews(ctx context.Context, repo string, n int) ([]github.Review, error)
	// Checks lists the checks of a PR's current head (required only when
	// asked), or of sha when n is 0.
	Checks(ctx context.Context, repo string, n int, sha string, required bool) ([]github.Check, error)
	// WorkflowRuns lists the GitHub Actions runs on commit sha, every
	// workflow (github.checks with workflow=, JOB-75).
	WorkflowRuns(ctx context.Context, repo, sha string) ([]github.WorkflowRun, error)
	// Issue reads an issue's state and labels (github.label, #603).
	Issue(ctx context.Context, repo string, n int) (github.IssueInfo, error)
	// Tag reports whether tag exists in repo (github.tag, JOB-11).
	Tag(ctx context.Context, repo, tag string) (bool, error)
	// Contains reports whether head is sha or a descendant of it in repo
	// (deploy, JOB-11).
	Contains(ctx context.Context, repo, sha, head string) (bool, error)
}

// LinearSource reads a Linear issue for linear.state and linear.label.
type LinearSource interface {
	IssueContext(ctx context.Context, key string) (linear.Context, error)
}

// RunSource reads the state of a run ticket (the run condition).
type RunSource interface {
	Run(ctx context.Context, id string) (RunState, error)
}

// QuestionSource reads whether a question ticket is closed (the question
// condition, JOB-202).
type QuestionSource interface {
	// Question reports the question's resolution and whether it is closed.
	Question(ctx context.Context, id string) (QuestionState, error)
}

// QuestionState is what the question condition needs of a question ticket.
type QuestionState struct {
	Closed     bool
	Resolution string
}

// ReplySource answers a thread.reply condition. The controller owns it
// because the answer depends on its subscription filters (bots, ignore
// lists, authors, Linear identities).
type ReplySource interface {
	// Reply reports whether a human replied on cd.Thread after since,
	// from cd.From when set, with a note naming the author.
	Reply(ctx context.Context, cd domain.WaitCond, since time.Time) (met bool, note string, err error)
}

// RunState is what the run condition needs of a run ticket.
type RunState struct {
	Closed bool   // the ticket is closed
	State  string // store.RunDone, store.RunCancelled, ...
}

// Ended reports whether the run finished or was cancelled or closed.
func (r RunState) Ended() bool {
	return r.Closed || r.State == domain.RunDone || r.State == domain.RunCancelled
}

// Sources is everything one pass observes through.
type Sources struct {
	GitHub    Source
	Linear    LinearSource
	Runs      RunSource
	Questions QuestionSource
	Replies   ReplySource
	// Commit is the build the controller runs, for deploy conditions.
	// Empty: it cannot tell what is deployed.
	Commit string
}
