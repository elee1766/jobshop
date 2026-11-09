package control

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

// TestDedupeTurns: a resumed session over three turns with one rule
// unchanged, one edited and one removed gets the exact text below each
// turn; a reset (context loss) sends everything again; inject once and
// always (#397).
func TestDedupeTurns(t *testing.T) {
	t.Parallel()
	rules := func(extra string) string {
		return rulesBase + "prompt_rules:\n  - {name: keep, text: KEEP}\n" + extra
	}
	c, _ := rulesController(t, rules("  - {name: edit, text: EDIT-1}\n  - {name: gone, text: GONE}\n"+
		"  - {name: once, text: ONCE-1, inject: once}\n  - {name: every, text: EVERY, inject: always}\n"))
	ctx := t.Context()
	ts := turnSpec{pool: "chat", kind: turnMessage}
	body := PromptPart{Source: "messages", Text: "M"}
	r := &runner{c: c, pool: "chat", log: c.log}

	turn := func() string { return r.useParts(c.assemble(ctx, ts, body)) }

	assert.Equal(t, "KEEP\n\nEDIT-1\n\nGONE\n\nONCE-1\n\nEVERY\n\nM", turn())

	setConfig(t, c, rules("  - {name: edit, text: EDIT-2}\n"+
		"  - {name: once, text: ONCE-2, inject: once}\n  - {name: every, text: EVERY, inject: always}\n"))
	assert.Equal(t, "Updated prompt rule \"edit\":\n\nEDIT-2\n\nEVERY\n\n"+
		"Prompt rule \"gone\" no longer applies.\n\nM", turn())
	assert.Equal(t, []string{"edit", "every"}, r.turnRules)

	assert.Equal(t, "EVERY\n\nM", turn(), "nothing changed: only the always rule")

	r.resetInjected()
	assert.Equal(t, "KEEP\n\nEDIT-2\n\nONCE-2\n\nEVERY\n\nM", turn(), "after context loss everything again")

	// session prompt shows the skipped parts with their status.
	_, all, _ := dedupe(c.config(), r.injected, c.assemble(ctx, ts, body))
	assert.Equal(t, "=== rule:keep (unchanged, not sent) ===\n\n=== rule:edit (unchanged, not sent) ===\n\n"+
		"=== rule:once (unchanged, not sent) ===\n\n=== rule:every (always) ===\nEVERY\n\n=== messages ===\nM\n", renderParts(all))
}

// setConfig swaps c's config for cfgYAML, as a reload would.
func setConfig(t *testing.T, c *Controller, cfgYAML string) {
	t.Helper()
	require.NoError(t, os.WriteFile(c.root.Config(), []byte(cfgYAML), 0o600))
	cfg, err := config.Load(c.root)
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	c.live.Store(lc)
}
