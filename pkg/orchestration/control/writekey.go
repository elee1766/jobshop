package control

import (
	"context"
	"fmt"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store/writekey"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Idempotency keys for tracker writes (#280). The CLI claims a key before
// it sends a write and stores the result after, so a retry after an
// unknown outcome returns the stored record instead of posting twice.
// Keys are stored per project: a session's key cannot collide with
// another project's.

func writeKeyName(sc scope, key string) string {
	p := sc.project
	if p == "" {
		p = "admin"
	}
	return p + "/" + strings.TrimSpace(key)
}

func (c *Controller) writeKeyClaim(ctx context.Context, sc scope, p rpc.WriteKeyParams) (any, error) {
	if strings.TrimSpace(p.Key) == "" {
		return nil, fmt.Errorf("%w: writekey.claim needs a key", domain.ErrInvalid)
	}
	k, err := writekey.Claim(ctx, c.store, writeKeyName(sc, p.Key), p.Op, sc.actor)
	if err != nil {
		return nil, err
	}
	k.Scoped = k.Key
	k.Key = p.Key
	return k, nil
}

func (c *Controller) writeKeyFinish(ctx context.Context, sc scope, p rpc.WriteKeyParams) (any, error) {
	if strings.TrimSpace(p.Key) == "" || p.Result == "" {
		return nil, fmt.Errorf("%w: writekey.finish needs a key and a result", domain.ErrInvalid)
	}
	k, err := writekey.Finish(ctx, c.store, writeKeyName(sc, p.Key), p.Result)
	if err != nil {
		return nil, err
	}
	k.Key = p.Key
	return k, nil
}
