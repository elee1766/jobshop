package linearsvc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/linear"
)

func TestIntakeMatch(t *testing.T) {
	t.Parallel()
	is := linear.Summary{State: linear.State{Name: "Todo", Type: "unstarted"}, Labels: []string{"Agent", "x"}}
	assert.True(t, intakeMatch(is, nil, nil, nil))
	assert.True(t, intakeMatch(is, []string{"todo"}, []string{"agent"}, nil))
	assert.False(t, intakeMatch(is, []string{"Backlog"}, nil, nil))
	assert.False(t, intakeMatch(is, nil, []string{"bug"}, nil))
	assert.False(t, intakeMatch(is, nil, nil, []string{"X"}))
	is.State.Type = "canceled"
	assert.False(t, intakeMatch(is, nil, nil, nil))

	// JOB-4: target and proposed parents never start a run.
	skip := []string{"Blocked", "target", "proposed"}
	for _, l := range []string{"target", "Proposed"} {
		p := linear.Summary{State: linear.State{Name: "Todo", Type: "unstarted"}, Labels: []string{"repo", l}}
		assert.False(t, intakeMatch(p, []string{"Todo"}, nil, skip), l)
	}
	sub := linear.Summary{State: linear.State{Name: "Todo", Type: "unstarted"}, Labels: []string{"repo"}}
	assert.True(t, intakeMatch(sub, []string{"Todo"}, nil, skip), "a target's sub-issue is taken")
}
