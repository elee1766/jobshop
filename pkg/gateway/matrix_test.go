package gateway

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMatrixPostRetries: a 429 on send is retried, so the reply is
// delivered instead of dropped (#196).
func TestMatrixPostRetries(t *testing.T) {
	var sends atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/account/whoami"):
			_, _ = w.Write([]byte(`{"user_id":"@bot:x","device_id":"D"}`))
		case strings.Contains(r.URL.Path, "/send/"):
			if sends.Add(1) == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"errcode":"M_LIMIT_EXCEEDED","error":"slow down"}`))
				return
			}
			_, _ = w.Write([]byte(`{"event_id":"$e"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	m, err := NewMatrix(t.Context(), srv.URL, "tok", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, matrixRetries, m.cli.DefaultHTTPRetries)
	m.cli.DefaultHTTPBackoff = 10 * time.Millisecond
	require.NoError(t, m.Post(t.Context(), Thread{Room: "!r:x", Root: "$root"}, "hi"))
	assert.Equal(t, int32(2), sends.Load())
}
