// Package question holds what a question ticket carries and how it reads
// (JOB-101, split out of pkg/control by JOB-147). A worker asks a human
// with `report --question`; the controller files a needs-human message
// ticket whose meta.question names the asked ticket and the asking
// session, the asker's context and suggested options. This package builds
// and reads that content. It does not touch the store: the controller
// files, answers and closes questions (grants, stale-option checks and
// every write stay there).
package question

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

const (
	// Resolution closes a question ticket that got its answer.
	Resolution = "answered"
	// AnswerTitlePrefix starts the message that carries an answer to the
	// asking session.
	AnswerTitlePrefix = "Answer to your question "
	// To addresses a question message. No session has it as id, key or
	// pool, so the message is never delivered to an agent.
	To = "human"

	// MaxContext bounds the asker's context.
	MaxContext = 16 << 10

	// recentComments is how many recent comments on the asked ticket the
	// context carries, and maxComment clips each of them.
	recentComments = 3
	maxComment     = 400
	// maxOptionTitle and maxOptionDescription bound an option.
	maxOptionTitle, maxOptionDescription = 120, 1000
)

// Meta is meta.question on a question ticket.
type Meta struct {
	Ticket  string `json:"ticket"`            // the ticket the question is about
	Session string `json:"session,omitempty"` // who asked; the answer goes there
	// Context is the asker's context with the controller's links and
	// recent comments (JOB-101); Options the suggested answers.
	Context string               `json:"context,omitempty"`
	Options []rpc.QuestionOption `json:"options,omitempty"`
}

// Of is meta.question of a question ticket.
func Of(t store.Ticket) (Meta, bool) {
	var m map[string]json.RawMessage
	var q Meta
	if json.Unmarshal(t.Meta, &m) != nil || json.Unmarshal(m[store.MetaQuestion], &q) != nil || q.Ticket == "" {
		return q, false
	}
	return q, true
}

// Options checks the asker's options and numbers them 1..n.
func Options(in []rpc.QuestionOption) ([]rpc.QuestionOption, error) {
	if len(in) > rpc.MaxQuestionOptions {
		return nil, fmt.Errorf("%w: at most %d options, got %d", store.ErrInvalid, rpc.MaxQuestionOptions, len(in))
	}
	var out []rpc.QuestionOption
	for i, o := range in {
		o.ID = i + 1
		o.Title, o.Description = strings.TrimSpace(o.Title), strings.TrimSpace(o.Description)
		if o.Title == "" {
			return nil, fmt.Errorf("%w: option %d has no title", store.ErrInvalid, o.ID)
		}
		if strings.Contains(o.Title, "\n") || len(o.Title) > maxOptionTitle {
			return nil, fmt.Errorf("%w: option %d title must be one line of at most %d bytes", store.ErrInvalid, o.ID, maxOptionTitle)
		}
		if len(o.Description) > maxOptionDescription {
			return nil, fmt.Errorf("%w: option %d description is over %d bytes", store.ErrInvalid, o.ID, maxOptionDescription)
		}
		out = append(out, o)
	}
	return out, nil
}

// Asker is who asked: a session, or an actor when no session did.
type Asker struct{ Session, Actor string }

// Context is the asker's context followed by the links and recent
// comments the controller fills in, so the question reads on its own in
// Desk, the CLI and Linear. t is the ticket asked about, comments its
// thread oldest first.
func Context(asked string, t store.Ticket, by Asker, comments []store.Comment) string {
	var b strings.Builder
	if asked = strings.TrimSpace(asked); asked != "" {
		b.WriteString(asked)
		b.WriteString("\n\n")
	}
	b.WriteString("**Links**\n\n")
	fmt.Fprintf(&b, "- Ticket: %s (%s)\n", t.ID, t.Title)
	if t.Kind == store.KindStep {
		if f, err := store.DecodeFields[store.StepFields](t); err == nil && f.Run != "" {
			fmt.Fprintf(&b, "- Run: %s, step %s\n", f.Run, f.Step)
		}
	}
	for _, l := range metaLinks(t.Meta) {
		fmt.Fprintf(&b, "- %s\n", l)
	}
	if by.Session != "" {
		fmt.Fprintf(&b, "- Asked by session: %s\n", by.Session)
	} else {
		fmt.Fprintf(&b, "- Asked by: %s\n", by.Actor)
	}
	if len(comments) > 0 {
		b.WriteString("\n**Recent comments**\n\n")
		for _, cm := range comments[max(0, len(comments)-recentComments):] {
			body := strings.Join(strings.Fields(cm.Body), " ")
			if len(body) > maxComment {
				body = ClipBytes(body, maxComment) + "..."
			}
			fmt.Fprintf(&b, "- %s %s: %s\n", cm.CreatedAt.UTC().Format("2006-01-02 15:04"), cm.Author, body)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// metaLinks are the issue and PR links in meta.github and meta.linear.
func metaLinks(raw json.RawMessage) []string {
	var m struct {
		GitHub struct {
			Repo  string `json:"repo"`
			Issue any    `json:"issue"`
			PR    any    `json:"pr"`
		} `json:"github"`
		Linear struct {
			Issue     string `json:"issue"`
			Workspace string `json:"workspace"`
		} `json:"linear"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &m) != nil {
		return nil
	}
	var out []string
	if g := m.GitHub; g.Repo != "" {
		if g.Issue != nil {
			out = append(out, fmt.Sprintf("Issue: https://github.com/%s/issues/%v", g.Repo, g.Issue))
		}
		if g.PR != nil {
			out = append(out, fmt.Sprintf("PR: https://github.com/%s/pull/%v", g.Repo, g.PR))
		}
		if g.Issue == nil && g.PR == nil {
			out = append(out, "Repo: https://github.com/"+g.Repo)
		}
	}
	if l := m.Linear; l.Issue != "" {
		if l.Workspace != "" {
			out = append(out, "Linear: https://linear.app/"+url.PathEscape(l.Workspace)+"/issue/"+url.PathEscape(l.Issue))
		} else {
			out = append(out, "Linear: "+l.Issue)
		}
	}
	return out
}

// OptionsText lists options for a ticket body or prompt.
func OptionsText(opts []rpc.QuestionOption) string {
	var b strings.Builder
	for _, o := range opts {
		fmt.Fprintf(&b, "%d. %s", o.ID, o.Title)
		if o.Recommended {
			b.WriteString(" (recommended)")
		}
		if o.Description != "" {
			b.WriteString(": " + strings.ReplaceAll(o.Description, "\n", " "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ChosenText is what the asking session gets for an answer that picks an
// option.
func ChosenText(o rpc.QuestionOption, note string) string {
	s := fmt.Sprintf("Chose option %d: %s", o.ID, o.Title)
	if o.Description != "" {
		s += "\n" + o.Description
	}
	if note != "" {
		s += "\n\nNote: " + note
	}
	return s
}

// ClipBytes cuts s to at most n bytes on a rune boundary.
func ClipBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

// ChatPrompt is the question part of an attention chat's first prompt,
// with the asker's context and options (JOB-101) last. id is the question
// ticket.
func ChatPrompt(id string, qm Meta) string {
	s := " The operator talks to you here, not to the asking session. When they give an answer, relay it with `jobshop answer " + id +
		" \"<text>\"`: write it as clear text that stands on its own (`yes, merge PR 12`), never a bare option number or letter, and do not offer numbered choices of your own." +
		" After answering, the controller tells you here how the asking session reacted: resolved (the ticket is in progress, closed or ready), parked (a wait or defer, with its reason) or re-asked (a new question from the same session, which you show the operator and answer in this chat the same way)." +
		" Report each in a line or two. The chat ends on its own once the answer settled and the operator leaves."
	if len(qm.Options) > 0 {
		s += "\n\nThe asker suggested these options; the operator may pick one or answer freely. When they pick one, relay it with `jobshop answer " + id +
			" --option N \"<their note, if any>\"`.\n\n" + OptionsText(qm.Options)
	}
	if qm.Context != "" {
		s += "\n\nThe asker's context:\n\n" + qm.Context + "\n"
	}
	return s
}
