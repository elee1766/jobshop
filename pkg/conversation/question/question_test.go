package question_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/conversation/question"
	"github.com/elee1766/jobshop/pkg/domain"
)

func TestOptions(t *testing.T) {
	t.Parallel()
	out, err := question.Options([]domain.QuestionOption{{Title: " A ", Description: " do a "}, {Title: "B"}})
	require.NoError(t, err)
	assert.Equal(t, []domain.QuestionOption{{ID: 1, Title: "A", Description: "do a"}, {ID: 2, Title: "B"}}, out)

	for name, in := range map[string][]domain.QuestionOption{
		"no title":   {{Title: " "}},
		"two lines":  {{Title: "a\nb"}},
		"long title": {{Title: strings.Repeat("x", 121)}},
		"long desc":  {{Title: "a", Description: strings.Repeat("x", 1001)}},
		"too many":   make([]domain.QuestionOption, domain.MaxQuestionOptions+1),
	} {
		_, err := question.Options(in)
		require.ErrorIs(t, err, domain.ErrInvalid, name)
	}
}

func TestOf(t *testing.T) {
	t.Parallel()
	q, ok := question.Of(domain.Ticket{Meta: json.RawMessage(`{"question":{"ticket":"t1","session":"s1","context":"c"}}`)})
	require.True(t, ok)
	assert.Equal(t, question.Meta{Ticket: "t1", Session: "s1", Context: "c"}, q)
	_, ok = question.Of(domain.Ticket{Meta: json.RawMessage(`{"question":{"session":"s1"}}`)})
	assert.False(t, ok, "a question names the ticket it asks about")
	_, ok = question.Of(domain.Ticket{})
	assert.False(t, ok)
}

func TestContext(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 11, 8, 9, 30, 0, 0, time.UTC)
	step := domain.Ticket{
		ID: "js-2", Title: "Implement", Kind: domain.KindStep, Fields: json.RawMessage(`{"run":"js-1","step":"implement"}`),
		Meta: json.RawMessage(`{"github":{"repo":"a/b","pr":4},"linear":{"workspace":"acme","issue":"ENG-5"}}`),
	}
	var comments []domain.Comment
	for _, body := range []string{"one", "two", "three", "  four\n\n words  ", strings.Repeat("é", 300)} {
		comments = append(comments, domain.Comment{Author: "ann", Body: body, CreatedAt: at})
	}
	got := question.Context(" why stuck ", step, question.Asker{Session: "js-s-1"}, comments)
	assert.True(t, strings.HasPrefix(got, "why stuck\n\n**Links**\n\n- Ticket: js-2 (Implement)\n- Run: js-1, step implement\n"))
	assert.Contains(t, got, "- PR: https://github.com/a/b/pull/4\n")
	assert.Contains(t, got, "- Linear: https://linear.app/acme/issue/ENG-5\n")
	assert.Contains(t, got, "- Asked by session: js-s-1\n")
	assert.NotContains(t, got, "one", "only the last three comments")
	assert.Contains(t, got, "- 2026-11-08 09:30 ann: four words\n")
	assert.True(t, strings.HasSuffix(got, "..."), "the longest comment is clipped on a rune boundary")
	assert.True(t, strings.ToValidUTF8(got, "") == got)

	plain := question.Context("", domain.Ticket{ID: "t", Title: "T"}, question.Asker{Actor: "admin"}, nil)
	assert.Equal(t, "**Links**\n\n- Ticket: t (T)\n- Asked by: admin", plain)
}

func TestTexts(t *testing.T) {
	t.Parallel()
	opts := []domain.QuestionOption{{ID: 1, Title: "A", Description: "line\nbreak", Recommended: true}, {ID: 2, Title: "B"}}
	assert.Equal(t, "1. A (recommended): line break\n2. B\n", question.OptionsText(opts))
	assert.Equal(t, "Chose option 1: A\nline\nbreak\n\nNote: n", question.ChosenText(opts[0], "n"))
	assert.Equal(t, "Chose option 2: B", question.ChosenText(opts[1], ""))
	assert.Equal(t, "ab", question.ClipBytes("abé", 3))
	assert.Equal(t, "abé", question.ClipBytes("abé", 4))

	prompt := question.ChatPrompt("q1", question.Meta{Context: "ctx", Options: opts})
	assert.Contains(t, prompt, "`jobshop answer q1 \"<text>\"`")
	assert.Contains(t, prompt, "`jobshop answer q1 --option N")
	assert.Contains(t, prompt, "1. A (recommended)")
	assert.True(t, strings.HasSuffix(prompt, "The asker's context:\n\nctx\n"))
}
