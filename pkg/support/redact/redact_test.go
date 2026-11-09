package redact

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Secrets are built at run time from a fixed seed, so no literal in this
// file looks like a key (push protection) and runs are repeatable.
var gen = rand.New(rand.NewPCG(578, 1))

const (
	alnum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	hexd  = "0123456789abcdef"
)

func rnd(set string, n int) string {
	var b strings.Builder
	for range n {
		b.WriteByte(set[gen.IntN(len(set))])
	}
	return b.String()
}

func uuid() string {
	return rnd(hexd, 8) + "-" + rnd(hexd, 4) + "-" + rnd(hexd, 4) + "-" + rnd(hexd, 4) + "-" + rnd(hexd, 12)
}

func ghToken() string { return "ghp" + "_" + rnd(alnum, 36) }

func TestSecretEnvValues(t *testing.T) {
	got := SecretEnvValues([]string{"PATH=/usr/bin:/bin", "GH_TOKEN=abcdefgh", "DB_PASSWORD=abcdefghijkl", "API_KEY=short", "NAME=jobshop-user"})
	assert.Equal(t, []string{"abcdefghijkl", "abcdefgh"}, got, "secret-named, 8+ bytes, longest first")
}

func TestJSON(t *testing.T) {
	tok := ghToken()
	env := []string{"FAKE_API_TOKEN=launch-secret-value-123"}
	raw := json.RawMessage(`{"api_key":"sk-keyvalue","text":"key launch-secret-value-123 and ` + tok + `","n":1}`)
	got := string(JSON(raw, env))
	assert.True(t, json.Valid([]byte(got)))
	assert.NotContains(t, got, "sk-keyvalue", "secret-named key")
	assert.NotContains(t, got, "launch-secret-value-123", "env value")
	assert.NotContains(t, got, tok, "scanner finding")
	assert.Contains(t, got, `"n":1`)
}

// ACP usage counters have "token" in their names; JOB-160 kept them
// readable in journals. Only whole numbers under a "usage" key survive.
func TestJSONUsageCounters(t *testing.T) {
	// js-s-0034cxpaw8002 seq 164, a real prompt response.
	resp := `{"id":"h5","jsonrpc":"2.0","result":{"stopReason":"end_turn","usage":{"cachedReadTokens":109517,"cachedWriteTokens":7328,"inputTokens":6453,"outputTokens":766,"totalTokens":124064}}}`
	assert.JSONEq(t, resp, string(JSON(json.RawMessage(resp), nil)), "numeric accounting survives")

	sk := "sk-" + rnd(alnum, 20)
	for name, raw := range map[string]string{
		"string value":     `{"usage":{"inputTokens":"` + sk + `","totalTokens":"12"}}`,
		"object value":     `{"usage":{"inputTokens":{"v":"` + sk + `"}}}`,
		"array value":      `{"usage":{"outputTokens":["` + sk + `"]}}`,
		"negative":         `{"usage":{"totalTokens":-5}}`,
		"fraction":         `{"usage":{"totalTokens":1.5}}`,
		"outside usage":    `{"result":{"inputTokens":6453}}`,
		"other token key":  `{"usage":{"accessToken":6453}}`,
		"counter in array": `{"usage":[{"inputTokens":6453}]}`,
	} {
		got := string(JSON(json.RawMessage(raw), nil))
		assert.NotContains(t, got, sk, name)
		assert.NotRegexp(t, `:(-5|1\.5|6453)\b`, got, name)
		assert.Contains(t, got, Mask, name)
	}
}

func TestText(t *testing.T) {
	tok := ghToken()
	got := Text("tok "+tok, nil)
	assert.NotContains(t, got, tok)
	assert.Contains(t, got, Mask)
	assert.Equal(t, "plain", Text("plain", nil))
}

func TestFind(t *testing.T) {
	got := Find("a\nb\ntok " + ghToken())
	if assert.Len(t, got, 1) {
		assert.Equal(t, 3, got[0].Line, "1-based line")
		assert.NotEmpty(t, got[0].Rule)
	}
	assert.Empty(t, Find("no secrets here"))
}

// TestRules: what the default and built-in rules catch, and benign
// shapes they must leave alone (#578).
func TestRules(t *testing.T) {
	pw := rnd(alnum, 16)
	caught := []struct {
		name, text, secret string
	}{
		{"anthropic oauth", "", "sk-ant-oat01-" + rnd(alnum+"_-", 95) + "AA"},
		{"matrix", "", "syt_" + "am9ic2hvcA" + "_" + rnd(alnum, 20) + "_" + rnd(alnum, 6)},
		{"linear oauth", "", "lin_oauth_" + rnd(hexd, 64)},
		{"bitwarden sm", "", "0." + uuid() + "." + rnd(alnum, 30) + ":" + rnd(alnum, 22) + "=="},
		{"postgres dsn password", "DSN postgres://app:" + pw + "@db.internal:5432/app", pw},
		{"github app", "", "ghs" + "_" + rnd(alnum, 36)},
		{"openai project", "", "sk-proj-" + rnd(alnum, 58) + "T3Blbk" + "FJ" + rnd(alnum, 58)},
		{"anthropic api03", "", "sk-ant-api03-" + rnd(alnum, 93) + "AA"},
		{"webhook secret env", "WEBHOOK_SECRET=" + pw + rnd(alnum, 16), pw},
	}
	for _, c := range caught {
		t.Run(c.name, func(t *testing.T) {
			text := c.text
			if text == "" {
				text = "value " + c.secret + " end"
			}
			assert.NotEmpty(t, Find(text), "finding")
			got := Text(text, nil)
			assert.NotContains(t, got, c.secret, "masked")
			assert.Contains(t, got, Mask)
		})
	}
	// A URI rule masks the password, not the whole URI.
	assert.Contains(t, Text("postgres://app:"+pw+"@db.internal:5432/app", nil), "@db.internal:5432/app")

	benign := map[string]string{
		"git sha":     rnd(hexd, 40),
		"sha256":      "sha256:" + rnd(hexd, 64),
		"tx hash":     "0x" + rnd(hexd, 64),
		"address":     "0x" + rnd(hexd, 40),
		"nix store":   "/nix/store/" + rnd("0123456789abcdfghijklmnpqrsvwxyz", 32) + "-jobshop-0.1",
		"go.sum line": "github.com/stretchr/testify v1.12.1 h1:" + rnd(alnum+"+/", 43) + "=",
		"uuid":        uuid(),
		"traceparent": "00-" + rnd(hexd, 32) + "-" + rnd(hexd, 16) + "-01",
		"ticket id":   "js-00337t1b5c002",
	}
	for name, s := range benign {
		assert.Empty(t, Find("see "+s+" here"), "%s: %s", name, s)
	}
}

// TestCompile: operator rules are caught, disabled rules stop matching,
// and bad rules fail with the rule id.
func TestCompile(t *testing.T) {
	tok := "acme" + "_" + rnd(alnum, 32)
	sc, err := Compile([]Rule{{ID: "acme-token", Regex: `\bacme_[A-Za-z0-9]{32}\b`, Keywords: []string{"acme_"}}}, nil)
	require.NoError(t, err)
	got := sc.Find("x " + tok)
	require.Len(t, got, 1)
	assert.Equal(t, "acme-token", got[0].Rule)
	assert.NotContains(t, sc.Redactor(nil).Text("x "+tok), tok)
	assert.Empty(t, Find("x "+tok), "not in the built-in scanner")

	gh := ghToken()
	require.NotEmpty(t, Find(gh))
	ids := map[string]bool{}
	for _, f := range Find(gh) {
		ids[f.Rule] = true
	}
	var off []string
	for id := range ids {
		off = append(off, id)
	}
	sc, err = Compile(nil, off)
	require.NoError(t, err)
	assert.Empty(t, sc.Find(gh), "disabled")

	for _, c := range []struct {
		rules    []Rule
		disabled []string
		want     string
	}{
		{[]Rule{{ID: "bad", Regex: `(`, Keywords: []string{"x"}}}, nil, "bad"},
		{[]Rule{{ID: "github-pat", Regex: `x`, Keywords: []string{"x"}}}, nil, "github-pat"},
		{[]Rule{{ID: "a", Regex: `x`, Keywords: []string{"x"}}, {ID: "a", Regex: `y`, Keywords: []string{"y"}}}, nil, `"a"`},
		{[]Rule{{ID: "nokw", Regex: `x`}}, nil, "nokw"},
		{nil, []string{"no-such-rule"}, "no-such-rule"},
	} {
		_, err := Compile(c.rules, c.disabled)
		require.Error(t, err)
		assert.Contains(t, err.Error(), c.want)
	}
}
