package control

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Activity feed (#255): recent GitHub and Linear activity for every repo
// and team the host works, for Desk and `jobshop activity`. The poller
// keeps a bounded window in <root>/activity.json, a cache file and not a
// store table: the data is external and can be refetched, and store
// writes would add history rows and change events for every item.
// Links to runs and tickets are computed when listing, so they follow
// meta changes.

const (
	defaultActivityInterval = time.Minute
	defaultActivityWindow   = 7 * 24 * time.Hour
	defaultActivityMax      = 1000
	activityBodyMax         = 2000
	activityFile            = "activity.json"
	githubAPI               = "https://api.github.com"
)

// repoPoll is the per-repo poll state kept across restarts.
type repoPoll struct {
	EventsETag    string    `json:"events_etag,omitempty"`
	RunsETag      string    `json:"runs_etag,omitempty"`
	DefaultBranch string    `json:"default_branch,omitempty"`
	Next          time.Time `json:"next,omitzero"` // X-Poll-Interval
}

// activityState is the cache file.
type activityState struct {
	Repos map[string]*repoPoll `json:"repos"`
	Teams map[string]time.Time `json:"teams"` // last Linear poll time
	Items []rpc.ActivityItem   `json:"items"`
	seen  map[string]struct{}  // item ids
	errs  map[string]string    // last error logged per source
	mu    *sync.Mutex          // guards Items for activity.list
	path  string
}

// activitySource fetches one repo or team. Tests replace it.
type activitySource interface {
	GitHub(ctx context.Context, repo string, st *repoPoll) ([]rpc.ActivityItem, error)
	Linear(ctx context.Context, team string, since time.Time) ([]rpc.ActivityItem, error)
}

func (c *Controller) loadActivity() *activityState {
	st := &activityState{
		Repos: map[string]*repoPoll{}, Teams: map[string]time.Time{},
		seen: map[string]struct{}{}, errs: map[string]string{}, mu: &sync.Mutex{},
		path: filepath.Join(c.root.Dir, activityFile),
	}
	b, err := os.ReadFile(st.path)
	if err == nil {
		if err := json.Unmarshal(b, st); err != nil {
			c.log.Warn("activity: cache unreadable, starting empty", "file", st.path, "err", err)
		}
	}
	if st.Repos == nil {
		st.Repos = map[string]*repoPoll{}
	}
	if st.Teams == nil {
		st.Teams = map[string]time.Time{}
	}
	for _, it := range st.Items {
		st.seen[it.ID] = struct{}{}
	}
	return st
}

func (s *activityState) save() error {
	s.mu.Lock()
	b, err := json.Marshal(s)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// unseen returns the items not yet in the cache.
func (s *activityState) unseen(items []rpc.ActivityItem) []rpc.ActivityItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []rpc.ActivityItem
	for _, it := range items {
		if _, ok := s.seen[it.ID]; !ok {
			out = append(out, it)
		}
	}
	return out
}

// add merges new items and trims to the window and per-source bound.
// It returns how many were new.
func (s *activityState) add(items []rpc.ActivityItem, now time.Time, window time.Duration, maxPer int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, it := range items {
		if _, ok := s.seen[it.ID]; ok || now.Sub(it.At) > window {
			continue
		}
		s.seen[it.ID] = struct{}{}
		s.Items = append(s.Items, it)
		n++
	}
	sort.SliceStable(s.Items, func(i, j int) bool { return s.Items[i].At.After(s.Items[j].At) })
	per := map[string]int{}
	kept := s.Items[:0]
	for _, it := range s.Items {
		src := it.Source + ":" + it.Repo
		if now.Sub(it.At) > window || per[src] >= maxPer {
			delete(s.seen, it.ID)
			continue
		}
		per[src]++
		kept = append(kept, it)
	}
	s.Items = kept
	return n
}

// activityErr logs a source's error only when it changes, so a broken source
// does not log every poll; a later success clears it.
func (c *Controller) activityErr(s *activityState, src string, err error) {
	if err == nil {
		if _, ok := s.errs[src]; ok {
			c.log.Info("activity: source recovered", "source", src)
			delete(s.errs, src)
		}
		return
	}
	if s.errs[src] == err.Error() {
		return
	}
	s.errs[src] = err.Error()
	c.log.Warn("activity: poll", "source", src, "err", err)
}

// pollActivity runs one pass over every repo and team due.
func (c *Controller) pollActivity(ctx context.Context, s *activityState, src activitySource) int {
	cfg := c.config()
	a := cfg.Activity
	if a == nil {
		return 0
	}
	now := c.now()
	window := config.Duration(a.Window, defaultActivityWindow)
	maxPer := a.MaxItems
	if maxPer == 0 {
		maxPer = defaultActivityMax
	}
	added := 0
	for _, repo := range cfg.ActivityRepos() {
		rp := s.Repos[repo]
		if rp == nil {
			rp = &repoPoll{}
			s.Repos[repo] = rp
		}
		if now.Before(rp.Next) {
			continue
		}
		items, err := src.GitHub(ctx, repo, rp)
		c.activityErr(s, "github:"+repo, err)
		fresh := s.unseen(items)
		added += s.add(items, now, window, maxPer)
		if rs, ok := src.(revertSource); ok {
			c.checkReverts(ctx, rs, fresh)
		}
	}
	for _, team := range cfg.ActivityTeams() {
		since, ok := s.Teams[team]
		if !ok || now.Sub(since) > window {
			since = now.Add(-window)
		}
		items, err := src.Linear(ctx, team, since)
		c.activityErr(s, "linear:"+team, err)
		if err == nil {
			s.Teams[team] = now
		}
		added += s.add(items, now, window, maxPer)
	}
	// Drop poll state for repos and teams no longer configured.
	for r := range s.Repos {
		if !slices.Contains(cfg.ActivityRepos(), r) {
			delete(s.Repos, r)
		}
	}
	for t := range s.Teams {
		if !slices.Contains(cfg.ActivityTeams(), t) {
			delete(s.Teams, t)
		}
	}
	if err := s.save(); err != nil {
		c.activityErr(s, "cache", err)
	}
	if added > 0 && c.events != nil {
		c.events.publish([]rpc.Event{{Kind: rpc.EventKindActivity, Time: now}})
	}
	return added
}

// pumpActivity polls until the controller stops. The state is loaded
// once, on the first pass with activity configured.
func (c *Controller) pumpActivity() {
	src := c.activitySrc
	if src == nil {
		src = &liveActivity{c: c}
	}
	for {
		if c.config().Activity != nil {
			c.activityMu.Lock()
			if c.activity == nil {
				c.activity = c.loadActivity()
			}
			s := c.activity
			c.activityMu.Unlock()
			c.pollActivity(c.ctx, s, src)
		}
		every := defaultActivityInterval
		if a := c.config().Activity; a != nil {
			every = config.Duration(a.Interval, defaultActivityInterval)
		}
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(every):
		}
	}
}

// activityLinks maps "repo#n" and Linear keys to the runs and tickets
// that reference them. A step stands for its run.
func (c *Controller) activityLinks(ctx context.Context) (map[string][]string, error) {
	all, err := c.store.List(ctx, store.Filter{Limit: math.MaxInt32})
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	addLink := func(k, id string) {
		if !slices.Contains(out[k], id) {
			out[k] = append(out[k], id)
		}
	}
	for _, t := range all {
		if t.Kind != store.KindRun && t.Kind != store.KindStep && t.Kind != store.KindTask {
			continue
		}
		var m struct {
			GitHub struct {
				Repo  string `json:"repo"`
				Issue any    `json:"issue"`
				PR    any    `json:"pr"`
			} `json:"github"`
			Linear struct {
				Issue string `json:"issue"`
			} `json:"linear"`
		}
		if len(t.Meta) == 0 || json.Unmarshal(t.Meta, &m) != nil {
			continue
		}
		id := t.ID
		if t.Kind == store.KindStep && t.ParentID != "" {
			id = t.ParentID
		}
		if m.GitHub.Repo != "" {
			for _, v := range []any{m.GitHub.Issue, m.GitHub.PR} {
				if n := refNumber(v); n > 0 {
					addLink(fmt.Sprintf("%s#%d", m.GitHub.Repo, n), id)
				}
			}
		}
		if m.Linear.Issue != "" {
			addLink(m.Linear.Issue, id)
		}
	}
	return out, nil
}

func linkKey(it rpc.ActivityItem) string {
	if it.Source == rpc.ActivityLinear {
		return it.Ref
	}
	if it.Ref == "" {
		return ""
	}
	return it.Repo + "#" + it.Ref
}

func (c *Controller) activityList(ctx context.Context, sc scope, p rpc.ActivityListParams) (any, error) {
	if err := c.requireGrant(sc, "activity", "activity list"); err != nil {
		return nil, err
	}
	switch p.Bots {
	case "", "all", "humans", "bots":
	default:
		return nil, fmt.Errorf("%w: bots must be humans, bots or all", store.ErrInvalid)
	}
	cfg := c.config()
	res := rpc.ActivityListResult{Items: []rpc.ActivityItem{}, Sources: []string{}}
	for _, r := range cfg.ActivityRepos() {
		res.Sources = append(res.Sources, rpc.ActivityGitHub+":"+r)
	}
	for _, t := range cfg.ActivityTeams() {
		res.Sources = append(res.Sources, rpc.ActivityLinear+":"+t)
	}
	if cfg.Activity == nil {
		return res, nil
	}
	c.activityMu.Lock()
	s := c.activity
	c.activityMu.Unlock()
	if s == nil {
		return res, nil
	}
	links, err := c.activityLinks(ctx)
	if err != nil {
		return nil, err
	}
	now := c.now()
	limit := p.Limit
	if limit <= 0 {
		limit = 200
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, it := range s.Items {
		if p.Since > 0 && now.Sub(it.At) > p.Since {
			continue
		}
		if (p.Source != "" && it.Source != p.Source) || (p.Repo != "" && !strings.EqualFold(it.Repo, p.Repo)) ||
			(p.Kind != "" && it.Kind != p.Kind) || (p.Actor != "" && !strings.EqualFold(it.Actor, p.Actor)) ||
			(p.Bots == "humans" && it.Bot) || (p.Bots == "bots" && !it.Bot) {
			continue
		}
		it.Tickets = links[linkKey(it)]
		if p.Ticket != "" && !slices.Contains(it.Tickets, p.Ticket) {
			continue
		}
		res.Items = append(res.Items, it)
		if len(res.Items) >= limit {
			break
		}
	}
	return res, nil
}

// liveActivity reads GitHub over HTTP (ETag, X-Poll-Interval) and Linear
// with pkg/linear.
type liveActivity struct {
	c      *Controller
	http   *http.Client
	linURL string // "" is linear.DefaultAPIURL

	mu      sync.Mutex
	ghTok   string
	ghTokAt time.Time
}

// githubAuth returns the API base and token for repo: the identity
// covering it (#200), else the host's gh token (as subscriptions use).
func (l *liveActivity) githubAuth(ctx context.Context, repo string) (string, string, error) {
	cfg := l.c.config()
	if name, ok := cfg.IdentityFor(cfg.IdentityNames(), repo); ok {
		id := cfg.Identities[name]
		tok, err := l.c.minter.Token(ctx, name, id, repo)
		if err != nil {
			return "", "", err
		}
		base := id.APIURL
		if base == "" {
			base = githubAPI
		}
		return strings.TrimRight(base, "/"), tok.Token, nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.ghTok != "" && time.Since(l.ghTokAt) < 10*time.Minute {
		return githubAPI, l.ghTok, nil
	}
	tok := os.Getenv("GH_TOKEN")
	if tok == "" {
		out, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()
		if err != nil {
			return "", "", fmt.Errorf("no identity covers %s and gh auth token failed: %w", repo, err)
		}
		tok = strings.TrimSpace(string(out))
	}
	l.ghTok, l.ghTokAt = tok, time.Now()
	return githubAPI, tok, nil
}

// get fetches url with If-None-Match. A 304 returns nil body, no error.
func (l *liveActivity) get(ctx context.Context, u, tok, etag string) ([]byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+tok)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	hc := l.http
	if hc == nil {
		hc = &http.Client{Timeout: time.Minute}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotModified {
		return nil, resp.Header, nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("GET %s: %s: %s", u, resp.Status, bytes.TrimSpace(body[:min(len(body), 300)]))
	}
	return body, resp.Header, nil
}

func (l *liveActivity) GitHub(ctx context.Context, repo string, st *repoPoll) ([]rpc.ActivityItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	base, tok, err := l.githubAuth(ctx, repo)
	if err != nil {
		return nil, err
	}
	return l.fetchGitHub(ctx, base, tok, repo, st)
}

// fetchGitHub reads repo events and completed default-branch runs, each
// with its ETag: an unchanged repo is two 304s, which cost no rate limit.
func (l *liveActivity) fetchGitHub(ctx context.Context, base, tok, repo string, st *repoPoll) ([]rpc.ActivityItem, error) {
	var out []rpc.ActivityItem
	body, h, err := l.get(ctx, fmt.Sprintf("%s/repos/%s/events?per_page=100", base, repo), tok, st.EventsETag)
	if err != nil {
		return nil, err
	}
	if s, err := strconv.Atoi(h.Get("X-Poll-Interval")); err == nil && s > 0 {
		st.Next = l.c.now().Add(time.Duration(s) * time.Second)
	}
	if body != nil {
		items, err := parseRepoEvents(repo, body)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		st.EventsETag = h.Get("ETag")
	}
	if st.DefaultBranch == "" {
		b, _, err := l.get(ctx, fmt.Sprintf("%s/repos/%s", base, repo), tok, "")
		if err != nil {
			return out, err
		}
		var r struct {
			DefaultBranch string `json:"default_branch"`
		}
		if err := json.Unmarshal(b, &r); err != nil {
			return out, err
		}
		st.DefaultBranch = r.DefaultBranch
	}
	u := fmt.Sprintf("%s/repos/%s/actions/runs?status=completed&per_page=30&branch=%s", base, repo, url.QueryEscape(st.DefaultBranch))
	body, h, err = l.get(ctx, u, tok, st.RunsETag)
	if err != nil {
		return out, err
	}
	if body != nil {
		items, err := parseWorkflowRuns(repo, body)
		if err != nil {
			return out, err
		}
		out = append(out, items...)
		st.RunsETag = h.Get("ETag")
	}
	return out, nil
}

// PullRequest reads one PR for revert detection (#272).
func (l *liveActivity) PullRequest(ctx context.Context, repo string, n int) (ghPR, error) {
	base, tok, err := l.githubAuth(ctx, repo)
	if err != nil {
		return ghPR{}, err
	}
	b, _, err := l.get(ctx, fmt.Sprintf("%s/repos/%s/pulls/%d", base, repo, n), tok, "")
	if err != nil {
		return ghPR{}, err
	}
	var pr ghPR
	return pr, json.Unmarshal(b, &pr)
}

// CommitPulls lists the PRs a commit is associated with.
func (l *liveActivity) CommitPulls(ctx context.Context, repo, sha string) ([]int, error) {
	base, tok, err := l.githubAuth(ctx, repo)
	if err != nil {
		return nil, err
	}
	b, _, err := l.get(ctx, fmt.Sprintf("%s/repos/%s/commits/%s/pulls", base, repo, url.PathEscape(sha)), tok, "")
	if err != nil {
		return nil, err
	}
	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(b, &prs); err != nil {
		return nil, err
	}
	out := make([]int, len(prs))
	for i, p := range prs {
		out[i] = p.Number
	}
	return out, nil
}

func (l *liveActivity) Linear(ctx context.Context, team string, since time.Time) ([]rpc.ActivityItem, error) {
	cfg := l.c.config()
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), team)
	if !ok {
		return nil, fmt.Errorf("no identity covers team %s", team)
	}
	tok, err := l.c.minter.Token(ctx, name, cfg.Identities[name], team)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	chs, err := linear.New(l.linURL, tok.Token).Changes(ctx, team, since)
	if err != nil {
		return nil, err
	}
	out := make([]rpc.ActivityItem, 0, len(chs))
	for _, ch := range chs {
		it := rpc.ActivityItem{
			ID: "linear:" + ch.ItemID, Source: rpc.ActivityLinear, Repo: team, Kind: ch.Kind,
			Actor: ch.Actor, Bot: ch.Bot || ch.Self, At: ch.At, URL: ch.URL, Ref: ch.Issue,
		}
		switch ch.Kind {
		case "comment":
			it.Title, it.Body = ch.Issue+": "+ch.Title, truncate(ch.Text)
		case "issue":
			it.Title = ch.Issue + " created: " + ch.Title
		default:
			it.Title = fmt.Sprintf("%s %s %s: %s", ch.Issue, ch.Kind, ch.Text, ch.Title)
		}
		out = append(out, it)
	}
	return out, nil
}

func truncate(s string) string {
	if len(s) <= activityBodyMax {
		return s
	}
	return strings.ToValidUTF8(s[:activityBodyMax], "") + " [truncated]"
}

func ghBot(login string) bool { return strings.HasSuffix(login, "[bot]") }

// parseRepoEvents maps GET /repos/{r}/events to items. Unknown event
// types are skipped.
func parseRepoEvents(repo string, body []byte) ([]rpc.ActivityItem, error) {
	var evs []struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Actor struct {
			Login string `json:"login"`
		} `json:"actor"`
		CreatedAt time.Time       `json:"created_at"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(body, &evs); err != nil {
		return nil, fmt.Errorf("events: %w", err)
	}
	web := "https://github.com/" + repo
	var out []rpc.ActivityItem
	for _, e := range evs {
		var p struct {
			Action  string `json:"action"`
			Number  int    `json:"number"`
			Ref     string `json:"ref"`
			RefType string `json:"ref_type"`
			Head    string `json:"head"`
			Before  string `json:"before"`
			Commits []struct {
				Message string `json:"message"`
			} `json:"commits"`
			PullRequest *struct {
				Number  int    `json:"number"`
				Title   string `json:"title"`
				HTMLURL string `json:"html_url"`
				Merged  bool   `json:"merged"`
			} `json:"pull_request"`
			Issue *struct {
				Number      int    `json:"number"`
				Title       string `json:"title"`
				HTMLURL     string `json:"html_url"`
				PullRequest any    `json:"pull_request"`
			} `json:"issue"`
			Comment *struct {
				Body    string `json:"body"`
				HTMLURL string `json:"html_url"`
			} `json:"comment"`
			Review *struct {
				State   string `json:"state"`
				Body    string `json:"body"`
				HTMLURL string `json:"html_url"`
			} `json:"review"`
			Release *struct {
				TagName string `json:"tag_name"`
				Name    string `json:"name"`
				HTMLURL string `json:"html_url"`
			} `json:"release"`
		}
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			continue
		}
		it := rpc.ActivityItem{
			ID: "gh:" + repo + ":" + e.ID, Source: rpc.ActivityGitHub, Repo: repo,
			Actor: e.Actor.Login, Bot: ghBot(e.Actor.Login), At: e.CreatedAt,
		}
		prRef := func() {
			n := p.Number
			if p.PullRequest != nil && p.PullRequest.Number > 0 {
				n = p.PullRequest.Number
			}
			it.Ref, it.PR, it.URL = strconv.Itoa(n), true, fmt.Sprintf("%s/pull/%d", web, n)
			if p.PullRequest != nil && p.PullRequest.HTMLURL != "" {
				it.URL = p.PullRequest.HTMLURL
			}
		}
		prTitle := func() string {
			if p.PullRequest != nil && p.PullRequest.Title != "" {
				return ": " + p.PullRequest.Title
			}
			return ""
		}
		switch e.Type {
		case "PushEvent":
			branch := strings.TrimPrefix(p.Ref, "refs/heads/")
			it.Kind, it.Title = "push", "pushed to "+branch
			if len(p.Commits) > 0 {
				it.Title = fmt.Sprintf("pushed %d commit(s) to %s", len(p.Commits), branch)
				msgs := make([]string, 0, len(p.Commits))
				for _, cm := range p.Commits {
					msgs = append(msgs, cm.Message)
				}
				it.Body = truncate(strings.Join(msgs, "\n"))
			}
			if p.Head != "" {
				it.URL = web + "/commit/" + p.Head
				if p.Before != "" && strings.Trim(p.Before, "0") != "" {
					it.URL = fmt.Sprintf("%s/compare/%.12s...%.12s", web, p.Before, p.Head)
				}
			}
		case "PullRequestEvent":
			prRef()
			action := p.Action
			if action == "merged" || (action == "closed" && p.PullRequest != nil && p.PullRequest.Merged) {
				it.Kind, action = "merge", "merged"
			} else {
				it.Kind = "pr"
			}
			it.Title = fmt.Sprintf("PR #%s %s%s", it.Ref, action, prTitle())
		case "PullRequestReviewEvent":
			prRef()
			it.Kind = "review"
			state := ""
			if p.Review != nil {
				state = strings.ToLower(strings.ReplaceAll(p.Review.State, "_", " "))
				it.Body = truncate(p.Review.Body)
				if p.Review.HTMLURL != "" {
					it.URL = p.Review.HTMLURL
				}
			}
			it.Title = strings.TrimSpace(fmt.Sprintf("review %s on PR #%s%s", state, it.Ref, prTitle()))
		case "PullRequestReviewCommentEvent":
			prRef()
			it.Kind, it.Title = "review_comment", fmt.Sprintf("review comment on PR #%s%s", it.Ref, prTitle())
			if p.Comment != nil {
				it.Body, it.URL = truncate(p.Comment.Body), p.Comment.HTMLURL
			}
		case "IssueCommentEvent", "IssuesEvent":
			if p.Issue == nil {
				continue
			}
			it.Ref, it.PR, it.URL = strconv.Itoa(p.Issue.Number), p.Issue.PullRequest != nil, p.Issue.HTMLURL
			what := "issue"
			if it.PR {
				what = "PR"
			}
			if e.Type == "IssuesEvent" {
				it.Kind, it.Title = "issue", fmt.Sprintf("%s #%d %s: %s", what, p.Issue.Number, p.Action, p.Issue.Title)
			} else {
				if p.Action != "created" {
					continue // edits and deletes are noise
				}
				it.Kind, it.Title = "comment", fmt.Sprintf("comment on %s #%d: %s", what, p.Issue.Number, p.Issue.Title)
				if p.Comment != nil {
					it.Body, it.URL = truncate(p.Comment.Body), p.Comment.HTMLURL
				}
			}
		case "CreateEvent", "DeleteEvent":
			if p.RefType != "branch" && p.RefType != "tag" {
				continue
			}
			verb := "created"
			if e.Type == "DeleteEvent" {
				verb = "deleted"
			}
			it.Kind, it.Title = "branch", fmt.Sprintf("%s %s %s", verb, p.RefType, p.Ref)
			if verb == "created" {
				it.URL = web + "/tree/" + p.Ref
			}
		case "ReleaseEvent":
			if p.Release == nil {
				continue
			}
			it.Kind, it.Title, it.URL = "release", fmt.Sprintf("release %s %s", p.Release.TagName, p.Action), p.Release.HTMLURL
		default:
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

// parseWorkflowRuns maps completed actions runs to ci items.
func parseWorkflowRuns(repo string, body []byte) ([]rpc.ActivityItem, error) {
	var r struct {
		WorkflowRuns []struct {
			ID           int64     `json:"id"`
			RunAttempt   int       `json:"run_attempt"`
			Name         string    `json:"name"`
			DisplayTitle string    `json:"display_title"`
			Conclusion   string    `json:"conclusion"`
			HeadBranch   string    `json:"head_branch"`
			HTMLURL      string    `json:"html_url"`
			UpdatedAt    time.Time `json:"updated_at"`
			Actor        struct {
				Login string `json:"login"`
			} `json:"actor"`
		} `json:"workflow_runs"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("actions runs: %w", err)
	}
	out := make([]rpc.ActivityItem, 0, len(r.WorkflowRuns))
	for _, w := range r.WorkflowRuns {
		if w.Conclusion == "" {
			continue
		}
		out = append(out, rpc.ActivityItem{
			ID:     fmt.Sprintf("gh:%s:run:%d:%d", repo, w.ID, w.RunAttempt),
			Source: rpc.ActivityGitHub, Repo: repo, Kind: "ci",
			Actor: w.Actor.Login, Bot: ghBot(w.Actor.Login), At: w.UpdatedAt,
			Title: fmt.Sprintf("%s %s on %s: %s", w.Name, w.Conclusion, w.HeadBranch, w.DisplayTitle),
			URL:   w.HTMLURL,
		})
	}
	return out, nil
}
