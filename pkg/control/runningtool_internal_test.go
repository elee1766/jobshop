package control

import (
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
)

// TestRunningTool: the latest open tool call names the running tool; a
// completed or failed call drops out, and an update without a title
// keeps the one it had (JOB-54).
func TestRunningTool(t *testing.T) {
	r := &runner{}
	assert.Empty(t, r.runningTool())
	r.trackTool("a", "Read file", acp.ToolCallStatusPending)
	r.trackTool("b", "go test", acp.ToolCallStatusInProgress)
	assert.Equal(t, "go test", r.runningTool())
	r.trackTool("b", "", acp.ToolCallStatusInProgress)
	assert.Equal(t, "go test", r.runningTool())
	r.trackTool("b", "", acp.ToolCallStatusCompleted)
	assert.Equal(t, "Read file", r.runningTool())
	r.trackTool("a", "", acp.ToolCallStatusFailed)
	assert.Empty(t, r.runningTool())
}
