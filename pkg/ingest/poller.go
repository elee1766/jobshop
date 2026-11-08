package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/store"
)

// Poller reads GitHub repos that send no webhook and writes inbox rows
// shaped like webhook deliveries (#341): the same kind, a payload subset
// with the fields the drain reads, and the drain's natural key, so a
// webhook and a poll of the same comment dedupe to one row. The endpoint
// on each row is the poller name.
//
// Streams, each with its own cursor (newest updated_at handled) and
// ETag: "issues" (GET /repos/R/issues, issues and pull requests),
// "comments" (GET /repos/R/issues/comments) and "reviews" (each pull
// request the issues stream saw updated). A repo with no cursor starts
// at now: older activity is not news.
type Poller struct {
	Name   string
	Events []string
	// Client returns the GitHub client for repo.
	Client func(repo string) (*github.Client, error)
	Store  PollStore
	Now    func() time.Time
	// MaxPages bounds pages per stream and pass; 0 is 10.
	MaxPages int
}

// PollStore is the part of the store a poller uses.
type PollStore interface {
	Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error)
}

// PollResult counts one repo pass.
type PollResult struct {
	Items int // rows written, before dedupe
	Fresh int // new inbox rows
	// New are the new inbox rows, for the controller's wake (#679).
	New []store.InboxItem
}

func (p *Poller) wants(ev string) bool { return slices.Contains(p.Events, ev) }

// Poll reads one repo. Cursors move only after the rows are stored, so a
// failure re-reads the same window and the inbox dedupes it.
func (p *Poller) Poll(ctx context.Context, repo string) (PollResult, error) {
	var res PollResult
	cl, err := p.Client(repo)
	if err != nil {
		return res, err
	}
	g := ghGet{cl: cl, max: p.MaxPages}

	var items []store.InboxItem
	var cursors []store.PollCursor
	var prs []int64
	if p.wants("issues") || p.wants("pull_request") || p.wants("pull_request_review") {
		cur, err := p.cursor(ctx, repo, "issues")
		if err != nil {
			return res, err
		}
		if cur.Found {
			var issues []ghIssue
			etag, err := g.list(ctx, listURL(repo, "issues", cur.Since, "state=all"), cur.ETag, &issues)
			if err != nil {
				return res, err
			}
			if etag != "" || len(issues) > 0 {
				cur.ETag = etag
			}
			for _, is := range issues {
				if is.UpdatedAt.After(cur.Since) {
					cur.Since = is.UpdatedAt
				}
				it, ok, err := p.issueItem(repo, is)
				if err != nil {
					return res, err
				}
				if ok {
					items = append(items, it)
				}
				if is.PullRequest != nil {
					prs = append(prs, is.Number)
				}
			}
		}
		cursors = append(cursors, cur)
	}
	if p.wants("pull_request_review") {
		cur, err := p.cursor(ctx, repo, "reviews")
		if err != nil {
			return res, err
		}
		if cur.Found {
			since := cur.Since
			for _, n := range prs {
				var revs []ghReview
				if _, err := g.list(ctx, fmt.Sprintf("/repos/%s/pulls/%d/reviews?per_page=100", repo, n), "", &revs); err != nil {
					return res, err
				}
				for _, rv := range revs {
					if !rv.SubmittedAt.After(since) {
						continue
					}
					if rv.SubmittedAt.After(cur.Since) {
						cur.Since = rv.SubmittedAt
					}
					it, err := p.reviewItem(repo, n, rv)
					if err != nil {
						return res, err
					}
					items = append(items, it)
				}
			}
		}
		cursors = append(cursors, cur)
	}
	if p.wants("issue_comment") {
		cur, err := p.cursor(ctx, repo, "comments")
		if err != nil {
			return res, err
		}
		if cur.Found {
			var cms []ghIssueComment
			etag, err := g.list(ctx, listURL(repo, "issues/comments", cur.Since, ""), cur.ETag, &cms)
			if err != nil {
				return res, err
			}
			if etag != "" || len(cms) > 0 {
				cur.ETag = etag
			}
			for _, cm := range cms {
				if cm.UpdatedAt.After(cur.Since) {
					cur.Since = cm.UpdatedAt
				}
				it, err := p.commentItem(repo, cm)
				if err != nil {
					return res, err
				}
				items = append(items, it)
			}
		}
		cursors = append(cursors, cur)
	}

	if len(items) > 0 {
		fresh, err := InboxPutBatch(ctx, p.Store, items)
		if err != nil {
			return res, fmt.Errorf("poller %s: inbox: %w", p.Name, err)
		}
		res.Items, res.Fresh, res.New = len(items), len(fresh), fresh
	}
	for _, c := range cursors {
		if err := SetPollCursor(ctx, p.Store, c); err != nil {
			return res, err
		}
	}
	return res, nil
}

// cursor loads a stream's cursor. A missing one starts at now and is
// stored by the caller, so the first pass reads nothing.
func (p *Poller) cursor(ctx context.Context, repo, stream string) (store.PollCursor, error) {
	c, err := GetPollCursor(ctx, p.Store, p.Name, repo, stream)
	if err != nil {
		return c, err
	}
	if !c.Found {
		c.Since = p.Now().UTC().Truncate(time.Second)
	}
	return c, nil
}

// listURL asks for items updated at or after since, oldest first. The
// boundary item comes again next pass; its delivery id dedupes it.
func listURL(repo, path string, since time.Time, extra string) string {
	q := url.Values{}
	q.Set("sort", "updated")
	q.Set("direction", "asc")
	q.Set("per_page", "100")
	q.Set("since", since.UTC().Format(time.RFC3339))
	u := fmt.Sprintf("/repos/%s/%s?%s", repo, path, q.Encode())
	if extra != "" {
		u += "&" + extra
	}
	return u
}

type ghUser struct {
	Login string `json:"login"`
}

type ghIssue struct {
	Number      int64     `json:"number"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	HTMLURL     string    `json:"html_url"`
	User        ghUser    `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PullRequest *struct {
		MergedAt *time.Time `json:"merged_at"`
	} `json:"pull_request"`
}

type ghIssueComment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	IssueURL  string    `json:"issue_url"`
	User      ghUser    `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ghReview struct {
	ID          int64     `json:"id"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	HTMLURL     string    `json:"html_url"`
	User        ghUser    `json:"user"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// item builds a row from a webhook-shaped payload, with the drain's
// subject and natural key.
func (p *Poller) item(kind, deliveryID string, payload map[string]any) (store.InboxItem, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return store.InboxItem{}, err
	}
	it := store.InboxItem{
		Source: SourceGitHub, Endpoint: p.Name, Kind: kind, ReceivedAt: p.Now(), Payload: b,
	}
	if err := githubSubject(&it, b); err != nil {
		return it, err
	}
	// Comments and reviews take their delivery id from the natural key,
	// as the subscription poller's rows do.
	if it.NaturalKey != "" {
		it.DeliveryID = "poll:" + it.NaturalKey
	} else {
		it.DeliveryID = deliveryID
	}
	return it, nil
}

func repoObj(repo string) map[string]any { return map[string]any{"full_name": repo} }

// issueItem maps an issue or pull request. Polling sees the latest
// state, not each change, so the action is opened for one created since
// it was last seen and updated otherwise.
func (p *Poller) issueItem(repo string, is ghIssue) (store.InboxItem, bool, error) {
	kind, key := "issues", "issue"
	if is.PullRequest != nil {
		kind, key = "pull_request", "pull_request"
	}
	if !p.wants(kind) {
		return store.InboxItem{}, false, nil
	}
	action := "updated"
	if is.CreatedAt.Equal(is.UpdatedAt) {
		action = "opened"
	}
	obj := map[string]any{
		"number": is.Number, "title": is.Title, "state": is.State, "html_url": is.HTMLURL,
		"user": is.User, "created_at": is.CreatedAt, "updated_at": is.UpdatedAt,
	}
	if is.PullRequest != nil && is.PullRequest.MergedAt != nil {
		obj["merged_at"] = is.PullRequest.MergedAt
	}
	id := fmt.Sprintf("poll:github:%s:%s:%d:%s", kind, strings.ToLower(repo), is.Number, is.UpdatedAt.UTC().Format(time.RFC3339))
	it, err := p.item(kind, id, map[string]any{
		"action": action, "repository": repoObj(repo), "sender": is.User, key: obj,
	})
	return it, err == nil, err
}

func (p *Poller) commentItem(repo string, cm ghIssueComment) (store.InboxItem, error) {
	n, _ := strconv.ParseInt(cm.IssueURL[strings.LastIndex(cm.IssueURL, "/")+1:], 10, 64)
	action := "edited"
	if cm.CreatedAt.Equal(cm.UpdatedAt) {
		action = "created"
	}
	return p.item("issue_comment", "", map[string]any{
		"action": action, "repository": repoObj(repo), "sender": cm.User,
		"issue": map[string]any{"number": n},
		"comment": map[string]any{
			"id": cm.ID, "body": cm.Body, "html_url": cm.HTMLURL, "user": cm.User,
			"created_at": cm.CreatedAt, "updated_at": cm.UpdatedAt,
		},
	})
}

func (p *Poller) reviewItem(repo string, pr int64, rv ghReview) (store.InboxItem, error) {
	return p.item("pull_request_review", "", map[string]any{
		"action": "submitted", "repository": repoObj(repo), "sender": rv.User,
		"pull_request": map[string]any{"number": pr},
		"review": map[string]any{
			"id": rv.ID, "body": rv.Body, "state": rv.State, "html_url": rv.HTMLURL,
			"user": rv.User, "submitted_at": rv.SubmittedAt,
		},
	})
}

// ghGet lists GitHub collections with ETag and pagination.
type ghGet struct {
	cl  *github.Client
	max int
}

// list GETs path and its Link next pages into out (a slice pointer). The
// first request sends etag as If-None-Match; a 304 leaves out empty and
// returns etag unchanged. It returns the first page's ETag.
func (g ghGet) list(ctx context.Context, path, etag string, out any) (string, error) {
	limit := g.max
	if limit <= 0 {
		limit = 10
	}
	items, resp, err := g.cl.List(ctx, github.Request{Path: path, ETag: etag, Route: "/repos/{repo}/poll"}, limit)
	if err != nil {
		return "", err
	}
	if resp.NotModified() {
		return etag, nil
	}
	b, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	return resp.Header.Get("ETag"), json.Unmarshal(b, out)
}
