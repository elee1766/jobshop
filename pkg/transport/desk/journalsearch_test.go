package desk

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func TestJournalSearchProxy(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodJournalSearch: rpc.JournalSearchResult{Truncated: true}}}
	h := New(Config{}, ctl, nil).Handler()
	var got rpc.JournalSearchResult
	get(t, h, "/api/journal/search?q=status:failed&since=2h&limit=5", &got)
	assert.True(t, got.Truncated)
	assert.Equal(t, rpc.MethodJournalSearch+` {"query":"status:failed","since":7200000000000,"limit":5}`, ctl.calls[len(ctl.calls)-1])
	require.Equal(t, http.StatusBadRequest, get(t, h, "/api/journal/search?q=x&since=soon", nil).Code)
}
