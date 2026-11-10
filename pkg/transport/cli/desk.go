package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/elee1766/jobshop/pkg/transport/desk"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

type deskCmd struct {
	LogFlags
	Action       string   `arg:"" optional:"" help:"View command: navigate, filter, select, attach, expand or collapse. Empty serves Desk."`
	Args         []string `arg:"" optional:"" help:"View arguments as key=value, for example view=trace ref=js-123."`
	Session      string   `help:"Named chat session to target. Defaults to the calling agent's session."`
	Listen       string   `help:"Address to serve on. Non-loopback needs --allow-remote." default:"127.0.0.1:7480"`
	Host         []string `help:"Extra Host header values to accept (e.g. a tailnet name); loopback is always accepted."`
	AllowRemote  bool     `help:"Allow a non-loopback listen address. Desk has no login; expose it only on a trusted network."`
	ChatPool     string   `help:"type: chat pool for Desk chat (escalation threads, new sessions). Empty disables sending." default:"operator"`
	UserHeader   string   `help:"Header an authenticating proxy sets to the user (e.g. Tailscale-User-Login). Writes record desk:<user>. Empty: no user."`
	BasePath     string   `help:"URL prefix Desk is served under behind a proxy, e.g. /desk/. The proxy may keep or strip it." default:"/"`
	TrustedProxy []string `help:"Peer address or CIDR (repeatable) whose --user-header and X-Forwarded-Prefix Desk trusts. Required with --user-header. --base-path wins over X-Forwarded-Prefix."`
	Push         bool     `help:"Send Web Push notifications for new attention items. Browsers allow push only over HTTPS (or localhost)."`
	PushSubject  string   `name:"push-subject" help:"VAPID contact push services see (mailto: or https: URL)." default:"mailto:jobshop@localhost"`
	Favicon      string   `help:"Image file (SVG, PNG or ICO) used as the browser tab icon instead of the built-in one." type:"existingfile"`
}

// Run serves Jobshop Desk over the admin socket (#202).
func (c deskCmd) Run(g *Globals) error {
	if c.Action != "" {
		args := map[string]string{}
		for _, arg := range c.Args {
			key, value, ok := strings.Cut(arg, "=")
			if !ok || key == "" {
				return fmt.Errorf("desk: want key=value, got %q", arg)
			}
			args[key] = value
		}
		return g.call(rpc.MethodDeskCommand, rpc.DeskCommandParams{Session: c.Session, Action: c.Action, Args: args}, nil)
	}
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
	base := desk.CleanBasePath(c.BasePath)
	if base == "" {
		return fmt.Errorf("desk: --base-path %q: want a path such as /desk/", c.BasePath)
	}
	proxies, err := desk.ParseTrustedProxies(c.TrustedProxy)
	if err != nil {
		return err
	}
	if c.UserHeader != "" && len(proxies) == 0 {
		return errors.New("desk: --user-header needs at least one --trusted-proxy")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	tel, err := startTelemetry(ctx, r, "desk", c.LogFlags)
	if err != nil {
		return err
	}
	defer func() { _ = tel.Shutdown(context.WithoutCancel(ctx)) }()
	log := tel.Log.With("component", "desk")
	cfg := desk.Config{Hosts: c.Host, ChatPool: c.ChatPool, UserHeader: c.UserHeader, BasePath: base, TrustedProxies: proxies, Root: r.Dir}
	if c.Favicon != "" {
		if cfg.Favicon, err = os.ReadFile(c.Favicon); err != nil {
			return fmt.Errorf("desk: --favicon: %w", err)
		}
	}
	if c.Push {
		// Under the root, with the key and subscriptions: only Desk reads it.
		cfg.Push = &desk.Push{File: filepath.Join(r.Dir, "desk", "push.json"), Subject: c.PushSubject}
	}
	d := desk.New(cfg, desk.Socket{Path: sock}, log)
	if !desk.Built() {
		log.Warn("desk UI is not built into this binary: serving a placeholder page")
	}
	srv := &http.Server{
		Handler:           d.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	ln, err := net.Listen("tcp", c.Listen)
	if err != nil {
		return err
	}
	if err := d.Start(ctx); err != nil {
		return err
	}
	srv.BaseContext = func(net.Listener) context.Context { return ctx }
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	log.Info("desk: serving", "url", "http://"+ln.Addr().String()+base)
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
