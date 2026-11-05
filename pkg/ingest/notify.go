package ingest

import "sync"

// Notifier is the inbox's change signal (#679). The controller advances
// it after every inbox append and every admission; consumers (the
// admission pump, `ingest.next` long polls from external workers) block
// in Wait until the seq passes their cursor, then run a pass. The seq
// counts changes since the controller started; a cursor from an older
// controller is past it and returns at once.
type Notifier struct {
	mu  sync.Mutex
	seq int64
	ch  chan struct{}
}

// NewNotifier starts at seq 0.
func NewNotifier() *Notifier { return &Notifier{ch: make(chan struct{})} }

// Seq is the current change count.
func (n *Notifier) Seq() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.seq
}

// Advance records a change and wakes every waiter.
func (n *Notifier) Advance() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.seq++
	close(n.ch)
	n.ch = make(chan struct{})
}

// Changed is closed by the next Advance. Take it before a pass, so a
// change during the pass is not missed.
func (n *Notifier) Changed() <-chan struct{} {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.ch
}

// Wait returns the seq once it differs from after (past it, or below it
// after a controller restart), or when done fires.
func (n *Notifier) Wait(after int64, done <-chan struct{}) int64 {
	for {
		n.mu.Lock()
		seq, ch := n.seq, n.ch
		n.mu.Unlock()
		if seq != after {
			return seq
		}
		select {
		case <-ch:
		case <-done:
			return n.Seq()
		}
	}
}
