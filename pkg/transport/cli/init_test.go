package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	packpkg "github.com/gfx-labs/jobshop/pkg/pack"
)

func TestInitRootPack(t *testing.T) {
	dir := t.TempDir()
	cmd := initCmd{Dir: dir, Project: "work"}
	require.NoError(t, cmd.Run())
	c, err := config.Load(config.Root{Dir: dir})
	require.NoError(t, err)
	require.Equal(t, "pack.yaml", c.Pack)
	require.Equal(t, "work", c.DefaultProject)
	require.Contains(t, c.Projects, "work")
	require.ErrorIs(t, cmd.Run(), os.ErrExist)

	// A pre-existing pack must survive, without a dangling new service file.
	dir = t.TempDir()
	pack := filepath.Join(dir, packpkg.FileName)
	require.NoError(t, os.WriteFile(pack, []byte("existing work\n"), 0o600))
	require.ErrorIs(t, (initCmd{Dir: dir, Project: "work"}).Run(), os.ErrExist)
	b, err := os.ReadFile(pack)
	require.NoError(t, err)
	require.Equal(t, "existing work\n", string(b))
	_, err = os.Stat((config.Root{Dir: dir}).Config())
	require.ErrorIs(t, err, os.ErrNotExist)
}
