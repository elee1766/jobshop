package journal

// Recall over session journals (#265) on the journals store collection
// (#579). Each session has one record, keyed by session id, holding a
// summary of its journal; the controller rewrites it after every turn.
// Text feeds the collection's full-text index; ticket and project are
// indexed for recall's filters.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
)

// Spec is the journals collection.
var Spec = store.CollectionSpec{
	Name: "journals",
	Index: []store.Field{
		{Name: "ticket", Path: "ticket"},
		{Name: "project", Path: "project"},
	},
	Search: []store.Field{{Name: "text", Path: "text"}},
}

// Store is the part of the store recall needs.
type Store interface {
	Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error)
}

// Doc is one session's journal summary.
type Doc struct {
	Session string
	Ticket  string
	Project string
	At      time.Time // the newest journal frame
	Text    string
}

// doc is the record's JSON.
type doc struct {
	Ticket  string `json:"ticket"`
	Project string `json:"project"`
	At      int64  `json:"at"` // unix ms
	Text    string `json:"text"`
}

// putTries bounds the compare-and-set loop under races.
const putTries = 10

// Index writes d as its session's summary, replacing an older one. An
// empty Text removes the session's record.
func Index(ctx context.Context, st Store, d Doc) error {
	if d.Session == "" {
		return fmt.Errorf("%w: journal needs a session", store.ErrInvalid)
	}
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return err
	}
	if strings.TrimSpace(d.Text) == "" {
		_, err := col.Delete(ctx, store.RecordFilter{Keys: []string{d.Session}})
		return err
	}
	b, err := json.Marshal(doc{Ticket: d.Ticket, Project: d.Project, At: d.At.UnixMilli(), Text: d.Text})
	if err != nil {
		return err
	}
	for range putTries {
		r, created, err := col.Put(ctx, store.Record{Key: d.Session, Value: b, CreatedAt: d.At})
		if err != nil || created {
			return err
		}
		r.Value = b
		if _, err = col.Update(ctx, r); !errors.Is(err, store.ErrConflict) {
			return err
		}
	}
	return fmt.Errorf("%w: journal %s: lost every race", store.ErrConflict, d.Session)
}

// snippetMax bounds a recall snippet.
const snippetMax = 200

// Recall searches session journals, best match first. project and
// ticketID narrow the hits when set.
func Recall(ctx context.Context, st Store, text, project, ticketID string, limit int) ([]store.RecallHit, error) {
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	// Filters run here, after the search: fetch more than asked.
	rs, err := col.Search(ctx, text, limit*10)
	if err != nil {
		return nil, err
	}
	out := []store.RecallHit{}
	for _, r := range rs {
		var v doc
		if err := json.Unmarshal(r.Value, &v); err != nil {
			return nil, fmt.Errorf("journal %s: %w", r.Key, err)
		}
		if (project != "" && v.Project != project) || (ticketID != "" && v.Ticket != ticketID) {
			continue
		}
		out = append(out, store.RecallHit{
			Kind: store.RecallJournal, Ref: r.Key, TicketID: v.Ticket, Project: v.Project,
			At: time.UnixMilli(v.At).UTC(), Snippet: snippet(v.Text, text), Score: float64(len(out)),
		})
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

// snippet returns up to snippetMax bytes of text around the first query
// word it contains, or its start.
func snippet(text, query string) string {
	lower := strings.ToLower(text)
	at := 0
	for _, w := range strings.Fields(strings.ToLower(query)) {
		if i := strings.Index(lower, w); i >= 0 {
			at = i
			break
		}
	}
	start := max(0, at-snippetMax/4)
	end := min(len(text), start+snippetMax)
	s := strings.TrimSpace(strings.ToValidUTF8(text[start:end], ""))
	if start > 0 {
		s = "..." + s
	}
	if end < len(text) {
		s += "..."
	}
	return s
}

// Load returns the indexed summary of session's journal, or "" when it
// has none.
func Load(ctx context.Context, st Store, session string) (string, error) {
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return "", err
	}
	r, err := col.Get(ctx, session)
	if errors.Is(err, store.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var v doc
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return "", fmt.Errorf("journal %s: %w", session, err)
	}
	return v.Text, nil
}
