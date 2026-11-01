package control_test

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/host"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestConversationExportDeadEndpoint exports to an address nothing
// listens on: the host gets the export spec, and the fake agent's turn
// still runs and closes its ticket.
func TestConversationExportDeadEndpoint(t *testing.T) {
	root, start := poolRoot(t, "1m")
	// A port that was free a moment ago: connections are refused.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	dead := "http://" + ln.Addr().String() + "/v1/logs"
	require.NoError(t, ln.Close())
	cfg := fmt.Sprintf("projects:\n  js: {}\ntelemetry:\n  otlp_logs_endpoint: %q\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\n    env: {FAKE_API_TOKEN: \"launch-secret-0123\"}\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n    conversation_export: true\n",
		dead, os.Args[0], root.Dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var a, b store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &a))
	waitClosed(t, admin, a.ID)
	sessions := openSessions(t, admin)
	require.Len(t, sessions, 1)
	spec, err := host.ReadSpec(root.SessionDir(sessions[0].ID))
	require.NoError(t, err)
	require.NotNil(t, spec.ConvExport)
	assert.Equal(t, dead, spec.ConvExport.LogsURL)
	assert.Equal(t, "dev", spec.ConvExport.Pool)

	// A second turn on the same host: export failures never stall it.
	start2 := time.Now()
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &b))
	waitClosed(t, admin, b.ID)
	assert.Less(t, time.Since(start2), 15*time.Second)
	assert.Len(t, openSessions(t, admin), 1, "session not crashed")
}
