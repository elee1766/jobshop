package ids

import (
	"fmt"
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

// TestShapesDisjoint: each generator's ids match only its own shape, so
// ids from different generators under one prefix never collide (JOB-57).
func TestShapesDisjoint(t *testing.T) {
	for _, name := range Names {
		g, err := Open(Spec{Generator: name})
		require.NoError(t, err)
		var id string
		if IsSeq(g) {
			id = SeqID("js", "", 42)
		} else {
			id, err = g.Next("js", "")
			require.NoError(t, err)
		}
		body := strings.TrimPrefix(id, "js-")
		for _, other := range all {
			same := fmt.Sprintf("%T", other) == fmt.Sprintf("%T", g)
			require.Equal(t, same, other.Match(body), "%s body %q vs %T", name, body, other)
		}
	}
}

func TestSortedGenerators(t *testing.T) {
	for _, name := range []string{ULID, UUIDv7} {
		g, err := Open(Spec{Generator: name})
		require.NoError(t, err)
		require.True(t, g.Sorted())
		prev := ""
		for range 5000 {
			id, err := g.Next("js", "")
			require.NoError(t, err)
			require.Greater(t, id, prev, name)
			prev = id
		}
	}
}

func TestSpecCheck(t *testing.T) {
	require.NoError(t, Spec{}.Check())
	require.NoError(t, Spec{Generator: Words, Words: 4, Separator: "_"}.Check())
	require.Error(t, Spec{Generator: "nope"}.Check())
	require.Error(t, Spec{Generator: Seq, Words: 3}.Check())
	require.Error(t, Spec{Generator: Words, Words: 1}.Check())
	require.Error(t, Spec{Generator: Words, Separator: "/"}.Check())
	require.Equal(t, int64(10), SeqStart("js", []string{"js-9", "js-s-3", "js-0031w02j0r012", "jsx-99"}))
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
