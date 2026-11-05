package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const trigPool = "pools:\n  dev:\n    agent: a\n    project: js\n    dir: /tmp\n"

// TestTriggerBodyFile: a pack trigger's body_file is read relative to the
// pack and var-substituted; a host one is relative to jobshop.yaml and
// literal; config sources name the file (#382).
func TestTriggerBodyFile(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"p/pack.yaml":    trigPool + "triggers:\n  t1:\n    when: {pool: dev, ready_below: 1}\n    ticket:\n      pool: dev\n      title: hi\n      body_file: bodies/t1.md\n",
		"p/bodies/t1.md": "repo {{.repo}}\n",
		"host.md":        "host ${repo}\n",
		FileName: hostBase + "packs:\n  - path: p\n    prefix: x-\n    vars: {repo: o/r}\n" +
			"triggers:\n  t2:\n    when: {pool: x-dev, ready_below: 1}\n    ticket:\n      pool: x-dev\n      title: hi\n      body_file: host.md\n",
	})
	c, err := LoadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	assert.Equal(t, "repo o/r\n", c.Triggers["x-t1"].Ticket.Body)
	assert.Equal(t, "host ${repo}\n", c.Triggers["t2"].Ticket.Body)
	assert.Equal(t, []string{filepath.Join(dir, "p", "bodies", "t1.md")}, c.Sources["triggers.x-t1.ticket.body"])
	assert.Equal(t, []string{filepath.Join(dir, "host.md")}, c.Sources["triggers.t2.ticket.body"])
}

func TestTriggerBodyFileBoth(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"b.md": "x",
		FileName: hostBase + trigPool +
			"triggers:\n  t:\n    when: {pool: dev, ready_below: 1}\n    ticket:\n      pool: dev\n      title: hi\n      body: inline\n      body_file: b.md\n",
	})
	_, err := LoadFile(filepath.Join(dir, FileName))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "set body or body_file, not both")
}
