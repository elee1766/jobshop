package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
)

type ingestCmd struct {
	Serve   ingestServeCmd   `cmd:"" help:"Serve the webhook endpoints: verify, filter and spool deliveries. Needs no controller or main database."`
	Drain   ingestDrainCmd   `cmd:"" help:"Move spooled deliveries into the inbox through the controller (admin). The controller also drains on its own."`
	Put     ingestPutCmd     `cmd:"" help:"Spool one JSON delivery from stdin for a generic endpoint, with no HTTP or signature (trusted local senders). Exits 0 for new or duplicate deliveries."`
	Work    ingestWorkCmd    `cmd:"" help:"Admit pending inbox items through the controller (admin) with ingest.workers external (#443): claim, run the admission hooks here, report."`
	Workers ingestWorkersCmd `cmd:"" help:"List external ingest workers' heartbeats and inbox counts by state."`
}

type ingestWorkersCmd struct{}

func (ingestWorkersCmd) Run(g *Globals) error {
	return g.call(rpc.MethodIngestWorkers, nil, nil)
}

type ingestWorkCmd struct {
	Once     bool          `help:"Admit until nothing is pending and exit."`
	Worker   string        `help:"Worker id (default <host>-<pid>)."`
	Every    time.Duration `default:"1s" help:"Pause when nothing is pending, without --once. Each claim is the worker's heartbeat."`
	Dispatch bool          `help:"Also dispatch admitted items through the controller (#633). Safe alongside other dispatchers; set ingest.dispatch external to stop the controller's own."`
	NoAdmit  bool          `name:"no-admit" help:"Only dispatch (with --dispatch), for ingest.workers internal."`
}

func (c ingestWorkCmd) Run(g *Globals) error {
	r, err := g.root()
	if err != nil {
		return err
	}
	if c.NoAdmit && !c.Dispatch {
		return errors.New("--no-admit needs --dispatch")
	}
	host, _ := os.Hostname()
	id := c.Worker
	if id == "" {
		id = fmt.Sprintf("%s-%d", host, os.Getpid())
	}
	w := control.IngestWorker{
		Call: func(_ context.Context, method string, params, out any) error { return g.call(method, params, out) },
		Dir:  r.Dir, Worker: id, Host: host, Dispatch: c.Dispatch, NoAdmit: c.NoAdmit,
		Log: slog.New(slog.NewTextHandler(os.Stderr, nil)).With("component", "ingest-work", "worker", id),
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if c.Once {
		n, d, err := w.Pass(ctx)
		if err != nil {
			return err
		}
		if c.Dispatch {
			_, err = fmt.Fprintf(stdout, "admitted %d dispatched %d\n", n, d)
			return err
		}
		_, err = fmt.Fprintf(stdout, "admitted %d\n", n)
		return err
	}
	return w.Run(ctx, c.Every, time.After)
}

type ingestPutCmd struct {
	Endpoint   string `required:"" help:"Generic endpoint name (ingest.endpoints)."`
	DeliveryID string `name:"delivery-id" help:"Delivery id for dedupe (default the endpoint's map.delivery_id, else a hash of the body)."`
	Kind       string `help:"Event kind (default the endpoint's map.kind)."`
	Spool      string `help:"Spool database (default ingest.spool, else <root>/ingest/spool.db)."`
}

func (c ingestPutCmd) Run(g *Globals) error {
	r, err := g.root()
	if err != nil {
		return err
	}
	cfg, err := config.Load(r)
	if err != nil {
		return err
	}
	idx := slices.IndexFunc(cfg.IngestEndpoints(), func(e config.IngestEndpoint) bool { return e.Name == c.Endpoint })
	if idx < 0 {
		return fmt.Errorf("ingest put: no endpoint %q in %s", c.Endpoint, r.Config())
	}
	body, err := io.ReadAll(io.LimitReader(stdin, ingest.MaxBody+1))
	if err != nil {
		return err
	}
	if len(body) > ingest.MaxBody {
		return fmt.Errorf("ingest put: payload over %d bytes", ingest.MaxBody)
	}
	path := cfg.Ingest.SpoolPath(r)
	if c.Spool != "" {
		path = r.Path(c.Spool)
	}
	ctx := context.Background()
	sp, err := ingest.OpenSpool(ctx, path)
	if err != nil {
		return err
	}
	defer func() { _ = sp.Close() }()
	fresh, err := ingest.Put(ctx, sp, cfg.IngestEndpoints()[idx], c.DeliveryID, c.Kind, body, time.Now())
	if err != nil {
		return fmt.Errorf("ingest put: %w", err)
	}
	msg := "spooled"
	if !fresh {
		msg = "duplicate"
	}
	_, err = fmt.Fprintln(stdout, msg)
	return err
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
	// Each new row tells the controller to drain now (#679). One call at
	// a time, coalesced: the drain takes every row. A controller that is
	// down misses it; its drain timer picks the rows up.
	kick := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-kick:
				if err := g.call(rpc.MethodIngestNotify, nil, nil); err != nil {
					log.Debug("ingest: notify controller", "err", err)
				}
			}
		}
	}()
	h := &ingest.Handler{
		Spool: sp, Endpoints: eps, Secret: ingest.SecretFor(r, cfg), Log: log,
		OnSpool: func(ingest.Delivery) {
			select {
			case kick <- struct{}{}:
			default:
			}
		},
	}
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
