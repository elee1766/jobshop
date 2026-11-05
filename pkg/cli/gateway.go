package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/goccy/go-yaml"

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
	tok, err := config.ReadSecret(context.Background(), gc.TokenFile.In(r.Dir))
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
		EscalationRoom: gc.EscalationRoom, Digest: digest, ArtifactRoom: gc.ArtifactRoom, DeskURL: gc.DeskURL,
		Fetch:      socketController{sock: sock}.gatewayConfig(c.Name),
		FetchEvery: gatewayConfigEvery,
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

func (s socketController) Answer(ctx context.Context, p rpc.AnswerParams) (rpc.AnswerResult, error) {
	var res rpc.AnswerResult
	return res, s.call(ctx, rpc.MethodQuestionAnswer, p, &res)
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

// gatewayConfigEvery is how often a gateway reads its settings from the
// controller, so a reload reaches it without a restart (#307).
const gatewayConfigEvery = 30 * time.Second

// gatewayConfig returns a fetch for gateway name's rooms, senders,
// marker and escalation room from the controller's active config.
func (s socketController) gatewayConfig(name string) func(context.Context) (gateway.Config, error) {
	return func(ctx context.Context) (gateway.Config, error) {
		var res rpc.ConfigShowResult
		if err := s.call(ctx, rpc.MethodConfigShow, struct{}{}, &res); err != nil {
			return gateway.Config{}, err
		}
		gws, _ := res.Config["gateway"].(map[string]any)
		raw, ok := gws[name]
		if !ok {
			return gateway.Config{}, fmt.Errorf("no gateway %q in the controller's config", name)
		}
		b, err := yaml.Marshal(raw)
		if err != nil {
			return gateway.Config{}, err
		}
		var gc config.Gateway
		if err := yaml.Unmarshal(b, &gc); err != nil {
			return gateway.Config{}, err
		}
		return gateway.Config{Rooms: gc.Rooms, Senders: gc.Senders, Marker: gc.InterruptMarker, EscalationRoom: gc.EscalationRoom, ArtifactRoom: gc.ArtifactRoom, DeskURL: gc.DeskURL}, nil
	}
}
