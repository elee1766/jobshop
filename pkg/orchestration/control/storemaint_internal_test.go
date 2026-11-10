package control

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/support/clock"
)

// maintainCounter counts Maintain calls.
type maintainCounter struct {
	store.Store
	n int
}

func (m *maintainCounter) Maintain(context.Context) error { m.n++; return nil }

// TestMaintainStoreInterval: the first tick leaves the statistics Open
// built alone, then the step runs once per interval (JOB-191).
func TestMaintainStoreInterval(t *testing.T) {
	t.Parallel()
	clk := clock.NewFake(time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC))
	st := &maintainCounter{}
	c := &Controller{store: st, opt: Options{Clock: clk}, log: slog.New(slog.DiscardHandler)}
	ctx := t.Context()

	c.maintainStore(ctx, nil)
	assert.Zero(t, st.n, "Open just did it")
	clk.Advance(storeMaintainEvery - time.Minute)
	c.maintainStore(ctx, nil)
	assert.Zero(t, st.n, "not yet")
	clk.Advance(time.Minute)
	c.maintainStore(ctx, nil)
	assert.Equal(t, 1, st.n, "due")
	c.maintainStore(ctx, nil)
	assert.Equal(t, 1, st.n, "once per interval")
}
