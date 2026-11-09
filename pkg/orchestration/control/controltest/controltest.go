// Package controltest starts a controller for tests that drive it through
// its RPC API (JOB-47), as a user, Desk or a session would. A test outside
// pkg/orchestration/control imports it instead of reaching into controller internals.
//
// The controller runs on a temp root with a fake clock, an hourly ticker
// that never fires on its own, supervisor passes on request (Pass) and
// session hosts in-process (InProcHosts). Sessions whose agent is not in
// the config are external: nothing spawns for them, and a test writes
// their journal files itself.
package controltest

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/orchestration/control"
	"github.com/elee1766/jobshop/pkg/support/clock"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Epoch is the fake clock's start time.
var Epoch = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)

// DialWait bounds how long Dial and PassUntil wait. It only matters on
// failure.
const DialWait = 30 * time.Second

// NewRoot makes a temp controller root and writes cfg as its config,
// with every $ROOT replaced by the root dir (for pool dirs). The dir is
// short (os.MkdirTemp, not t.TempDir) because unix socket paths under it
// are limited to 108 bytes.
func NewRoot(t testing.TB, cfg string) config.Root {
	t.Helper()
	dir, err := os.MkdirTemp("", "js")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	WriteConfig(t, root, strings.ReplaceAll(cfg, "$ROOT", dir))
	return root
}

// WriteConfig replaces root's config. A running controller sees it on
// its next reload.
func WriteConfig(t testing.TB, root config.Root, cfg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
}

// Controller is one running controller and its admin client.
type Controller struct {
	Root  config.Root
	Admin *rpc.Client
	// Clock is the fake clock, nil when Options.Clock was set to another.
	Clock *clock.Fake

	t       testing.TB
	pass    chan chan struct{}
	cancel  context.CancelFunc
	done    chan error
	once    sync.Once
	metrics chan string
	murl    string
}

// Start runs a controller on root until the test ends or Stop. opt's zero
// fields get the test defaults: a fake clock at Epoch, Tick of an hour,
// in-process hosts and the test binary as HostExe. Starting again on the
// same root after Stop is a controller restart.
func Start(t testing.TB, root config.Root, opt control.Options) *Controller {
	t.Helper()
	c := &Controller{Root: root, t: t, pass: make(chan chan struct{}), done: make(chan error, 1), metrics: make(chan string, 1)}
	if opt.Clock == nil {
		c.Clock = clock.NewFake(Epoch)
		opt.Clock = c.Clock
	} else if f, ok := opt.Clock.(*clock.Fake); ok {
		c.Clock = f
	}
	if opt.Tick == 0 {
		opt.Tick = time.Hour
	}
	if opt.Hosts == nil {
		h := NewInProcHosts()
		opt.Hosts = h
		t.Cleanup(h.StopAll)
	}
	if opt.HostExe == "" {
		opt.HostExe = os.Args[0]
	}
	opt.Pass = c.pass
	if opt.MetricsReady == nil {
		opt.MetricsReady = func(addr string) { c.metrics <- addr }
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go func() { c.done <- control.Run(ctx, root, opt) }()
	t.Cleanup(c.Stop)
	c.Admin = c.Dial(root.AdminSocket())
	return c
}

// Option adjusts the controller options before StartWith starts it.
type Option func(*control.Options)

// WithVersion sets the version and commit the version method reports.
func WithVersion(version, commit string) Option {
	return func(o *control.Options) { o.Version, o.Commit = version, commit }
}

// StartWith is Start for a test outside pkg/orchestration/control that needs only the
// default controller or a few options: it names no control type, so an
// adapter's tests do not import the controller (JOB-174).
func StartWith(t testing.TB, root config.Root, opts ...Option) *Controller {
	t.Helper()
	var o control.Options
	for _, f := range opts {
		f(&o)
	}
	return Start(t, root, o)
}

// MetricsURL is the /metrics URL. The config must set
// telemetry.metrics_listen (127.0.0.1:0 picks a free port).
func (c *Controller) MetricsURL() string {
	c.t.Helper()
	if c.murl == "" {
		select {
		case addr := <-c.metrics:
			c.murl = "http://" + addr + "/metrics"
		case <-time.After(DialWait):
			c.t.Fatal("no metrics listener: set telemetry.metrics_listen")
		}
	}
	return c.murl
}

// Stop stops the controller and waits for it. Sessions' hosts keep
// running, as on a real controller restart.
func (c *Controller) Stop() {
	c.once.Do(func() {
		c.cancel()
		if err := <-c.done; err != nil && !errors.Is(err, context.Canceled) {
			c.t.Errorf("controller: %v", err)
		}
	})
}

// Dial connects to a controller socket (the admin socket, or a session's
// cli.sock from session.new), retrying until it listens.
func (c *Controller) Dial(path string) *rpc.Client {
	c.t.Helper()
	deadline := time.Now().Add(DialWait)
	for {
		cl, err := rpc.Dial(context.Background(), path)
		if err == nil {
			c.t.Cleanup(func() { _ = cl.Close() })
			return cl
		}
		if time.Now().After(deadline) {
			c.t.Fatalf("dial %s: %v", path, err)
		}
		<-time.After(10 * time.Millisecond)
	}
}

// Call is Admin.Call that fails the test on an error.
func (c *Controller) Call(method string, params, out any) {
	c.t.Helper()
	if err := c.Admin.Call(context.Background(), method, params, out); err != nil {
		c.t.Fatalf("%s: %v", method, err)
	}
}

// Pass runs n whole supervisor passes and returns when they are done.
// Use it instead of sleeping through ticks before a negative assertion.
func (c *Controller) Pass(n int) {
	c.t.Helper()
	for range n {
		done := make(chan struct{})
		select {
		case c.pass <- done:
		case <-time.After(DialWait):
			c.t.Fatal("controller took no pass")
		}
		select {
		case <-done:
		case <-time.After(DialWait):
			c.t.Fatal("pass did not finish")
		}
	}
}

// PassUntil runs passes until cond holds. Work that runs outside the pass
// (host events, auto steps) may need several, so it waits on the
// condition, not a pass count.
func (c *Controller) PassUntil(cond func() bool, msg string) {
	c.t.Helper()
	deadline := time.Now().Add(DialWait)
	for !cond() {
		if time.Now().After(deadline) {
			c.t.Fatal(msg)
		}
		c.Pass(1)
		<-time.After(10 * time.Millisecond)
	}
}
