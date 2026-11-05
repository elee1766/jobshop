package store

import "context"

// Close closes ticket id with resolution, fenced by c when non-nil. A
// helper over Release (#579), not a Store method.
func Close(ctx context.Context, s Store, actor, id string, c *Claim, resolution string) (Ticket, error) {
	return s.Release(ctx, actor, id, c, Outcome{Close: resolution})
}
