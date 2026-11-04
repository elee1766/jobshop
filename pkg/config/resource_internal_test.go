package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceChecks(t *testing.T) {
	for _, tt := range []struct {
		name string
		r    Resource
		want string
	}{
		{"counted", Resource{Max: 2}, ""},
		{"keyed default", Resource{Keyed: true}, ""},
		{"no max", Resource{}, "max must be >= 1"},
		{"keyed with max", Resource{Keyed: true, Max: 1}, "not max"},
		{"per key without keyed", Resource{Max: 1, CapacityPerKey: 2}, "needs keyed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Resources: map[string]Resource{"r": tt.r}}
			err := c.checkResources()
			if tt.want == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
	c := &Config{Resources: map[string]Resource{"pr": {Keyed: true, CapacityPerKey: 2}, "emu": {Max: 3}}}
	assert.Equal(t, map[string]int{"pr": 2, "emu": 3}, c.ResourceCaps())
	require.NoError(t, c.CheckResourceKey("pr:a#1"))
	require.NoError(t, c.CheckResourceKey("emu"))
	for key, want := range map[string]string{
		"pr":     "is keyed",
		"pr:":    "is keyed",
		"emu:x":  "not keyed",
		"nope":   "unknown resource",
		"nope:x": "unknown resource",
	} {
		err := c.CheckResourceKey(key)
		require.Error(t, err, key)
		assert.Contains(t, err.Error(), want, key)
	}
}
