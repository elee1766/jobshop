package rpc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/telemetry"
)

func TestMetaRoundTrip(t *testing.T) {
	const tp = "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"
	ctx := telemetry.WithTraceParent(context.Background(), tp)

	p, err := withMeta(ctx, CloseParams{ID: "js-1"})
	require.NoError(t, err)
	raw, err := json.Marshal(p)
	require.NoError(t, err)
	rest, meta := SplitMeta(raw)
	assert.Equal(t, tp, meta.TraceParent)
	assert.JSONEq(t, `{"id":"js-1"}`, string(rest))

	// No params still carries metadata; no trace leaves params alone.
	p, err = withMeta(ctx, nil)
	require.NoError(t, err)
	raw, err = json.Marshal(p)
	require.NoError(t, err)
	_, meta = SplitMeta(raw)
	assert.Equal(t, tp, meta.TraceParent)
	p, err = withMeta(context.Background(), CloseParams{ID: "x"})
	require.NoError(t, err)
	assert.Equal(t, CloseParams{ID: "x"}, p)
}
