package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/cron"
	"github.com/elee1766/jobshop/pkg/gateway"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type gatewayCmd struct {
	Name string `arg:"" help:"Gateway name under gateway: in jobshop.yaml."`
}

func (c gatewayCmd) Run(g *Globals) error {
	r, err := g.root()
	if err != nil {
		return err
	}
	cfg, err := config.Load(r)
	if err != nil {
		return err
	}
	gc, ok := cfg.Gateway[c.Name]
	if !ok {
		return fmt.Errorf("no gateway %q in %s", c.Name, r.Config())
	}
	tokenFile := gc.TokenFile
	if !filepath.IsAbs(tokenFile) {
		tokenFile = filepath.Join(r.Dir, tokenFile)
	}
	tok, err := os.ReadFile(tokenFile)
	if err != nil {
		return fmt.Errorf("gateway %s: token: %w", c.Name, err)
	}
	// Always the admin socket: JOBSHOP_SOCKET in an agent's env names its
	// session socket, which cannot send to other sessions.
	sock := r.AdminSocket()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil)).With("gateway", c.Name)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var ad gateway.Adapter
	if gc.Network == config.GatewayDiscord {
		ad, err = gateway.NewDiscord(strings.TrimSpace(string(tok)), log)
	} else {
		ad, err = gateway.NewMatrix(ctx, gc.Homeserver, strings.TrimSpace(string(tok)), gc.Rooms, log)
	}
	if err != nil {
		return err
	}
	var digest *cron.Schedule
	if gc.AttentionDigest != "" {
		if digest, err = cron.Parse(gc.AttentionDigest); err != nil {
			return err
		}
	}
	gw := gateway.New(gateway.Config{
		Name: c.Name, Pool: gc.Pool, Rooms: gc.Rooms, Senders: gc.Senders, Marker: gc.InterruptMarker,
		EscalationRoom: gc.EscalationRoom, Digest: digest,
	}, ad, socketController{sock: sock}, log)
	log.Info("gateway: running", "network", gc.Network, "pool", gc.Pool, "escalation_room", gc.EscalationRoom)
	return gw.Run(ctx)
}

// socketController dials the admin socket per call, so a controller
// restart costs one failed call, not a dead gateway.
type socketController struct{ sock string }

func (s socketController) call(ctx context.Context, method string, p, out any) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	c, err := rpc.Dial(ctx, s.sock)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	return c.Call(ctx, method, p, out)
}

func (s socketController) Send(ctx context.Context, p rpc.SessionSendParams) (rpc.SessionSendResult, error) {
	var res rpc.SessionSendResult
	return res, s.call(ctx, rpc.MethodSessionSend, p, &res)
}

func (s socketController) List(ctx context.Context, p rpc.ListParams) ([]store.Ticket, error) {
	var res []store.Ticket
	return res, s.call(ctx, rpc.MethodTicketList, p, &res)
}

func (s socketController) Show(ctx context.Context, id string) (rpc.ShowResult, error) {
	var res rpc.ShowResult
	return res, s.call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: id}, &res)
}

func (s socketController) Update(ctx context.Context, p rpc.UpdateParams) (store.Ticket, error) {
	var res store.Ticket
	return res, s.call(ctx, rpc.MethodTicketUpdate, p, &res)
}

func (s socketController) Events(ctx context.Context, p rpc.EventsSubscribeParams, fn func(rpc.Event)) (rpc.EventsSubscribeResult, error) {
	c, err := rpc.Dial(ctx, s.sock)
	if err != nil {
		return rpc.EventsSubscribeResult{Seq: p.SinceSeq}, err
	}
	defer func() { _ = c.Close() }()
	return c.Events(ctx, p, fn)
}

func (s socketController) Subscribe(ctx context.Context, p rpc.SessionSubscribeParams, fn func(rpc.SessionEvent)) (rpc.SessionSubscribeResult, error) {
	c, err := rpc.Dial(ctx, s.sock)
	if err != nil {
		return rpc.SessionSubscribeResult{Seq: p.SinceSeq}, err
	}
	defer func() { _ = c.Close() }()
	return c.Subscribe(ctx, p, fn)
}

func (s socketController) Attention(ctx context.Context) (rpc.AttentionResult, error) {
	var res rpc.AttentionResult
	return res, s.call(ctx, rpc.MethodAttention, nil, &res)
}
