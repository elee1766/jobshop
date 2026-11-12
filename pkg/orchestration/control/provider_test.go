package control_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	packpkg "github.com/gfx-labs/jobshop/pkg/pack"
	"github.com/gfx-labs/jobshop/pkg/support/clock"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// slotRoot is a root with a provider resource emu whose provider script
// logs each call to calls, fails a reset while failreset exists and a
// health check of slot <id> while failhealth-<id> exists.
func slotRoot(t *testing.T) config.Root {
	t.Helper()
	return slotRootWith(t, "idle_release: 10m")
}

// slotRootWith is slotRoot with extra provider settings.
func slotRootWith(t *testing.T, provider string) config.Root {
	t.Helper()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "prov.sh"), []byte(`echo "$1 $2 $JOBSHOP_RESOURCE" >> calls
if [ "$1" = reset ] && [ -f failreset ]; then exit 1; fi
if [ "$1" = health ] && [ -f "failhealth-$2" ]; then exit 1; fi
exit 0
`), 0o600))
	pack := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pack, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, packpkg.FileName), []byte(`workflows:
  device:
    vars: [issue]
    steps:
      - id: test
        type: command
        needs_resources: [emu]
        run: 'echo "{\"serial\":\"$ANDROID_SERIAL\",\"slot\":\"$JOBSHOP_SLOT_EMU\"}"'
  agent:
    vars: [issue]
    steps:
      - {id: work, pool: dev, prompt: w, title: "Work {{.issue}}", needs_resources: [emu]}
`), 0o600))
	cfg := fmt.Sprintf(`packs:
  - path: pk
resources:
  emu:
    provider: {command: "sh prov.sh", %s}
projects:
  js: {}
agents:
  fake:
    command: %q
pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: %q}
`, provider, os.Args[0], dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	return root
}

func providerCalls(t *testing.T, root config.Root) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root.Dir, "calls"))
	if os.IsNotExist(err) {
		return ""
	}
	require.NoError(t, err)
	return string(b)
}

func slotsOf(t *testing.T, admin *rpc.Client) map[string]domain.Slot {
	t.Helper()
	var r rpc.ResourceListResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceList, nil, &r))
	out := map[string]domain.Slot{}
	for _, s := range r.Slots {
		out[s.ID] = s
	}
	return out
}

// TestProviderSlots: a step needing a provider resource waits for a
// slot, gets the slot's env, and its slot is reset once after the claim
// (#407). A failed reset makes the slot gone; demand and idle remove go
// to the provider.
func TestProviderSlots(t *testing.T) {
	t.Parallel()
	root := slotRoot(t)
	clk := clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC))
	admin, _ := startAuto(t, root, clk)

	r := startRun(t, admin, rpc.RunStartParams{Workflow: "device", Vars: map[string]string{"issue": "1"}})
	pass(t, root.Dir, 2)
	assert.Equal(t, domain.StatusOpen, showRun(t, admin, r.Run.ID).Steps[0].Status, "no slot yet")
	passUntilWall(t, root, func() bool {
		return strings.Contains(providerCalls(t, root), "demand  emu")
	}, "demand sent")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceAdd,
		rpc.SlotParams{Resource: "emu", Slot: "e1", Env: map[string]string{"ANDROID_SERIAL": "emulator-5554"}}, nil))
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r.Run.ID).Steps[0].Status == domain.StatusClosed
	}, "step ran on the slot")
	sf := stepFields(t, showRun(t, admin, r.Run.ID).Steps[0])
	assert.JSONEq(t, `{"serial":"emulator-5554","slot":"e1"}`, sf.Output, "slot env injected")

	passUntilWall(t, root, func() bool { return slotsOf(t, admin)["e1"].State == domain.SlotReady }, "slot reset")
	assert.Equal(t, 1, strings.Count(providerCalls(t, root), "reset e1 emu"), "reset once")

	// A failed reset: the slot is gone.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "failreset"), nil, 0o600))
	r2 := startRun(t, admin, rpc.RunStartParams{Workflow: "device", Vars: map[string]string{"issue": "2"}})
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r2.Run.ID).Steps[0].Status == domain.StatusClosed
	}, "second step ran")
	passUntilWall(t, root, func() bool { return slotsOf(t, admin)["e1"].State == domain.SlotGone }, "failed reset: gone")
	require.NoError(t, os.Remove(filepath.Join(root.Dir, "failreset")))

	// Idle release: a ready slot idle past idle_release with nothing
	// waiting is offered back to the provider.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceAdd, rpc.SlotParams{Resource: "emu", Slot: "e1"}, nil))
	pass(t, root.Dir, 1)
	assert.NotContains(t, providerCalls(t, root), "remove e1")
	clk.Advance(11 * time.Minute)
	passUntilWall(t, root, func() bool { return strings.Contains(providerCalls(t, root), "remove e1 emu") }, "idle remove asked")
	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceRemove, rpc.SlotParams{Resource: "emu", Slot: "e1"}, nil))
	assert.NotContains(t, slotsOf(t, admin), "e1")
}

// TestProviderSlotHealth: with health_every, ready slots idle that long
// get a health call; a failing one leaves the pool and a step runs on the
// healthy one. A passing slot is not checked again until health_every
// after its check (JOB-204).
func TestProviderSlotHealth(t *testing.T) {
	t.Parallel()
	root := slotRootWith(t, "health_every: 5m")
	clk := clock.NewFake(time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC))
	admin, _ := startAuto(t, root, clk)
	for _, id := range []string{"e1", "e2"} {
		require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceAdd, rpc.SlotParams{Resource: "emu", Slot: id}, nil))
	}
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "failhealth-e1"), nil, 0o600))
	pass(t, root.Dir, 2)
	assert.NotContains(t, providerCalls(t, root), "health", "not idle long enough")

	clk.Advance(6 * time.Minute)
	passUntilWall(t, root, func() bool {
		s := slotsOf(t, admin)
		return s["e1"].State == domain.SlotGone && strings.Contains(providerCalls(t, root), "health e2 emu")
	}, "failed health check: e1 gone")
	assert.Equal(t, domain.SlotReady, slotsOf(t, admin)["e2"].State)
	pass(t, root.Dir, 2)
	assert.Equal(t, 1, strings.Count(providerCalls(t, root), "health e2 emu"), "checked once per health_every")
	assert.Equal(t, 1, strings.Count(providerCalls(t, root), "health e1 emu"), "a gone slot is not checked")

	r := startRun(t, admin, rpc.RunStartParams{Workflow: "device", Vars: map[string]string{"issue": "1"}})
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r.Run.ID).Steps[0].Status == domain.StatusClosed
	}, "step ran on the healthy slot")
	assert.JSONEq(t, `{"serial":"","slot":"e2"}`, stepFields(t, showRun(t, admin, r.Run.ID).Steps[0]).Output)

	// The provider adds e1 back: in the pool again, timed from the add.
	require.NoError(t, os.Remove(filepath.Join(root.Dir, "failhealth-e1")))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceAdd, rpc.SlotParams{Resource: "emu", Slot: "e1"}, nil))
	assert.Equal(t, domain.SlotReady, slotsOf(t, admin)["e1"].State)
	clk.Advance(6 * time.Minute)
	passUntilWall(t, root, func() bool {
		return strings.Count(providerCalls(t, root), "health e1 emu") == 2
	}, "re-added slot checked")
	pass(t, root.Dir, 1)
	assert.Equal(t, domain.SlotReady, slotsOf(t, admin)["e1"].State, "passing check keeps it")
}

// TestProviderSlotUnhealthy: a slot that goes unhealthy while held
// requeues its step without counting a failure (#407).
func TestProviderSlotUnhealthy(t *testing.T) {
	t.Parallel()
	root := slotRoot(t)
	admin, _ := startAuto(t, root, clock.NewFake(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceAdd, rpc.SlotParams{Resource: "emu", Slot: "e1"}, nil))
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "agent", Vars: map[string]string{"issue": "1"}})
	id := r.Steps[0].ID
	require.NoError(t, admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{ID: id, Holder: "s1"}, nil))
	assert.Equal(t, domain.SlotHeld, slotsOf(t, admin)["e1"].State)
	assert.Empty(t, readyIDs(t, admin))

	var res rpc.SlotResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceUnhealthy,
		rpc.SlotParams{Resource: "emu", Slot: "e1", Reason: "adb offline"}, &res))
	assert.Equal(t, id, res.Requeued)
	st := showRun(t, admin, r.Run.ID).Steps[0]
	assert.Equal(t, domain.StatusOpen, st.Status)
	assert.Zero(t, stepFields(t, st).Failures)
	assert.Empty(t, readyIDs(t, admin), "no healthy slot")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceAdd, rpc.SlotParams{Resource: "emu", Slot: "e2"}, nil))
	assert.Equal(t, []string{id}, readyIDs(t, admin))

	err := admin.Call(t.Context(), rpc.MethodResourceAdd, rpc.SlotParams{Resource: "nope", Slot: "x"}, nil)
	require.Error(t, err)
}
