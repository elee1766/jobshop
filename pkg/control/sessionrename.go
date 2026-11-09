package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// maxDisplayName caps a session display name, in runes.
const maxDisplayName = 120

// sessionRenameRPC sets a session's display name (#626). A session may
// always rename itself; another session needs admin or the
// session-rename grant.
func (c *Controller) sessionRenameRPC(ctx context.Context, sc scope, p rpc.SessionRenameParams) (any, error) {
	if p.ID == "" {
		p.ID = sc.session
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", domain.ErrInvalid)
	}
	if p.ID != sc.session {
		if err := c.requireGrant(sc, "session-rename", "session rename of another session"); err != nil {
			return nil, err
		}
	}
	name := strings.Join(strings.Fields(p.Name), " ")
	if utf8.RuneCountInString(name) > maxDisplayName {
		return nil, fmt.Errorf("%w: name is longer than %d characters", domain.ErrInvalid, maxDisplayName)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != domain.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", domain.ErrInvalid, p.ID, t.Kind)
	}
	if t.Status == domain.StatusClosed {
		return nil, fmt.Errorf("%w: session %s", domain.ErrClosed, p.ID)
	}
	if err := c.setDisplayName(ctx, sc.actor, p.ID, name); err != nil {
		return nil, err
	}
	return c.sessionInfo(ctx, p.ID)
}

// setDisplayName writes the name with a revision check, so a concurrent
// fields write by the runner is not lost.
func (c *Controller) setDisplayName(ctx context.Context, actor, id, name string) error {
	for range 5 {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return err
		}
		f, err := domain.DecodeFields[domain.SessionFields](t)
		if err != nil {
			return err
		}
		f.DisplayName = name
		b, err := json.Marshal(f)
		if err != nil {
			return err
		}
		_, err = c.store.Update(ctx, actor, id, store.Update{Fields: b, IfRevision: t.Revision})
		if !errors.Is(err, domain.ErrConflict) {
			return err
		}
	}
	return fmt.Errorf("%w: session %s kept changing", domain.ErrConflict, id)
}
