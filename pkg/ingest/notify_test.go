package ingest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotifier(t *testing.T) {
	n := NewNotifier()
	done := make(chan struct{})
	ch := n.Changed()
	n.Advance()
	select {
	case <-ch:
	default:
		t.Fatal("Changed not closed by Advance")
	}
	assert.Equal(t, int64(1), n.Wait(0, done), "past the cursor: returns at once")
	assert.Equal(t, int64(1), n.Wait(7, done), "a cursor from an older controller returns at once")
	got := make(chan int64)
	go func() { got <- n.Wait(1, done) }()
	n.Advance()
	assert.Equal(t, int64(2), <-got)
	close(done)
	assert.Equal(t, int64(2), n.Wait(2, done), "done ends the wait")
}
