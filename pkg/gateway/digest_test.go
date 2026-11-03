package gateway

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestDigest: nothing is posted while the attention list is empty; a
// non-empty list opens one thread in the escalation room (#185).
func TestDigest(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{}
	g := New(Config{Name: "ops", Pool: "chat", EscalationRoom: "!esc"}, ad, ctl, nil)

	g.digest(t.Context())
	assert.Equal(t, 0, ad.opened, "healthy: no digest")

	ctl.attention = []rpc.AttentionItem{
		{Kind: rpc.AttentionTicket, ID: "js-1", Title: "Escalate run js-r", Since: time.Now().Add(-3 * time.Hour), Repo: "acme/widget", Ref: "#7"},
		{Kind: rpc.AttentionRollback, ID: "abc123", Title: "self-deploy rolled back abc123", Since: time.Now().Add(-time.Minute)},
	}
	g.digest(t.Context())
	require.Equal(t, 1, ad.opened)
	p := ad.posts[0]
	assert.Contains(t, p, "!esc/")
	assert.Contains(t, p, "2 waiting on a human")
	assert.Contains(t, p, "- ticket js-1: Escalate run js-r (acme/widget #7), 3h")
	assert.Contains(t, p, "- rollback abc123: self-deploy rolled back abc123, 1m")
}
