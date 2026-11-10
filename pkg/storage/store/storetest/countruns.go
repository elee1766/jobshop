package storetest

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// testCountRuns checks CountRuns counts distinct parents of steps in the
// window and pools, and nothing else (JOB-191).
func testCountRuns(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	start := c.Now()
	// run -> steps: pool and step count differ, so counting rows or
	// pools instead of distinct runs gives another number.
	mk := func(name string, steps ...string) {
		t.Helper()
		run, err := store.Create(ctx, s, "test", store.NewTicket{
			Project: proj, Kind: domain.KindRun, Title: name,
			Fields: json.RawMessage(fmt.Sprintf(`{"workflow":"w","key":%q,"state":"running"}`, name)),
		})
		require.NoError(t, err)
		for i, pool := range steps {
			_, err := store.Create(ctx, s, "test", store.NewTicket{
				Project: proj, Kind: domain.KindStep, Pool: pool, ParentID: run.ID,
				Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":"s%d","index":%d}`, run.ID, i, i)),
			})
			require.NoError(t, err)
			c.Advance(time.Minute)
		}
	}
	mk("r1", "dev", "dev", "review") // two dev steps: one run
	mk("r2", "review")
	mid := c.Now()
	mk("r3", "dev")
	mk("r4") // a run with no steps
	end := c.Now()
	// Not a step: a task with a pool is not counted.
	_, err := store.Create(ctx, s, "test", store.NewTicket{Project: proj, Kind: domain.KindTask, Title: "t", Pool: "dev"})
	require.NoError(t, err)

	count := func(pools []string, r store.TimeRange) int {
		t.Helper()
		n, err := s.CountRuns(ctx, pools, r)
		require.NoError(t, err)
		return n
	}
	all := store.TimeRange{Since: start, Until: end.Add(time.Hour)}
	assert.Equal(t, 3, count(nil, all), "r1, r2 and r3 own steps")
	assert.Equal(t, 2, count([]string{"dev"}, all), "r1 and r3, once each")
	assert.Equal(t, 2, count([]string{"review"}, all))
	assert.Equal(t, 3, count([]string{"dev", "review"}, all))
	assert.Equal(t, 0, count([]string{"other"}, all))
	assert.Equal(t, 2, count(nil, store.TimeRange{Since: start, Until: mid}), "Until is exclusive")
	assert.Equal(t, 1, count(nil, store.TimeRange{Since: mid, Until: end}))
	assert.Equal(t, 3, count(nil, store.TimeRange{}), "an open range")
}
