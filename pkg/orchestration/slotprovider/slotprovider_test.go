package slotprovider_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/slotprovider"
)

// fakeHost serves a fixed resource report and runs background calls on a
// WaitGroup the test waits on.
type fakeHost struct {
	t   *testing.T
	dir string
	now time.Time
	rep domain.ResourceReport
	wg  sync.WaitGroup
}

func (h *fakeHost) Now() time.Time           { return h.now }
func (h *fakeHost) Go(f func())              { h.wg.Go(f) }
func (h *fakeHost) Context() context.Context { return h.t.Context() }
func (h *fakeHost) Dir() string              { return h.dir }
func (h *fakeHost) Env() []string            { return os.Environ() }
func (h *fakeHost) Resources(_ context.Context, pool string) (domain.ResourceReport, error) {
	if pool == "" {
		return h.rep, nil
	}
	return domain.ResourceReport{Demand: h.rep.Demand}, nil
}

func newClient(t *testing.T) (*slotprovider.Client, *fakeHost) {
	t.Helper()
	h := &fakeHost{t: t, dir: t.TempDir(), now: time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)}
	return slotprovider.New(h, slog.New(slog.DiscardHandler)), h
}

func cfgOf(p *config.Provider) *config.Config {
	return &config.Config{Resources: map[string]config.Resource{"emu": {Provider: p}}}
}

func calls(t *testing.T, dir string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "calls"))
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	var out []string
	for l := range strings.SplitSeq(strings.TrimSpace(string(b)), "\n") {
		out = append(out, strings.TrimSpace(l))
	}
	return out
}

// TestCallCommand: a command provider gets the op and slot as arguments,
// the resource in env and the call JSON on stdin; a failure carries the
// stderr tail.
func TestCallCommand(t *testing.T) {
	t.Parallel()
	c, h := newClient(t)
	p := &config.Provider{Command: `cat > in.json; echo "$1 $2 $JOBSHOP_RESOURCE $JOBSHOP_SLOT" >> calls; [ "$1" = reset ] || exit 0; echo boom >&2; exit 3`}
	require.NoError(t, c.Call(t.Context(), p, slotprovider.Call{Op: "remove", Resource: "emu", Slot: "e1", Env: map[string]string{"K": "v"}}))
	assert.Equal(t, []string{"remove e1 emu e1"}, calls(t, h.dir))
	b, err := os.ReadFile(filepath.Join(h.dir, "in.json"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"op":"remove","resource":"emu","slot":"e1","env":{"K":"v"}}`, string(b))

	err = c.Call(t.Context(), p, slotprovider.Call{Op: "reset", Resource: "emu", Slot: "e1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

// TestCallHTTP: an http provider gets the call as a JSON POST; a non-2xx
// status is an error.
func TestCallHTTP(t *testing.T) {
	t.Parallel()
	c, _ := newClient(t)
	var got slotprovider.Call
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	p := &config.Provider{URL: srv.URL}
	require.NoError(t, c.Call(t.Context(), p, slotprovider.Call{Op: "demand", Resource: "emu", Demand: &domain.SlotDemand{Waiting: 2}}))
	assert.Equal(t, "demand", got.Op)
	assert.Equal(t, 2, got.Demand.Waiting)

	status = http.StatusInternalServerError
	err := c.Call(t.Context(), p, slotprovider.Call{Op: "demand", Resource: "emu"})
	require.ErrorContains(t, err, "provider demand: 500")
}

// TestPassDemandAndIdle: demand goes out on first sight and on change,
// not again before demand_every; an idle slot is asked to be removed once
// per idle_release, and not while a claim waits.
func TestPassDemandAndIdle(t *testing.T) {
	t.Parallel()
	c, h := newClient(t)
	p := &config.Provider{
		Command:     `echo "$1 $2" >> calls;`,
		DemandEvery: "1h", IdleRelease: "10m",
	}
	cfg := cfgOf(p)
	h.rep = domain.ResourceReport{Slots: []domain.Slot{{Pool: "emu", ID: "e1", State: domain.SlotReady, UpdatedAt: h.now}}}

	c.Pass(t.Context(), cfg)
	h.wg.Wait()
	assert.Equal(t, []string{"demand"}, calls(t, h.dir), "first demand, slot not idle yet")

	h.now = h.now.Add(11 * time.Minute)
	c.Pass(t.Context(), cfg)
	h.wg.Wait()
	assert.Equal(t, []string{"demand", "remove e1"}, calls(t, h.dir), "idle remove, no repeat demand")

	h.now = h.now.Add(time.Minute)
	c.Pass(t.Context(), cfg)
	h.wg.Wait()
	assert.Len(t, calls(t, h.dir), 2, "remove not asked again within idle_release")

	h.rep.Demand = domain.SlotDemand{Waiting: 1}
	h.now = h.now.Add(20 * time.Minute)
	c.Pass(t.Context(), cfg)
	h.wg.Wait()
	assert.Equal(t, []string{"demand", "remove e1", "demand"}, calls(t, h.dir), "changed demand sent, no remove while waiting")
}

// TestHealthDue: a ready slot is due health_every after its last state
// change, then health_every after its last check; held slots and
// resources without health_every are never due (JOB-204).
func TestHealthDue(t *testing.T) {
	t.Parallel()
	c, h := newClient(t)
	cfg := &config.Config{Resources: map[string]config.Resource{
		"emu": {Provider: &config.Provider{Command: "p", HealthEvery: "5m"}},
		"gpu": {Provider: &config.Provider{Command: "p"}},
	}}
	t0 := h.now
	slots := []domain.Slot{
		{Pool: "emu", ID: "e1", State: domain.SlotReady, UpdatedAt: t0},
		{Pool: "emu", ID: "e2", State: domain.SlotHeld, UpdatedAt: t0},
		{Pool: "gpu", ID: "g1", State: domain.SlotReady, UpdatedAt: t0},
	}
	ids := func(due []domain.Slot) []string {
		var out []string
		for _, sl := range due {
			out = append(out, sl.Pool+"/"+sl.ID)
		}
		return out
	}
	assert.Empty(t, c.HealthDue(cfg, slots), "not idle long enough")

	h.now = t0.Add(5 * time.Minute)
	due := c.HealthDue(cfg, slots)
	assert.Equal(t, []string{"emu/e1"}, ids(due))
	c.Checked(due[0])
	h.now = t0.Add(9 * time.Minute)
	assert.Empty(t, c.HealthDue(cfg, slots), "checked 4m ago")
	h.now = t0.Add(10 * time.Minute)
	assert.Equal(t, []string{"emu/e1"}, ids(c.HealthDue(cfg, slots)))

	// Held, then reset: timed from the reset, not the old check.
	slots[0].State = domain.SlotHeld
	assert.Empty(t, c.HealthDue(cfg, slots))
	slots[0].State, slots[0].UpdatedAt = domain.SlotReady, h.now
	h.now = h.now.Add(4 * time.Minute)
	assert.Empty(t, c.HealthDue(cfg, slots))
}

// TestEnvAndPrompt: the env lists each held slot, then its env sorted;
// the prompt quotes values for a shell.
func TestEnvAndPrompt(t *testing.T) {
	t.Parallel()
	tk := domain.Ticket{ID: "t1", Status: domain.StatusInProgress, Fence: 2}
	rep := domain.ResourceReport{Slots: []domain.Slot{
		{Pool: "emu", ID: "e1", State: domain.SlotHeld, Ticket: "t1", Fence: 2, Env: map[string]string{"B": "it's", "A": "1"}},
		{Pool: "emu", ID: "e2", State: domain.SlotReady},
	}}
	env := slotprovider.Env(rep, tk)
	assert.Equal(t, []string{"JOBSHOP_SLOT_EMU=e1", "A=1", "B=it's"}, env)
	assert.Contains(t, slotprovider.Prompt(env), `export B='it'\''s'`)
	assert.Empty(t, slotprovider.Prompt(nil))
	assert.Empty(t, slotprovider.Env(domain.ResourceReport{}, tk))
}
