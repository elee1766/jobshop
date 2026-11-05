package redact

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A GitHub token pattern the default gitleaks rules detect.
const ghToken = "ghp_" + "Zx9Qw3Er5Ty7Ui9Op1As3Df5Gh7Jk9Lz1Xc3V"

func TestSecretEnvValues(t *testing.T) {
	got := SecretEnvValues([]string{"PATH=/usr/bin:/bin", "GH_TOKEN=abcdefgh", "DB_PASSWORD=abcdefghijkl", "API_KEY=short", "NAME=jobshop-user"})
	assert.Equal(t, []string{"abcdefghijkl", "abcdefgh"}, got, "secret-named, 8+ bytes, longest first")
}

func TestJSON(t *testing.T) {
	env := []string{"FAKE_API_TOKEN=launch-secret-value-123"}
	raw := json.RawMessage(`{"api_key":"sk-keyvalue","text":"key launch-secret-value-123 and ` + ghToken + `","n":1}`)
	got := string(JSON(raw, env))
	assert.True(t, json.Valid([]byte(got)))
	assert.NotContains(t, got, "sk-keyvalue", "secret-named key")
	assert.NotContains(t, got, "launch-secret-value-123", "env value")
	assert.NotContains(t, got, ghToken, "gitleaks finding")
	assert.Contains(t, got, `"n":1`)
}

func TestText(t *testing.T) {
	got := Text("tok "+ghToken, nil)
	assert.NotContains(t, got, ghToken)
	assert.Contains(t, got, Mask)
	assert.Equal(t, "plain", Text("plain", nil))
}

func TestFind(t *testing.T) {
	got := Find("a\nb\ntok " + ghToken)
	if assert.Len(t, got, 1) {
		assert.Equal(t, 3, got[0].Line, "1-based line")
		assert.NotEmpty(t, got[0].Rule)
	}
	assert.Empty(t, Find("no secrets here"))
}
