// Package logctx carries log attributes (run, ticket, session, pool ids)
// in a context, so call sites that log with a ctx do not repeat them
// (JOB-49).
package logctx

import (
	"context"
	"log/slog"
	"slices"
)

type key struct{}

// With returns ctx carrying the attribute k=v. A later With of the same
// key replaces the earlier value.
func With(ctx context.Context, k string, v any) context.Context {
	old := attrs(ctx)
	out := make([]slog.Attr, 0, len(old)+1)
	for _, a := range old {
		if a.Key != k {
			out = append(out, a)
		}
	}
	out = append(out, slog.Any(k, v))
	return context.WithValue(ctx, key{}, out)
}

// Attrs returns the attributes stored in ctx.
func Attrs(ctx context.Context) []slog.Attr { return slices.Clone(attrs(ctx)) }

func attrs(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	a, _ := ctx.Value(key{}).([]slog.Attr)
	return a
}

// Handler adds the attributes stored in the record's context.
type Handler struct{ slog.Handler }

// NewHandler wraps h. An h that is already a Handler is unwrapped first,
// so stacking wrappers does not add the attributes twice.
func NewHandler(h slog.Handler) Handler { return Handler{Unwrap(h)} }

// Unwrap returns the handler inside h if h is a Handler, else h. A caller
// that adds its own wrapper further out uses it to avoid adding the
// attributes twice.
func Unwrap(h slog.Handler) slog.Handler {
	if in, ok := h.(Handler); ok {
		return in.Handler
	}
	return h
}

// Handle adds the ctx attributes after the record's own.
func (h Handler) Handle(ctx context.Context, r slog.Record) error {
	if as := attrs(ctx); len(as) > 0 {
		r = r.Clone()
		r.AddAttrs(as...)
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs keeps the wrapper.
func (h Handler) WithAttrs(as []slog.Attr) slog.Handler { return Handler{h.Handler.WithAttrs(as)} }

// WithGroup keeps the wrapper.
func (h Handler) WithGroup(n string) slog.Handler { return Handler{h.Handler.WithGroup(n)} }
