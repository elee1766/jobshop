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
	"github.com/elee1766/jobshop/pkg/gateway"
	"github.com/elee1766/jobshop/pkg/rpc"
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
	ad, err := gateway.NewMatrix(ctx, gc.Homeserver, strings.TrimSpace(string(tok)), gc.Rooms, log)
	if err != nil {
		return err
	}
	gw := gateway.New(gateway.Config{
		Pool: gc.Pool, Rooms: gc.Rooms, Senders: gc.Senders, Marker: gc.InterruptMarker,
	}, ad, socketController{sock: sock}, log)
	log.Info("gateway: running", "network", gc.Network, "pool", gc.Pool)
	return gw.Run(ctx)
}

// socketController dials the admin socket per call, so a controller
// restart costs one failed call, not a dead gateway.
type socketController struct{ sock string }

func (s socketController) Send(ctx context.Context, p rpc.SessionSendParams) (rpc.SessionSendResult, error) {
	var res rpc.SessionSendResult
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	c, err := rpc.Dial(ctx, s.sock)
	if err != nil {
		return res, err
	}
	defer func() { _ = c.Close() }()
	err = c.Call(ctx, rpc.MethodSessionSend, p, &res)
	return res, err
}

func (s socketController) Subscribe(ctx context.Context, p rpc.SessionSubscribeParams, fn func(rpc.SessionEvent)) (rpc.SessionSubscribeResult, error) {
	c, err := rpc.Dial(ctx, s.sock)
	if err != nil {
		return rpc.SessionSubscribeResult{Seq: p.SinceSeq}, err
	}
	defer func() { _ = c.Close() }()
	return c.Subscribe(ctx, p, fn)
}
