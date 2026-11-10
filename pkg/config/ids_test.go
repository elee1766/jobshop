package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/ids"
)

// TestIDSpec: the global ids block, a per-project override, and bad
// specs rejected at load (JOB-57).
func TestIDSpec(t *testing.T) {
	c, err := Parse([]byte(`default_project: js
ids: {generator: ulid}
projects:
  js: {}
  hq: {ids: {generator: words, words: 2, separator: _}}
`))
	require.NoError(t, err)
	assert.Equal(t, ids.Spec{Generator: ids.ULID}, c.IDSpec("js"))
	assert.Equal(t, ids.Spec{Generator: ids.Words, Words: 2, Separator: "_"}, c.IDSpec("hq"))
	assert.Equal(t, ids.Spec{Generator: ids.ULID}, c.IDSpec("other"))

	c, err = Parse([]byte("projects:\n  js: {}\n"))
	require.NoError(t, err)
	assert.Equal(t, ids.Spec{Generator: ids.Sonyflake}, c.IDSpec("js"), "default")

	_, err = Parse([]byte("ids: {generator: nope}\n"))
	require.ErrorContains(t, err, "unknown generator")
	_, err = Parse([]byte("projects:\n  js: {ids: {generator: seq, words: 3}}\n"))
	require.ErrorContains(t, err, "project js")
}
