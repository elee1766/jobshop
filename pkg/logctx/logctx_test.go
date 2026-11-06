package logctx_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elee1766/jobshop/pkg/logctx"
)

func TestHandlerAddsContextAttrs(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(logctx.NewHandler(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}))).With("component", "x")
	ctx := logctx.With(t.Context(), "run", "r1")
	ctx = logctx.With(ctx, "ticket", "t1")
	ctx = logctx.With(ctx, "run", "r2")
	log.InfoContext(ctx, "hello", "n", 1)
	assert.Equal(t, "level=INFO msg=hello component=x n=1 ticket=t1 run=r2\n", buf.String())

	buf.Reset()
	log.InfoContext(t.Context(), "plain")
	assert.Equal(t, "level=INFO msg=plain component=x\n", buf.String())
}
