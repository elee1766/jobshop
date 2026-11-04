package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecretYAML: a path string, a file mapping and a command mapping all
// parse; "!" is not special (#435); bad mappings fail.
func TestSecretYAML(t *testing.T) {
	var v struct {
		S Secret `yaml:"s"`
	}
	for _, tc := range []struct {
		in   string
		want Secret
	}{
		{`s: /run/secrets/x`, Secret{Type: SecretFile, Path: "/run/secrets/x"}},
		{`s: "!pass show x"`, Secret{Type: SecretFile, Path: "!pass show x"}},
		{`s: {type: file, path: rel/x}`, Secret{Type: SecretFile, Path: "rel/x"}},
		{"s:\n  type: command\n  command: \"pass show gh\"\n", Secret{Type: SecretCommand, Command: "pass show gh"}},
		{`s: ""`, Secret{}},
	} {
		v.S = Secret{}
		require.NoError(t, yaml.UnmarshalWithOptions([]byte(tc.in), &v, yaml.Strict()), tc.in)
		assert.Equal(t, tc.want, v.S, tc.in)
	}
	for in, msg := range map[string]string{
		`s: {type: command}`:                     "needs command",
		`s: {type: file, command: x}`:            "needs path",
		`s: {type: shell, command: x}`:           `type "shell"`,
		`s: {type: command, command: x, pth: y}`: "pth",
	} {
		err := yaml.UnmarshalWithOptions([]byte(in), &v, yaml.Strict())
		require.Error(t, err, in)
		assert.Contains(t, err.Error(), msg, in)
	}

	// A file secret round-trips as a plain path, a command as a mapping.
	b, err := yaml.Marshal(map[string]Secret{"a": SecretPath("/x"), "b": {Type: SecretCommand, Command: "c"}})
	require.NoError(t, err)
	var back map[string]Secret
	require.NoError(t, yaml.Unmarshal(b, &back))
	assert.Equal(t, SecretPath("/x"), back["a"])
	assert.Equal(t, Secret{Type: SecretCommand, Command: "c"}, back["b"])
	assert.Contains(t, string(b), "a: /x")
}

func TestReadSecret(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "tok"), []byte("from-file\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "!bang"), []byte("bang-file"), 0o600))

	// String path, resolved against the root.
	b, err := ReadSecret(ctx, SecretPath("tok").In(dir))
	require.NoError(t, err)
	assert.Equal(t, "from-file\n", string(b))

	// File mapping.
	b, err = ReadSecret(ctx, Secret{Type: SecretFile, Path: "tok"}.In(dir))
	require.NoError(t, err)
	assert.Equal(t, "from-file\n", string(b))

	// A "!" path is a path, not a command.
	b, err = ReadSecret(ctx, SecretPath("!bang").In(dir))
	require.NoError(t, err)
	assert.Equal(t, "bang-file", string(b))

	// Command mapping: not joined to the root, stdout is the secret.
	cmd := Secret{Type: SecretCommand, Command: "printf from-cmd"}
	assert.Equal(t, cmd, cmd.In(dir))
	b, err = ReadSecret(ctx, cmd)
	require.NoError(t, err)
	assert.Equal(t, "from-cmd", string(b))
	assert.Equal(t, "secret command", cmd.String(), "errors never name the command")

	// A failing command reports stderr's last line, not the command or secret.
	_, err = ReadSecret(ctx, Secret{Type: SecretCommand, Command: "echo noise >&2; echo vault sealed >&2; exit 3 # tok-123"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault sealed")
	assert.NotContains(t, err.Error(), "noise")
	assert.NotContains(t, err.Error(), "tok-123")

	_, err = ReadSecret(ctx, Secret{})
	require.ErrorContains(t, err, "not set")
}

// TestReadSecretTimeout: a stuck command is killed and reported as a
// timeout. Not parallel: it shortens the package timeout.
func TestReadSecretTimeout(t *testing.T) {
	old := SecretCommandTimeout
	SecretCommandTimeout = 100 * time.Millisecond
	t.Cleanup(func() { SecretCommandTimeout = old })
	_, err := ReadSecret(context.Background(), Secret{Type: SecretCommand, Command: "sleep 30 # s3cret"})
	require.ErrorContains(t, err, "timed out")
	assert.NotContains(t, err.Error(), "s3cret")
}

// TestSecretFieldsAcceptMapping: every secret field takes the command
// form in a full config.
func TestSecretFieldsAcceptMapping(t *testing.T) {
	cmd := func(c string) string { return "{type: command, command: " + c + "}" }
	c, err := Parse([]byte(hostBase + `identities:
  app: {kind: github-app, app_id: 1, private_key_file: ` + cmd("key") + `, installations: {a: 1}, repos: ["a/*"]}
  pat: {kind: github-token, token_file: ` + cmd("pat") + `, repos: ["b/*"]}
  lapp: {kind: linear-app, client_id_file: ` + cmd("cid") + `, client_secret_file: ` + cmd("csec") + `, webhook_secret_file: ` + cmd("whs") + `, workspace: w, teams: [ENG]}
`))
	require.NoError(t, err)
	assert.Equal(t, "key", c.Identities["app"].PrivateKeyFile.Command)
	assert.Equal(t, "pat", c.Identities["pat"].TokenFile.Command)
	assert.Equal(t, "cid", c.Identities["lapp"].ClientIDFile.Command)
	assert.Equal(t, "csec", c.Identities["lapp"].ClientSecretFile.Command)
	assert.Equal(t, "whs", c.Identities["lapp"].WebhookSecretFile.Command)
}
