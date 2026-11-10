package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
)

// These tests pin representations that other processes and old data depend
// on. The types moved here from store and rpc (JOB-166): a change to a tag, a
// message or a constant below is a compatibility break, not a refactor.

func TestErrorMessagesKeepTheStorePrefix(t *testing.T) {
	for err, want := range map[error]string{
		domain.ErrNotFound:     "store: not found",
		domain.ErrConflict:     "store: revision conflict",
		domain.ErrNotHolder:    "store: caller does not hold the claim",
		domain.ErrNoWork:       "store: no claimable ticket",
		domain.ErrInvalid:      "store: invalid input",
		domain.ErrCycle:        "store: edge would create a blocking cycle",
		domain.ErrClosed:       "store: ticket is closed",
		domain.ErrAlreadyExist: "store: already exists",
	} {
		assert.EqualError(t, err, want)
	}
}

func TestTicketJSONShape(t *testing.T) {
	at := time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC)
	tk := domain.Ticket{
		ID: "js-1", Project: "js", Kind: domain.KindStep, Status: domain.StatusInProgress,
		Priority: 2, Fence: 3, Revision: 4, CreatedAt: at, UpdatedAt: at,
		Fields: json.RawMessage(`{"run":"r","step":"s"}`),
	}
	b, err := json.Marshal(tk)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"js-1","project":"js","kind":"step","status":"in_progress","priority":2,"fence":3,"revision":4,`+
		`"created_at":"2026-11-09T10:00:00Z","updated_at":"2026-11-09T10:00:00Z","fields":{"run":"r","step":"s"}}`, string(b))

	var back domain.Ticket
	require.NoError(t, json.Unmarshal(b, &back))
	assert.Equal(t, tk.ID, back.ID)
	assert.Equal(t, domain.KindStep, back.Kind)
}

func TestFieldsRoundTripAndStrictness(t *testing.T) {
	raw := json.RawMessage(`{"run":"r","step":"s","index":1,"retries":2}`)
	out, err := domain.NormalizeFields(domain.KindStep, raw)
	require.NoError(t, err)
	f, err := domain.DecodeFields[domain.StepFields](domain.Ticket{Kind: domain.KindStep, Fields: out})
	require.NoError(t, err)
	assert.Equal(t, "r", f.Run)
	assert.Equal(t, 2, f.Retries)

	// Unknown fields and missing required ones are still rejected as invalid.
	_, err = domain.NormalizeFields(domain.KindStep, json.RawMessage(`{"run":"r","step":"s","nope":1}`))
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = domain.NormalizeFields(domain.KindStep, json.RawMessage(`{"run":"r"}`))
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = domain.NormalizeFields("nope", nil)
	require.ErrorIs(t, err, domain.ErrInvalid)
}

func TestStoredVocabularyIsUnchanged(t *testing.T) {
	assert.Equal(t, []string{"task", "session", "message", "run", "step", "batch"},
		[]string{
			string(domain.KindTask), string(domain.KindSession), string(domain.KindMessage),
			string(domain.KindRun), string(domain.KindStep), string(domain.KindBatch),
		})
	assert.Equal(t, []string{"open", "in_progress", "closed"},
		[]string{string(domain.StatusOpen), string(domain.StatusInProgress), string(domain.StatusClosed)})
	assert.Equal(t, []string{"blocks", "waits_for", "tracks"},
		[]string{string(domain.EdgeBlocks), string(domain.EdgeWaitsFor), string(domain.EdgeTracks)})
	assert.Equal(t, []string{"pending", "admitted", "denied", "held", "dispatched", "failed"}, domain.InboxStates)
	assert.Equal(t, []string{"allow", "deny", "ask_human"}, []string{domain.GuardAllow, domain.GuardDeny, domain.GuardAskHuman})
	assert.Equal(t, "request_changes", domain.ResolutionRequestChanges)
	assert.Equal(t, 5, domain.MaxQuestionOptions)
	assert.Equal(t, []string{"correction", "failure", "workaround"}, domain.LearningKinds)
	assert.Equal(t, "github.checks", domain.WaitChecks)
}

func TestQuestionOptionJSON(t *testing.T) {
	b, err := json.Marshal([]domain.QuestionOption{{ID: 1, Title: "A", Recommended: true}, {ID: 2, Title: "B", Description: "d"}})
	require.NoError(t, err)
	assert.JSONEq(t, `[{"id":1,"title":"A","recommended":true},{"id":2,"title":"B","description":"d"}]`, string(b))
}

func TestWaitValidation(t *testing.T) {
	ok := &domain.Wait{Reason: "ci", Conds: []domain.WaitCond{{Type: domain.WaitChecks, Repo: "a/b", PR: 1, State: "success"}}}
	require.NoError(t, ok.Validate())
	for name, w := range map[string]*domain.Wait{
		"no conditions": {Reason: "r"},
		"no reason":     {Conds: ok.Conds},
		"bad mode":      {Reason: "r", Mode: "some", Conds: ok.Conds},
		"bad checks":    {Reason: "r", Conds: []domain.WaitCond{{Type: domain.WaitChecks, Repo: "a/b", PR: 1, State: "pending"}}},
		"unknown type":  {Reason: "r", Conds: []domain.WaitCond{{Type: "nope"}}},
	} {
		require.ErrorIs(t, w.Validate(), domain.ErrInvalid, name)
	}
	assert.True(t, (&domain.Wait{Mode: domain.WaitAny, Conds: []domain.WaitCond{{Met: true}, {}}}).Satisfied())
	assert.False(t, (&domain.Wait{Conds: []domain.WaitCond{{Met: true}, {}}}).Satisfied())
}

func TestStepRules(t *testing.T) {
	closed := func(res string, f domain.StepFields) domain.Ticket {
		b, err := json.Marshal(f)
		require.NoError(t, err)
		return domain.Ticket{Kind: domain.KindStep, Status: domain.StatusClosed, Resolution: res, Fields: b}
	}
	assert.True(t, domain.StepFinished(closed("done", domain.StepFields{})))
	assert.True(t, domain.StepFinished(closed(domain.StepSkipped, domain.StepFields{})))
	assert.False(t, domain.StepFinished(closed("failed", domain.StepFields{})))
	assert.True(t, domain.StepExhausted(closed("failed", domain.StepFields{Failures: 3, Retries: 3})))
	assert.False(t, domain.StepExhausted(closed("failed", domain.StepFields{Failures: 1, Retries: 3})))
	assert.True(t, domain.ReviewCapped(closed(domain.ResolutionRequestChanges, domain.StepFields{ReviewRounds: 4})))
	assert.False(t, domain.ReviewCapped(closed(domain.ResolutionRequestChanges, domain.StepFields{ReviewRounds: 3})))

	var step domain.StepFields
	assert.Equal(t, domain.DefaultReviewRounds, step.ReviewRoundLimit())
	one := 1
	step.MaxReviewRounds = &one
	assert.Equal(t, 1, step.ReviewRoundLimit())
}

func TestMetaReaders(t *testing.T) {
	raw := json.RawMessage(`{"hook":{"kind":"approval","n":1},"run":{"id":"r1"}}`)
	assert.Equal(t, "approval", domain.MetaString(raw, domain.MetaHook, "kind"))
	assert.Equal(t, "", domain.MetaString(raw, domain.MetaHook, "n"), "a non-string value reads as empty")
	assert.Equal(t, "", domain.MetaString(json.RawMessage(`not json`), domain.MetaHook, "kind"))
	assert.Len(t, domain.MetaMap(raw), 2)
	assert.Empty(t, domain.MetaMap(nil))
}

func TestStepFailureCountsOneAttempt(t *testing.T) {
	b, err := json.Marshal(domain.StepFields{Run: "r", Step: "s", Retries: 2})
	require.NoError(t, err)
	tk := domain.Ticket{Kind: domain.KindStep, Fields: b}
	f, exhausted, isStep, err := domain.StepFailure(&tk)
	require.NoError(t, err)
	assert.True(t, isStep)
	assert.False(t, exhausted)
	assert.Equal(t, 1, f.Failures)
	_, exhausted, _, err = domain.StepFailure(&tk)
	require.NoError(t, err)
	assert.True(t, exhausted, "the second failure reaches the retry limit of 2")

	_, _, isStep, err = domain.StepFailure(&domain.Ticket{Kind: domain.KindTask})
	require.NoError(t, err)
	assert.False(t, isStep)
}
