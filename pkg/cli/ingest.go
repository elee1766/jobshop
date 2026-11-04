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

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/ingest"
)

type ingestCmd struct {
	Serve ingestServeCmd `cmd:"" help:"Serve the webhook receiver: verify, filter and spool deliveries. Needs no controller or database."`
}

type ingestServeCmd struct {
	Listen   string `help:"TCP address (default ingest.listen, else 127.0.0.1:8740)."`
	SpoolDir string `name:"spool-dir" help:"Spool directory (default ingest.spool_dir, else <root>/ingest/spool)."`
}

func (c ingestServeCmd) Run(g *Globals) error {
	r, err := g.root()
	if err != nil {
		return err
	}
	cfg, err := config.Load(r)
	if err != nil {
		return err
	}
	addr := cfg.Ingest.ListenAddr()
	if c.Listen != "" {
		addr = c.Listen
	}
	dir := cfg.Ingest.Spool(r)
	if c.SpoolDir != "" {
		dir = r.Path(c.SpoolDir)
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, nil)).With("component", "ingest")
	warn := func(name string, err error) { log.Warn("ingest: secret", "name", name, "err", err) }
	h := &ingest.Handler{Spool: ingest.Spool{Dir: dir}, Log: log}
	// Secrets are read per request, so a rotated file needs no restart.
	// A source with no secret configured at start is not served.
	if len(ingest.GitHubSecrets(r, cfg, warn)) > 0 {
		h.GitHub = func() []ingest.Secret { return ingest.GitHubSecrets(r, cfg, warn) }
	}
	if len(ingest.LinearSecrets(r, cfg, warn)) > 0 {
		h.Linear = func() []ingest.Secret { return ingest.LinearSecrets(r, cfg, warn) }
	}
	if h.GitHub == nil && h.Linear == nil {
		return fmt.Errorf("ingest serve: no webhook secret: set ingest.github_secret_file, ingest.linear_secret_file or a linear-app webhook_secret_file in %s", r.Config())
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("ingest listen: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv := &http.Server{Handler: h.Mux(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	log.Info("ingest: serving", "addr", ln.Addr().String(), "spool", dir, "github", h.GitHub != nil, "linear", h.Linear != nil)
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
