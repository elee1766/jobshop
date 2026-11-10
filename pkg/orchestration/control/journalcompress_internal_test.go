package control

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	acp "github.com/coder/acp-go-sdk"
	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/journal"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// subscribePeer is a socket peer for sessionSubscribe: notifications sent
// on conn arrive at the other end of a pipe and are collected.
type subscribePeer struct {
	mu   sync.Mutex
	seqs []int64
}

func newSubscribePeer(t *testing.T) (*peer, *subscribePeer) {
	t.Helper()
	ar, aw := io.Pipe()
	br, bw := io.Pipe()
	t.Cleanup(func() { _ = aw.Close(); _ = bw.Close() })
	got := &subscribePeer{}
	// The far end decodes session.event notifications.
	_ = acp.NewConnection(func(_ context.Context, method string, params json.RawMessage) (any, *acp.RequestError) {
		if method == rpc.NotifySessionEvent {
			var ev rpc.SessionEvent
			if json.Unmarshal(params, &ev) == nil && ev.Seq > 0 {
				got.mu.Lock()
				got.seqs = append(got.seqs, ev.Seq)
				got.mu.Unlock()
			}
		}
		return nil, nil
	}, bw, ar)
	conn := acp.NewConnection(func(context.Context, string, json.RawMessage) (any, *acp.RequestError) { return nil, nil }, aw, br)
	return &peer{conn: conn}, got
}

func (s *subscribePeer) get() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]int64(nil), s.seqs...)
}

// TestJournalCompressSweep: a session closed longer than
// journal.compress.after ago has its journal compressed by the sweep;
// a newer one and an open one are left. The compressed journal is still
// served by session.journal, session.subscribe (since and tail) and the
// recall index.
func TestJournalCompressSweep(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanControllerYAML(t, journalYAML)
	ctx := t.Context()
	cfg := *c.config()
	cfg.Journal.Compress.After = "24h"
	cfg.Journal.Compress.Level = 3
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)

	mk := func() domain.Ticket {
		s := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
		writeJournal(t, c, s.ID, 10)
		return s
	}
	old, newer, open := mk(), mk(), mk()
	_, err = store.Close(ctx, c.store, "controller", old.ID, nil, "done")
	require.NoError(t, err)
	clk.advance(23 * time.Hour)
	_, err = store.Close(ctx, c.store, "controller", newer.ID, nil, "done")
	require.NoError(t, err)
	clk.advance(2 * time.Hour)

	c.maybeCompressJournals(t.Context(), c.config())
	c.wg.Wait()
	assert.False(t, c.compressor().Running())

	plain := func(id string) string { return filepath.Join(c.root.SessionDir(id), host.JournalFile) }
	zst := func(id string) string { return filepath.Join(c.root.SessionDir(id), host.CompressedJournalFile) }
	assert.NoFileExists(t, plain(old.ID))
	assert.FileExists(t, zst(old.ID))
	assert.FileExists(t, plain(newer.ID), "closed less than after ago")
	assert.FileExists(t, plain(open.ID), "still open")

	// The hour has not passed: no second sweep.
	before := c.compressor().LastSweep()
	assert.False(t, before.IsZero())
	c.maybeCompressJournals(t.Context(), c.config())
	assert.Equal(t, before, c.compressor().LastSweep())

	admin := scope{admin: true, actor: "admin"}
	out, err := c.sessionJournal(ctx, admin, rpc.SessionJournalParams{ID: old.ID, SinceSeq: 6})
	require.NoError(t, err)
	var seqs []int64
	for _, f := range out.(rpc.SessionJournalResult).Frames {
		seqs = append(seqs, f.Seq)
	}
	assert.Equal(t, []int64{7, 8, 9, 10}, seqs)
	assert.NotContains(t, string(out.(rpc.SessionJournalResult).Frames[0].Msg), "sk-supersecret")

	seq, _ := host.JournalTail(c.root.SessionDir(old.ID))
	assert.Equal(t, int64(10), seq)

	subscribe := func(p rpc.SessionSubscribeParams) []int64 {
		t.Helper()
		pr, got := newSubscribePeer(t)
		sc := admin
		sc.peer = pr
		res, err := c.sessionSubscribe(ctx, sc, p)
		require.NoError(t, err)
		r := res.(rpc.SessionSubscribeResult)
		assert.True(t, r.Closed)
		assert.Equal(t, int64(10), r.Seq)
		require.Eventually(t, func() bool { s := got.get(); return len(s) > 0 && s[len(s)-1] == 10 }, 5*time.Second, 10*time.Millisecond)
		return got.get()
	}
	assert.Equal(t, []int64{9}, subscribe(rpc.SessionSubscribeParams{ID: old.ID, SinceSeq: 8})[:1])
	for _, s := range subscribe(rpc.SessionSubscribeParams{ID: old.ID, Tail: 3}) {
		assert.Greater(t, s, int64(7), "only the last 3 frames")
	}

	// Recall indexes a compressed journal through the acphost reader.
	r := &runner{c: c, id: old.ID, pool: "dev", ctx: ctx, dir: c.root.SessionDir(old.ID), log: c.log}
	r.indexJournal(ctx, old.ID)
	hits, err := journal.Recall(ctx, c.store, "hi", "", "", 5)
	require.NoError(t, err)
	var refs []string
	for _, h := range hits {
		refs = append(refs, h.Ref)
	}
	assert.Contains(t, refs, old.ID)
}

// TestJournalCompressOff: with journal.compress.after empty nothing runs.
func TestJournalCompressOff(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanControllerYAML(t, journalYAML)
	s := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	writeJournal(t, c, s.ID, 3)
	_, err := store.Close(t.Context(), c.store, "controller", s.ID, nil, "done")
	require.NoError(t, err)
	clk.advance(1000 * time.Hour)
	c.maybeCompressJournals(t.Context(), c.config())
	c.wg.Wait()
	_, err = os.Stat(filepath.Join(c.root.SessionDir(s.ID), host.JournalFile))
	require.NoError(t, err)
	assert.True(t, c.compressor().LastSweep().IsZero())
}
