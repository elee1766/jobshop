package threads

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/linearsvc"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/store"
)

// Host is what the engine needs from the controller.
type Host interface {
	// Config is the active config.
	Config() *config.Config
	// Now is the controller clock.
	Now() time.Time
	// Log is the controller's logger.
	Log() *slog.Logger
	// Store is read for tickets and their comments; the engine writes
	// only meta.subscribe, the read positions (Save).
	Store() store.Store
	// LinearDo runs fn as Linear identity name through the Linear
	// service (outage backoff, request bounds).
	LinearDo(ctx context.Context, name string, fn func(context.Context, *linear.Client) error) error
	// Login is the login of GitHub identity name.
	Login(ctx context.Context, name string, id config.Identity) (string, error)
}

// Engine plans thread polls and filters comments. The zero value is not
// usable; use New.
type Engine struct {
	h Host

	// mu keeps a webhook pass and the pump from reading a thread at once
	// (#300).
	mu sync.Mutex
	// selfLoginErr is the last self-login error logged per identity, so a
	// refused lookup warns once, not on every poll (JOB-108).
	selfLoginErr sync.Map
}

// New returns an engine on h.
func New(h Host) *Engine { return &Engine{h: h} }

// LinearSource is the Linear comment source on this engine's host.
func (e *Engine) LinearSource() LinearSource { return LinearSource{h: e.h} }

// DefaultSource reads GitHub with gh and Linear through the host.
func (e *Engine) DefaultSource() Source {
	return Mux{GH: GHSource{}, Linear: e.LinearSource()}
}

// Put records one new comment of th for delivery (the inbox, #442).
type Put func(ctx context.Context, th Thread, cm Comment) error

// Poll is a pass over every followed thread, or only the one keyed only
// when it is set. It hands each new comment to put; delivery is not the
// engine's. It reports whether put stored anything. A thread's read
// position advances only when all its new comments were put.
func (e *Engine) Poll(ctx context.Context, src Source, only string, put Put) (wrote bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	cfg := e.h.Config()
	if cfg.Subscriptions == nil {
		return false
	}
	log := e.h.Log()
	window := config.Duration(cfg.Subscriptions.ResumeWindow, DefaultResume)
	followers, byThread, order, err := e.Followers(ctx, window, e.h.Now())
	if err != nil {
		log.WarnContext(ctx, "subscriptions: list", "err", err)
		return false
	}
	for _, k := range order {
		if only != "" && k != only {
			continue
		}
		fs := byThread[k]
		th, seen := Of(fs, k)
		cs, err := src.Comments(ctx, th, seen)
		if err != nil {
			if !linearsvc.Down(err) { // logged once per outage
				log.WarnContext(ctx, "subscriptions: fetch", "thread", k, "err", err)
			}
			continue
		}
		last := seen
		self := ""
		if th.Linear == "" {
			self = e.Self(ctx, th.Repo)
		}
		ok := true
		for _, cm := range cs {
			if !cm.At.After(seen) {
				continue
			}
			if self != "" && strings.EqualFold(cm.Author, self) {
				cm.Self = true
			}
			if err := put(ctx, th, cm); err != nil {
				log.WarnContext(ctx, "inbox: poller write", "thread", k, "err", err)
				ok = false
				break // read position stays; retried next pass
			}
			wrote = true
			if cm.At.After(last) {
				last = cm.At
			}
		}
		if !ok {
			continue
		}
		for _, f := range fs {
			for i := range f.Threads {
				if f.Threads[i].Key() == k && f.Threads[i].Seen.Before(last) {
					f.Threads[i].Seen = last
					f.Changed = true
				}
			}
		}
	}
	for _, f := range followers {
		if f.Changed {
			e.Save(ctx, f)
		}
	}
	return wrote
}

// Followers lists the tickets following threads, grouped by thread key
// in first-seen order. A ticket closed longer than window ago is left out.
func (e *Engine) Followers(ctx context.Context, window time.Duration, now time.Time) ([]*Follower, map[string][]*Follower, []string, error) {
	all, err := e.h.Store().List(ctx, store.Filter{Limit: math.MaxInt32})
	if err != nil {
		return nil, nil, nil, err
	}
	var followers []*Follower
	for _, t := range all {
		if t.Kind != domain.KindTask && t.Kind != domain.KindStep {
			continue
		}
		if t.Status == domain.StatusClosed && now.Sub(t.ClosedAt) > window {
			continue
		}
		f, err := e.follow(ctx, t, now)
		if err != nil {
			e.h.Log().WarnContext(ctx, "subscriptions: follow", "ticket", t.ID, "err", err)
			continue
		}
		if f != nil {
			followers = append(followers, f)
		}
	}
	byThread := map[string][]*Follower{}
	var order []string
	for _, f := range followers {
		for _, th := range f.Threads {
			if _, ok := byThread[th.Key()]; !ok {
				order = append(order, th.Key())
			}
			byThread[th.Key()] = append(byThread[th.Key()], f)
		}
	}
	return followers, byThread, order, nil
}

// follow returns the ticket's threads, adding implicit ones with the read
// position at now (older comments are already on the thread when the
// work starts). It saves new entries right away so a restart keeps them.
func (e *Engine) follow(ctx context.Context, t domain.Ticket, now time.Time) (*Follower, error) {
	sm, _ := DecodeMeta(t.Meta)
	var comments []domain.Comment
	var err error
	if t.Status != domain.StatusClosed || len(sm.Threads) > 0 {
		if comments, err = store.Comments(ctx, e.h.Store(), t.ID); err != nil {
			return nil, err
		}
	}
	f := &Follower{T: t, Threads: sm.Threads}
	for _, th := range Implicit(t, comments) {
		if slices.ContainsFunc(f.Threads, func(x Thread) bool { return x.Key() == th.Key() }) {
			continue
		}
		if t.Status == domain.StatusClosed {
			continue // a closed ticket only keeps what it followed
		}
		th.Seen = now
		f.Threads = append(f.Threads, th)
		f.Changed = true
	}
	if len(f.Threads) == 0 {
		return nil, nil
	}
	if f.Changed {
		e.Save(ctx, f)
		f.Changed = false
	}
	return f, nil
}

// Save writes f's threads to the ticket's meta.subscribe, merging with
// what is there now: entries added since are kept, read positions only
// advance. On success f has the stored ticket and threads.
func (e *Engine) Save(ctx context.Context, f *Follower) {
	st := e.h.Store()
	cur, err := st.Get(ctx, f.T.ID)
	if err != nil {
		return
	}
	sm, m := DecodeMeta(cur.Meta)
	for _, th := range f.Threads {
		i := slices.IndexFunc(sm.Threads, func(x Thread) bool { return x.Key() == th.Key() })
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
	m[MetaKey] = raw
	mb, err := json.Marshal(m)
	if err != nil {
		return
	}
	t, err := st.Update(ctx, "controller", cur.ID, store.Update{Meta: mb, IfRevision: cur.Revision})
	if err != nil {
		e.h.Log().WarnContext(ctx, "subscriptions: save", "ticket", cur.ID, "err", err)
		return
	}
	f.T = t
	f.Threads = sm.Threads
}

// Self is the login of the GitHub identity covering repo: a
// github-token's user, or a github-app's "<slug>[bot]" (#402). Without
// it the host's own posts would come back as news (#340), and listing
// the app in subscriptions.bots would wake sessions on their own
// comments. "" for none; a failed lookup is logged and filters nothing.
func (e *Engine) Self(ctx context.Context, repo string) string {
	cfg := e.h.Config()
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), repo)
	if !ok || !config.IsGitHub(cfg.Identities[name].Kind) {
		return ""
	}
	// An App's comments are [bot] and dropped anyway unless bots is set.
	if cfg.Identities[name].Kind == config.IdentityGitHubApp && (cfg.Subscriptions == nil || len(cfg.Subscriptions.Bots) == 0) {
		return ""
	}
	l, err := e.h.Login(ctx, name, cfg.Identities[name])
	if err != nil {
		// Polled per repo every minute: log a changed error, not each call.
		if prev, _ := e.selfLoginErr.Swap(name, err.Error()); prev != err.Error() {
			e.h.Log().WarnContext(ctx, "identity: self login", "identity", name, "err", err)
		}
		return ""
	}
	e.selfLoginErr.Delete(name)
	return l
}

// Ignored drops the host's own comments, bots not on subscriptions.bots,
// CI, configured logins and, when authors is set, every human not on it.
// Linear comments match entries by user id, displayName or email, and by
// full name only as a logged fallback (#595).
func (e *Engine) Ignored(cm Comment) bool {
	if cm.Self || cm.Author == "" {
		return true
	}
	s := e.h.Config().Subscriptions
	is := func(l string) bool { return strings.EqualFold(l, cm.Author) }
	if cm.Linear {
		is = func(l string) bool { return e.LinearMatch(l, cm) }
	}
	if s != nil && slices.ContainsFunc(s.Ignore, is) {
		return true
	}
	if cm.Bot || strings.HasSuffix(cm.Author, "[bot]") {
		if cm.Linear {
			return s == nil || !slices.ContainsFunc(s.Bots, is)
		}
		return s == nil || !slices.Contains(s.Bots, cm.Author)
	}
	if s == nil {
		return false
	}
	return len(s.Authors) > 0 && !slices.ContainsFunc(s.Authors, is)
}

// LinearMatch reports whether a subscriptions entry names cm's author:
// the user or bot actor id, the displayName or the email. The full name
// is neither unique nor stable, so a match on it alone logs a WARN
// naming the entry. A bot actor has only a name, so its name matches
// without the warning.
func (e *Engine) LinearMatch(entry string, cm Comment) bool {
	if entry == "" {
		return false
	}
	for _, v := range []string{cm.AuthorID, cm.Handle, cm.Email} {
		if v != "" && strings.EqualFold(entry, v) {
			return true
		}
	}
	name := cm.Name
	if name == "" {
		name = cm.Author
	}
	if !strings.EqualFold(entry, name) {
		return false
	}
	if !cm.Bot {
		e.h.Log().Warn("subscriptions: Linear entry matched by full name; use the user id, displayName or email", "entry", entry, "user", cm.AuthorID)
	}
	return true
}
