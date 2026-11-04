package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvalWhen(t *testing.T) {
	for s, want := range map[string]bool{
		"high != 'high'": false,
		"low != 'high'":  true,
		`a == "a"`:       true,
		" a == b ":       false,
	} {
		got, err := EvalWhen(s)
		require.NoError(t, err, s)
		assert.Equal(t, want, got, s)
	}
	_, err := EvalWhen("a = b")
	require.Error(t, err)
}

func TestPlacementChecks(t *testing.T) {
	c := &Config{
		Resources: map[string]Resource{"pr": {Keyed: true}, "big": {Max: 1}},
		Pools:     map[string]Pool{"a": {Project: "p"}, "b": {Project: "p"}, "x": {Project: "q"}},
	}
	w := &Workflow{Vars: []string{"pr", "risk"}}
	opt := func(pool string) PlacementOption { return PlacementOption{Pool: pool, Weight: 1} }
	for _, tt := range []struct {
		name string
		s    Step
		want string
	}{
		{"ok", Step{Placement: &Placement{Required: []string{"pr:${pr}"}, Options: []PlacementOption{opt("a"), {Pool: "b", When: "${risk} != high"}}}}, ""},
		{"with pool", Step{Pool: "a", Placement: &Placement{Options: []PlacementOption{opt("a")}}}, "set one or the other"},
		{"no options", Step{Placement: &Placement{}}, "at least one option"},
		{"bad prefer", Step{Placement: &Placement{PreferFor: "soon", Options: []PlacementOption{opt("a")}}}, "prefer_for"},
		{"bad pool", Step{Placement: &Placement{Options: []PlacementOption{opt("nope")}}}, "unknown pool"},
		{"bad when", Step{Placement: &Placement{Options: []PlacementOption{{Pool: "a", When: "${risk}"}}}}, "A == B"},
		{"unknown var", Step{Placement: &Placement{Options: []PlacementOption{{Pool: "a", When: "${z} == 1"}}}}, "not in workflow vars"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := c.checkPlacement(w, &tt.s)
			if tt.want == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}

	p := &Placement{Required: []string{"pr:${pr}"}, Options: []PlacementOption{
		{Pool: "b", Weight: 10, When: "${risk} != high"},
		{Pool: "a", Resources: []string{"big"}, Weight: 100},
	}}
	got, err := c.ResolvePlacement(w, p, map[string]any{"pr": "7", "risk": "low"})
	require.NoError(t, err)
	assert.Equal(t, []ResolvedOption{
		{Pool: "a", Resources: []string{"pr:7", "big"}, Weight: 100},
		{Pool: "b", Resources: []string{"pr:7"}, Weight: 10},
	}, got, "sorted by weight, required keys first")
	got, err = c.ResolvePlacement(w, p, map[string]any{"pr": "7", "risk": "high"})
	require.NoError(t, err)
	assert.Len(t, got, 1)
	_, err = c.ResolvePlacement(w, &Placement{Options: []PlacementOption{{Pool: "a"}, {Pool: "x"}}}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project")
}
