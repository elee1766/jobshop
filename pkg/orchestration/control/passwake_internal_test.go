package control

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gfx-labs/jobshop/pkg/orchestration/hooks"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// TestPassWakeCoalesces (JOB-193): invocations applied while a pass runs
// leave one wake, so the serve loop runs one further pass, not one each.
func TestPassWakeCoalesces(t *testing.T) {
	t.Parallel()
	c, _, _ := offTickController(t, offTickCommandYAML, io.Discard)
	// The pass holds tickMu; the serve loop is inside it, not receiving.
	unlock := c.lockTick(t.Context(), "tick")
	applied := hooks.Invocation{Run: "no-such-run", Plan: &store.Graph{Tickets: []store.NewTicket{{ID: "a"}}}}
	for range 5 {
		hookHost{c}.HookApplied(t.Context(), applied)
	}
	c.wakePass()
	unlock()
	assert.Len(t, c.offWake, 1, "five applied invocations and a direct wake leave one")

	// The loop takes it and runs one pass; nothing else is queued.
	<-c.offWake
	assert.Empty(t, c.offWake)
}

// TestPassWakeOnlyForNewSteps (JOB-193): an applied invocation that
// created no ticket (a comment, a finish) asks for no pass.
func TestPassWakeOnlyForNewSteps(t *testing.T) {
	t.Parallel()
	c, _, _ := offTickController(t, offTickCommandYAML, io.Discard)
	hookHost{c}.HookApplied(t.Context(), hooks.Invocation{})
	hookHost{c}.HookApplied(t.Context(), hooks.Invocation{Run: "no-such-run", Plan: &store.Graph{}})
	assert.Empty(t, c.offWake)
	hookHost{c}.HookApplied(t.Context(), hooks.Invocation{Plan: &store.Graph{Tickets: []store.NewTicket{{ID: "a"}}}})
	assert.Len(t, c.offWake, 1)
}

// TestPassWakeWithoutChannel (JOB-193): a controller built by hand has no
// wake channel and ignores a wake.
func TestPassWakeWithoutChannel(t *testing.T) {
	t.Parallel()
	c := &Controller{}
	c.wakePass()
}
