package tmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	vars := map[string]any{"repo": "acme/w", "team": "", "n": 3, "list": []any{"a", "b"}}
	for _, tt := range []struct{ in, want string }{
		{"no template ${x} $$", "no template ${x} $$"},
		{"on {{.repo}}", "on acme/w"},
		{"{{if .team}}team {{.team}}{{else}}no team{{end}}", "no team"},
		{`{{.team | default "go test ./..."}}`, "go test ./..."},
		{`{{.repo | default "x"}}`, "acme/w"},
		{`{{join ", " .list}}`, "a, b"},
		{`{{range .list}}- {{.}}{{end}}`, "- a- b"},
		{`{{quote .repo}} {{lower "AB"}}`, `"acme/w" ab`},
		{`literal {{"{{"}} and ${{"{{"}} github.sha }}`, "literal {{ and ${{ github.sha }}"},
		{"{{toYAML .list}}", "- a\n- b"},
	} {
		got, err := Render(tt.in, vars)
		require.NoError(t, err, tt.in)
		assert.Equal(t, tt.want, got, tt.in)
	}
	_, err := Render("{{.missing}}", vars)
	require.ErrorContains(t, err, `var "missing" is not set`)
	_, err = Render("{{if .x}", vars)
	require.ErrorContains(t, err, "template")
}

func TestValue(t *testing.T) {
	vars := map[string]any{"n": 3, "list": []any{"a"}}
	v, err := Value("{{.n}}", vars)
	require.NoError(t, err)
	assert.Equal(t, 3, v, "a whole-value reference keeps its type")
	v, err = Value("{{ .list }}", vars)
	require.NoError(t, err)
	assert.Equal(t, []any{"a"}, v)
	v, err = Value("n={{.n}}", vars)
	require.NoError(t, err)
	assert.Equal(t, "n=3", v)
	_, err = Value("{{.nope}}", vars)
	require.ErrorContains(t, err, `var "nope" is not set`)
}

func TestVars(t *testing.T) {
	vs, err := Vars(`{{.a}} {{if .b}}{{.c | default .d}}{{end}} {{range .e}}{{.inner}}{{end}} {{with .f}}{{.g}}{{end}}`)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, vs, "range and with bodies are not top-level vars")
	vs, err = Vars("plain")
	require.NoError(t, err)
	assert.Empty(t, vs)
}
