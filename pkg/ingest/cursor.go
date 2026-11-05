package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
)

// Poller cursors (#341) live in the poll_cursors store collection (#579),
// keyed "poller:repo:stream". Since is unix ms, as the old table kept it.

// CursorSpec is the poll_cursors collection.
var CursorSpec = store.CollectionSpec{Name: "poll_cursors"}

type cursorValue struct {
	Poller string `json:"poller"`
	Repo   string `json:"repo"`
	Stream string `json:"stream"`
	Since  int64  `json:"since"`
	ETag   string `json:"etag"`
}

func cursorKey(poller, repo, stream string) string { return poller + ":" + repo + ":" + stream }

// cursorTries bounds SetPollCursor's put-or-update loop under races.
const cursorTries = 10

// GetPollCursor returns a poller's cursor, or Found false.
func GetPollCursor(ctx context.Context, st PollStore, poller, repo, stream string) (store.PollCursor, error) {
	out := store.PollCursor{Poller: poller, Repo: repo, Stream: stream}
	col, err := st.Collection(ctx, CursorSpec)
	if err != nil {
		return out, err
	}
	r, err := col.Get(ctx, cursorKey(poller, repo, stream))
	if errors.Is(err, store.ErrNotFound) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	var v cursorValue
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return out, fmt.Errorf("poll cursor %s: %w", r.Key, err)
	}
	if v.Since != 0 {
		out.Since = time.UnixMilli(v.Since).UTC()
	}
	out.ETag, out.Found = v.ETag, true
	return out, nil
}

// SetPollCursor stores a poller's cursor.
func SetPollCursor(ctx context.Context, st PollStore, c store.PollCursor) error {
	col, err := st.Collection(ctx, CursorSpec)
	if err != nil {
		return err
	}
	v := cursorValue{Poller: c.Poller, Repo: c.Repo, Stream: c.Stream, ETag: c.ETag}
	if !c.Since.IsZero() {
		v.Since = c.Since.UnixMilli()
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	key := cursorKey(c.Poller, c.Repo, c.Stream)
	for range cursorTries {
		r, created, err := col.Put(ctx, store.Record{Key: key, Value: b})
		if err != nil || created {
			return err
		}
		r.Value = b
		_, err = col.Update(ctx, r)
		if !errors.Is(err, store.ErrConflict) {
			return err
		}
	}
	return fmt.Errorf("%w: poll cursor %s: lost every race", store.ErrConflict, key)
}
