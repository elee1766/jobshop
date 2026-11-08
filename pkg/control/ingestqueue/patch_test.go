package ingestqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyJSONPatch(t *testing.T) {
	t.Parallel()
	out, err := applyJSONPatch([]byte(`{"a":[1,2],"b":{"c":"x"}}`), []byte(`[
		{"op":"add","path":"/a/1","value":9},
		{"op":"remove","path":"/b/c"},
		{"op":"copy","from":"/a/0","path":"/b/d"},
		{"op":"move","from":"/a/2","path":"/e"},
		{"op":"test","path":"/e","value":2}
	]`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"a":[1,9],"b":{"d":1},"e":2}`, string(out))
	_, err = applyJSONPatch([]byte(`{"a":1}`), []byte(`[{"op":"test","path":"/a","value":2}]`))
	require.Error(t, err)
	_, err = applyJSONPatch([]byte(`{"a":1}`), []byte(`[{"op":"replace","path":"/nope","value":2}]`))
	require.Error(t, err)
}
