package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// A reload tells event subscribers, so Desk re-reads desk.embeds at once
// (JOB-151). A rejected config changes nothing and says nothing.
func TestReloadPublishesStateForDesk(t *testing.T) {
	c, _, dir := scanControllerYAML(t, "projects:\n  js: {}\nagents:\n  a: {command: x}\n# %[1]s\n")
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	c.poolQuarantined = map[string]string{}
	c.configSwapped = make(chan struct{}, 1)
	c.events = newHub(8)
	sub, _ := c.events.add()

	write := func(body string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte("projects:\n  js: {}\nagents:\n  a: {command: x}\n"+body), 0o600))
	}
	write("desk:\n  embeds:\n    - {id: a, label: A, url: \"https://a.example/\"}\n")
	require.NoError(t, c.reload())
	select {
	case ev := <-sub.ch:
		assert.Equal(t, rpc.EventKindState, ev.Kind)
	default:
		t.Fatal("no state event after a reload")
	}
	require.Len(t, c.config().Desk.Embeds, 1)

	write("desk:\n  embeds:\n    - {id: a, label: A, url: \"javascript:alert(1)\"}\n")
	require.Error(t, c.reload())
	select {
	case ev := <-sub.ch:
		t.Fatalf("event %q after a rejected reload", ev.Kind)
	default:
	}
	assert.Equal(t, "https://a.example/", c.config().Desk.Embeds[0].URL, "the old config stays")
}
