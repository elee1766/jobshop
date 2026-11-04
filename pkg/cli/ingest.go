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
	Serve ingestServeCmd `cmd:"" help:"Serve the webhook endpoints: verify, filter and spool deliveries. Needs no controller or main database."`
}

type ingestServeCmd struct {
	Listen string `help:"TCP address (default ingest.listen, else 127.0.0.1:8740)."`
	Spool  string `help:"Spool database (default ingest.spool, else <root>/ingest/spool.db)."`
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
	path := cfg.Ingest.SpoolPath(r)
	if c.Spool != "" {
		path = r.Path(c.Spool)
	}
	// Endpoints are read once: adding one needs a restart. Secrets are
	// read per request, so rotating a secret file does not.
	eps := cfg.IngestEndpoints()
	if len(eps) == 0 {
		return fmt.Errorf("ingest serve: no endpoints: add ingest.endpoints in %s", r.Config())
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, nil)).With("component", "ingest")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	sp, err := ingest.OpenSpool(ctx, path)
	if err != nil {
		return err
	}
	defer func() { _ = sp.Close() }()
	h := &ingest.Handler{Spool: sp, Endpoints: eps, Secret: ingest.SecretFor(r, cfg), Log: log}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("ingest listen: %w", err)
	}
	srv := &http.Server{Handler: h.Mux(), ReadHeaderTimeout: 10 * time.Second}
	// Serve returns as soon as Shutdown starts; wait for it so handlers
	// in flight finish their writes before the spool closes.
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	for _, e := range eps {
		log.Info("ingest: endpoint", "name", e.Name, "source", e.Source, "path", e.EndpointPath(), "repos", len(e.Repos))
	}
	log.Info("ingest: serving", "addr", ln.Addr().String(), "spool", path)
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-done
	return nil
}
