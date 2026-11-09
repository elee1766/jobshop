package cli

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// sessionJournalCmd reads a session's journal through the controller
// (#529): redacted, and allowed for a session's own journal or a pool
// with the session-journal grant, unlike `session attach`, which reads
// the file directly.
type sessionJournalCmd struct {
	ID     string `arg:"" help:"Session id."`
	Since  int64  `help:"Start after this journal seq."`
	Limit  int    `help:"Frames per page (default 500, max 5000)."`
	Kinds  string `help:"Comma-separated msg.method or sessionUpdate values, e.g. tool_call_update,agent_message_chunk."`
	From   string `help:"Only frames from agent or controller." enum:",agent,controller" default:""`
	Follow bool   `help:"Keep polling for new frames until interrupted."`
}

// journalPollEvery spaces --follow polls.
const journalPollEvery = time.Second

func (c sessionJournalCmd) Run(g *Globals) error {
	p := rpc.SessionJournalParams{ID: c.ID, SinceSeq: c.Since, Limit: c.Limit, From: c.From}
	for _, k := range strings.Split(c.Kinds, ",") {
		if k = strings.TrimSpace(k); k != "" {
			p.Kinds = append(p.Kinds, k)
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	enc := json.NewEncoder(stdout)
	for {
		var res rpc.SessionJournalResult
		if err := g.call(rpc.MethodSessionJournal, p, &res); err != nil {
			return err
		}
		for _, f := range res.Frames {
			if err := enc.Encode(f); err != nil {
				return err
			}
		}
		p.SinceSeq = res.NextSeq
		if !res.Done {
			continue
		}
		if !c.Follow {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(journalPollEvery):
		}
	}
}
