package session

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSlots struct {
	used map[int]bool
	err  error
}

func (f fakeSlots) UsedSlots(context.Context, string) (map[int]bool, error) { return f.used, f.err }

// TestAllocSlot: the lowest free slot wins, a full range reports none,
// and a source error passes through.
func TestAllocSlot(t *testing.T) {
	ctx := t.Context()
	s, ok, err := AllocSlot(ctx, fakeSlots{used: map[int]bool{0: true, 2: true}}, "dev", 4)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1, s)

	_, ok, err = AllocSlot(ctx, fakeSlots{used: map[int]bool{0: true, 1: true}}, "dev", 2)
	require.NoError(t, err)
	assert.False(t, ok)

	boom := errors.New("boom")
	_, _, err = AllocSlot(ctx, fakeSlots{err: boom}, "dev", 2)
	require.ErrorIs(t, err, boom)
}

func TestFreeSlots(t *testing.T) {
	n, err := FreeSlots(t.Context(), fakeSlots{used: map[int]bool{1: true}}, "dev", 3)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}
