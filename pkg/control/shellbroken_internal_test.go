package control

import (
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
)

func denied() map[string]any {
	return map[string]any{"error": "Error: Permission denied (os error 13)", "output": "Error: Permission denied (os error 13)"}
}

// TestShellBroken: two shell calls in a row that jcode could not start
// break the shell (JOB-94); a call that ran, a non-shell call or a
// command's own EACCES output does not count.
func TestShellBroken(t *testing.T) {
	exec := acp.ToolKindExecute
	read := acp.ToolKindRead
	r := &runner{shellDead: make(chan struct{}, 1)}
	start := func(id acp.ToolCallId, k *acp.ToolKind) {
		assert.False(t, r.noteShellCall(id, k, acp.ToolCallStatusPending, nil, nil))
	}

	start("a", &exec)
	// The failing update omits the kind, as jcode's does.
	assert.False(t, r.noteShellCall("a", nil, acp.ToolCallStatusFailed, denied(), nil), "one denial is not enough")
	start("b", &read)
	assert.False(t, r.noteShellCall("b", nil, acp.ToolCallStatusFailed, denied(), nil), "a read call does not count")
	start("c", &exec)
	assert.False(t, r.noteShellCall("c", nil, acp.ToolCallStatusCompleted, nil, nil), "a working call resets")
	start("d", &exec)
	assert.False(t, r.noteShellCall("d", nil, acp.ToolCallStatusFailed, map[string]any{"output": "ls: cannot open directory 'x': Permission denied\nError: Permission denied (os error 13)"}, nil), "command output is not a spawn failure")
	start("e", &exec)
	assert.False(t, r.noteShellCall("e", nil, acp.ToolCallStatusFailed, denied(), nil))
	assert.False(t, r.shellBroken())
	start("f", &exec)
	text := []acp.ToolCallContent{acp.ToolContent(acp.TextBlock("Error: Permission denied (os error 13)"))}
	assert.True(t, r.noteShellCall("f", nil, acp.ToolCallStatusFailed, nil, text), "second denial in a row, from content")
	assert.True(t, r.shellBroken())
	start("g", &exec)
	assert.False(t, r.noteShellCall("g", nil, acp.ToolCallStatusFailed, denied(), nil), "reported once")

	r.resetShellCalls()
	assert.False(t, r.shellBroken(), "a new host starts clean")
}
