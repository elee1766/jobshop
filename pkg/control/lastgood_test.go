package control_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestStartWithBadPack: a live pack using a var this binary does not set
// keeps the controller up on the last good config, listed in attention,
// until a reload succeeds (#227).
func TestStartWithBadPack(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	write := func(name, s string) {
		p := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(s), 0o600))
	}
	host := "projects:\n  js: {}\nagents:\n  a: {command: x}\npacks:\n  - path: p\n"
	write("p/pack.yaml", "pools:\n  dev: {agent: a, dir: /w, max_sessions: 0, prompt: hi}\n")
	write(config.FileName, host)
	root := config.Root{Dir: dir}

	// First start saves the last good config.
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	dial(t, root.AdminSocket())
	cancel()
	require.NoError(t, <-done)
	_, err = os.Stat(root.LastGood())
	require.NoError(t, err)

	// The pack now needs a var this binary does not set.
	write("p/pack.yaml", "pools:\n  dev: {agent: a, dir: /w, max_sessions: 0, prompt: \"hi {{.pack_newvar}}\"}\n")
	ctx, cancel = context.WithCancel(t.Context())
	done = make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })
	admin := dial(t, root.AdminSocket())

	var show rpc.ConfigShowResult
	require.NoError(t, admin.Call(ctx, rpc.MethodConfigShow, struct{}{}, &show))
	assert.Contains(t, show.Config["pools"], "dev", "serves the last good pools")
	var att rpc.AttentionResult
	require.NoError(t, admin.Call(ctx, rpc.MethodAttention, struct{}{}, &att))
	require.Len(t, att.Items, 1)
	assert.Equal(t, rpc.AttentionConfig, att.Items[0].Kind)
	assert.Contains(t, att.Items[0].Title, `var "pack_newvar" is not set`)

	// Fixed pack: a reload clears the item.
	write("p/pack.yaml", "pools:\n  dev: {agent: a, dir: /w, max_sessions: 0, prompt: hi}\n")
	require.NoError(t, admin.Call(ctx, rpc.MethodReload, struct{}{}, nil))
	require.NoError(t, admin.Call(ctx, rpc.MethodAttention, struct{}{}, &att))
	assert.Empty(t, att.Items)

	// Host config errors still fail fast.
	write(config.FileName, host+"bogus: 1\n")
	err = control.Run(t.Context(), config.Root{Dir: dir}, control.Options{Tick: time.Hour})
	require.Error(t, err)
}
