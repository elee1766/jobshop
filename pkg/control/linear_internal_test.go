package control

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestRunMetaLinear: a linear workflow stores meta.linear with the key
// and only the code repo in meta.github, so GitHub readers ignore it.
func TestRunMetaLinear(t *testing.T) {
	w := config.Workflow{Issue: "issue", Repo: "acme/app", Tracker: config.TrackerLinear, LinearWorkspace: "acme"}
	mb, err := runMeta(w, map[string]any{"issue": "eng-42"})
	require.NoError(t, err)
	var m map[string]map[string]any
	require.NoError(t, json.Unmarshal(mb, &m))
	assert.Equal(t, map[string]any{"workspace": "acme", "issue": "ENG-42"}, m["linear"])
	assert.Equal(t, map[string]any{"repo": "acme/app"}, m["github"])
	assert.Empty(t, implicitThreads(store.Ticket{Meta: mb}, nil), "no GitHub thread for a Linear key")

	_, err = runMeta(w, map[string]any{"issue": "42"})
	require.ErrorContains(t, err, "not a Linear issue key")
}
