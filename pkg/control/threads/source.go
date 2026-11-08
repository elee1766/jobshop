package threads

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/linear"
)

// Source reads a thread's comments newer than since. Tests replace it;
// the controller uses gh and Linear. It is also what the thread.reply
// wait reads through, so both apply the same filters (waitengine.ReplySource).
type Source interface {
	Comments(ctx context.Context, th Thread, since time.Time) ([]Comment, error)
}

// Mux sends Linear threads to Linear and the rest to GH.
type Mux struct{ GH, Linear Source }

// Comments implements Source.
func (m Mux) Comments(ctx context.Context, th Thread, since time.Time) ([]Comment, error) {
	if th.Linear != "" {
		return m.Linear.Comments(ctx, th, since)
	}
	return m.GH.Comments(ctx, th, since)
}

// GHSource reads GitHub threads with the gh CLI (pkg/github), which has
// the host's token.
type GHSource struct{}

// Comments implements Source.
func (GHSource) Comments(ctx context.Context, th Thread, since time.Time) ([]Comment, error) {
	cs, err := github.CLI{}.Comments(ctx, th.GitHub(), since)
	if err != nil {
		return nil, err
	}
	out := make([]Comment, len(cs))
	for i, cm := range cs {
		out[i] = Comment{Comment: cm}
	}
	return out, nil
}

// LinearSource reads a Linear issue's comments as the identity covering
// its team (#200). Comments by that identity's own user are jobshop's
// sessions talking, so they count as bot comments: seen, not delivered.
type LinearSource struct {
	h Host
}

// Comments implements Source.
func (s LinearSource) Comments(ctx context.Context, th Thread, since time.Time) ([]Comment, error) {
	team, ok := linear.Team(th.Linear)
	if !ok {
		return nil, fmt.Errorf("linear: bad issue key %q", th.Linear)
	}
	cfg := s.h.Config()
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), team)
	if !ok {
		return nil, fmt.Errorf("linear: no identity covers team %s", team)
	}
	var (
		is      linear.Issue
		appSelf string
	)
	err := s.h.LinearDo(ctx, name, func(ctx context.Context, cl *linear.Client) error {
		var err error
		if is, err = cl.GetIssue(ctx, th.Linear); err != nil {
			return err
		}
		// A linear-app's own comments can arrive as its app actor, not
		// isMe: match the viewer's name so a listed bot never wakes on
		// itself (#402).
		if cfg.Identities[name].Kind == config.IdentityLinearApp && cfg.Subscriptions != nil && len(cfg.Subscriptions.Bots) > 0 {
			v, err := cl.Viewer(ctx)
			if err != nil {
				return err
			}
			appSelf = v.Name
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var out []Comment
	byID := map[string]linear.Comment{}
	for _, cm := range is.Comments {
		byID[cm.ID] = cm
	}
	for _, cm := range is.Comments {
		if cm.Bot && appSelf != "" && cm.Author == appSelf {
			cm.Self = true
		}
		out = appendLinearComment(out, is, cm, byID, since)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, nil
}

// appendLinearComment adds cm when it is new or edited after since
// (#595). An edit is delivered at its editedAt with kind "comment
// (edited)": the inbox key is the id plus that time, so each edit is
// delivered once. A new comment edited since takes the later time, so
// the read position passes the edit and it is not sent again. Edits by
// self or bots are not delivered.
func appendLinearComment(out []Comment, is linear.Issue, cm linear.Comment, byID map[string]linear.Comment, since time.Time) []Comment {
	at, kind := cm.CreatedAt, "comment"
	edited := cm.EditedAt != nil && cm.EditedAt.After(since)
	switch {
	case cm.CreatedAt.After(since):
		if edited && cm.EditedAt.After(at) {
			at = *cm.EditedAt
		}
	case edited && !cm.Self && !cm.Bot:
		at, kind = *cm.EditedAt, "comment (edited)"
	default:
		return out
	}
	url := cm.URL
	if url == "" {
		url = is.URL
	}
	return append(out, Comment{
		Comment: github.Comment{ID: cm.ID, Author: cm.Display(), Bot: cm.Bot, Kind: kind, Body: cm.Body, URL: url, At: at},
		Self:    cm.Self, Parent: cm.ParentID,
		Linear: true, AuthorID: cm.AuthorID, Name: cm.Author, Handle: cm.Handle, Email: cm.Email,
		Reply: replyLine(cm, byID), Title: is.Title, IssueState: is.State.Name,
	})
}

// replyLine is the "in reply to" line for a reply: the parent's author
// and first line when it was fetched, else its URL.
func replyLine(cm linear.Comment, byID map[string]linear.Comment) string {
	if cm.ParentID == "" {
		return ""
	}
	p, ok := byID[cm.ParentID]
	if !ok {
		ref := cm.ParentURL
		if ref == "" {
			ref = cm.ParentID
		}
		return "in reply to " + ref
	}
	first, _, _ := strings.Cut(strings.TrimSpace(p.Body), "\n")
	if r := []rune(first); len(r) > replyQuoteMax {
		first = string(r[:replyQuoteMax]) + "..."
	}
	return fmt.Sprintf("in reply to %s: %q", p.Display(), first)
}
