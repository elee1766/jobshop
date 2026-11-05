package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// Fake keys: the AWS one is gitleaks' own true-positive sample
// (cmd/generate/config/rules/aws.go). Split so this file is not a finding
// and push protection does not trip.
const (
	fakeAWSKey = "AKIA" + "LALEMEL33243OLIB"
	fakeGHPAT  = "ghp_" + "Zx9Qw3Er5Ty7Ui9Op1As3Df5Gh7Jk9Lz1Xc3V"
)

// TestGHSecretRefused (#574): a gh write whose body holds a key exits 1,
// gh is not run, and stderr names the rule, not the key.
func TestGHSecretRefused(t *testing.T) {
	dir := t.TempDir()
	gh, marker := tokenGH(t, dir)
	t.Setenv("JOBSHOP_GH", gh)
	t.Setenv("GH_TOKEN", "")
	f := &fakeController{decision: rpc.GuardAllow}
	g := &Globals{Socket: f.serve(t)}
	body := filepath.Join(dir, "body.md")
	require.NoError(t, os.WriteFile(body, []byte("ok\nkey "+fakeGHPAT+"\n"), 0o600))
	for _, args := range [][]string{
		{"pr", "comment", "1", "-R", "o/n", "--body", "aws " + fakeAWSKey},
		{"issue", "create", "-R", "o/n", "-t", "t", "-F", body},
		{"api", "repos/o/n/issues/1/comments", "-f", "body=" + fakeGHPAT},
		{"pr", "comment", "1", "-R", "o/n", "--body-file", "-"},
	} {
		t.Run(args[0]+" "+args[1], func(t *testing.T) {
			var out, errb bytes.Buffer
			code := runGH(g, args, strings.NewReader("stdin "+fakeGHPAT), &out, &errb)
			assert.Equal(t, 1, code)
			assert.NoFileExists(t, marker, "gh not run")
			assert.Contains(t, errb.String(), "rule ")
			assert.NotContains(t, errb.String(), fakeAWSKey)
			assert.NotContains(t, errb.String(), fakeGHPAT)
		})
	}
	f.mu.Lock()
	assert.Empty(t, f.guards, "refused before the guard")
	f.mu.Unlock()
}

// TestGHSecretAllow: the override needs an admin caller; reads ignore the
// check and refuse the flag.
func TestGHSecretAllow(t *testing.T) {
	dir := t.TempDir()
	gh, marker := tokenGH(t, dir)
	t.Setenv("JOBSHOP_GH", gh)
	t.Setenv("GH_TOKEN", "x")
	f := &fakeController{decision: rpc.GuardAllow}
	g := &Globals{Socket: f.serve(t)}
	var out, errb bytes.Buffer
	code := runGH(g, []string{"pr", "comment", "1", "-R", "o/n", "--body", fakeAWSKey, allowSecretsFlag}, nil, &out, &errb)
	assert.Equal(t, 1, code, "fake controller's whoami is not admin")
	assert.NoFileExists(t, marker)

	// A clean write runs.
	errb.Reset()
	assert.Equal(t, 5, runGH(g, []string{"pr", "comment", "1", "-R", "o/n", "--body", "fine"}, nil, &out, &errb), errb.String())
	assert.FileExists(t, marker)
}

func TestGHOutText(t *testing.T) {
	texts, in, err := ghOutText([]string{"api", "x", "-fa=1", "-F", "b=2", "--raw-field", "c=3", "--input", "-"}, strings.NewReader("IN"))
	require.NoError(t, err)
	assert.Equal(t, "IN", string(in))
	var got []string
	for _, x := range texts {
		got = append(got, x.text)
	}
	assert.ElementsMatch(t, []string{"1", "2", "3", "IN"}, got)
}
