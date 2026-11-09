package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestReportQuestionArgs (JOB-101): -q needs a context of two lines or
// more; options parse from --option or --options-json, at most five, and
// --recommend marks one.
func TestReportQuestionArgs(t *testing.T) {
	var p rpc.ReportParams
	for _, bad := range []reportArgs{
		{Question: "which?"},
		{Question: "which?", Context: "   "},
		{Question: "which?", Context: "just one line"},
		{Context: "a\nb"},
		{Question: "q", Context: "a\nb", Option: []string{"|no title"}},
		{Question: "q", Context: "a\nb", Option: []string{"a", "b", "c", "d", "e", "f"}},
		{Question: "q", Context: "a\nb", Option: []string{"a"}, OptionsJSON: `[{"title":"b"}]`},
		{Question: "q", Context: "a\nb", Option: []string{"a"}, Recommend: 2},
		{Question: "q", Context: "-", OptionsJSON: "-"},
	} {
		assert.Error(t, bad.question(&p), "%+v", bad)
	}

	p = rpc.ReportParams{}
	a := reportArgs{Question: "which base?", Context: "rebased on both\nmain has the API", Option: []string{"main|ship next minor", "release | backport"}, Recommend: 1}
	require.NoError(t, a.question(&p))
	assert.Equal(t, "rebased on both\nmain has the API", p.Context)
	assert.Equal(t, []domain.QuestionOption{
		{Title: "main", Description: "ship next minor", Recommended: true},
		{Title: "release", Description: "backport"},
	}, p.Options)

	p = rpc.ReportParams{}
	a = reportArgs{Question: "q", Context: "a\nb", OptionsJSON: `[{"title":"x","description":"long\ntext"}]`}
	require.NoError(t, a.question(&p))
	assert.Equal(t, []domain.QuestionOption{{Title: "x", Description: "long\ntext"}}, p.Options)

	// A report with no question needs none of it.
	p = rpc.ReportParams{}
	require.NoError(t, reportArgs{Note: "on it"}.question(&p))
	assert.Empty(t, p.Context)
}
