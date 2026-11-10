// Package transport is the request mechanics the GitHub and Linear clients
// share: renewing a refused token once, telling the caller's observer a
// response was read, and bounding a response body. It is internal to
// pkg/integrations. What differs between the providers (URLs, headers,
// GraphQL shapes, error types, redirect policy) stays in each client.
package transport

import (
	"context"
	"io"
	"net/http"

	"github.com/gfx-labs/jobshop/pkg/integrations/apicall"
)

// Send sends one request with tok. status is the HTTP status of res, which
// Authorized reads to spot a refused token. Only a nil err means res is a
// response.
type Send[R any] func(tok string) (res R, status int, err error)

// Authorized sends with tok. When the provider answers 401 it asks src to
// renew the refused token and sends once more, only if the token changed: a
// source with nothing newer, or one whose renew fails, leaves the refused
// response as the answer. A response read without an error is reported to
// ctx's apicall observer.
func Authorized[R any](ctx context.Context, src apicall.Source, tok string, send Send[R]) (R, error) {
	res, status, err := send(tok)
	if err == nil && status == http.StatusUnauthorized {
		if fresh, rerr := src.Renew(ctx, tok); rerr == nil && fresh != tok {
			res, _, err = send(fresh)
		}
	}
	if err == nil {
		apicall.Received(ctx)
	}
	return res, err
}

// ReadLimited reads at most limit bytes of r. over reports that r held more,
// and the body is then one byte past the limit and not a whole response.
func ReadLimited(r io.Reader, limit int) (body []byte, over bool, err error) {
	body, err = io.ReadAll(io.LimitReader(r, int64(limit)+1))
	return body, err == nil && len(body) > limit, err
}
