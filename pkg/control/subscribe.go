package control

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Subscriptions (#187): a ticket follows GitHub issue and PR threads. Its
// own meta.github issue or pr and PR links in its comments are followed
// implicitly; sessions add others with `jobshop subscribe`. The read
// position per thread is meta.subscribe.threads[].seen, so it survives
// restarts. New human comments go to one follower per thread: the
// session working it, else the open ticket as a comment, else the most
// recently closed one, which is reopened.

const (
	subscribePack         = "subscribe"
	defaultSubInterval    = time.Minute
	defaultSubResume      = 7 * 24 * time.Hour
	subCommentBodyMax     = 4000
	subscribeAuthorPrefix = "github:"
	linearAuthorPrefix    = "linear:"
)

// prLink finds PR URLs in comments, e.g. the pr-work "PR <url> head <sha>".
var prLink = regexp.MustCompile(`https://github\.com/([\w.-]+/[\w.-]+)/pull/(\d+)`)

// thread is one followed GitHub thread, or a Linear issue when Linear is
// set (#233), and its read position.
type thread struct {
	Repo   string    `json:"repo,omitempty"`
	Number int       `json:"number,omitempty"`
	PR     bool      `json:"pr,omitempty"`
	Linear string    `json:"linear,omitempty"`
	Seen   time.Time `json:"seen"`
}

func (t thread) key() string {
	if t.Linear != "" {
		return "linear:" + t.Linear
	}
	return fmt.Sprintf("%s#%d", t.Repo, t.Number)
}

// authorPrefix marks ticket comments quoted from the thread's tracker.
func (t thread) authorPrefix() string {
	if t.Linear != "" {
		return linearAuthorPrefix
	}
	return subscribeAuthorPrefix
}

// ghComment is a comment or review on a thread.
type ghComment struct {
	ID     string
	Author string
	Bot    bool
	Kind   string // comment, review, review comment
	State  string // review state
	Body   string
	URL    string
	At     time.Time
	// Parent is the comment replied to (Linear threads only).
	Parent string
}

// threadSource reads a thread's comments newer than since. Tests replace
// it; the controller uses gh.
type threadSource interface {
	Comments(ctx context.Context, th thread, since time.Time) ([]ghComment, error)
}

// subMeta is meta.subscribe.
type subMeta struct {
	Threads []thread `json:"threads,omitempty"`
}

func decodeSubMeta(meta json.RawMessage) (subMeta, map[string]json.RawMessage) {
	var m map[string]json.RawMessage
	var s subMeta
	if len(meta) == 0 || json.Unmarshal(meta, &m) != nil || m == nil {
		return s, map[string]json.RawMessage{}
	}
	if raw, ok := m[subscribePack]; ok {
		_ = json.Unmarshal(raw, &s)
	}
	return s, m
}

// implicitThreads are the threads a ticket follows without asking: its
// meta.github issue or pr, its meta.linear issue, and PRs linked in its
// comments.
func implicitThreads(t store.Ticket, comments []store.Comment) []thread {
	var out []thread
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
	if len(t.Meta) > 0 && json.Unmarshal(t.Meta, &m) == nil && linear.KeyPattern.MatchString(m.Linear.Issue) {
		out = append(out, thread{Linear: m.Linear.Issue})
	}
	if len(t.Meta) > 0 && json.Unmarshal(t.Meta, &m) == nil && m.GitHub.Repo != "" {
		if n := refNumber(m.GitHub.Issue); n > 0 {
			out = append(out, thread{Repo: m.GitHub.Repo, Number: n})
		}
		if n := refNumber(m.GitHub.PR); n > 0 {
			out = append(out, thread{Repo: m.GitHub.Repo, Number: n, PR: true})
		}
	}
	for _, c := range comments {
		if strings.HasPrefix(c.Author, subscribeAuthorPrefix) || strings.HasPrefix(c.Author, linearAuthorPrefix) {
			continue // quoted tracker text, not the ticket's own links
		}
		for _, sm := range prLink.FindAllStringSubmatch(c.Body, -1) {
			n, _ := strconv.Atoi(sm[2])
			out = append(out, thread{Repo: sm[1], Number: n, PR: true})
		}
	}
	return out
}

func refNumber(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(strings.TrimPrefix(x, "#"))
		return n
	}
	return 0
}

// follower is a ticket and the threads it follows.
type follower struct {
	t       store.Ticket
	threads []thread
	changed bool // threads gained an entry not yet saved
}

// pollSubscriptions runs one pass: register new threads, fetch new
// comments, deliver them.
func (c *Controller) pollSubscriptions(ctx context.Context, src threadSource) {
	cfg := c.config()
	if cfg.Subscriptions == nil {
		return
	}
	now := c.now()
	window := config.Duration(cfg.Subscriptions.ResumeWindow, defaultSubResume)
	all, err := c.store.List(ctx, store.Filter{Limit: math.MaxInt32})
	if err != nil {
		c.log.Warn("subscriptions: list", "err", err)
		return
	}
	var followers []*follower
	for _, t := range all {
		if t.Kind != store.KindTask && t.Kind != store.KindStep {
			continue
		}
		if t.Status == store.StatusClosed && now.Sub(t.ClosedAt) > window {
			continue
		}
		f, err := c.follow(ctx, t, now)
		if err != nil {
			c.log.Warn("subscriptions: follow", "ticket", t.ID, "err", err)
			continue
		}
		if f != nil {
			followers = append(followers, f)
		}
	}
	byThread := map[string][]*follower{}
	var order []string
	for _, f := range followers {
		for _, th := range f.threads {
			if _, ok := byThread[th.key()]; !ok {
				order = append(order, th.key())
			}
			byThread[th.key()] = append(byThread[th.key()], f)
		}
	}
	for _, k := range order {
		fs := byThread[k]
		th, seen := threadOf(fs, k)
		cs, err := src.Comments(ctx, th, seen)
		if err != nil {
			c.log.Warn("subscriptions: fetch", "thread", k, "err", err)
			continue
		}
		var human []ghComment
		last := seen
		for _, cm := range cs {
			if !cm.At.After(seen) {
				continue
			}
			if cm.At.After(last) {
				last = cm.At
			}
			if !c.ignored(cm) {
				human = append(human, cm)
			}
		}
		if len(human) > 0 {
			if err := c.deliverThread(ctx, th, human, fs, now, window); err != nil {
				c.log.Warn("subscriptions: deliver", "thread", k, "err", err)
				continue // read position stays; retried next pass
			}
		}
		for _, f := range fs {
			for i := range f.threads {
				if f.threads[i].key() == k && f.threads[i].Seen.Before(last) {
					f.threads[i].Seen = last
					f.changed = true
				}
			}
		}
	}
	for _, f := range followers {
		if f.changed {
			c.saveThreads(ctx, f)
		}
	}
}

// follow returns the ticket's threads, adding implicit ones with the read
// position at now (older comments are already on the thread when the
// work starts). It saves new entries right away so a restart keeps them.
func (c *Controller) follow(ctx context.Context, t store.Ticket, now time.Time) (*follower, error) {
	sm, _ := decodeSubMeta(t.Meta)
	var comments []store.Comment
	var err error
	if t.Status != store.StatusClosed || len(sm.Threads) > 0 {
		if comments, err = c.store.Comments(ctx, t.ID); err != nil {
			return nil, err
		}
	}
	f := &follower{t: t, threads: sm.Threads}
	for _, th := range implicitThreads(t, comments) {
		if slices.ContainsFunc(f.threads, func(x thread) bool { return x.key() == th.key() }) {
			continue
		}
		if t.Status == store.StatusClosed {
			continue // a closed ticket only keeps what it followed
		}
		th.Seen = now
		f.threads = append(f.threads, th)
		f.changed = true
	}
	if len(f.threads) == 0 {
		return nil, nil
	}
	if f.changed {
		c.saveThreads(ctx, f)
		f.changed = false
	}
	return f, nil
}

// threadOf returns the thread k and its read position, the newest seen
// among followers.
func threadOf(fs []*follower, k string) (thread, time.Time) {
	var th thread
	found := false
	for _, f := range fs {
		for _, x := range f.threads {
			if x.key() != k {
				continue
			}
			if !found || x.Seen.After(th.Seen) {
				found = true
				pr := th.PR || x.PR
				th = x
				th.PR = pr
			}
			th.PR = th.PR || x.PR
		}
	}
	return th, th.Seen
}

func (c *Controller) saveThreads(ctx context.Context, f *follower) {
	cur, err := c.store.Get(ctx, f.t.ID)
	if err != nil {
		return
	}
	sm, m := decodeSubMeta(cur.Meta)
	// Merge: keep entries added since (subscribe), advance read positions.
	for _, th := range f.threads {
		i := slices.IndexFunc(sm.Threads, func(x thread) bool { return x.key() == th.key() })
		if i < 0 {
			sm.Threads = append(sm.Threads, th)
		} else if th.Seen.After(sm.Threads[i].Seen) {
			sm.Threads[i].Seen = th.Seen
		}
	}
	raw, err := json.Marshal(sm)
	if err != nil {
		return
	}
	m[subscribePack] = raw
	mb, err := json.Marshal(m)
	if err != nil {
		return
	}
	t, err := c.store.Update(ctx, "controller", cur.ID, store.Update{Meta: mb, IfRevision: cur.Revision})
	if err != nil {
		c.log.Warn("subscriptions: save", "ticket", cur.ID, "err", err)
		return
	}
	f.t = t
	f.threads = sm.Threads
}

// ignored drops bots, CI and configured logins.
func (c *Controller) ignored(cm ghComment) bool {
	if cm.Bot || strings.HasSuffix(cm.Author, "[bot]") || cm.Author == "" {
		return true
	}
	s := c.config().Subscriptions
	return s != nil && slices.ContainsFunc(s.Ignore, func(l string) bool { return strings.EqualFold(l, cm.Author) })
}

// owner picks the follower that gets a thread's news: the one being
// worked, else an open one, else the most recently closed.
func owner(fs []*follower) *follower {
	rank := func(t store.Ticket) int {
		switch t.Status {
		case store.StatusInProgress:
			return 0
		case store.StatusOpen:
			return 1
		}
		return 2
	}
	best := slices.Clone(fs)
	sort.SliceStable(best, func(i, j int) bool {
		a, b := best[i].t, best[j].t
		if rank(a) != rank(b) {
			return rank(a) < rank(b)
		}
		if a.Status == store.StatusClosed {
			return a.ClosedAt.After(b.ClosedAt)
		}
		return a.CreatedAt.After(b.CreatedAt)
	})
	return best[0]
}

func threadText(th thread, cs []ghComment) string {
	var b strings.Builder
	switch {
	case th.Linear != "":
		fmt.Fprintf(&b, "New on Linear issue %s:\n", th.Linear)
	case th.PR:
		fmt.Fprintf(&b, "New on GitHub PR %s:\n", th.key())
	default:
		fmt.Fprintf(&b, "New on GitHub issue %s:\n", th.key())
	}
	for _, cm := range cs {
		body := cm.Body
		if len(body) > subCommentBodyMax {
			body = strings.ToValidUTF8(body[:subCommentBodyMax], "") + " [truncated]"
		}
		kind := cm.Kind
		if cm.State != "" {
			kind += " (" + strings.ToLower(cm.State) + ")"
		}
		fmt.Fprintf(&b, "\n[%s] %s %s %s\n%s\n", cm.At.UTC().Format(time.RFC3339), cm.Author, kind, cm.URL, body)
	}
	return b.String()
}

// deliverThread hands new human comments to the thread's owner.
func (c *Controller) deliverThread(ctx context.Context, th thread, cs []ghComment, fs []*follower, now time.Time, window time.Duration) error {
	o := owner(fs)
	t, err := c.store.Get(ctx, o.t.ID)
	if err != nil {
		return err
	}
	cs = c.intakeComments(ctx, t.Project, th, cs)
	if cs = c.linearAnswers(ctx, th, cs); len(cs) == 0 {
		return nil
	}
	text := threadText(th, cs)
	author := th.authorPrefix() + cs[len(cs)-1].Author
	if _, err := c.store.AddComment(ctx, author, t.ID, text); err != nil {
		return err
	}
	switch t.Status {
	case store.StatusInProgress:
		// The working session gets it now: interrupt, deliver, then the
		// held ticket is prompted again (runner.work).
		fields, err := json.Marshal(store.MessageFields{To: t.Holder, From: author, Interrupt: true})
		if err != nil {
			return err
		}
		_, err = c.store.Create(ctx, "controller", store.NewTicket{
			Project: t.Project, Kind: store.KindMessage,
			Title: fmt.Sprintf("%s has new comments for %s", threadName(th), t.ID),
			Body:  text, Fields: fields,
		})
		return err
	case store.StatusOpen:
		return nil // the comment is in the prompt when it is claimed
	}
	if now.Sub(t.ClosedAt) > window {
		return nil
	}
	if t.Kind == store.KindStep && t.ParentID != "" {
		// A finished run stays finished; its step keeps the comment and
		// the operator decides (run resume).
		if run, err := c.store.Get(ctx, t.ParentID); err != nil || run.Status == store.StatusClosed {
			return nil
		}
	}
	return c.resumeFollower(ctx, t, th)
}

// resumeFollower reopens a closed ticket for a human comment and, when
// the session that last held it is still open, assigns it there so the
// same conversation continues.
func (c *Controller) resumeFollower(ctx context.Context, t store.Ticket, th thread) error {
	last := c.lastHolder(ctx, t.ID)
	if _, err := c.store.Reopen(ctx, "controller", t.ID); err != nil {
		return err
	}
	c.log.Info("subscriptions: resumed ticket", "ticket", t.ID, "thread", th.key(), "session", last)
	if last == "" {
		return nil
	}
	if s, err := c.store.Get(ctx, last); err != nil || s.Kind != store.KindSession || s.Status == store.StatusClosed {
		return nil
	}
	if _, err := c.store.Update(ctx, "controller", t.ID, store.Update{Holder: &last}); err != nil {
		c.log.Warn("subscriptions: assign", "ticket", t.ID, "session", last, "err", err)
	}
	return nil
}

// lastHolder is the session that last claimed id.
func (c *Controller) lastHolder(ctx context.Context, id string) string {
	hist, err := c.store.History(ctx, id)
	if err != nil {
		return ""
	}
	for i := len(hist) - 1; i >= 0; i-- {
		if hist[i].Op == "claim" {
			return hist[i].Actor
		}
	}
	return ""
}

// subscribe adds a thread to a ticket's meta.subscribe (session or admin).
func (c *Controller) subscribe(ctx context.Context, sc scope, p rpc.SubscribeParams) (any, error) {
	if !config.RepoPattern.MatchString(p.Repo) || p.Number <= 0 {
		return nil, fmt.Errorf("%w: thread must be owner/name and a positive number", store.ErrInvalid)
	}
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	th := thread{Repo: p.Repo, Number: p.Number, PR: p.PR, Seen: c.now()}
	sm, _ := decodeSubMeta(t.Meta)
	if slices.ContainsFunc(sm.Threads, func(x thread) bool { return x.key() == th.key() }) {
		return t, nil
	}
	f := &follower{t: t, threads: append(sm.Threads, th)}
	c.saveThreads(ctx, f)
	if !slices.ContainsFunc(f.threads, func(x thread) bool { return x.key() == th.key() }) {
		return nil, fmt.Errorf("%w: %s changed, try again", store.ErrConflict, id)
	}
	return f.t, nil
}

// pumpSubscriptions polls on its own interval until the controller stops.
func (c *Controller) pumpSubscriptions() {
	src := c.subSource
	if src == nil {
		src = muxSource{gh: ghSource{}, linear: linearSource{c: c}}
	}
	for {
		every := defaultSubInterval
		if s := c.config().Subscriptions; s != nil {
			every = config.Duration(s.Interval, defaultSubInterval)
			c.pollSubscriptions(c.ctx, src)
		}
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(every):
		}
	}
}

func threadName(th thread) string {
	if th.Linear != "" {
		return "Linear " + th.Linear
	}
	return "GitHub " + th.key()
}

// muxSource sends Linear threads to linear and the rest to gh.
type muxSource struct{ gh, linear threadSource }

func (m muxSource) Comments(ctx context.Context, th thread, since time.Time) ([]ghComment, error) {
	if th.Linear != "" {
		return m.linear.Comments(ctx, th, since)
	}
	return m.gh.Comments(ctx, th, since)
}

// linearSource reads a Linear issue's comments as the identity covering
// its team (#200). Comments by that identity's own user are jobshop's
// sessions talking, so they count as bot comments: seen, not delivered.
type linearSource struct {
	c   *Controller
	url string // "" is linear.DefaultAPIURL; tests set a fake
}

func (s linearSource) Comments(ctx context.Context, th thread, since time.Time) ([]ghComment, error) {
	team, ok := linear.Team(th.Linear)
	if !ok {
		return nil, fmt.Errorf("linear: bad issue key %q", th.Linear)
	}
	cfg := s.c.config()
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), team)
	if !ok {
		return nil, fmt.Errorf("linear: no identity covers team %s", team)
	}
	tok, err := s.c.minter.Token(ctx, name, cfg.Identities[name], team)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	is, err := linear.New(s.url, tok.Token).GetIssue(ctx, th.Linear)
	if err != nil {
		return nil, err
	}
	var out []ghComment
	for _, cm := range is.Comments {
		if !cm.CreatedAt.After(since) {
			continue
		}
		url := cm.URL
		if url == "" {
			url = is.URL
		}
		out = append(out, ghComment{ID: cm.ID, Author: cm.Author, Bot: cm.Bot || cm.Self, Kind: "comment", Body: cm.Body, URL: url, At: cm.CreatedAt, Parent: cm.ParentID})
	}
	return out, nil
}

// ghSource reads threads with the gh CLI, which has the host's token.
type ghSource struct{}

func (ghSource) Comments(ctx context.Context, th thread, since time.Time) ([]ghComment, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	base := fmt.Sprintf("repos/%s", th.Repo)
	s := since.UTC().Format(time.RFC3339)
	const user = `author: .user.login, bot: (.user.type == "Bot")`
	queries := [][2]string{{
		fmt.Sprintf("%s/issues/%d/comments?since=%s&per_page=100", base, th.Number, s),
		`.[] | {id: ("c" + (.id|tostring)), kind: "comment", body, url: .html_url, at: .updated_at, ` + user + `}`,
	}}
	if th.PR {
		queries = append(queries,
			[2]string{
				fmt.Sprintf("%s/pulls/%d/reviews?per_page=100", base, th.Number),
				`.[] | select(.submitted_at != null) | {id: ("r" + (.id|tostring)), kind: "review", state, body, url: .html_url, at: .submitted_at, ` + user + `}`,
			},
			[2]string{
				fmt.Sprintf("%s/pulls/%d/comments?since=%s&per_page=100", base, th.Number, s),
				`.[] | {id: ("p" + (.id|tostring)), kind: "review comment", body, url: .html_url, at: .updated_at, ` + user + `}`,
			})
	}
	var out []ghComment
	for _, q := range queries {
		cmd := exec.CommandContext(ctx, "gh", "api", "--paginate", q[0], "--jq", q[1])
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		b, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("gh api %s: %w: %s", q[0], err, strings.TrimSpace(stderr.String()))
		}
		cs, err := parseGHLines(b)
		if err != nil {
			return nil, err
		}
		out = append(out, cs...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, nil
}

// parseGHLines reads the JSON lines the gh --jq filters print.
func parseGHLines(b []byte) ([]ghComment, error) {
	var out []ghComment
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var r struct {
			ID     string    `json:"id"`
			Kind   string    `json:"kind"`
			State  string    `json:"state"`
			Body   string    `json:"body"`
			URL    string    `json:"url"`
			At     time.Time `json:"at"`
			Author string    `json:"author"`
			Bot    bool      `json:"bot"`
		}
		if err := json.Unmarshal(line, &r); err != nil {
			return nil, fmt.Errorf("gh api output: %w", err)
		}
		if r.Kind == "review" && strings.TrimSpace(r.Body) == "" && r.State == "COMMENTED" {
			continue // the review's line comments arrive on their own
		}
		out = append(out, ghComment{ID: r.ID, Author: r.Author, Bot: r.Bot, Kind: r.Kind, State: r.State, Body: r.Body, URL: r.URL, At: r.At})
	}
	return out, sc.Err()
}
