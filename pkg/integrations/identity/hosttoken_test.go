package identity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/integrations/identity"
)

// A cached token is kept until the API refuses it, then read again.
func TestHostToken(t *testing.T) {
	t.Setenv("GH_TOKEN", "first")
	var h identity.HostToken
	tok, err := h.Token(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "first", tok)

	t.Setenv("GH_TOKEN", "second")
	tok, err = h.Token(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "first", tok, "cached")

	tok, err = h.Renew(t.Context(), "other")
	require.NoError(t, err)
	assert.Equal(t, "first", tok, "a different stale token leaves the cache")

	tok, err = h.Renew(t.Context(), "first")
	require.NoError(t, err)
	assert.Equal(t, "second", tok)
}
