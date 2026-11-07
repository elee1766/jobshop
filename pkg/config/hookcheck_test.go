package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigHookValidation(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		FileName: `pack: pack.cue
automation:
  http: true
  shell: true
  credentials:
    private:
      origins: [https://example.com]
      header: Authorization
      secret: missing-token
jev:
  enabled: true
  endpoint: https://example.com/decide
  api_key: missing-key
`,
		CUEPackFileName: `package example
projects: js: {}
workflows: example: {
    on_start: "hooks.star:begin"
    on_step_completed: "hooks.star:completed"
}
`,
		"helper.star": "message = \"ready\"\n",
	})
	valid := `load("helper.star", "message")
def begin(run):
    return []
def completed(run, step):
    return []
`
	write := func(body string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "hooks.star"), []byte(body), 0o600))
	}
	write(valid)
	cfg, err := LoadFile(filepath.Join(dir, FileName))
	require.NoError(t, err, "offline checking must not read missing credential files")
	require.Contains(t, cfg.Watch.Files, filepath.Join(dir, "helper.star"), "imported helpers must participate in reloads")
	sentinel := filepath.Join(dir, "should-not-exist")
	for _, tc := range []struct{ body, want string }{
		{"def begin(:\n", "hooks.star"},
		{"def other(run):\n    return []\n", "exported function not found"},
		{"def begin(run, extra):\n    return []\ndef completed(run, step):\n    return []\n", "begin"},
		{"def begin(run):\n    return []\ndef completed(run):\n    return []\n", "completed"},
		{fmt.Sprintf("shell([\"/bin/touch\", %q])\n", sentinel) + valid, "unavailable during module loading"},
	} {
		write(tc.body)
		_, err := LoadFile(filepath.Join(dir, FileName))
		require.ErrorIs(t, err, ErrPack)
		require.ErrorContains(t, err, tc.want)
	}
	_, err = os.Stat(sentinel)
	require.ErrorIs(t, err, os.ErrNotExist)
}
