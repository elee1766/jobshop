package transport_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/integrations/apicall"
	"github.com/gfx-labs/jobshop/pkg/integrations/internal/transport"
)

type source struct {
	fresh string
	err   error
	stale []string
}

func (s *source) Token(context.Context) (string, error) { return "", nil }

func (s *source) Renew(_ context.Context, stale string) (string, error) {
	s.stale = append(s.stale, stale)
	return s.fresh, s.err
}

// sender answers 401 to every token but ok and records the tokens it saw.
func sender(ok string, saw *[]string) transport.Send[string] {
	return func(tok string) (string, int, error) {
		*saw = append(*saw, tok)
		if tok == ok {
			return "ok", http.StatusOK, nil
		}
		return "refused", http.StatusUnauthorized, nil
	}
}

func TestAuthorized(t *testing.T) {
	t.Run("renews a refused token once and reports the response", func(t *testing.T) {
		var saw []string
		done := 0
		ctx := apicall.WithObserver(t.Context(), apicall.Observer{Done: func() { done++ }})
		src := &source{fresh: "new"}
		got, err := transport.Authorized(ctx, src, "old", sender("new", &saw))
		require.NoError(t, err)
		assert.Equal(t, "ok", got)
		assert.Equal(t, []string{"old", "new"}, saw)
		assert.Equal(t, []string{"old"}, src.stale)
		assert.Equal(t, 1, done)
	})
	t.Run("no newer token keeps the refused response", func(t *testing.T) {
		var saw []string
		src := &source{fresh: "old"}
		got, err := transport.Authorized(t.Context(), src, "old", sender("new", &saw))
		require.NoError(t, err)
		assert.Equal(t, "refused", got)
		assert.Equal(t, []string{"old"}, saw)
	})
	t.Run("a failed renew keeps the refused response", func(t *testing.T) {
		var saw []string
		src := &source{err: errors.New("mint failed")}
		got, err := transport.Authorized(t.Context(), src, "old", sender("new", &saw))
		require.NoError(t, err)
		assert.Equal(t, "refused", got)
		assert.Equal(t, []string{"old"}, saw)
	})
	t.Run("a send error is not renewed or reported", func(t *testing.T) {
		done := 0
		ctx := apicall.WithObserver(t.Context(), apicall.Observer{Done: func() { done++ }})
		src := &source{fresh: "new"}
		boom := errors.New("boom")
		_, err := transport.Authorized(ctx, src, "old", func(string) (string, int, error) { return "", http.StatusUnauthorized, boom })
		require.ErrorIs(t, err, boom)
		assert.Empty(t, src.stale)
		assert.Zero(t, done)
	})
}

func TestReadLimited(t *testing.T) {
	b, over, err := transport.ReadLimited(strings.NewReader("abcd"), 4)
	require.NoError(t, err)
	assert.False(t, over)
	assert.Equal(t, "abcd", string(b))

	_, over, err = transport.ReadLimited(strings.NewReader("abcde"), 4)
	require.NoError(t, err)
	assert.True(t, over)
}
