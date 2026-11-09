package control

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClipSummary keeps the header and the newest text within the bound.
func TestClipSummary(t *testing.T) {
	assert.Equal(t, "short", clipSummary("short", 100))
	s := "Header\n" + strings.Repeat("old ", 100) + "newest"
	got := clipSummary(s, 80)
	assert.LessOrEqual(t, len(got), 80)
	assert.True(t, strings.HasPrefix(got, "Header\n"))
	assert.True(t, strings.HasSuffix(got, "newest"))
}
