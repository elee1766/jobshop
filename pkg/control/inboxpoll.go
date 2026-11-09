package control

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/control/threads"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// Pollers only write the inbox (#335, #442). Each find gets a stable
// synthetic delivery id and, for comments, the natural key the drain gives
// the webhook copy. The dispatcher (dispatch.go) delivers admitted rows.

// pollComment is a subscription poller row's payload: the thread and the
// comment, as admission hooks see and may rewrite them.
type pollComment struct {
	Thread  threads.Thread  `json:"thread"`
	Comment pollCommentBody `json:"comment"`
}

// pollCommentBody is ghComment with JSON names.
type pollCommentBody struct {
	ID     string    `json:"id"`
	Author string    `json:"author"`
	Bot    bool      `json:"bot,omitempty"`
	Self   bool      `json:"self,omitempty"`
	Kind   string    `json:"kind"`
	State  string    `json:"state,omitempty"`
	Body   string    `json:"body"`
	URL    string    `json:"url,omitempty"`
	At     time.Time `json:"at"`
	Parent string    `json:"parent,omitempty"`

	Linear     bool   `json:"linear,omitempty"`
	AuthorID   string `json:"author_id,omitempty"`
	Name       string `json:"name,omitempty"`
	Handle     string `json:"handle,omitempty"`
	Email      string `json:"email,omitempty"`
	Reply      string `json:"reply,omitempty"`
	Title      string `json:"title,omitempty"`
	IssueState string `json:"issue_state,omitempty"`
}

func newPollCommentBody(cm threads.Comment) pollCommentBody {
	return pollCommentBody{
		ID: cm.ID, Author: cm.Author, Bot: cm.Bot, Self: cm.Self, Kind: cm.Kind, State: cm.State,
		Body: cm.Body, URL: cm.URL, At: cm.At, Parent: cm.Parent,
		Linear: cm.Linear, AuthorID: cm.AuthorID, Name: cm.Name, Handle: cm.Handle, Email: cm.Email,
		Reply: cm.Reply, Title: cm.Title, IssueState: cm.IssueState,
	}
}

func (b pollCommentBody) comment() threads.Comment {
	return threads.Comment{
		Comment: github.Comment{ID: b.ID, Author: b.Author, Bot: b.Bot, Kind: b.Kind, State: b.State, Body: b.Body, URL: b.URL, At: b.At},
		Self:    b.Self, Parent: b.Parent,
		Linear: b.Linear, AuthorID: b.AuthorID, Name: b.Name, Handle: b.Handle, Email: b.Email,
		Reply: b.Reply, Title: b.Title, IssueState: b.IssueState,
	}
}

// inboxComment records a subscription poller comment. A webhook copy
// already stored under the same natural key is not dispatched (the
// dispatcher only delivers poller rows), so the poller then stores its
// own row without the natural key.
func (c *Controller) inboxComment(ctx context.Context, th threads.Thread, cm threads.Comment) error {
	it := domain.InboxItem{ReceivedAt: c.now(), Actor: cm.Author}
	seenless := th
	seenless.Seen = time.Time{}
	p := pollComment{Thread: seenless, Comment: newPollCommentBody(cm)}
	if th.Linear != "" {
		team, num, _ := strings.Cut(th.Linear, "-")
		it.Source, it.RepoOrTeam, it.Kind, it.CommentID = ingest.SourceLinear, team, "Comment", cm.ID
		_ = json.Unmarshal([]byte(num), &it.Number)
		it.NaturalKey = ingest.LinearKey(cm.ID, cm.At)
		it.DeliveryID = "poll:" + it.NaturalKey
	} else {
		it.Source, it.RepoOrTeam, it.Number = ingest.SourceGitHub, th.Repo, int64(th.Number)
		kind, id := ghCommentKind(cm.ID)
		it.Kind, it.CommentID = kind, id
		if id != "" {
			it.NaturalKey = ingest.GitHubKey(th.Repo, strings.TrimPrefix(kind, "pull_request_"), id, cm.At)
			it.DeliveryID = "poll:" + it.NaturalKey
		} else {
			it.DeliveryID = "poll:github:" + th.Repo + ":" + cm.ID + ":" + cm.At.UTC().Format("2006-01-02T15:04:05Z")
		}
	}
	var err error
	if it.Payload, err = json.Marshal(p); err != nil {
		return err
	}
	return c.queue().PutPolled(ctx, it)
}

// ghCommentKind maps the subscription poller's prefixed ids (c, r, p) to
// the webhook event kind and the bare id. The natural key kinds are
// comment, review and review_comment, the same as the drain's.
func ghCommentKind(id string) (kind, bare string) {
	if len(id) < 2 {
		return "", ""
	}
	switch id[0] {
	case 'c':
		return "comment", id[1:]
	case 'r':
		return "pull_request_review", id[1:]
	case 'p':
		return "pull_request_review_comment", id[1:]
	}
	return "", ""
}

// activityPrefix starts an activity poller row's delivery id.
const activityPrefix = "poll:activity:"

// inboxActivity records activity poller items (repo events, CI runs and
// Linear changes). Their ids are already stable; they have no natural
// key, since the events API does not give the comment version. The
// payload is the rpc.ActivityItem.
func (c *Controller) inboxActivity(ctx context.Context, items []rpc.ActivityItem) {
	if len(items) == 0 {
		return
	}
	its := make([]domain.InboxItem, 0, len(items))
	for _, a := range items {
		if a.Source != ingest.SourceGitHub && a.Source != ingest.SourceLinear {
			continue
		}
		p, _ := json.Marshal(a)
		it := domain.InboxItem{
			Source: a.Source, DeliveryID: activityPrefix + a.ID, Kind: a.Kind, RepoOrTeam: a.Repo,
			Actor: a.Actor, ReceivedAt: c.now(), Payload: p,
		}
		_ = json.Unmarshal([]byte(a.Ref), &it.Number)
		its = append(its, it)
	}
	if err := c.queue().PutBatch(ctx, its); err != nil {
		c.log.WarnContext(ctx, "inbox: activity write", "items", len(its), "err", err)
	}
}
