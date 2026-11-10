package cli

import (
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// learnCmd is `jobshop learn` (#264).
type learnCmd struct {
	Add        learnAddCmd        `cmd:"" help:"Record a learning of any kind the config allows (learnings.kinds)."`
	Correction learnCorrectionCmd `cmd:"" help:"Record a correction: a human or reviewer told you to do it differently."`
	Failure    learnFailureCmd    `cmd:"" help:"Record a failure and its cause."`
	Workaround learnWorkaroundCmd `cmd:"" help:"Record a workaround for a tool or environment problem."`
	List       learnListCmd       `cmd:"" help:"List learnings, newest first."`
	Update     learnUpdateCmd     `cmd:"" help:"Set the status or PR of learnings (curator, #273)."`
	Expire     learnExpireCmd     `cmd:"" help:"Set expired on old open learnings that have no PR, so curators stop reading them (admin)."`
	Evaluate   learnEvaluateCmd   `cmd:"" help:"Curator gate: decide whether merged memory edits helped (#274)."`
	Edits      learnEditsCmd      `cmd:"" help:"List merged memory edits with their baseline and evaluation (#274)."`
}

type learnKindCmd struct {
	Summary string `arg:"" help:"One line: what happened and what to do instead. Use - to read stdin."`
	Fix     string `help:"The fix or the rule to follow next time."`
	Scope   string `help:"global, repo:<owner/name> or pool:<name> (default: the ticket's repo or pool)."`
	Ticket  string `short:"t" help:"Ticket it is about (default: the session's current ticket)."`
}

func (c learnKindCmd) send(g *Globals, kind string) error {
	s, err := readBody(c.Summary)
	if err != nil {
		return err
	}
	return g.call(rpc.MethodLearnAdd, rpc.LearnParams{Kind: kind, Summary: s, Fix: c.Fix, Scope: c.Scope, Ticket: c.Ticket}, nil)
}

type learnCorrectionCmd struct{ learnKindCmd }

func (c learnCorrectionCmd) Run(g *Globals) error { return c.send(g, domain.LearningCorrection) }

type learnFailureCmd struct{ learnKindCmd }

func (c learnFailureCmd) Run(g *Globals) error { return c.send(g, domain.LearningFailure) }

type learnWorkaroundCmd struct{ learnKindCmd }

func (c learnWorkaroundCmd) Run(g *Globals) error { return c.send(g, domain.LearningWorkaround) }

// learnAddCmd is `jobshop learn add <kind> <summary>` (#524); the
// controller checks the kind against learnings.kinds.
type learnAddCmd struct {
	Kind string `arg:"" help:"Learning kind, e.g. friction or blocker. Must match learnings.kinds."`
	learnKindCmd
}

func (c learnAddCmd) Run(g *Globals) error { return c.send(g, c.Kind) }

type learnListCmd struct {
	Since       time.Duration `help:"Only learnings newer than this, e.g. 24h."`
	Scope       string        `help:"global, repo:<owner/name> or pool:<name>."`
	Kind        string        `help:"Kind or glob, e.g. failure or review-*."`
	Status      string        `enum:",open,applied,rejected,expired,duplicate" default:"" help:"open, applied, rejected, expired or duplicate."`
	Ticket      string        `short:"t" help:"Only learnings about this ticket."`
	DuplicateOf int64         `name:"duplicate-of" help:"Only the duplicates of this learning id."`
	Limit       int           `default:"50" help:"Most learnings to print."`
}

func (c learnListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodLearnList, rpc.LearnListParams{
		Since: c.Since, Scope: c.Scope, Kind: c.Kind, Status: c.Status, Ticket: c.Ticket, DuplicateOf: c.DuplicateOf, Limit: c.Limit,
	}, nil)
}

// learnUpdateCmd is `jobshop learn update` (#273): the curator records
// the PR addressing learnings, then applied (merged) or rejected (closed
// unmerged or reverted; needs the PR). expired needs no PR; --status open
// reopens an expired or duplicate one. --duplicate-of marks the ids as
// duplicates of an open learning kept instead (JOB-199).
type learnUpdateCmd struct {
	IDs         []int64   `arg:"" name:"id" help:"Learning ids."`
	Status      string    `enum:",open,applied,rejected,expired" default:"" help:"open, applied, rejected or expired."`
	PR          string    `name:"pr" help:"PR URL that addresses them."`
	DuplicateOf int64     `name:"duplicate-of" help:"Mark the ids as duplicates of this open learning, kept for the curator. Fails if it is not open or is one of the ids."`
	MergedAt    time.Time `name:"merged-at" help:"With --status applied: when the PR merged (RFC3339, default now). Starts the curator gate's window."`
}

func (c learnUpdateCmd) Run(g *Globals) error {
	return g.call(rpc.MethodLearnUpdate, rpc.LearnUpdateParams{IDs: c.IDs, Status: c.Status, PR: c.PR, DuplicateOf: c.DuplicateOf, MergedAt: c.MergedAt}, nil)
}

// learnExpireCmd is `jobshop learn expire` (JOB-198).
type learnExpireCmd struct {
	OlderThan time.Duration `name:"older-than" required:"" help:"Expire open learnings recorded before now minus this, e.g. 120h."`
	Scope     string        `help:"Only this scope: global, repo:<owner/name> or pool:<name>."`
	Kind      string        `help:"Only this kind or glob, e.g. failure or review-*."`
}

func (c learnExpireCmd) Run(g *Globals) error {
	return g.call(rpc.MethodLearnExpire, rpc.LearnExpireParams{OlderThan: c.OlderThan, Scope: c.Scope, Kind: c.Kind}, nil)
}

// learnEvaluateCmd is `jobshop learn evaluate` (#274): the evaluate
// step's decision on pending memory edits.
type learnEvaluateCmd struct {
	ID   int64         `arg:"" optional:"" help:"Memory edit id: evaluate it even when not due."`
	Runs int           `default:"20" help:"An edit is due after this many runs in its pools since the merge."`
	Age  time.Duration `default:"72h" help:"Or after this long since the merge."`
}

func (c learnEvaluateCmd) Run(g *Globals) error {
	return g.call(rpc.MethodLearnEvaluate, rpc.LearnEvaluateParams{ID: c.ID, Runs: c.Runs, Age: c.Age}, nil)
}

type learnEditsCmd struct {
	State string `enum:",pending,helped,not_helped" default:"" help:"pending, helped or not_helped."`
	Limit int    `default:"50" help:"Most edits to print."`
}

func (c learnEditsCmd) Run(g *Globals) error {
	return g.call(rpc.MethodLearnEdits, rpc.LearnEditsParams{State: c.State, Limit: c.Limit}, nil)
}
