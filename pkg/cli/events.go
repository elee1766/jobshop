package cli

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/elee1766/jobshop/pkg/rpc"
)

type eventsCmd struct {
	Since  int64    `help:"Replay history after this seq first."`
	Kind   []string `help:"Only these kinds (ticket, run, session); repeatable."`
	Label  []string `help:"Only tickets with one of these labels; repeatable."`
	Follow bool     `help:"Reconnect from the last seq after a lagged event or error." default:"true" negatable:""`
}

// Run prints events.subscribe as JSON lines until Ctrl-C (#147).
func (c eventsCmd) Run(g *Globals) error {
	sock := g.Socket
	if sock == "" {
		r, err := g.root()
		if err != nil {
			return err
		}
		sock = r.AdminSocket()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	enc := json.NewEncoder(stdout)
	since := c.Since
	for {
		cl, err := rpc.Dial(ctx, sock)
		if err != nil {
			return err
		}
		res, err := cl.Events(ctx, rpc.EventsSubscribeParams{SinceSeq: since, Kinds: c.Kind, Labels: c.Label}, func(ev rpc.Event) {
			_ = enc.Encode(ev)
		})
		_ = cl.Close()
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return err
		}
		if !c.Follow || !res.Lagged {
			return nil
		}
		since = res.Seq
	}
}
