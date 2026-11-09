package control

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
<<<<<<< HEAD
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/store"
=======
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/storage/store"
>>>>>>> 7fea25a3 (refactor(storage): group store and journal packages under pkg/storage (JOB-173))
)

// TestRebaseMergeState: a rebase close on a step whose run's PR is dirty
// routes back without a failure, a clean PR counts a failure with a
// comment, and an unreadable PR keeps the rebase (#511).
func TestRebaseMergeState(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		prs      map[int]github.Pull
		failures int
		comment  bool
	}{
		{"dirty", map[int]github.Pull{300: {Number: 300, MergeableState: "dirty"}}, 0, false},
		{"behind", map[int]github.Pull{300: {Number: 300, MergeableState: "behind"}}, 0, false},
		{"clean", map[int]github.Pull{300: {Number: 300, MergeableState: "clean"}}, 1, true},
		{"read error", nil, 0, false},
		{"unknown", map[int]github.Pull{300: {Number: 300, MergeableState: "unknown"}}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, _, fa := activityController(t)
			c.activitySrc = &fakeReverts{fakeActivity: fa, prs: tc.prs}
			run := mkTicket(t, c, store.NewTicket{
				Kind: domain.KindRun, Title: "run", Fields: json.RawMessage(`{"workflow":"pr-work","state":"running"}`),
				Meta: json.RawMessage(`{"github":{"repo":"elee1766/jobshop","issue":511}}`),
			})
			impl := mkTicket(t, c, store.NewTicket{
				Kind: domain.KindStep, Title: "implement", ParentID: run.ID,
				Fields: json.RawMessage(`{"step":"implement","run":"` + run.ID + `"}`),
			})
			review := mkTicket(t, c, store.NewTicket{
				Kind: domain.KindStep, Title: "review-merge", ParentID: run.ID,
				Fields: json.RawMessage(`{"step":"review-merge","run":"` + run.ID + `"}`),
			})
			_, err := store.AddComment(t.Context(), c.store, "s1", impl.ID, "PR https://github.com/elee1766/jobshop/pull/300 head abc")
			require.NoError(t, err)

			res, why := c.rebaseResolution(t.Context(), review)
			got, err := store.Close(t.Context(), c.store, "admin", review.ID, nil, res)
			require.NoError(t, err)
			sf, err := domain.DecodeFields[domain.StepFields](got)
			require.NoError(t, err)
			assert.Equal(t, tc.failures, sf.Failures)
			assert.Equal(t, tc.comment, strings.HasPrefix(why, "close -r rebase counted as failed"))
		})
	}
}
