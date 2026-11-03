package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elee1766/jobshop/pkg/desk"
)

type deskCmd struct {
	Listen      string   `help:"Address to serve on. Non-loopback needs --allow-remote." default:"127.0.0.1:7480"`
	Host        []string `help:"Extra Host header values to accept (e.g. a tailnet name); loopback is always accepted."`
	DeployState string   `help:"Self-deploy state dir with deployed and failed files." default:"/var/lib/jobshop-deploy"`
	AllowRemote bool     `help:"Allow a non-loopback listen address. Desk has no login; expose it only on a trusted network."`
}

// Run serves Jobshop Desk over the admin socket (#202).
func (c deskCmd) Run(g *Globals) error {
	r, err := g.root()
	if err != nil {
		return err
	}
	// Always the admin socket, as for the gateway: JOBSHOP_SOCKET in an
	// agent's env names its session socket, which cannot list pools.
	sock := r.AdminSocket()
	if !desk.Loopback(c.Listen) && !c.AllowRemote {
		return fmt.Errorf("desk: %s is not a loopback address; pass --allow-remote to expose it", c.Listen)
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, nil)).With("component", "desk")
	srv := &http.Server{
		Handler:           desk.New(desk.Config{Hosts: c.Host, DeployState: c.DeployState}, desk.Socket{Path: sock}, log).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	ln, err := net.Listen("tcp", c.Listen)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv.BaseContext = func(net.Listener) context.Context { return ctx }
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	log.Info("desk: serving", "url", "http://"+ln.Addr().String())
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
