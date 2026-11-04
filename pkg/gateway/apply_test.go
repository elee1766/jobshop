package gateway

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
)

type joinAdapter struct {
	fakeAdapter
	mu     sync.Mutex
	joined []string
}

func (j *joinAdapter) Join(_ context.Context, room string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.joined = append(j.joined, room)
	return nil
}

func TestApplyFollowsConfig(t *testing.T) {
	ad := &joinAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16)}
	var mu sync.Mutex
	next := Config{Rooms: []string{"!room"}, Senders: []string{"@op"}}
	g := New(Config{
		Pool: "operator", Rooms: []string{"!room"}, Senders: []string{"@op"},
		FetchEvery: 5 * time.Millisecond,
		Fetch: func(context.Context) (Config, error) {
			mu.Lock()
			defer mu.Unlock()
			return next, nil
		},
	}, ad, ctl, nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx) }()
	t.Cleanup(func() { cancel(); assert.NoError(t, <-done) })

	m := Message{Thread: Thread{Room: "!new", Root: "$r"}, ID: "$a", Sender: "@claude", Text: "hi"}
	assert.False(t, g.Allowed(m))

	mu.Lock()
	next = Config{Rooms: []string{"!room", "!new"}, Senders: []string{"@op", "@claude"}, Marker: "/"}
	mu.Unlock()
	require.Eventually(t, func() bool { return g.Allowed(m) }, 5*time.Second, 5*time.Millisecond)
	assert.Equal(t, "/", g.settings().Marker)
	assert.Equal(t, "operator", g.settings().Pool, "pool is not changed by Apply")
	ad.mu.Lock()
	assert.Equal(t, []string{"!new"}, ad.joined)
	ad.mu.Unlock()

	// A removed sender is refused again.
	mu.Lock()
	next = Config{Rooms: []string{"!room", "!new"}, Senders: []string{"@op"}}
	mu.Unlock()
	require.Eventually(t, func() bool { return !g.Allowed(m) }, 5*time.Second, 5*time.Millisecond)
}
