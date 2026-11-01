package ids

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNextIsFixedWidthAndOrdered(t *testing.T) {
	g, err := New()
	require.NoError(t, err)

	prev := ""
	for range 10000 {
		id, err := g.Next("js")
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(id, "js-"))
		require.Len(t, id, len("js-")+13)
		require.Greater(t, id, prev, "ids must sort by creation order as strings")
		prev = id
	}
}
