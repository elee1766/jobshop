package control

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Pollers write what they find to the inbox as well (#335), with a
// stable synthetic delivery id and, for comments, the natural key the
// drain gives the webhook copy, so the two dedupe to one item. They keep
// dispatching directly for now; dispatch moves onto inbox rows later
// (#318 item 5). A failed write is logged and does not stop the poll.

// inboxComment records a subscription poller comment.
func (c *Controller) inboxComment(ctx context.Context, th thread, cm ghComment) {
	it := store.InboxItem{ReceivedAt: c.now(), Actor: cm.Author}
	payload := map[string]any{"body": cm.Body, "url": cm.URL, "author": cm.Author, "at": cm.At, "kind": cm.Kind}
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
	it.Payload, _ = json.Marshal(payload)
	if _, _, err := c.store.InboxPut(ctx, it); err != nil {
		c.log.Warn("inbox: poller write", "delivery", it.DeliveryID, "err", err)
	}
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

// inboxActivity records activity poller items (repo events, CI runs and
// Linear changes). Their ids are already stable; they have no natural
// key, since the events API does not give the comment version.
func (c *Controller) inboxActivity(ctx context.Context, items []rpc.ActivityItem) {
	if len(items) == 0 {
		return
	}
	its := make([]store.InboxItem, 0, len(items))
	for _, a := range items {
		if a.Source != ingest.SourceGitHub && a.Source != ingest.SourceLinear {
			continue
		}
		p, _ := json.Marshal(a)
		it := store.InboxItem{
			Source: a.Source, DeliveryID: "poll:activity:" + a.ID, Kind: a.Kind, RepoOrTeam: a.Repo,
			Actor: a.Actor, ReceivedAt: c.now(), Payload: p,
		}
		_ = json.Unmarshal([]byte(a.Ref), &it.Number)
		its = append(its, it)
	}
	if _, err := c.store.InboxPutBatch(ctx, its); err != nil {
		c.log.Warn("inbox: activity write", "items", len(its), "err", err)
	}
}
