package scaleplan_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/scaleplan"
)

func TestForKeys(t *testing.T) {
	t.Parallel()
	idle := func(id, repo string) scaleplan.Session { return scaleplan.Session{ID: id, Repo: repo, Idle: true} }
	busy := func(id, repo string) scaleplan.Session { return scaleplan.Session{ID: id, Repo: repo} }
	for _, tc := range []struct {
		name       string
		desired    int
		ready      map[string]int
		sessions   []scaleplan.Session
		want       scaleplan.Plan
		wantStarts []string
	}{
		{
			"starts follow each repo's demand", 3,
			map[string]int{"": 1, "o/b": 2},
			nil,
			scaleplan.Plan{Start: 3},
			[]string{"", "o/b", "o/b"},
		},
		{
			"a short total is shared one at a time", 2,
			map[string]int{"": 3, "o/b": 3},
			nil,
			scaleplan.Plan{Start: 2},
			[]string{"", "o/b"},
		},
		{
			"an idle session of a repo with no work is cordoned", 1,
			map[string]int{"": 0, "o/b": 1},
			[]scaleplan.Session{idle("d", "")},
			scaleplan.Plan{Cordon: []string{"d"}, Start: 1},
			[]string{"o/b"},
		},
		{
			"a working session counts as its repo's demand", 2,
			map[string]int{"": 0, "o/b": 1},
			[]scaleplan.Session{busy("b1", "o/b")},
			scaleplan.Plan{Start: 1},
			[]string{"o/b"},
		},
		{
			"surplus held by the window keeps sessions in rotation", 2,
			map[string]int{"": 0, "o/b": 0},
			[]scaleplan.Session{idle("d", ""), idle("b", "o/b")},
			scaleplan.Plan{},
			nil,
		},
		{
			"min_sessions with no sessions goes to the default key", 1,
			map[string]int{"": 0, "o/b": 0},
			nil,
			scaleplan.Plan{Start: 1},
			[]string{""},
		},
		{
			"a key dropped from the pool drains", 1,
			map[string]int{"": 1},
			[]scaleplan.Session{idle("x", "o/gone")},
			scaleplan.Plan{Cordon: []string{"x"}, Start: 1},
			[]string{""},
		},
		{
			"a scale cordon is lifted in its own repo", 1,
			map[string]int{"": 0, "o/b": 1},
			[]scaleplan.Session{{ID: "b", Repo: "o/b", Idle: true, Cordon: domain.CordonScale}},
			scaleplan.Plan{Uncordon: []string{"b"}},
			nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			by := map[string][]scaleplan.Session{}
			for _, s := range tc.sessions {
				by[s.Repo] = append(by[s.Repo], s)
			}
			plan, starts := scaleplan.ForKeys(tc.desired, tc.ready, by)
			assert.Equal(t, tc.want, plan)
			assert.Equal(t, tc.wantStarts, starts)
		})
	}
}
