package cue_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/pack/cue"
)

func TestCheckSchema(t *testing.T) {
	t.Parallel()
	schema := map[string]any{"type": "object", "properties": map[string]any{"n": map[string]any{"type": "integer"}}, "required": []any{"n"}}
	require.NoError(t, cue.CheckSchema(schema, map[string]any{"n": 1}, true))
	require.ErrorContains(t, cue.CheckSchema(schema, map[string]any{"n": "x"}, true), "step output")
	require.Error(t, cue.CheckSchema(schema, map[string]any{}, true), "required key missing")
	require.NoError(t, cue.CheckSchema(nil, map[string]any{"anything": true}, true), "no schema accepts any output")
}

func TestCheckSchemaCompileOnly(t *testing.T) {
	t.Parallel()
	require.NoError(t, cue.CheckSchema(map[string]any{"type": "object"}, nil, false))
	require.ErrorContains(t, cue.CheckSchema(map[string]any{"type": "nonsense"}, nil, false), "output_schema")
	// Compile-only ignores the output.
	require.NoError(t, cue.CheckSchema(map[string]any{"type": "object"}, map[string]any{"x": 1}, false))
}
