package github_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/apicall"
	"github.com/elee1766/jobshop/pkg/github"
)

// renewing is a source whose token is refused until renewed.
type renewing struct {
	tok   string
	stale []string
}

func (r *renewing) Token(context.Context) (string, error) { return r.tok, nil }

func (r *renewing) Renew(_ context.Context, stale string) (string, error) {
	r.stale = append(r.stale, stale)
	r.tok = "fresh"
	return r.tok, nil
}

// One client renews a refused token once, follows Link pages on the API
// host, returns a 304 for an ETag, refuses a URL off the API and reports
// each token it sends to the caller's observer.
func TestClient(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fresh" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"message":"Bad credentials"}`)
			return
		}
		switch {
		case r.Header.Get("If-None-Match") == `W/"1"`:
			w.WriteHeader(http.StatusNotModified)
		case r.URL.Query().Get("page") == "2":
			_, _ = io.WriteString(w, `[3]`)
		case r.URL.Path == "/repos/o/r/issues":
			w.Header().Set("ETag", `W/"1"`)
			w.Header().Set("Link", "<"+srv.URL+"/repos/o/r/issues?page=2>; rel=\"next\"")
			_, _ = io.WriteString(w, `[1,2]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	src := &renewing{tok: "revoked"}
	cl := github.New(srv.URL, src)
	var sent []string
	ctx := apicall.WithObserver(t.Context(), apicall.Observer{Token: func(tok string) { sent = append(sent, tok) }})

	items, resp, err := cl.List(ctx, github.Request{Path: "/repos/o/r/issues"}, 5)
	require.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, `W/"1"`, resp.Header.Get("ETag"))
	assert.Equal(t, []string{"revoked"}, src.stale, "renewed once, naming the refused token")
	assert.Equal(t, []string{"revoked", "fresh", "fresh"}, sent)

	resp, err = cl.Do(ctx, github.Request{Path: "/repos/o/r/issues", ETag: `W/"1"`})
	require.NoError(t, err)
	assert.True(t, resp.NotModified())

	_, err = cl.Do(ctx, github.Request{Path: "/repos/o/r/missing"})
	assert.True(t, github.IsNotFound(err))

	_, err = cl.Do(ctx, github.Request{Path: "https://elsewhere.example/repos/o/r"})
	require.ErrorContains(t, err, "leaves the API")
}

// GraphQL turns a rate limit into a *StatusError that says when to retry,
// and tells it from a 403 for a permission.
func TestGraphQLStatusError(t *testing.T) {
	for name, tc := range map[string]struct {
		status  int
		header  map[string]string
		body    string
		limited bool
		retry   bool
	}{
		"429 with retry-after":    {429, map[string]string{"Retry-After": "60"}, `{"message":"slow down"}`, true, true},
		"403 primary limit":       {403, map[string]string{"X-RateLimit-Remaining": "0", "X-RateLimit-Reset": "4102444800"}, `{"message":"API rate limit exceeded"}`, true, true},
		"403 secondary limit":     {403, nil, `{"message":"You have exceeded a secondary rate limit"}`, true, false},
		"403 permission":          {403, nil, `{"message":"Resource not accessible by integration"}`, false, false},
		"200 graphql rate limit":  {200, nil, `{"errors":[{"type":"RATE_LIMITED","message":"API rate limit exceeded"}]}`, true, false},
		"502 without a json body": {502, nil, `<html>`, false, false},
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				for k, v := range tc.header {
					w.Header().Set(k, v)
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer srv.Close()
			err := github.New(srv.URL, apicall.Static("t")).GraphQL(t.Context(), `query{viewer{login}}`, nil, &struct{}{})
			var se *github.StatusError
			require.ErrorAs(t, err, &se)
			assert.Equal(t, tc.limited, se.RateLimited())
			assert.Equal(t, tc.retry, !se.RetryAt.IsZero())
		})
	}
}
