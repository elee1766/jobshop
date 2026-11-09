package browse

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

var (
	githubLogin = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}(\[bot\])?$`)
	githubRef   = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9_.-]+)#([1-9][0-9]{0,8})$`)
	// A Linear key in text, for the PR to issue link.
	keyInText = regexp.MustCompile(`\b([A-Z][A-Z0-9]{1,9})-([1-9][0-9]{0,6})\b`)
)

const githubListFields = `__typename
      ... on PullRequest { number title url state isDraft merged updatedAt author { login __typename }
        reviewDecision labels(first: 10) { nodes { name } }
        commits(last: 1) { nodes { commit { statusCheckRollup { state } } } } }
      ... on Issue { number title url state updatedAt author { login __typename } labels(first: 10) { nodes { name } } }`

const githubListQuery = `query($q: String!, $first: Int!, $after: String) {
  search(query: $q, type: ISSUE, first: $first, after: $after) {
    issueCount
    pageInfo { hasNextPage endCursor }
    nodes { ` + githubListFields + ` }
  }
}`

type ghActor struct {
	Login string `json:"login"`
	Type  string `json:"__typename"`
}

type ghNode struct {
	Type      string    `json:"__typename"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	State     string    `json:"state"`
	IsDraft   bool      `json:"isDraft"`
	Merged    bool      `json:"merged"`
	UpdatedAt time.Time `json:"updatedAt"`
	Author    *ghActor  `json:"author"`
	Review    string    `json:"reviewDecision"`
	Labels    struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				Rollup *struct {
					State string `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
}

// ghState is open, draft, merged or closed.
func ghState(n ghNode) string {
	switch {
	case n.Merged || n.State == "MERGED":
		return "merged"
	case n.State == "CLOSED":
		return "closed"
	case n.IsDraft:
		return "draft"
	}
	return "open"
}

func ghChecks(state string) string {
	switch state {
	case "SUCCESS":
		return "pass"
	case "FAILURE", "ERROR":
		return "fail"
	case "PENDING", "EXPECTED":
		return "run"
	}
	return ""
}

func ghReview(d string) string {
	switch d {
	case "APPROVED":
		return "approved"
	case "CHANGES_REQUESTED":
		return "changes"
	case "REVIEW_REQUIRED":
		return "required"
	}
	return ""
}

func (n ghNode) item(repo string) rpc.BrowseItem {
	st := ghState(n)
	it := rpc.BrowseItem{
		Provider: rpc.BrowseGitHub, Scope: repo, Ref: repo + "#" + strconv.Itoa(n.Number), ID: "#" + strconv.Itoa(n.Number),
		Kind: "issue", Title: n.Title, State: st, StateType: st, Labels: []string{}, UpdatedAt: n.UpdatedAt, URL: n.URL,
		Review: ghReview(n.Review),
	}
	if n.Type == "PullRequest" {
		it.Kind = "pr"
		if len(n.Commits.Nodes) > 0 && n.Commits.Nodes[0].Commit.Rollup != nil {
			it.Checks = ghChecks(n.Commits.Nodes[0].Commit.Rollup.State)
		}
	}
	if n.Author != nil {
		it.Owner, it.OwnerBot = n.Author.Login, n.Author.Type == "Bot"
	}
	for _, l := range n.Labels.Nodes {
		it.Labels = append(it.Labels, l.Name)
	}
	return it
}

// githubSearch builds the provider search for a list. The scope qualifier
// is always the configured repo, and the person's words get none of their
// own: a word with a colon has it replaced, so repo:, org: or user: in the
// box cannot widen the search.
func githubSearch(p rpc.BrowseListParams, kind string) (string, error) {
	parts := []string{"repo:" + p.Scope}
	switch kind {
	case "pr":
		parts = append(parts, "is:pr")
	case "issue":
		parts = append(parts, "is:issue")
	}
	switch p.View {
	case "", "open":
		parts = append(parts, "is:open")
	case "merged":
		parts = append(parts, "is:merged")
	case "closed":
		parts = append(parts, "is:closed", "is:unmerged")
	case "all":
	default:
		return "", fmt.Errorf("%w: view %q is not open, merged, closed or all", store.ErrInvalid, p.View)
	}
	if p.Author != "" {
		if !githubLogin.MatchString(p.Author) {
			return "", fmt.Errorf("%w: author is not a GitHub login", store.ErrInvalid)
		}
		a := strings.TrimSuffix(p.Author, "[bot]")
		if a != p.Author {
			a = "app/" + a
		}
		parts = append(parts, "author:"+a)
	}
	if t := plainWords(searchTerm(p.Query)); t != "" {
		parts = append(parts, t+" in:title,body")
	}
	parts = append(parts, "sort:updated-desc")
	return strings.Join(parts, " "), nil
}

// plainWords drops every word with a qualifier, a quote or a minus sign in
// front, so what is left only matches text.
func plainWords(q string) string {
	var out []string
	for w := range strings.FieldsSeq(q) {
		if strings.ContainsAny(w, ":\"") || strings.HasPrefix(w, "-") {
			w = strings.NewReplacer(":", " ", "\"", " ").Replace(strings.TrimLeft(w, "-"))
			out = append(out, strings.Fields(w)...)
			continue
		}
		out = append(out, w)
	}
	return strings.Join(out, " ")
}

func (s *Service) listGitHub(ctx context.Context, p rpc.BrowseListParams) (rpc.BrowseListResult, error) {
	if err := s.githubRepo(p.Scope); err != nil {
		return rpc.BrowseListResult{}, err
	}
	kind := p.Kind
	if kind == "" {
		kind = "pr"
	}
	if kind != "pr" && kind != "issue" && kind != "all" {
		return rpc.BrowseListResult{}, fmt.Errorf("%w: kind must be pr, issue or all", store.ErrInvalid)
	}
	q, err := githubSearch(p, kind)
	if err != nil {
		return rpc.BrowseListResult{}, err
	}
	first := bound(p.Limit)
	key := fmt.Sprintf("github:list:%s|%d|%s", q, first, p.Cursor)
	res, st := fetch(s, ctx, key, "github:"+p.Scope, listTTL, p.Refresh, func(ctx context.Context) (rpc.BrowseListResult, error) {
		cl, err := s.h.GitHub(p.Scope)
		if err != nil {
			return rpc.BrowseListResult{}, err
		}
		vars := map[string]any{"q": q, "first": first}
		if p.Cursor != "" {
			vars["after"] = p.Cursor
		}
		var r struct {
			Search struct {
				IssueCount int            `json:"issueCount"`
				PageInfo   linearPageInfo `json:"pageInfo"`
				Nodes      []ghNode       `json:"nodes"`
			} `json:"search"`
		}
		if err := cl.GraphQL(ctx, githubListQuery, vars, &r); err != nil {
			return rpc.BrowseListResult{}, err
		}
		out := rpc.BrowseListResult{Items: make([]rpc.BrowseItem, 0, len(r.Search.Nodes)), HasMore: r.Search.PageInfo.HasNextPage, Total: r.Search.IssueCount}
		for _, n := range r.Search.Nodes {
			if n.Number > 0 {
				out.Items = append(out.Items, n.item(p.Scope))
			}
		}
		if out.HasMore {
			out.Next = r.Search.PageInfo.EndCursor
		}
		return out, nil
	})
	if res.Items == nil {
		res.Items = []rpc.BrowseItem{}
	}
	res.Status = st
	return res, nil
}

const githubPRQuery = `query($o: String!, $n: String!, $num: Int!, $last: Int!, $before: String) {
  repository(owner: $o, name: $n) { pullRequest(number: $num) {
    number title url state isDraft merged body createdAt updatedAt
    author { login __typename } baseRefName headRefName additions deletions changedFiles mergeable reviewDecision
    labels(first: 20) { nodes { name } }
    closingIssuesReferences(first: 10) { nodes { number title url state repository { nameWithOwner } } }
    files(first: 100) { totalCount nodes { path additions deletions changeType } }
    commits(last: 1) { nodes { commit { statusCheckRollup { state contexts(first: 50) { totalCount nodes {
      __typename
      ... on CheckRun { name status conclusion detailsUrl startedAt completedAt }
      ... on StatusContext { context state targetUrl createdAt }
    } } } } } }
    timelineItems(last: $last, before: $before, itemTypes: [ISSUE_COMMENT, PULL_REQUEST_REVIEW, PULL_REQUEST_REVIEW_THREAD]) {
      pageInfo { hasPreviousPage startCursor }
      nodes { __typename
        ... on IssueComment { id author { login __typename } body createdAt lastEditedAt url }
        ... on PullRequestReview { id author { login __typename } state body createdAt url }
        ... on PullRequestReviewThread { id isResolved isOutdated path line
          comments(first: 20) { totalCount nodes { id author { login __typename } body createdAt url replyTo { id } } } }
      }
    }
  } }
}`

const githubIssueQuery = `query($o: String!, $n: String!, $num: Int!, $last: Int!, $before: String) {
  repository(owner: $o, name: $n) { issue(number: $num) {
    number title url state body createdAt updatedAt author { login __typename }
    labels(first: 20) { nodes { name } }
    timelineItems(last: $last, before: $before, itemTypes: [ISSUE_COMMENT]) {
      pageInfo { hasPreviousPage startCursor }
      nodes { __typename ... on IssueComment { id author { login __typename } body createdAt lastEditedAt url } }
    }
  } }
}`

type ghComment struct {
	ID        string     `json:"id"`
	Author    *ghActor   `json:"author"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	CreatedAt time.Time  `json:"createdAt"`
	EditedAt  *time.Time `json:"lastEditedAt"`
	URL       string     `json:"url"`
	ReplyTo   *struct {
		ID string `json:"id"`
	} `json:"replyTo"`
}

type ghTimeline struct {
	Type       string `json:"__typename"`
	IsResolved bool   `json:"isResolved"`
	IsOutdated bool   `json:"isOutdated"`
	Path       string `json:"path"`
	Line       int    `json:"line"`
	Comments   *struct {
		Total int         `json:"totalCount"`
		Nodes []ghComment `json:"nodes"`
	} `json:"comments"`
	ghComment
}

type ghDetail struct {
	ghNode
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"createdAt"`
	Base        string    `json:"baseRefName"`
	Head        string    `json:"headRefName"`
	Additions   int       `json:"additions"`
	Deletions   int       `json:"deletions"`
	ChangedFile int       `json:"changedFiles"`
	Mergeable   string    `json:"mergeable"`
	Closing     struct {
		Nodes []struct {
			Number     int    `json:"number"`
			Title      string `json:"title"`
			URL        string `json:"url"`
			State      string `json:"state"`
			Repository struct {
				Name string `json:"nameWithOwner"`
			} `json:"repository"`
		} `json:"nodes"`
	} `json:"closingIssuesReferences"`
	Files struct {
		Total int `json:"totalCount"`
		Nodes []struct {
			Path      string `json:"path"`
			Additions int    `json:"additions"`
			Deletions int    `json:"deletions"`
			Change    string `json:"changeType"`
		} `json:"nodes"`
	} `json:"files"`
	CommitsFull struct {
		Nodes []struct {
			Commit struct {
				Rollup *struct {
					State    string `json:"state"`
					Contexts struct {
						Total int `json:"totalCount"`
						Nodes []struct {
							Type        string    `json:"__typename"`
							Name        string    `json:"name"`
							Status      string    `json:"status"`
							Conclusion  string    `json:"conclusion"`
							DetailsURL  string    `json:"detailsUrl"`
							StartedAt   time.Time `json:"startedAt"`
							CompletedAt time.Time `json:"completedAt"`
							Context     string    `json:"context"`
							State       string    `json:"state"`
							TargetURL   string    `json:"targetUrl"`
							CreatedAt   time.Time `json:"createdAt"`
						} `json:"nodes"`
					} `json:"contexts"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
	Timeline struct {
		PageInfo struct {
			HasPrevious bool   `json:"hasPreviousPage"`
			StartCursor string `json:"startCursor"`
		} `json:"pageInfo"`
		Nodes []ghTimeline `json:"nodes"`
	} `json:"timelineItems"`
}

func (s *Service) showGitHub(ctx context.Context, p rpc.BrowseShowParams) (rpc.BrowseDetail, error) {
	m := githubRef.FindStringSubmatch(strings.TrimSpace(p.Ref))
	if m == nil {
		return rpc.BrowseDetail{}, fmt.Errorf("%w: %q is not owner/name#number", store.ErrInvalid, p.Ref)
	}
	repo := m[1]
	num, _ := strconv.Atoi(m[2])
	if err := s.githubRepo(repo); err != nil {
		return rpc.BrowseDetail{}, err
	}
	kind := p.Kind
	switch kind {
	case "", "pr", "issue":
	default:
		return rpc.BrowseDetail{}, fmt.Errorf("%w: kind must be pr or issue", store.ErrInvalid)
	}
	ck := fmt.Sprintf("github:show:%s#%d|%s|%s|%t", repo, num, kind, p.Cursor, p.Brief)
	d, st := fetch(s, ctx, ck, "github:"+repo, detailTTL, p.Refresh, func(ctx context.Context) (rpc.BrowseDetail, error) {
		cl, err := s.h.GitHub(repo)
		if err != nil {
			return rpc.BrowseDetail{}, err
		}
		owner, name, _ := strings.Cut(repo, "/")
		last := threadPage
		if p.Brief {
			last = 1
		}
		vars := map[string]any{"o": owner, "n": name, "num": num, "last": last}
		if p.Cursor != "" {
			vars["before"] = p.Cursor
		}
		try := func(query, field string) (*ghDetail, error) {
			var r struct {
				Repository map[string]*ghDetail `json:"repository"`
			}
			if err := cl.GraphQL(ctx, query, vars, &r); err != nil {
				if strings.Contains(err.Error(), "Could not resolve to") {
					return nil, errNotFound
				}
				return nil, err
			}
			if r.Repository == nil || r.Repository[field] == nil {
				return nil, errNotFound
			}
			return r.Repository[field], nil
		}
		var det *ghDetail
		switch kind {
		case "issue":
			det, err = try(githubIssueQuery, "issue")
		case "pr":
			det, err = try(githubPRQuery, "pullRequest")
		default:
			// A number is a PR or an issue: a PR first, which is what the
			// tab is about, then the issue.
			if det, err = try(githubPRQuery, "pullRequest"); err == errNotFound {
				det, err = try(githubIssueQuery, "issue")
			}
		}
		if err != nil {
			return rpc.BrowseDetail{}, err
		}
		return s.githubDetail(repo, det, p), nil
	})
	if st.State == rpc.BrowseError && d.Item.Ref == "" {
		d.Item = rpc.BrowseItem{Provider: rpc.BrowseGitHub, Scope: repo, Ref: repo + "#" + strconv.Itoa(num), ID: "#" + strconv.Itoa(num), Kind: "pr", Labels: []string{}}
	}
	d = fillDetail(d)
	d.Status = st
	return d, nil
}

func (s *Service) githubDetail(repo string, g *ghDetail, p rpc.BrowseShowParams) rpc.BrowseDetail {
	// The issue query has no commits, so the PR shape is told by them.
	isPR := g.Base != "" || g.Files.Total > 0 || len(g.CommitsFull.Nodes) > 0 || g.Head != ""
	n := g.ghNode
	if isPR {
		n.Type = "PullRequest"
	} else {
		n.Type = "Issue"
	}
	d := rpc.BrowseDetail{Item: n.item(repo), Threads: []rpc.BrowseThread{}}
	if len(g.CommitsFull.Nodes) > 0 && g.CommitsFull.Nodes[0].Commit.Rollup != nil {
		d.Item.Checks = ghChecks(g.CommitsFull.Nodes[0].Commit.Rollup.State)
	}
	if p.Cursor != "" {
		d.Threads = githubThreads(g.Timeline.Nodes)
		if g.Timeline.PageInfo.HasPrevious {
			d.Older = g.Timeline.PageInfo.StartCursor
		}
		return d
	}
	d.Body, d.BodyTruncated = clip(g.Body, maxBody)
	d.Props = rpc.BrowseProps{
		Author: d.Item.Owner, Base: g.Base, Head: g.Head, Additions: g.Additions, Deletions: g.Deletions,
		ChangedFiles: g.ChangedFile, Labels: d.Item.Labels, CreatedAt: g.CreatedAt, Mergeable: strings.ToLower(g.Mergeable),
	}
	d.Linked = s.githubLinked(repo, g)
	d.FilesTotal = g.Files.Total
	for _, f := range g.Files.Nodes {
		d.Files = append(d.Files, rpc.BrowseFile{Path: f.Path, Additions: f.Additions, Deletions: f.Deletions, Change: strings.ToLower(f.Change)})
	}
	if len(g.CommitsFull.Nodes) > 0 && g.CommitsFull.Nodes[0].Commit.Rollup != nil {
		ctxs := g.CommitsFull.Nodes[0].Commit.Rollup.Contexts
		d.ChecksTotal = ctxs.Total
		for _, c := range ctxs.Nodes {
			d.Checks = append(d.Checks, ghCheck(c.Type, c.Name, c.Status, c.Conclusion, c.DetailsURL, c.Context, c.State, c.TargetURL, c.StartedAt, c.CompletedAt, c.CreatedAt))
		}
	}
	if !p.Brief {
		d.Threads = githubThreads(g.Timeline.Nodes)
		if g.Timeline.PageInfo.HasPrevious {
			d.Older = g.Timeline.PageInfo.StartCursor
		}
	}
	return d
}

func ghCheck(typ, name, status, conclusion, details, ctx, state, target string, started, completed, created time.Time) rpc.BrowseCheck {
	c := rpc.BrowseCheck{Name: name, URL: details, StartedAt: started}
	if typ == "StatusContext" {
		c.Name, c.URL, c.StartedAt = ctx, target, created
		c.State = ghChecks(state)
		if c.State == "" {
			c.State = "skip"
		}
		return c
	}
	switch {
	case status != "COMPLETED":
		c.State = "run"
	case conclusion == "SUCCESS":
		c.State = "pass"
	case conclusion == "SKIPPED" || conclusion == "NEUTRAL":
		c.State = "skip"
	default:
		c.State = "fail"
	}
	if !started.IsZero() && completed.After(started) {
		c.Seconds = int(completed.Sub(started).Seconds())
	}
	return c
}

// githubLinked is what the PR says it closes, plus the Linear issues its
// title and branch name. A key counts only when its team is one Desk reads:
// an arbitrary word with a dash is not an association.
func (s *Service) githubLinked(repo string, g *ghDetail) []rpc.BrowseLinked {
	out := []rpc.BrowseLinked{}
	for _, c := range g.Closing.Nodes {
		r := c.Repository.Name
		if r == "" {
			r = repo
		}
		out = append(out, rpc.BrowseLinked{Kind: "issue", Ref: r + "#" + strconv.Itoa(c.Number), Title: c.Title, State: strings.ToLower(c.State), URL: c.URL, Relation: "closes"})
	}
	teams := s.Scope().Linear
	seen := map[string]bool{}
	for _, text := range []string{g.Title, g.Head} {
		for _, m := range keyInText.FindAllStringSubmatch(strings.ToUpper(text), -1) {
			key := m[1] + "-" + m[2]
			if seen[key] || !slices.ContainsFunc(teams, func(t rpc.BrowseLinearScope) bool { return t.Team == m[1] }) {
				continue
			}
			seen[key] = true
			out = append(out, rpc.BrowseLinked{Kind: "issue", Ref: key, Relation: "mentions"})
		}
	}
	return out
}

// githubThreads orders timeline items by latest activity, newest first. A
// review thread keeps the provider's own reply links; an issue comment and a
// review are each one message. Nothing is nested that the provider did not
// nest.
func githubThreads(nodes []ghTimeline) []rpc.BrowseThread {
	out := make([]rpc.BrowseThread, 0, len(nodes))
	conv := func(c ghComment) rpc.BrowseComment {
		bc := rpc.BrowseComment{ID: c.ID, At: c.CreatedAt, URL: c.URL}
		if c.Author != nil {
			bc.Author, bc.Bot = c.Author.Login, c.Author.Type == "Bot"
		}
		var cut bool
		bc.Body, cut = clip(c.Body, maxComment)
		bc.Truncated = cut
		if c.EditedAt != nil {
			bc.EditedAt = *c.EditedAt
		}
		switch c.State {
		case "APPROVED":
			bc.State = "approved"
		case "CHANGES_REQUESTED":
			bc.State = "changes"
		case "COMMENTED", "DISMISSED":
			bc.State = "commented"
		}
		return bc
	}
	for _, n := range nodes {
		switch n.Type {
		case "IssueComment":
			out = append(out, rpc.BrowseThread{ID: n.ID, Kind: "comment", URL: n.URL, Comments: []rpc.BrowseComment{conv(n.ghComment)}})
		case "PullRequestReview":
			if n.Body == "" && n.State == "COMMENTED" {
				continue // the thread entries carry the inline comments
			}
			out = append(out, rpc.BrowseThread{ID: n.ID, Kind: "review", URL: n.URL, Comments: []rpc.BrowseComment{conv(n.ghComment)}})
		case "PullRequestReviewThread":
			if n.Comments == nil || len(n.Comments.Nodes) == 0 {
				continue
			}
			th := rpc.BrowseThread{ID: n.ID, Kind: "diff", Path: n.Path, Line: n.Line, Resolved: n.IsResolved, Outdated: n.IsOutdated, Comments: []rpc.BrowseComment{}}
			byID := map[string]rpc.BrowseComment{}
			for _, c := range n.Comments.Nodes {
				bc := conv(c)
				if c.ReplyTo != nil {
					if to, ok := byID[c.ReplyTo.ID]; ok {
						bc.ReplyTo, bc.ReplyToAuthor = to.ID, to.Author
					}
				}
				byID[c.ID] = bc
				th.Comments = append(th.Comments, bc)
			}
			th.URL = th.Comments[0].URL
			th.Hidden = max(0, n.Comments.Total-len(n.Comments.Nodes))
			out = append(out, th)
		}
	}
	slices.SortStableFunc(out, func(a, b rpc.BrowseThread) int { return out2last(b).Compare(out2last(a)) })
	return out
}
