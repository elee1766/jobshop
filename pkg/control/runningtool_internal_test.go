package control

import (
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
)

// TestRunningTool: the latest open tool call names the running tool; a
// completed or failed call drops out, and an update without a title or
// input keeps what it had (JOB-54).
func TestRunningTool(t *testing.T) {
	r := &runner{}
	assert.Zero(t, r.runningTool())
	r.trackTool("a", openTool{title: "Read file"}, acp.ToolCallStatusPending)
	r.trackTool("b", openTool{title: "Running shell command"}, acp.ToolCallStatusPending)
	r.trackTool("b", openTool{target: "go test ./...", intent: "Run tests"}, acp.ToolCallStatusInProgress)
	assert.Equal(t, openTool{"Running shell command", "go test ./...", "Run tests"}, r.runningTool())
	r.trackTool("b", openTool{}, acp.ToolCallStatusCompleted)
	assert.Equal(t, "Read file", r.runningTool().title)
	r.trackTool("a", openTool{}, acp.ToolCallStatusFailed)
	assert.Zero(t, r.runningTool())
}
