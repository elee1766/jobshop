package threads

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/store"
)

func lc(id, handle string) Comment {
	return Comment{Comment: github.Comment{Author: "ann (Ann Lee)"}, Linear: true, AuthorID: id, Handle: handle}
}

// TestLinearDeliveryAuthor: one author gets their displayName, a reply to
// a comment outside the fetched set links its URL.
func TestLinearDeliveryAuthor(t *testing.T) {
	t.Parallel()
	th := Thread{Linear: "ENG-5"}
	one := []Comment{lc("u1", "ann"), lc("u1", "ann")}
	assert.Equal(t, "linear:ann", DeliveryAuthor(th, one))
	two := append(slices.Clone(one), lc("u2", "ann"))
	assert.Equal(t, "linear:several", DeliveryAuthor(th, two), "same handle, different id")

	byID := map[string]linear.Comment{}
	assert.Equal(t, "in reply to https://linear.app/c/c0", replyLine(linear.Comment{ParentID: "c0", ParentURL: "https://linear.app/c/c0"}, byID))
	byID["c0"] = linear.Comment{Author: "Bob", Body: strings.Repeat("x", 130) + "\nmore"}
	assert.Equal(t, `in reply to Bob: "`+strings.Repeat("x", 120)+`..."`, replyLine(linear.Comment{ParentID: "c0"}, byID))
}

func TestKeyAndName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "acme/widget#7", Thread{Repo: "acme/widget", Number: 7}.Key())
	assert.Equal(t, "linear:ENG-5", Thread{Linear: "ENG-5"}.Key())
	assert.Equal(t, "GitHub acme/widget#7", Thread{Repo: "acme/widget", Number: 7}.Name())
	assert.Equal(t, "Linear ENG-5", Thread{Linear: "ENG-5"}.Name())
	assert.Equal(t, GitHubAuthorPrefix, Thread{Repo: "a/b", Number: 1}.AuthorPrefix())
	assert.Equal(t, LinearAuthorPrefix, Thread{Linear: "ENG-5"}.AuthorPrefix())
}

func TestText(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	cs := []Comment{{Comment: github.Comment{Author: "alice", Kind: "review", State: "CHANGES_REQUESTED", URL: "u", At: at, Body: "rename foo"}}}
	assert.Equal(t, "New on GitHub PR acme/widget#7:\n\n[2026-11-08T10:00:00Z] alice review (changes_requested) u\nrename foo\n",
		Text(Thread{Repo: "acme/widget", Number: 7, PR: true}, cs))
	long := []Comment{{Comment: github.Comment{Author: "a", Body: strings.Repeat("x", commentBodyMax+1)}}}
	assert.Contains(t, Text(Thread{Repo: "a/b", Number: 1}, long), " [truncated]")
}

func TestOwner(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	closedOld := &Follower{T: store.Ticket{ID: "a", Status: store.StatusClosed, ClosedAt: now.Add(-2 * time.Hour)}}
	closedNew := &Follower{T: store.Ticket{ID: "b", Status: store.StatusClosed, ClosedAt: now.Add(-time.Hour)}}
	open := &Follower{T: store.Ticket{ID: "c", Status: store.StatusOpen}}
	work := &Follower{T: store.Ticket{ID: "d", Status: store.StatusInProgress}}
	assert.Equal(t, "b", Owner([]*Follower{closedOld, closedNew}).T.ID)
	assert.Equal(t, "c", Owner([]*Follower{closedNew, open}).T.ID)
	assert.Equal(t, "d", Owner([]*Follower{open, closedNew, work}).T.ID)
}
