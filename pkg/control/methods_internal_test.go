package control

import (
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// rpc.Methods is what Desk's generic RPC endpoint accepts (#517); it must
// list exactly the methods the controller serves.
func TestRPCMethodsMatchHandlers(t *testing.T) {
	want := slices.Sorted(maps.Keys(methods))
	assert.Equal(t, want, rpc.Methods())
}

// rpc.Specs feeds the plugin schemas and desk-plugin.d.ts (#542); each
// entry's params must be the type the handler decodes.
func TestRPCSpecsMatchHandlerParams(t *testing.T) {
	for name, m := range methods {
		spec, ok := rpc.Specs[name]
		if !assert.True(t, ok, name) {
			continue
		}
		got := reflect.TypeOf(spec.Params)
		if m.params.Kind() == reflect.Struct && m.params.NumField() == 0 {
			assert.Equal(t, 0, got.NumField(), "%s: handler takes no params", name)
			continue
		}
		assert.Equal(t, m.params, got, name)
	}
}
