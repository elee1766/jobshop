package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestActivityRecord: a session's jobshop gh call lands in the feed with
// the session as actor and the PR number as ref.
func TestActivityRecord(t *testing.T) {
	t.Parallel()
	c, _, _ := activityController(t)
	sc := scope{session: "js-s1", pool: "dev"}
	_, err := c.activityRecord(t.Context(), sc, rpc.ActivityRecordParams{
		Action: "merge", Repo: "o/n", Target: "o/n#12", Decision: domain.GuardDeny, ExitCode: 1, Command: "gh pr merge 12",
	})
	require.NoError(t, err)
	res := listActivity(t, c, rpc.ActivityListParams{Kind: "gh-merge"})
	require.Len(t, res.Items, 1)
	it := res.Items[0]
	assert.Equal(t, "js-s1", it.Actor)
	assert.Equal(t, "12", it.Ref)
	assert.Equal(t, "o/n", it.Repo)
	assert.Contains(t, it.Title, "deny")
	assert.Equal(t, "gh pr merge 12", it.Body)

	_, err = c.activityRecord(t.Context(), sc, rpc.ActivityRecordParams{Action: "merge"})
	require.ErrorIs(t, err, domain.ErrInvalid)
}
