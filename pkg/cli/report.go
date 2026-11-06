package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// reportArgs are the flags of every `jobshop report <state>` (#186).
type reportArgs struct {
	Note     string `arg:"" optional:"" help:"Short note: what is done, what you wait for, why it is blocked or dropped."`
	Question string `short:"q" help:"Ask a human. Files a needs-human question; the answer is sent to this session. Needs --context; offer --option when there are concrete alternatives."`
	// Context and options make a question answerable without opening the
	// ticket (JOB-101).
	Context     string   `help:"Markdown context for -q, required with it (- reads stdin): what you were doing, what you found, why you cannot decide alone, and what happens under each answer. At least two lines. The controller adds links and recent comments."`
	Option      []string `sep:"none" placeholder:"TITLE|DESCRIPTION" help:"A suggested answer for -q, repeatable (at most 5): a short title, then | and what choosing it does. Desk shows each as a button; the human may still answer freely."`
	OptionsJSON string   `name:"options-json" placeholder:"JSON" help:"Options for -q as JSON (- reads stdin): [{\"title\":...,\"description\":...,\"recommended\":true}]. For long text; not with --option."`
	Recommend   int      `placeholder:"N" help:"Mark option N (1-based) as your recommendation."`
	ID          string   `short:"t" name:"ticket" help:"Ticket id (default: the session's current ticket)."`
}

// minContextLines is how many non-empty lines a question's context needs:
// one line is a restated question, not context.
const minContextLines = 2

func (a reportArgs) send(g *Globals, state, until string) error {
	p := rpc.ReportParams{ID: a.ID, State: state, Note: a.Note, Question: a.Question}
	if err := a.question(&p); err != nil {
		return err
	}
	if until != "" {
		t, err := parseUntil(until, time.Now())
		if err != nil {
			return err
		}
		p.Until = &t
	}
	return g.call(rpc.MethodReport, p, nil)
}

// question fills the context and options of p, checking them.
func (a reportArgs) question(p *rpc.ReportParams) error {
	if a.Context == "-" && a.OptionsJSON == "-" {
		return errors.New("--context and --options-json cannot both read stdin")
	}
	if strings.TrimSpace(a.Question) == "" {
		if a.Context != "" || len(a.Option) > 0 || a.OptionsJSON != "" || a.Recommend != 0 {
			return errors.New("--context, --option, --options-json and --recommend go with -q")
		}
		return nil
	}
	ctx, err := readBody(a.Context)
	if err != nil {
		return err
	}
	if n := contextLines(ctx); n < minContextLines {
		return fmt.Errorf("-q needs --context with at least %d lines (got %d): what you were doing, what you found, why you cannot decide alone, and what happens under each answer", minContextLines, n)
	}
	p.Context = ctx
	if p.Options, err = a.options(); err != nil {
		return err
	}
	if a.Recommend != 0 {
		if a.Recommend < 1 || a.Recommend > len(p.Options) {
			return fmt.Errorf("--recommend %d: there are %d options", a.Recommend, len(p.Options))
		}
		p.Options[a.Recommend-1].Recommended = true
	}
	return nil
}

func (a reportArgs) options() ([]rpc.QuestionOption, error) {
	if a.OptionsJSON != "" && len(a.Option) > 0 {
		return nil, errors.New("use --option or --options-json, not both")
	}
	var out []rpc.QuestionOption
	if a.OptionsJSON != "" {
		b, err := readBody(a.OptionsJSON)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(b), &out); err != nil {
			return nil, fmt.Errorf("--options-json: %w", err)
		}
	}
	for _, o := range a.Option {
		title, desc, _ := strings.Cut(o, "|")
		out = append(out, rpc.QuestionOption{Title: strings.TrimSpace(title), Description: strings.TrimSpace(desc)})
	}
	if len(out) > rpc.MaxQuestionOptions {
		return nil, fmt.Errorf("at most %d options, got %d", rpc.MaxQuestionOptions, len(out))
	}
	for i, o := range out {
		if strings.TrimSpace(o.Title) == "" {
			return nil, fmt.Errorf("option %d has no title", i+1)
		}
	}
	return out, nil
}

func contextLines(s string) int {
	n := 0
	for l := range strings.Lines(s) {
		if strings.TrimSpace(l) != "" {
			n++
		}
	}
	return n
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
	Text     string `arg:"" optional:"" help:"Answer text, or with --option a note. Use - to read stdin."`
	Option   int    `placeholder:"N" help:"Pick the question's option N (1-based); the asking session gets the option's title and description, plus the note."`
}

func (c answerCmd) Run(g *Globals) error {
	text, err := readBody(c.Text)
	if err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" && c.Option == 0 {
		return errors.New("give the answer text, or --option N")
	}
	return g.call(rpc.MethodQuestionAnswer, rpc.AnswerParams{ID: c.Question, Text: text, Option: c.Option}, nil)
}
