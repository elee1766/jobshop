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
	"github.com/elee1766/jobshop/pkg/rpc"
)

type ingestCmd struct {
	Serve ingestServeCmd `cmd:"" help:"Serve the webhook endpoints: verify, filter and spool deliveries. Needs no controller or main database."`
	Drain ingestDrainCmd `cmd:"" help:"Move spooled deliveries into the inbox through the controller (admin). The controller also drains on its own."`
}

type ingestDrainCmd struct {
	Once  bool          `help:"Drain once and exit."`
	Spool string        `help:"Spool database to drain (default the controller's ingest.spool)."`
	Every time.Duration `default:"5s" help:"Pause between passes without --once."`
}

func (c ingestDrainCmd) Run(g *Globals) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	for {
		var r rpc.IngestDrainResult
		if err := g.call(rpc.MethodIngestDrain, rpc.IngestDrainParams{Spool: c.Spool}, &r); err != nil {
			return err
		}
		if r.Moved > 0 || r.Bad > 0 || c.Once {
			if _, err := fmt.Fprintf(stdout, "moved %d (new %d), bad %d\n", r.Moved, r.Fresh, r.Bad); err != nil {
				return err
			}
		}
		if c.Once {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.Every):
		}
	}
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
