package control_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestReportIncidents lists a step with a failed attempt and its last
// comment; a zero window falls back to 24h.
func TestReportIncidents(t *testing.T) {
	admin := runRoot(t)
	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "7"}})
	s0 := r.Steps[0]
	closeStep(t, admin, s0.ID, "failed")

	var res rpc.IncidentsResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReportIncidents, rpc.IncidentsParams{}, &res))
	require.WithinDuration(t, time.Now().Add(-24*time.Hour), res.Since, time.Minute)
	require.Len(t, res.Steps, 1)
	st := res.Steps[0]
	require.Equal(t, s0.ID, st.ID)
	require.Equal(t, r.Run.ID, st.Run)
	require.Equal(t, "implement", st.Step)
	require.Equal(t, 1, st.Failures)
	require.Equal(t, domain.StatusOpen, st.Status, "reopened for retry")
	require.NotEmpty(t, st.LastError)
	require.NotNil(t, res.Logs)
	require.NotNil(t, res.Sessions)
}
