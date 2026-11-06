package control

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/config"
)

// TestPoolHasLabels (JOB-81): every wanted label must be on the pool, and
// k=v matches the label k:v.
func TestPoolHasLabels(t *testing.T) {
	t.Parallel()
	p := config.Pool{Labels: []string{"instance:sap", "pack:github", "review"}}
	for _, tc := range []struct {
		want []string
		ok   bool
	}{
		{[]string{"review"}, true},
		{[]string{"instance=sap"}, true},
		{[]string{"instance:sap"}, true},
		{[]string{"instance=sap", "review"}, true},
		{[]string{"instance=other"}, false},
		{[]string{"review", "sonnet"}, false},
		{[]string{"inst"}, false},
	} {
		assert.Equal(t, tc.ok, poolHasLabels(p, tc.want), "%v", tc.want)
	}
	assert.False(t, poolHasLabels(config.Pool{}, []string{"review"}))
}
