package control

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/store"
)

func TestTicketPromptComments(t *testing.T) {
	t.Parallel()
	tk := store.Ticket{ID: "js-1", Title: "Resolve #70", Body: "the body"}
	plain := ticketPromptText("pool prompt", tk, nil)
	assert.NotContains(t, plain, "# Recent comments", "no comments: no section")
	assert.Equal(t, plain, ticketPromptText("pool prompt", tk, []store.Comment{}))

	at := time.Date(2026, 11, 1, 20, 0, 0, 0, time.UTC)
	var cs []store.Comment
	for i, b := range []string{"first", "second", "third", "fourth", strings.Repeat("x", 3000)} {
		cs = append(cs, store.Comment{Author: "js-s", Body: b, CreatedAt: at.Add(time.Duration(i) * time.Minute)})
	}
	p := ticketPromptText("pool prompt", tk, cs)
	head, _, ok := strings.Cut(plain, "The jobshop CLI")
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(p, head), "body before comments")
	assert.Contains(t, p, "# Recent comments")
	assert.NotContains(t, p, "first", "only the last 3")
	assert.NotContains(t, p, "second")
	assert.Contains(t, p, "[2026-11-01T20:02:00Z] js-s:\nthird")
	assert.Contains(t, p, strings.Repeat("x", 2000)+" [truncated]")
	assert.NotContains(t, p, strings.Repeat("x", 2001))
	assert.Contains(t, p, "refinery finding")
	assert.Less(t, strings.Index(p, "# Recent comments"), strings.Index(p, "The jobshop CLI"))
}
