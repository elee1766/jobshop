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

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// slotRoot is a root with a provider resource emu whose provider script
// logs each call to calls and fails a reset while failreset exists.
func slotRoot(t *testing.T) config.Root {
	t.Helper()
	dir, err := os.MkdirTemp("", "js") // short path for the unix socket
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	root := config.Root{Dir: dir}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "prov.sh"), []byte(`echo "$1 $2 $JOBSHOP_RESOURCE" >> calls
if [ "$1" = reset ] && [ -f failreset ]; then exit 1; fi
exit 0
`), 0o600))
	pack := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pack, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName), []byte(`workflows:
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
    provider: {command: "sh prov.sh", idle_release: 10m}
projects:
  js: {}
agents:
  fake:
    command: %q
pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: %q}
`, os.Args[0], dir)
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

func slotsOf(t *testing.T, admin *rpc.Client) map[string]store.Slot {
	t.Helper()
	var r rpc.ResourceListResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceList, nil, &r))
	out := map[string]store.Slot{}
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
	assert.Equal(t, store.StatusOpen, showRun(t, admin, r.Run.ID).Steps[0].Status, "no slot yet")
	passUntilWall(t, root, func() bool {
		return strings.Contains(providerCalls(t, root), "demand  emu")
	}, "demand sent")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceAdd,
		rpc.SlotParams{Resource: "emu", Slot: "e1", Env: map[string]string{"ANDROID_SERIAL": "emulator-5554"}}, nil))
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r.Run.ID).Steps[0].Status == store.StatusClosed
	}, "step ran on the slot")
	sf := stepFields(t, showRun(t, admin, r.Run.ID).Steps[0])
	assert.JSONEq(t, `{"serial":"emulator-5554","slot":"e1"}`, sf.Output, "slot env injected")

	passUntilWall(t, root, func() bool { return slotsOf(t, admin)["e1"].State == store.SlotReady }, "slot reset")
	assert.Equal(t, 1, strings.Count(providerCalls(t, root), "reset e1 emu"), "reset once")

	// A failed reset: the slot is gone.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "failreset"), nil, 0o600))
	r2 := startRun(t, admin, rpc.RunStartParams{Workflow: "device", Vars: map[string]string{"issue": "2"}})
	passUntilWall(t, root, func() bool {
		return showRun(t, admin, r2.Run.ID).Steps[0].Status == store.StatusClosed
	}, "second step ran")
	passUntilWall(t, root, func() bool { return slotsOf(t, admin)["e1"].State == store.SlotGone }, "failed reset: gone")
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
	assert.Equal(t, store.SlotHeld, slotsOf(t, admin)["e1"].State)
	assert.Empty(t, readyIDs(t, admin))

	var res rpc.SlotResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceUnhealthy,
		rpc.SlotParams{Resource: "emu", Slot: "e1", Reason: "adb offline"}, &res))
	assert.Equal(t, id, res.Requeued)
	st := showRun(t, admin, r.Run.ID).Steps[0]
	assert.Equal(t, store.StatusOpen, st.Status)
	assert.Zero(t, stepFields(t, st).Failures)
	assert.Empty(t, readyIDs(t, admin), "no healthy slot")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodResourceAdd, rpc.SlotParams{Resource: "emu", Slot: "e2"}, nil))
	assert.Equal(t, []string{id}, readyIDs(t, admin))

	err := admin.Call(t.Context(), rpc.MethodResourceAdd, rpc.SlotParams{Resource: "nope", Slot: "x"}, nil)
	require.Error(t, err)
}
