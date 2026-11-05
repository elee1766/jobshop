package control

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// rpc.Methods is what Desk's generic RPC endpoint accepts (#517); it must
// list exactly the methods the controller serves.
func TestRPCMethodsMatchHandlers(t *testing.T) {
	want := slices.Sorted(maps.Keys(methods))
	assert.Equal(t, want, slices.Sorted(slices.Values(rpc.Methods())))
}
