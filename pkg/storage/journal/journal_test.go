package journal

import (
	"encoding/json"
	"testing"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/support/redact"
)

func TestMapper(t *testing.T) {
	m := &Mapper{
		Src:    Source{Project: "js", Pool: "dev", Session: "s1"},
		Redact: redact.New([]string{"GH_TOKEN=launch-secret-123"}),
		Lookup: func(string) Work { return Work{Run: "r1", Workflow: "pr-work", Step: "implement"} },
	}
	frames := []host.Frame{
		{Seq: 1, T: 1000, From: "controller", Msg: json.RawMessage(`{"method":"session/prompt","params":{"_meta":{"jobshop/ticket":"t1"}}}`)},
		{Seq: 2, T: 2000, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"tool_call","toolCallId":"c1","title":"go test","status":"pending"}}}`)},
		{Seq: 3, T: 3000, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"tool_call_update","toolCallId":"c1","status":"failed","rawOutput":"launch-secret-123"}}}`)},
	}
	m.Observe(frames[0])
	m.Observe(frames[1])
	row, err := m.Map(frames[2])
	require.NoError(t, err)
	var r map[string]any
	require.NoError(t, json.Unmarshal(row.Line, &r))
	assert.Equal(t, int64(3), row.Seq)
	assert.Equal(t, "1970-01-01T00:00:03Z", r["_time"])
	assert.Equal(t, "tool_call_update", r["update"])
	assert.Equal(t, "failed", r["status"])
	assert.Equal(t, "go test", r["tool"], "title from the earlier tool_call")
	assert.Equal(t, "t1", r["ticket"])
	assert.Equal(t, "pr-work", r["workflow"])
	assert.Equal(t, "implement", r["step"])
	assert.InDelta(t, 3, r["seq"], 0)
	assert.NotContains(t, r["_msg"], "launch-secret-123")

	// A prompt without a ticket (a message turn) clears it.
	row, err = m.Map(host.Frame{Seq: 4, T: 4000, From: "controller", Msg: json.RawMessage(`{"method":"session/prompt","params":{}}`)})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(row.Line, &r))
	assert.NotContains(t, string(row.Line), `"ticket"`)
}
