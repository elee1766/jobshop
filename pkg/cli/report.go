package cli

import (
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// reportArgs are the flags of every `jobshop report <state>` (#186).
type reportArgs struct {
	Note     string `arg:"" optional:"" help:"Short note: what is done, what you wait for, why it is blocked or dropped."`
	Question string `short:"q" help:"Ask a human. Files a needs-human question; the answer is sent to this session."`
	ID       string `short:"t" name:"ticket" help:"Ticket id (default: the session's current ticket)."`
}

func (a reportArgs) send(g *Globals, state, until string) error {
	p := rpc.ReportParams{ID: a.ID, State: state, Note: a.Note, Question: a.Question}
	if until != "" {
		t, err := parseUntil(until, time.Now())
		if err != nil {
			return err
		}
		p.Until = &t
	}
	return g.call(rpc.MethodReport, p, nil)
}

type reportProgressCmd struct{ reportArgs }

func (c reportProgressCmd) Run(g *Globals) error { return c.send(g, rpc.ReportProgress, "") }

type reportWaitingCmd struct {
	reportArgs
	Until string `help:"Release the ticket deferred until then: a duration (2h) or RFC3339 time."`
}

func (c reportWaitingCmd) Run(g *Globals) error { return c.send(g, rpc.ReportWaiting, c.Until) }

type reportReadyCmd struct{ reportArgs }

func (c reportReadyCmd) Run(g *Globals) error { return c.send(g, rpc.ReportReady, "") }

type reportBlockedCmd struct{ reportArgs }

func (c reportBlockedCmd) Run(g *Globals) error { return c.send(g, rpc.ReportBlocked, "") }

type reportFailedCmd struct{ reportArgs }

func (c reportFailedCmd) Run(g *Globals) error { return c.send(g, rpc.ReportFailed, "") }

type reportDroppedCmd struct{ reportArgs }

func (c reportDroppedCmd) Run(g *Globals) error { return c.send(g, rpc.ReportDropped, "") }

type answerCmd struct {
	Question string `arg:"" help:"Question ticket id."`
	Text     string `arg:"" help:"Answer text. Use - to read stdin."`
}

func (c answerCmd) Run(g *Globals) error {
	text, err := readBody(c.Text)
	if err != nil {
		return err
	}
	return g.call(rpc.MethodQuestionAnswer, rpc.AnswerParams{ID: c.Question, Text: text}, nil)
}
