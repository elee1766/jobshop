package rpcschema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc/rpcschema"
)

// The committed client declarations must match the rpc types. Regenerate
// with: go run ./pkg/transport/rpc/rpcschema/cmd/deskgen
func TestDTSCurrent(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", rpcschema.DTSPath))
	require.NoError(t, err)
	assert.Equal(t, rpcschema.TypeScript(rpcschema.Head), string(b),
		"packages/jobshop-client/rpc.d.ts is stale: go run ./pkg/transport/rpc/rpcschema/cmd/deskgen")
}

func TestSchemaTicketCreate(t *testing.T) {
	defs := map[string]any{}
	s := rpcschema.Schema(rpc.Specs[rpc.MethodTicketCreate].Params, defs)
	ref, _ := s["$ref"].(string)
	require.Equal(t, "#/$defs/rpc.CreateParams", ref)
	b, err := json.Marshal(defs["rpc.CreateParams"])
	require.NoError(t, err)
	var d struct {
		Properties map[string]map[string]any `json:"properties"`
		Required   []string                  `json:"required"`
	}
	require.NoError(t, json.Unmarshal(b, &d))
	assert.Equal(t, "string", d.Properties["title"]["type"])
	assert.Contains(t, d.Required, "title")
	assert.NotContains(t, d.Required, "meta")
	assert.Equal(t, map[string]any{}, d.Properties["meta"], "raw JSON is any")
}

// Recursive types (TraceNode.Children) terminate through $defs.
func TestSchemaRecursive(t *testing.T) {
	defs := map[string]any{}
	rpcschema.Schema(rpc.TraceNode{}, defs)
	assert.Contains(t, defs, "rpc.TraceNode")
	_, err := json.Marshal(defs)
	require.NoError(t, err)
}
