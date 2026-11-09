package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Each check touches its own file and waits for the other's: that only
// finishes when both run at once. Run one after another, each times out.
const parallelTriggerYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  plan: {agent: fake, dir: %[1]q}
triggers:
  a:
    when: {check: "touch %[1]s/a; for i in $(seq 100); do [ -e %[1]s/b ] && exit 1; sleep 0.05; done; exit 2", check_timeout: 10s}
    ticket: {title: a, pool: plan}
  b:
    when: {check: "touch %[1]s/b; for i in $(seq 100); do [ -e %[1]s/a ] && exit 1; sleep 0.05; done; exit 2", check_timeout: 10s}
    ticket: {title: b, pool: plan}
`

// TestTriggerListParallel: trigger.list runs trigger checks at the same
// time and keeps the result sorted by name (#522).
func TestTriggerListParallel(t *testing.T) {
	t.Parallel()
	c, _, _ := scanControllerYAML(t, parallelTriggerYAML)
	res, err := c.triggerList(t.Context(), scope{admin: true}, struct{}{})
	require.NoError(t, err)
	infos := res.([]rpc.TriggerInfo)
	require.Len(t, infos, 2)
	assert.Equal(t, "a", infos[0].Name)
	assert.Equal(t, "b", infos[1].Name)
	for _, in := range infos {
		assert.Equal(t, "check exited 1", in.Why, in.Name+": saw the other check running")
	}
}
