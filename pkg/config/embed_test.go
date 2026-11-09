package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeskEmbedsParse(t *testing.T) {
	c, err := Parse([]byte(hostBase + `desk:
  embeds:
    - id: grafana
      label: Grafana
      url: https://grafana.example/d/abc?kiosk
      sandbox: [allow-forms, allow-same-origin]
      allow: [fullscreen]
    - id: local
      label: local app
      url: http://127.0.0.1:3000/
`))
	require.NoError(t, err)
	require.Len(t, c.Desk.Embeds, 2)
	assert.Equal(t, "grafana", c.Desk.Embeds[0].ID, "config order is kept")
	assert.Equal(t, []string{"allow-scripts", "allow-forms", "allow-same-origin"}, c.Desk.Embeds[0].SandboxTokens())
	assert.Equal(t, "https://grafana.example", c.Desk.Embeds[0].Origin())
	assert.Equal(t, []string{"allow-scripts"}, c.Desk.Embeds[1].SandboxTokens(), "restrictive default")
	assert.Equal(t, "http://127.0.0.1:3000", c.Desk.Embeds[1].Origin())

	none, err := Parse([]byte(hostBase))
	require.NoError(t, err)
	assert.Empty(t, none.Desk.Embeds, "no embeds: Desk is unchanged")
}

func TestDeskEmbedsRejected(t *testing.T) {
	ok := func(extra string) string {
		return "id: a\nlabel: A\nurl: https://a.example/\n" + extra
	}
	cases := map[string]struct{ entry, want string }{
		"empty id":            {"id: \"\"\nlabel: A\nurl: https://a.example/", "id"},
		"uppercase id":        {"id: Grafana\nlabel: A\nurl: https://a.example/", "id"},
		"path id":             {"id: a/b\nlabel: A\nurl: https://a.example/", "id"},
		"built in id":         {"id: system\nlabel: A\nurl: https://a.example/", "reserved"},
		"embed id":            {"id: embed\nlabel: A\nurl: https://a.example/", "reserved"},
		"empty label":         {"id: a\nlabel: \"\"\nurl: https://a.example/", "label"},
		"blank label":         {"id: a\nlabel: \"  \"\nurl: https://a.example/", "label"},
		"padded label":        {"id: a\nlabel: \" A\"\nurl: https://a.example/", "label"},
		"long label":          {"id: a\nlabel: " + strings.Repeat("x", 40) + "\nurl: https://a.example/", "label"},
		"no url":              {"id: a\nlabel: A", "url is required"},
		"javascript":          {"id: a\nlabel: A\nurl: \"javascript:alert(1)\"", "scheme"},
		"data":                {"id: a\nlabel: A\nurl: \"data:text/html,<b>\"", "scheme"},
		"file":                {"id: a\nlabel: A\nurl: file:///etc/passwd", "scheme"},
		"relative":            {"id: a\nlabel: A\nurl: /other", "scheme"},
		"protocol relative":   {"id: a\nlabel: A\nurl: //a.example/", "scheme"},
		"userinfo":            {"id: a\nlabel: A\nurl: https://user:pw@a.example/", "credentials"},
		"user only":           {"id: a\nlabel: A\nurl: https://user@a.example/", "credentials"},
		"opaque":              {"id: a\nlabel: A\nurl: \"https:a.example\"", "host"},
		"no host":             {"id: a\nlabel: A\nurl: \"https:///x\"", "host"},
		"space":               {"id: a\nlabel: A\nurl: \"https://a.example/a b\"", "ASCII"},
		"non ascii host":      {"id: a\nlabel: A\nurl: \"https://ä.example/\"", "ASCII"},
		"host chars":          {"id: a\nlabel: A\nurl: \"https://a_b.example/\"", "host"},
		"token in query":      {"id: a\nlabel: A\nurl: \"https://a.example/?token=abc\"", "credential"},
		"token in fragment":   {"id: a\nlabel: A\nurl: \"https://a.example/#access_token=abc\"", "credential"},
		"encoded secret key":  {"id: a\nlabel: A\nurl: \"https://a.example/?api%5Fkey=1\"", "credential"},
		"plain http remote":   {"id: a\nlabel: A\nurl: http://a.example/", "plain http"},
		"bad sandbox":         {ok("sandbox: [allow-top-navigation]"), "sandbox"},
		"typo sandbox":        {ok("sandbox: [allow-script]"), "sandbox"},
		"orphan escape":       {ok("sandbox: [allow-popups-to-escape-sandbox]"), "allow-popups"},
		"bad allow":           {ok("allow: [camera]"), "allow"},
		"unknown field":       {ok("referrer: x"), "referrer"},
		"scalar not sequence": {ok("sandbox: allow-forms"), ""},
	}
	for name, tc := range cases {
		_, err := Parse([]byte(hostBase + "desk:\n  embeds:\n    - " + strings.ReplaceAll(strings.TrimSpace(tc.entry), "\n", "\n      ") + "\n"))
		if assert.Error(t, err, name) && tc.want != "" {
			assert.Contains(t, err.Error(), tc.want, name)
		}
	}
	_, err := Parse([]byte(hostBase + "desk:\n  embeds:\n    - {id: a, label: A, url: \"https://a.example/\"}\n    - {id: a, label: B, url: \"https://b.example/\"}\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate id")

	var many strings.Builder
	many.WriteString("desk:\n  embeds:\n")
	for i := range MaxDeskEmbeds + 1 {
		many.WriteString("    - {id: a" + strings.Repeat("b", i) + ", label: A, url: \"https://a.example/\"}\n")
	}
	_, err = Parse([]byte(hostBase + many.String()))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at most")
}

func TestDeskEmbedsAllowed(t *testing.T) {
	for _, url := range []string{
		"https://a.example/",
		"http://localhost:8080/x?y=1",
		"http://127.0.0.1:3000",
		"http://[::1]:3000/",
		"http://app.localhost/",
		"https://a.example/path?q=1#frag",
	} {
		assert.NoError(t, DeskEmbed{ID: "a", Label: "A", URL: url}.Check(), url)
	}
	assert.NoError(t, DeskEmbed{ID: "a", Label: "A", URL: "http://intranet.example/", AllowHTTP: true}.Check(), "explicit plain http")
}

func TestDeskEmbedsNotInPacks(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"p/pack.yaml":  "desk:\n  embeds:\n    - {id: a, label: A, url: \"https://a.example/\"}\n",
		"jobshop.yaml": hostBase + "packs:\n  - path: p\n",
	})
	_, err := LoadFile(dir + "/jobshop.yaml")
	assert.Error(t, err, "a pack cannot add an embed")
}
