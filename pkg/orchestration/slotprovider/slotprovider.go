// Package slotprovider is the provider side of resource slots (#407),
// split out of pkg/orchestration/control (JOB-158). The store keeps slot state and takes
// a slot in the claim transaction; this package calls the provider: the
// call transport (a command or an http POST with a timeout), the demand
// throttle, idle release and the slot env an agent gets. See
// docs/notes/JOB-158-slotprovider-split.md.
//
// The controller keeps the admin RPC (add, remove, fail slots), resetting
// slots after claims and every store write. It hands in the clock, slot
// reads and the place to run background calls through Host.
package slotprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
)

// outboundCeiling bounds an http call whose caller sets no deadline. Every
// call sets its own, shorter timeout (the provider's), so this only stops
// a call that would otherwise hang for good.
const outboundCeiling = 24 * time.Hour

var providerHTTP = telemetry.HTTPClient("provider", outboundCeiling)

// outputMax bounds the stderr kept from a failed command.
const outputMax = 64 << 10

// Host is what the client needs from the controller.
type Host interface {
	Now() time.Time
	// Resources reads slots and demand; pool "" is every pool.
	Resources(ctx context.Context, pool string) (domain.ResourceReport, error)
	// Go runs f in the background, tracked so the controller waits for it
	// on shutdown.
	Go(f func())
	// Context is the controller's, not the tick's: a demand or remove
	// call outlives the pass.
	Context() context.Context
	// Dir is the working directory of a command provider.
	Dir() string
	// Env is the environment a command provider starts with.
	Env() []string
}

// Call is the JSON a provider gets on stdin or as the POST body.
type Call struct {
	Op       string             `json:"op"` // reset, demand, remove
	Resource string             `json:"resource"`
	Slot     string             `json:"slot,omitempty"`
	Env      map[string]string  `json:"env,omitempty"`
	Demand   *domain.SlotDemand `json:"demand,omitempty"`
}

// Client calls providers and tracks demand and idle release. Its state is
// pass-only (Pass is not called concurrently).
type Client struct {
	h   Host
	log *slog.Logger

	demandAt   map[string]time.Time
	demandLast map[string]domain.SlotDemand
	// removing marks slots already asked to be removed, so a slow
	// provider is not asked every pass.
	removing map[string]time.Time
}

// New returns a Client on h.
func New(h Host, log *slog.Logger) *Client {
	return &Client{
		h: h, log: log,
		demandAt:   map[string]time.Time{},
		demandLast: map[string]domain.SlotDemand{},
		removing:   map[string]time.Time{},
	}
}

// Call runs one provider call and reports whether it succeeded.
func (c *Client) Call(ctx context.Context, p *config.Provider, call Call) error {
	ctx, cancel := context.WithTimeout(ctx, p.TimeoutDuration())
	defer cancel()
	body, err := json.Marshal(call)
	if err != nil {
		return err
	}
	if p.URL != "" {
		req, err := http.NewRequestWithContext(telemetry.WithRoute(ctx, "/provider/"+call.Op), http.MethodPost, p.URL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := providerHTTP.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			return fmt.Errorf("provider %s: %s", call.Op, resp.Status)
		}
		return nil
	}
	args := []string{"-c", p.Command + ` "$@"`, "provider", call.Op}
	if call.Slot != "" {
		args = append(args, call.Slot)
	}
	cmd, done := telemetry.ExecContext(ctx, "provider-"+call.Op, "sh", args...)
	cmd.Dir = c.h.Dir()
	cmd.Stdin = bytes.NewReader(body)
	cmd.Env = append(c.h.Env(), "JOBSHOP_RESOURCE="+call.Resource, "JOBSHOP_SLOT="+call.Slot)
	var errOut limitedBuffer
	cmd.Stderr = &errOut
	cmd.WaitDelay = 5 * time.Second
	err = cmd.Run()
	done(err)
	if err != nil {
		if tail := strings.TrimSpace(errOut.String()); tail != "" {
			return fmt.Errorf("%w: %s", err, lastLines(tail, 5))
		}
		return err
	}
	return nil
}

// Pass sends demand and asks for idle slots to be removed, for every
// provider resource in cfg. The caller has already reset slots.
func (c *Client) Pass(ctx context.Context, cfg *config.Config) {
	now := c.h.Now()
	rep, err := c.h.Resources(ctx, "")
	if err != nil {
		return
	}
	slots := rep.Slots
	// Forget remove marks of slots that are gone or no longer idle.
	idleNow := map[string]bool{}
	for _, sl := range slots {
		if sl.State == domain.SlotReady {
			idleNow[sl.Pool+"/"+sl.ID] = true
		}
	}
	maps.DeleteFunc(c.removing, func(k string, _ time.Time) bool { return !idleNow[k] })
	for _, name := range cfg.ResourceProviders() {
		p := cfg.Resources[name].Provider
		prep, err := c.h.Resources(ctx, name)
		if err != nil {
			continue
		}
		d := prep.Demand
		if last, ok := c.demandLast[name]; !ok || last != d || now.Sub(c.demandAt[name]) >= p.DemandEveryDuration() {
			c.demandAt[name], c.demandLast[name] = now, d
			c.h.Go(func() { c.async(p, Call{Op: "demand", Resource: name, Demand: &d}) })
		}
		idle := p.IdleReleaseDuration()
		if idle <= 0 || d.Waiting > 0 {
			continue
		}
		for _, sl := range slots {
			k := sl.Pool + "/" + sl.ID
			if sl.Pool != name || sl.State != domain.SlotReady || now.Sub(sl.UpdatedAt) < idle {
				continue
			}
			if at, ok := c.removing[k]; ok && now.Sub(at) < idle {
				continue
			}
			c.removing[k] = now
			c.h.Go(func() { c.async(p, Call{Op: "remove", Resource: name, Slot: sl.ID, Env: sl.Env}) })
		}
	}
}

// async is a fire-and-forget provider call (demand, remove): the provider
// answers a remove with `resource remove`.
func (c *Client) async(p *config.Provider, call Call) {
	ctx := c.h.Context()
	if err := c.Call(ctx, p, call); err != nil && ctx.Err() == nil {
		c.log.Warn("resource provider", "resource", call.Resource, "op", call.Op, "slot", call.Slot, "err", err)
	}
}

// Env is the env of the slots t's claim holds in rep, sorted, as
// KEY=value.
func Env(rep domain.ResourceReport, t domain.Ticket) []string {
	var out []string
	for _, sl := range domain.TicketSlots(rep, t) {
		out = append(out, "JOBSHOP_SLOT_"+strings.ToUpper(sl.Pool)+"="+sl.ID)
		for _, k := range slices.Sorted(maps.Keys(sl.Env)) {
			out = append(out, k+"="+sl.Env[k])
		}
	}
	return out
}

// Prompt tells an agent which slots it holds and their env. A host
// outlives claims, so the env goes in the prompt, not the process.
func Prompt(env []string) string {
	if len(env) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("# Resource slots\n\nThis claim holds resource slots. Export this env in the shell you use for the work:\n\n```sh\n")
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		fmt.Fprintf(&b, "export %s=%s\n", k, shellQuote(v))
	}
	b.WriteString("```\n\n")
	return b.String()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func lastLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, " | ")
}

// limitedBuffer keeps the first outputMax bytes and drops the rest.
type limitedBuffer struct {
	b         bytes.Buffer
	truncated bool
}

func (l *limitedBuffer) Write(p []byte) (int, error) {
	if room := outputMax - l.b.Len(); room < len(p) {
		l.truncated = true
		if room > 0 {
			l.b.Write(p[:room])
		}
		return len(p), nil
	}
	return l.b.Write(p)
}

func (l *limitedBuffer) String() string {
	if l.truncated {
		return l.b.String() + "\n[truncated]"
	}
	return l.b.String()
}
