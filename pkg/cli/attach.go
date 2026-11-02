package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// stdin is swapped by tests.
var stdin io.Reader = os.Stdin

// runRW is `session attach --rw`: output comes from session.subscribe
// rather than the journal file, since sending needs the controller anyway.
// Each stdin line is a session.send; "!text" interrupts the running turn.
// EOF or Ctrl-C detaches; the session keeps running.
func (c sessionAttachCmd) runRW(g *Globals) error {
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
	// Subscribe holds its connection's call open, so sends use another.
	sub, err := rpc.Dial(ctx, sock)
	if err != nil {
		return err
	}
	defer func() { _ = sub.Close() }()
	send, err := rpc.Dial(ctx, sock)
	if err != nil {
		return err
	}
	defer func() { _ = send.Close() }()
	// Refuse before reading input when the pool does not allow it.
	if err := send.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{ID: c.ID, Attach: true}, nil); err != nil {
		return err
	}
	var since int64
	if c.Since != nil {
		since = *c.Since
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		res, err := sub.Subscribe(ctx, rpc.SessionSubscribeParams{ID: c.ID, SinceSeq: since}, func(ev rpc.SessionEvent) {
			renderEvent(stdout, ev, c.JSON)
		})
		if err == nil && res.Closed {
			_, _ = fmt.Fprintf(stdout, "\n[session closed: %s]\n", res.Resolution)
		}
		done <- err
	}()
	lines := make(chan string)
	go func() {
		defer close(lines)
		sc := bufio.NewScanner(stdin)
		sc.Buffer(make([]byte, 64*1024), 1<<20)
		for sc.Scan() {
			select {
			case lines <- sc.Text():
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		select {
		case err := <-done:
			if ctx.Err() != nil {
				return nil
			}
			return err
		case line, ok := <-lines:
			if !ok {
				// EOF detaches; the session keeps running.
				cancel()
				<-done
				return nil
			}
			text, interrupt := strings.CutPrefix(line, "!")
			if strings.TrimSpace(text) == "" {
				continue
			}
			p := rpc.SessionSendParams{ID: c.ID, Text: text, Interrupt: interrupt, Attach: true}
			if err := send.Call(ctx, rpc.MethodSessionSend, p, nil); err != nil && ctx.Err() == nil {
				_, _ = fmt.Fprintf(os.Stderr, "send: %v\n", err)
			}
		}
	}
}

// renderEvent prints one session event: message text as it streams, one
// line per tool state change and turn boundary.
func renderEvent(w io.Writer, ev rpc.SessionEvent, raw bool) {
	if raw {
		_ = json.NewEncoder(w).Encode(ev)
		return
	}
	switch ev.Type {
	case rpc.EventTurnStart:
		_, _ = fmt.Fprintf(w, "\n> %s\n", ev.Text)
	case rpc.EventText:
		_, _ = fmt.Fprint(w, ev.Text)
	case rpc.EventTool:
		_, _ = fmt.Fprintf(w, "\n[tool %s: %s]\n", ev.Tool.Status, ev.Tool.Title)
	case rpc.EventTurnEnd:
		if ev.Error != "" {
			_, _ = fmt.Fprintf(w, "\n[turn error: %s]\n", ev.Error)
		} else {
			_, _ = fmt.Fprintf(w, "\n[turn end: %s]\n", ev.StopReason)
		}
	case rpc.EventState:
		_, _ = fmt.Fprintf(w, "[%s %s %s]\n", ev.Session, ev.Status, ev.Resolution)
	}
}
