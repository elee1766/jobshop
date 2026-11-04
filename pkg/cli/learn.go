package cli

import (
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// learnCmd is `jobshop learn` (#264).
type learnCmd struct {
	Correction learnCorrectionCmd `cmd:"" help:"Record a correction: a human or reviewer told you to do it differently."`
	Failure    learnFailureCmd    `cmd:"" help:"Record a failure and its cause."`
	Workaround learnWorkaroundCmd `cmd:"" help:"Record a workaround for a tool or environment problem."`
	List       learnListCmd       `cmd:"" help:"List learnings, newest first."`
	Update     learnUpdateCmd     `cmd:"" help:"Set the status or PR of learnings (curator, #273)."`
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

func (c learnCorrectionCmd) Run(g *Globals) error { return c.send(g, store.LearningCorrection) }

type learnFailureCmd struct{ learnKindCmd }

func (c learnFailureCmd) Run(g *Globals) error { return c.send(g, store.LearningFailure) }

type learnWorkaroundCmd struct{ learnKindCmd }

func (c learnWorkaroundCmd) Run(g *Globals) error { return c.send(g, store.LearningWorkaround) }

type learnListCmd struct {
	Since  time.Duration `help:"Only learnings newer than this, e.g. 24h."`
	Scope  string        `help:"global, repo:<owner/name> or pool:<name>."`
	Kind   string        `enum:",correction,failure,workaround" default:"" help:"correction, failure or workaround."`
	Status string        `enum:",open,applied,rejected" default:"" help:"open, applied or rejected."`
	Ticket string        `short:"t" help:"Only learnings about this ticket."`
	Limit  int           `default:"50" help:"Most learnings to print."`
}

func (c learnListCmd) Run(g *Globals) error {
	return g.call(rpc.MethodLearnList, rpc.LearnListParams{
		Since: c.Since, Scope: c.Scope, Kind: c.Kind, Status: c.Status, Ticket: c.Ticket, Limit: c.Limit,
	}, nil)
}

// learnUpdateCmd is `jobshop learn update` (#273): the curator records
// the PR addressing learnings, then applied (merged) or rejected (closed
// unmerged or reverted; needs the PR).
type learnUpdateCmd struct {
	IDs    []int64 `arg:"" name:"id" help:"Learning ids."`
	Status string  `enum:",open,applied,rejected" default:"" help:"open, applied or rejected."`
	PR     string  `name:"pr" help:"PR URL that addresses them."`
}

func (c learnUpdateCmd) Run(g *Globals) error {
	return g.call(rpc.MethodLearnUpdate, rpc.LearnUpdateParams{IDs: c.IDs, Status: c.Status, PR: c.PR}, nil)
}
