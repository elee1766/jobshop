package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPoolLabels (JOB-81): a pool's labels load, a host adds to a pack
// pool's instead of replacing them, and a packs entry gives its pools
// instance:<name> and pack:<pack> from its prefix.
func TestPoolLabels(t *testing.T) {
	c := loadFixture(t, "packs:\n  - path: gh\n    vars: "+ghVars+"\n  - path: gh\n    prefix: gz-\n    vars: {repo: acme/gizmo, checkout: /src/gizmo, agent: a, tracker: github}\n"+
		"pools:\n  gz-dev: {labels: [review, sonnet]}\n  own: {agent: a, dir: /tmp, labels: [solo]}\n")
	assert.Equal(t, []string{"instance:gh", "pack:gh"}, c.Pools["dev"].Labels, "no prefix: instance is the pack")
	assert.Equal(t, []string{"instance:gz", "pack:gh", "review", "sonnet"}, c.Pools["gz-dev"].Labels, "host adds to the pack's")
	assert.Equal(t, []string{"instance:gz", "pack:gh"}, c.Pools["gz-plan"].Labels)
	assert.Equal(t, []string{"solo"}, c.Pools["own"].Labels, "a host pool is in no instance")
}

// TestPoolLabelsInvalid: labels must be strings with no whitespace or
// comma, and are not silently coerced.
func TestPoolLabelsInvalid(t *testing.T) {
	for name, tc := range map[string]struct{ yaml, want string }{
		"not a string": {"[1]", "labels"},
		"a map":        {"[{a: b}]", "labels"},
		"empty":        {`[""]`, "must not be empty"},
		"space":        {`["a b"]`, "whitespace"},
		"comma":        {`["a,b"]`, "comma"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "pools:\n  p: {agent: a, dir: /tmp, labels: " + tc.yaml + "}\n"})
			_, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
			require.ErrorContains(t, err, tc.want)
		})
	}
}
