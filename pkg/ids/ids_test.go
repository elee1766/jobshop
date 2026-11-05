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
		id, err := g.Next("js", "")
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(id, "js-"))
		require.Len(t, id, len("js-")+13)
		require.Greater(t, id, prev, "ids must sort by creation order as strings")
		prev = id
	}
}

func TestIsSession(t *testing.T) {
	g, err := New()
	require.NoError(t, err)
	sess, err := g.Next("js", SessionTag)
	require.NoError(t, err)
	require.Regexp(t, `^js-s-[0-9a-z]{13}$`, sess)
	tick, err := g.Next("js", "")
	require.NoError(t, err)

	require.True(t, IsSession(sess))
	require.True(t, IsSession("my-proj-s-00337y40kc002"))
	require.False(t, IsSession(tick))
	require.False(t, IsSession("js-003371shfc002"), "pre-#615 session ids have no tag")
	require.False(t, IsSession("js-s-short"))
	require.False(t, IsSession(""))
}
