package control

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

func TestHookAcceptedManifestBindsAuthority(t *testing.T) {
	t.Parallel()
	changes := map[string]func(*testing.T, config.Root, *config.Config){
		"selector": func(_ *testing.T, _ config.Root, cfg *config.Config) {
			w := cfg.Workflows["demo-dynamic"]
			w.OnStart = strings.TrimSuffix(w.OnStart, ":begin") + ":alternate"
			cfg.Workflows["demo-dynamic"] = w
		},
		"owner": func(_ *testing.T, _ config.Root, cfg *config.Config) {
			cfg.Workflows["other"] = cfg.Workflows["demo-dynamic"]
			delete(cfg.Workflows, "demo-dynamic")
		},
		"instance": func(_ *testing.T, _ config.Root, cfg *config.Config) {
			cfg.Instances[0].Pools = append(cfg.Instances[0].Pools, "sibling-dev")
		},
		"grant": func(_ *testing.T, _ config.Root, cfg *config.Config) {
			cfg.Automation.Shell = true
		},
		"credential": func(t *testing.T, root config.Root, _ *config.Config) {
			require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "credential"), []byte("rotated"), 0o600))
		},
		"source": func(t *testing.T, root config.Root, _ *config.Config) {
			path := filepath.Join(root.Dir, "pack", "hooks.star")
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path, append(data, []byte("\n# changed\n")...), 0o600))
		},
		"asset_root": func(t *testing.T, root config.Root, _ *config.Config) {
			path := filepath.Join(root.Dir, "pack")
			require.NoError(t, os.Rename(path, path+"-original"))
			require.NoError(t, os.Mkdir(path+"-replacement", 0o700))
			data, err := os.ReadFile(filepath.Join(path+"-original", "hooks.star"))
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(path+"-replacement", "hooks.star"), data, 0o600))
			require.NoError(t, os.Symlink(path+"-replacement", path))
		},
		"root": nil,
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			root := config.Root{Dir: t.TempDir()}
			require.NoError(t, os.Mkdir(filepath.Join(root.Dir, "pack"), 0o700))
			host := `projects: {js: {}}
agents: {fake: {command: "false"}}
packs: [{path: pack, prefix: demo-}]
automation:
  credentials:
    api:
      origins: [https://example.com]
      header: X-API-Key
      secret: credential
`
			pack := `pools:
  dev: {project: js, agent: fake, max_sessions: 0, dir: .}
workflows:
  dynamic:
    project: js
    on_start: hooks.star:begin
    on_step_completed: hooks.star:completed
`
			script := "def begin(run):\n    return []\ndef alternate(run):\n    return []\ndef completed(run, step):\n    return []\n"
			require.NoError(t, os.WriteFile(root.Config(), []byte(host), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pack", "pack.yaml"), []byte(pack), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "pack", "hooks.star"), []byte(script), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "credential"), []byte("accepted"), 0o600))
			cfg, err := config.Load(root)
			require.NoError(t, err)
			accepted, err := newLiveConfig(cfg, root)
			require.NoError(t, err)
			require.NoError(t, config.SaveLastGood(root.LastGood(), accepted.cfg))
			saved, err := config.LoadLastGood(root.LastGood())
			require.NoError(t, err)
			_, err = newLiveConfig(saved, root)
			require.NoError(t, err, "unchanged accepted metadata must survive the real envelope")
			if change != nil {
				change(t, root, saved)
			}
			if name == "root" {
				newRoot := config.Root{Dir: t.TempDir()}
				require.NoError(t, os.WriteFile(filepath.Join(newRoot.Dir, "credential"), []byte("accepted"), 0o600))
				root = newRoot
			}
			_, err = newLiveConfig(saved, root)
			require.ErrorContains(t, err, "accepted hook manifest changed")
		})
	}
}
