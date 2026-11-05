package cli

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocsText(t *testing.T) {
	out, err := docsText("packs", "abc123")
	require.NoError(t, err)
	require.Equal(t, "https://github.com/elee1766/jobshop/blob/abc123/docs/packs.md\n"+
		"https://github.com/elee1766/jobshop/tree/abc123/packs\n", out)

	out, err = docsText("", "master")
	require.NoError(t, err)
	require.Contains(t, out, "packs")
	require.Contains(t, out, "readme")
	require.Contains(t, out, "desk-plugins")

	out, err = docsText("desk-plugins", "abc123")
	require.NoError(t, err)
	require.Contains(t, out, "/blob/abc123/docs/desk-plugins.md\n")

	_, err = docsText("nope", "master")
	require.Error(t, err)
}

func TestDocsRev(t *testing.T) {
	require.Equal(t, "master", docsRev("unknown", nil))
	require.Equal(t, "deadbeef", docsRev("deadbeef", nil))
	bi := &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "0123456789abcdef"}}}
	require.Equal(t, "0123456789ab", docsRev("unknown", bi))
	bi.Settings = append(bi.Settings, debug.BuildSetting{Key: "vcs.modified", Value: "true"})
	require.Equal(t, "master", docsRev("unknown", bi))
}
