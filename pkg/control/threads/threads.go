// Package threads is the thread-following engine behind subscriptions
// (#187): which GitHub and Linear threads a ticket follows, how their
// comments are read, which of them count as news, and the text a
// delivery carries. It reads and plans; it never delivers.
//
// The controller keeps delivery (comments on the owner ticket, messages
// to the working session, reopening a closed ticket) and every write to
// meta.subscribe. The engine reaches the store, the clock, the config,
// Linear and the identity minter through Host (JOB-156).
package threads

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/linear"
)

const (
	// MetaKey is the key of a ticket's meta that holds its Meta.
	MetaKey = "subscribe"
	// DefaultInterval is the poll interval when subscriptions.interval is
	// not set.
	DefaultInterval = time.Minute
	// DefaultResume is how long after closing a new comment still reopens
	// a ticket when subscriptions.resume_window is not set.
	DefaultResume = 7 * 24 * time.Hour

	// GitHubAuthorPrefix marks ticket comments quoted from GitHub.
	GitHubAuthorPrefix = "github:"
	// LinearAuthorPrefix marks ticket comments quoted from Linear.
	LinearAuthorPrefix = "linear:"

	commentBodyMax = 4000
	replyQuoteMax  = 120
	severalAuthors = "several"
)

// Thread is one followed GitHub thread, or a Linear issue when Linear is
// set (#233), and its read position.
type Thread struct {
	Repo   string    `json:"repo,omitempty"`
	Number int       `json:"number,omitempty"`
	PR     bool      `json:"pr,omitempty"`
	Linear string    `json:"linear,omitempty"`
	Seen   time.Time `json:"seen"`
}

// Key identifies the thread: owner/name#number, or linear:<issue>.
func (t Thread) Key() string {
	if t.Linear != "" {
		return "linear:" + t.Linear
	}
	return fmt.Sprintf("%s#%d", t.Repo, t.Number)
}

// AuthorPrefix marks ticket comments quoted from the thread's tracker.
func (t Thread) AuthorPrefix() string {
	if t.Linear != "" {
		return LinearAuthorPrefix
	}
	return GitHubAuthorPrefix
}

// Name is the thread as a message title names it.
func (t Thread) Name() string {
	if t.Linear != "" {
		return "Linear " + t.Linear
	}
	return "GitHub " + t.Key()
}

// GitHub is the thread as pkg/github reads it.
func (t Thread) GitHub() github.Thread {
	return github.Thread{Repo: t.Repo, Number: t.Number, PR: t.PR}
}

// Comment is a comment or review on a thread: the github.Comment plus
// what the controller adds (the host's own identity, Linear fields).
type Comment struct {
	github.Comment
	Self bool // the host's own identity; always dropped (#402)
	// Parent is the comment replied to (Linear threads only).
	Parent string
	// Linear-only fields (#595). Linear marks the comment as from a Linear
	// thread; Author is then "displayName (name)", Name the full name,
	// Handle the displayName, AuthorID the user or bot actor id. Reply is
	// the "in reply to ..." line, Title and IssueState the issue's.
	Linear     bool
	AuthorID   string
	Name       string
	Handle     string
	Email      string
	Reply      string
	Title      string
	IssueState string
}

// Meta is meta.subscribe.
type Meta struct {
	Threads []Thread `json:"threads,omitempty"`
}

// DecodeMeta reads meta.subscribe and returns the whole meta object too,
// for a writer to put the key back into.
func DecodeMeta(meta json.RawMessage) (Meta, map[string]json.RawMessage) {
	var m map[string]json.RawMessage
	var s Meta
	if len(meta) == 0 || json.Unmarshal(meta, &m) != nil || m == nil {
		return s, map[string]json.RawMessage{}
	}
	if raw, ok := m[MetaKey]; ok {
		_ = json.Unmarshal(raw, &s)
	}
	return s, m
}

// Implicit are the threads a ticket follows without asking: its
// meta.github issue or pr, its meta.linear issue, and PRs linked in its
// comments.
func Implicit(t domain.Ticket, comments []domain.Comment) []Thread {
	var out []Thread
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
		out = append(out, Thread{Linear: m.Linear.Issue})
	}
	if len(t.Meta) > 0 && json.Unmarshal(t.Meta, &m) == nil && m.GitHub.Repo != "" {
		if n := github.RefNumber(m.GitHub.Issue); n > 0 {
			out = append(out, Thread{Repo: m.GitHub.Repo, Number: n})
		}
		if n := github.RefNumber(m.GitHub.PR); n > 0 {
			out = append(out, Thread{Repo: m.GitHub.Repo, Number: n, PR: true})
		}
	}
	for _, c := range comments {
		if strings.HasPrefix(c.Author, GitHubAuthorPrefix) || strings.HasPrefix(c.Author, LinearAuthorPrefix) {
			continue // quoted tracker text, not the ticket's own links
		}
		// PR URLs, e.g. the pr-work "PR <url> head <sha>".
		for _, ref := range github.PRRefs(c.Body) {
			out = append(out, Thread{Repo: ref.Repo, Number: ref.Number, PR: true})
		}
	}
	return out
}

// Follower is a ticket and the threads it follows.
type Follower struct {
	T       domain.Ticket
	Threads []Thread
	Changed bool // Threads gained an entry not yet saved
}

// Of returns the thread k and its read position, the newest seen among
// followers.
func Of(fs []*Follower, k string) (Thread, time.Time) {
	var th Thread
	found := false
	for _, f := range fs {
		for _, x := range f.Threads {
			if x.Key() != k {
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

// Owner picks the follower that gets a thread's news: the one being
// worked, else an open one, else the most recently closed.
func Owner(fs []*Follower) *Follower {
	rank := func(t domain.Ticket) int {
		switch t.Status {
		case domain.StatusInProgress:
			return 0
		case domain.StatusOpen:
			return 1
		}
		return 2
	}
	best := slices.Clone(fs)
	sort.SliceStable(best, func(i, j int) bool {
		a, b := best[i].T, best[j].T
		if rank(a) != rank(b) {
			return rank(a) < rank(b)
		}
		if a.Status == domain.StatusClosed {
			return a.ClosedAt.After(b.ClosedAt)
		}
		return a.CreatedAt.After(b.CreatedAt)
	})
	return best[0]
}
