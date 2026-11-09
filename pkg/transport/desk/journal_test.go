package desk

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

func TestSessionJournalProxy(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionJournal: rpc.SessionJournalResult{NextSeq: 7, Done: true}}}
	h := New(Config{}, ctl, nil).Handler()
	var got rpc.SessionJournalResult
	get(t, h, "/api/session/s1/journal?since_seq=3&limit=10&kinds=tool_call_update,%20agent_message_chunk&from=agent", &got)
	assert.Equal(t, int64(7), got.NextSeq)
	assert.Equal(t, rpc.MethodSessionJournal+` {"id":"s1","since_seq":3,"limit":10,"kinds":["tool_call_update","agent_message_chunk"],"from":"agent"}`, ctl.calls[len(ctl.calls)-1])
	require.Equal(t, http.StatusBadRequest, get(t, h, "/api/session/s1/journal?limit=x", nil).Code)
}
