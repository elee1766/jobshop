package control

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/threads"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Subscriptions (#187): a ticket follows GitHub issue and PR threads. Its
// own meta.github issue or pr and PR links in its comments are followed
// implicitly; sessions add others with `jobshop subscribe`. The read
// position per thread is meta.subscribe.threads[].seen, so it survives
// restarts. New human comments go to one follower per thread: the
// session working it, else the open ticket as a comment, else the most
// recently closed one, which is reopened.
//
// Discovery, the comment sources, the poll plan, the reply filters and
// the message text live in pkg/control/threads (JOB-156). This file is
// the controller's side: the host it hands in and delivery, which sends
// messages, resumes sessions and writes meta.subscribe.

// threadEngine is the thread engine, made on first use so controllers
// built without Run get one too.
func (c *Controller) threadEngine() *threads.Engine {
	c.threadsOnce.Do(func() { c.threadsEng = threads.New(threadsHost{c}) })
	return c.threadsEng
}

// threadsHost is the controller as threads.Host.
type threadsHost struct{ c *Controller }

func (h threadsHost) Config() *config.Config { return h.c.config() }
func (h threadsHost) Now() time.Time         { return h.c.now() }
func (h threadsHost) Log() *slog.Logger      { return h.c.log }
func (h threadsHost) Store() store.Store     { return h.c.store }

func (h threadsHost) LinearDo(ctx context.Context, name string, fn func(context.Context, *linear.Client) error) error {
	return h.c.linearSvc().Do(ctx, name, fn)
}

func (h threadsHost) Login(ctx context.Context, name string, id config.Identity) (string, error) {
	return h.c.minter.Login(ctx, name, id)
}

// threadSrc is the comment source: the test fake, else gh and Linear.
func (c *Controller) threadSrc() threads.Source {
	if c.subSource != nil {
		return c.subSource
	}
	return c.threadEngine().DefaultSource()
}

// githubSelf is the login of the GitHub identity covering repo (#402).
func (c *Controller) githubSelf(ctx context.Context, repo string) string {
	return c.threadEngine().Self(ctx, repo)
}

// pollSubscriptions runs one pass: register new threads, fetch new
// comments, deliver them.
func (c *Controller) pollSubscriptions(ctx context.Context, src threads.Source) {
	c.pollThreads(ctx, src, "")
}

// pollSubscriptionThread runs one pass over the single thread th, for a
// webhook that says it changed (#300). Off when subscriptions are.
func (c *Controller) pollSubscriptionThread(ctx context.Context, th threads.Thread) {
	c.pollThreads(ctx, c.threadSrc(), th.Key())
}

// pollThreads is a pass over every followed thread, or only the one keyed
// only when it is set. It writes each new comment to the inbox (#442);
// delivery is the dispatcher's, from admitted rows only, so admission
// hooks see and may rewrite or refuse every comment first.
func (c *Controller) pollThreads(ctx context.Context, src threads.Source, only string) {
	if c.threadEngine().Poll(ctx, src, only, c.inboxComment) {
		// Admit and dispatch now rather than on the pump's next wake.
		c.ingestPass(ctx)
	}
}

// deliverThread hands new human comments to the thread's owner.
func (c *Controller) deliverThread(ctx context.Context, th threads.Thread, cs []threads.Comment, fs []*threads.Follower, now time.Time, window time.Duration) error {
	o := threads.Owner(fs)
	t, err := c.store.Get(ctx, o.T.ID)
	if err != nil {
		return err
	}
	cs = c.linearAnswers(ctx, th, cs)
	if len(cs) == 0 {
		return nil
	}
	text := threads.Text(th, cs)
	author := threads.DeliveryAuthor(th, cs)
	if _, err := store.AddComment(ctx, c.store, author, t.ID, text); err != nil {
		return err
	}
	switch t.Status {
	case domain.StatusInProgress:
		// The working session gets it now: interrupt, deliver, then the
		// held ticket is prompted again (runner.work).
		fields, err := json.Marshal(domain.MessageFields{To: t.Holder, From: author, Interrupt: true})
		if err != nil {
			return err
		}
		_, err = store.Create(ctx, c.store, "controller", store.NewTicket{
			Project: t.Project, Kind: domain.KindMessage,
			Title: fmt.Sprintf("%s has new comments for %s", th.Name(), t.ID),
			Body:  text, Fields: fields,
		})

		if err == nil {
			c.wakeRunner(t.Holder)
		}
		return err
	case domain.StatusOpen:
		return nil // the comment is in the prompt when it is claimed
	}
	if now.Sub(t.ClosedAt) > window {
		return nil
	}
	if t.Kind == domain.KindStep && t.ParentID != "" {
		// A finished run stays finished; its step keeps the comment and
		// the operator decides (run resume).
		if run, err := c.store.Get(ctx, t.ParentID); err != nil || run.Status == domain.StatusClosed {
			return nil
		}
	}
	return c.resumeFollower(ctx, t, th)
}

// resumeFollower reopens a closed ticket for a human comment and, when
// the session that last held it is still open, assigns it there so the
// same conversation continues.
func (c *Controller) resumeFollower(ctx context.Context, t domain.Ticket, th threads.Thread) error {
	last := c.lastHolder(ctx, t.ID)
	if _, err := store.Reopen(ctx, c.store, "controller", t.ID); err != nil {
		return err
	}
	c.log.InfoContext(ctx, "subscriptions: resumed ticket", "ticket", t.ID, "thread", th.Key(), "session", last)
	if last == "" {
		return nil
	}
	if s, err := c.store.Get(ctx, last); err != nil || s.Kind != domain.KindSession || s.Status == domain.StatusClosed {
		return nil
	}
	if _, err := c.store.Update(ctx, "controller", t.ID, store.Update{Holder: &last}); err != nil {
		c.log.WarnContext(ctx, "subscriptions: assign", "ticket", t.ID, "session", last, "err", err)
	}
	return nil
}

// lastHolder is the session that last claimed id.
func (c *Controller) lastHolder(ctx context.Context, id string) string {
	return store.LastClaimant(ctx, c.store, id)
}

// subscribe adds a thread to a ticket's meta.subscribe (session or admin).
func (c *Controller) subscribe(ctx context.Context, sc scope, p rpc.SubscribeParams) (any, error) {
	if !config.RepoPattern.MatchString(p.Repo) || p.Number <= 0 {
		return nil, fmt.Errorf("%w: thread must be owner/name and a positive number", domain.ErrInvalid)
	}
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	th := threads.Thread{Repo: p.Repo, Number: p.Number, PR: p.PR, Seen: c.now()}
	sm, _ := threads.DecodeMeta(t.Meta)
	if slices.ContainsFunc(sm.Threads, func(x threads.Thread) bool { return x.Key() == th.Key() }) {
		return t, nil
	}
	f := &threads.Follower{T: t, Threads: append(sm.Threads, th)}
	c.threadEngine().Save(ctx, f)
	if !slices.ContainsFunc(f.Threads, func(x threads.Thread) bool { return x.Key() == th.Key() }) {
		return nil, fmt.Errorf("%w: %s changed, try again", domain.ErrConflict, id)
	}
	return f.T, nil
}

// pumpSubscriptions polls on its own interval until the controller stops.
func (c *Controller) pumpSubscriptions() {
	src := c.threadSrc()
	for {
		every := threads.DefaultInterval
		if s := c.config().Subscriptions; s != nil {
			every = config.Duration(s.Interval, threads.DefaultInterval)
			c.pollSubscriptions(c.ctx, src)
		}
		select {
		case <-c.ctx.Done():
			return
		case <-c.after(every):
		}
	}
}
