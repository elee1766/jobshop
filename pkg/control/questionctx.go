package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Question context and options (JOB-101). The asker writes what it was
// doing and why it cannot decide; the controller adds the links a human
// needs to check it (ticket, run and step, issue or PR, the asking
// session) and the ticket's last comments, so the question reads on its
// own in Desk, the CLI and Linear.

const (
	// questionComments is how many recent comments on the asked ticket
	// the context carries.
	questionComments = 3
	// maxQuestionComment clips each of them.
	maxQuestionComment = 400
	// maxQuestionContext bounds the asker's context.
	maxQuestionContext = 16 << 10
	// maxOptionTitle and maxOptionDescription bound an option.
	maxOptionTitle, maxOptionDescription = 120, 1000
)

// questionOptions checks the asker's options and numbers them 1..n.
func questionOptions(in []rpc.QuestionOption) ([]rpc.QuestionOption, error) {
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

// questionContext is the asker's context followed by the links and
// recent comments the controller fills in.
func (c *Controller) questionContext(ctx context.Context, sc scope, t store.Ticket, asked string) string {
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
	if sc.session != "" {
		fmt.Fprintf(&b, "- Asked by session: %s\n", sc.session)
	} else {
		fmt.Fprintf(&b, "- Asked by: %s\n", sc.actor)
	}
	cs, err := store.Comments(ctx, c.store, t.ID)
	if err == nil && len(cs) > 0 {
		b.WriteString("\n**Recent comments**\n\n")
		for _, cm := range cs[max(0, len(cs)-questionComments):] {
			body := strings.Join(strings.Fields(cm.Body), " ")
			if len(body) > maxQuestionComment {
				body = clipBytes(body, maxQuestionComment) + "..."
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

// optionsText lists options for a ticket body or prompt.
func optionsText(opts []rpc.QuestionOption) string {
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

// chosenText is what the asking session gets for an answer that picks
// an option.
func chosenText(o rpc.QuestionOption, note string) string {
	s := fmt.Sprintf("Chose option %d: %s", o.ID, o.Title)
	if o.Description != "" {
		s += "\n" + o.Description
	}
	if note != "" {
		s += "\n\nNote: " + note
	}
	return s
}

// clipBytes cuts s to at most n bytes on a rune boundary.
func clipBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
