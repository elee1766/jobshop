package cli

import (
	"runtime/debug"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionInfo(t *testing.T) {
	const rev = "0123456789abcdef0123456789abcdef01234567"
	vcs := func(rev, modified string) []debug.BuildSetting {
		return []debug.BuildSetting{{Key: "vcs.revision", Value: rev}, {Key: "vcs.modified", Value: modified}}
	}
	tests := []struct {
		name                 string
		version, commit      string
		bi                   *debug.BuildInfo
		wantVersion, wantCmt string
	}{
		{"no build info", "dev", "unknown", nil, "dev", "unknown"},
		{
			"devel build, clean", "dev", "unknown",
			&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}, Settings: vcs(rev, "false")},
			"dev", "0123456789ab",
		},
		{
			"dirty tree", "dev", "unknown",
			&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}, Settings: vcs(rev, "true")},
			"dev", "0123456789ab-dirty",
		},
		{
			"go install with module version", "dev", "unknown",
			&debug.BuildInfo{Main: debug.Module{Version: "v0.3.0"}},
			"v0.3.0", "unknown",
		},
		{
			"short revision", "dev", "unknown",
			&debug.BuildInfo{Settings: vcs("abc", "false")},
			"dev", "abc",
		},
		{
			"ldflags win", "1.2.3", "deadbeef",
			&debug.BuildInfo{Main: debug.Module{Version: "v0.3.0"}, Settings: vcs(rev, "true")},
			"1.2.3", "deadbeef",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, c := versionInfo(tt.version, tt.commit, tt.bi)
			assert.Equal(t, tt.wantVersion, v)
			assert.Equal(t, tt.wantCmt, c)
		})
	}
}

// --orphaned parses and reaches the ticket.list params.
func TestTicketListOrphaned(t *testing.T) {
	var r root
	p, err := kong.New(&r, kong.Exit(func(int) { t.Fatal("exit") }))
	require.NoError(t, err)
	_, err = p.Parse([]string{"ticket", "list", "--orphaned", "--project", "js"})
	require.NoError(t, err)
	got := r.Ticket.List.params()
	assert.True(t, got.Orphaned)
	assert.Equal(t, "js", got.Project)

	r = root{}
	_, err = p.Parse([]string{"ticket", "list"})
	require.NoError(t, err)
	assert.False(t, r.Ticket.List.params().Orphaned)
}
