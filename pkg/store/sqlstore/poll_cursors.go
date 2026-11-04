package sqlstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// PollCursor returns a poller's cursor, or Found false.
func (s *Store) PollCursor(ctx context.Context, poller, repo, stream string) (store.PollCursor, error) {
	out := store.PollCursor{Poller: poller, Repo: repo, Stream: stream}
	r, err := s.q.GetPollCursor(ctx, sqlcdb.GetPollCursorParams{Poller: poller, Repo: repo, Stream: stream})
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.Since, out.ETag, out.Found = fromMS(r.Since), r.Etag, true
	return out, nil
}

// SetPollCursor stores a poller's cursor.
func (s *Store) SetPollCursor(ctx context.Context, c store.PollCursor) error {
	return s.q.UpsertPollCursor(ctx, sqlcdb.UpsertPollCursorParams{
		Poller: c.Poller, Repo: c.Repo, Stream: c.Stream, Since: ms(c.Since), Etag: c.ETag, Now: ms(s.now()),
	})
}
