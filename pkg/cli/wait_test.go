package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
)

func TestParseWaitCond(t *testing.T) {
	now := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	cd, err := parseWaitCond("github.review:pr=https://github.com/acme/widget/pull/5,login=bob,head=abc", now)
	require.NoError(t, err)
	assert.Equal(t, domain.WaitCond{Type: domain.WaitReview, Repo: "acme/widget", PR: 5, Login: "bob", Head: "abc"}, cd)
	cd, err = parseWaitCond("github.checks:pr=acme/widget#5,state=success,required=true", now)
	require.NoError(t, err)
	assert.True(t, cd.Required)
	cd, err = parseWaitCond("github.checks:repo=acme/widget,sha=abc1234,state=success,workflow=ci", now)
	require.NoError(t, err)
	assert.Equal(t, domain.WaitCond{Type: domain.WaitChecks, Repo: "acme/widget", SHA: "abc1234", State: "success", Workflow: "ci"}, cd)
	cd, err = parseWaitCond("time:at=2h", now)
	require.NoError(t, err)
	assert.Equal(t, now.Add(2*time.Hour), cd.At)
	cd, err = parseWaitCond("thread.reply:thread=https://github.com/acme/widget/issues/7", now)
	require.NoError(t, err)
	assert.Equal(t, "acme/widget#7", cd.Thread)
	for _, bad := range []string{"time", "time:at", "github.pr:nope=1", "github.review:pr=https://github.com/a/b/issues/1"} {
		_, err := parseWaitCond(bad, now)
		assert.Error(t, err, bad)
	}
}
