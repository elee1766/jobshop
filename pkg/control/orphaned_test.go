package control_test

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// TestOrphanedTickets covers both reasons: a ticket with no pool and one
// whose pool is not configured are listed and warned about once; setting
// a configured pool drops the ticket out.
func TestOrphanedTickets(t *testing.T) {
	var logs syncBuf
	root, start := poolRootOpt(t, "1m", control.Options{Log: slog.New(slog.NewTextHandler(&logs, nil))})
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var none, gone domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Project: "js", Title: "none"}, &none))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Project: "js", Title: "gone", Pool: "gone"}, &gone))

	orphaned := func(project string) []string {
		var out []domain.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Orphaned: true, Project: project}, &out))
		ids := []string{}
		for _, o := range out {
			ids = append(ids, o.ID)
		}
		return ids
	}
	assert.ElementsMatch(t, []string{none.ID, gone.ID}, orphaned(""))
	assert.Empty(t, orphaned("other"))

	warned := func(id, reason string) int {
		n := 0
		for _, l := range strings.Split(logs.String(), "\n") {
			if strings.Contains(l, "ticket has no claimable pool") && strings.Contains(l, "ticket="+id) &&
				strings.Contains(l, "reason="+reason) {
				n++
			}
		}
		return n
	}
	passEventually(t, root.Dir, func() bool {
		return warned(none.ID, "no_pool") == 1 && warned(gone.ID, "unknown_pool") == 1
	}, "condition not met after supervisor passes")
	pass(t, root.Dir, 3)
	assert.Equal(t, 1, warned(none.ID, "no_pool"), "warn once, not every tick")
	assert.Equal(t, 1, warned(gone.ID, "unknown_pool"))

	dev := "dev"
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketUpdate, rpc.UpdateParams{ID: none.ID, Pool: &dev}, nil))
	assert.Equal(t, []string{gone.ID}, orphaned(""))
}
