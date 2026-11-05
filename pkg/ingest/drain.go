package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
)

// DrainBatch is how many spool rows one drain step moves.
const DrainBatch = 100

// DrainResult counts one drain run.
type DrainResult struct {
	Moved int `json:"moved"` // spool rows removed after the inbox commit
	Fresh int `json:"fresh"` // of those, new inbox items (the rest deduped)
	Bad   int `json:"bad"`   // rows moved to spool_bad
}

// Drain moves spool rows into the inbox in id order until the spool is
// empty (#335). Each batch is stored in the inbox and only then deleted
// from the spool, so a crash between the two, or a failure part way
// through a batch, re-drains it and the inbox dedupes it. A row that
// does not parse moves to spool_bad with the error; it is never dropped.
func Drain(ctx context.Context, sp *Spool, in InboxStore, now func() time.Time, log *slog.Logger) (DrainResult, error) {
	var res DrainResult
	for {
		rows, err := sp.Rows(ctx, DrainBatch)
		if err != nil {
			return res, err
		}
		if len(rows) == 0 {
			return res, nil
		}
		items := make([]store.InboxItem, 0, len(rows))
		ids := make([]int64, 0, len(rows))
		for _, r := range rows {
			it, err := InboxItem(r)
			if err != nil {
				log.Warn("ingest drain: bad spool row", "id", r.ID, "endpoint", r.Endpoint, "delivery", r.DeliveryID, "err", err)
				if err := sp.MarkBad(ctx, r.ID, err.Error(), now()); err != nil {
					return res, err
				}
				res.Bad++
				continue
			}
			items = append(items, it)
			ids = append(ids, r.ID)
		}
		fresh, err := InboxPutBatch(ctx, in, items)
		if err != nil {
			return res, fmt.Errorf("ingest drain: inbox: %w", err)
		}
		if err := sp.Delete(ctx, ids); err != nil {
			return res, fmt.Errorf("ingest drain: spool delete: %w", err)
		}
		res.Moved += len(ids)
		res.Fresh += fresh
	}
}

// InboxItem maps a spool row to an inbox item with its subject and
// natural key.
func InboxItem(r Row) (store.InboxItem, error) {
	d, err := r.Delivery()
	if err != nil {
		return store.InboxItem{}, err
	}
	if d.Source != SourceGitHub && d.Source != SourceLinear && d.Source != SourceGeneric {
		return store.InboxItem{}, fmt.Errorf("unknown source %q", d.Source)
	}
	it := store.InboxItem{
		Source: d.Source, DeliveryID: d.Source + ":" + d.DeliveryID, Endpoint: d.Endpoint,
		Kind: d.Kind, ReceivedAt: d.ReceivedAt, Payload: d.Payload,
	}
	switch d.Source {
	case SourceGitHub:
		err = githubSubject(&it, d.Payload)
	case SourceGeneric:
		err = genericSubject(&it, d)
	default:
		err = linearSubject(&it, d.Payload)
	}
	return it, err
}

// GitHub payload fields used for the subject and natural key.
type ghPayload struct {
	Repository *struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Sender *struct {
		Login string `json:"login"`
	} `json:"sender"`
	Issue *struct {
		Number int64 `json:"number"`
	} `json:"issue"`
	PullRequest *struct {
		Number int64 `json:"number"`
	} `json:"pull_request"`
	Comment *struct {
		ID        int64     `json:"id"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"comment"`
	Review *struct {
		ID          int64     `json:"id"`
		SubmittedAt time.Time `json:"submitted_at"`
	} `json:"review"`
}

func githubSubject(it *store.InboxItem, payload json.RawMessage) error {
	var p ghPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("github payload: %w", err)
	}
	if p.Repository != nil {
		it.RepoOrTeam = p.Repository.FullName
	}
	if p.Sender != nil {
		it.Actor = p.Sender.Login
	}
	switch {
	case p.Issue != nil:
		it.Number = p.Issue.Number
	case p.PullRequest != nil:
		it.Number = p.PullRequest.Number
	}
	switch {
	case it.Kind == "pull_request_review" && p.Review != nil:
		it.CommentID = fmt.Sprint(p.Review.ID)
		it.NaturalKey = GitHubKey(it.RepoOrTeam, "review", it.CommentID, p.Review.SubmittedAt)
	case it.Kind == "pull_request_review_comment" && p.Comment != nil:
		it.CommentID = fmt.Sprint(p.Comment.ID)
		it.NaturalKey = GitHubKey(it.RepoOrTeam, "review_comment", it.CommentID, p.Comment.UpdatedAt)
	case it.Kind == "issue_comment" && p.Comment != nil:
		it.CommentID = fmt.Sprint(p.Comment.ID)
		it.NaturalKey = GitHubKey(it.RepoOrTeam, "comment", it.CommentID, p.Comment.UpdatedAt)
	}
	return nil
}

func linearSubject(it *store.InboxItem, payload json.RawMessage) error {
	var ev LinearEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return fmt.Errorf("linear payload: %w", err)
	}
	key := ev.IssueKey()
	if team, n, ok := strings.Cut(key, "-"); ok {
		it.RepoOrTeam = team
		_, _ = fmt.Sscan(n, &it.Number)
	}
	if ev.Actor != nil {
		it.Actor = ev.Actor.ID
	}
	if ev.Type == "Comment" && ev.Data.ID != "" {
		it.CommentID = ev.Data.ID
		it.NaturalKey = LinearKey(ev.Data.ID, ev.Data.CreatedAt)
	}
	return nil
}

// GitHubKey is the natural key of a GitHub comment, review or review
// comment at one version (#335). kind is comment, review or
// review_comment: their ids are separate sequences. at is updated_at
// (submitted_at for a review), so an edit is a new event.
func GitHubKey(repo, kind, id string, at time.Time) string {
	return fmt.Sprintf("github:%s:%s:%s:%s", kind, strings.ToLower(repo), id, at.UTC().Format(time.RFC3339))
}

// LinearKey is the natural key of a Linear comment.
func LinearKey(id string, created time.Time) string {
	return fmt.Sprintf("linear:comment:%s:%s", id, created.UTC().Format(time.RFC3339))
}
