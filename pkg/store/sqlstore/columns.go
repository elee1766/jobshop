package sqlstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"sort"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqldb"
)

// columns are the fields keys queries filter and sort on, copied to
// plain columns so SQL needs no JSON functions (#405). A key that is
// absent, or not the expected JSON type, is NULL, as json_extract was.
type columns struct {
	key, to, state, workflow sql.NullString
	index                    sql.NullInt64
	waiting                  int64
}

func fieldColumns(fields json.RawMessage) columns {
	var c columns
	var m map[string]json.RawMessage
	if json.Unmarshal(fields, &m) != nil {
		return c
	}
	str := func(k string) sql.NullString {
		var s string
		if v, ok := m[k]; ok && json.Unmarshal(v, &s) == nil && bytes.HasPrefix(bytes.TrimSpace(v), []byte(`"`)) {
			return sql.NullString{String: s, Valid: true}
		}
		return sql.NullString{}
	}
	c.key, c.to, c.state, c.workflow = str("key"), str("to"), str("state"), str("workflow")
	var f float64
	if v, ok := m["index"]; ok && json.Unmarshal(v, &f) == nil {
		c.index = sql.NullInt64{Int64: int64(f), Valid: true}
	}
	if v, ok := m["wait"]; ok && !bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
		c.waiting = 1
	}
	return c
}

// syncMeta rewrites ticket_meta, the lookup rows for List's meta filter,
// from meta: one row per string value meta[pack][key].
func syncMeta(ctx context.Context, q sqldb.Querier, id string, meta json.RawMessage) error {
	if err := q.ClearTicketMeta(ctx, id); err != nil {
		return err
	}
	var m map[string]map[string]json.RawMessage
	if len(meta) == 0 || json.Unmarshal(meta, &m) != nil {
		return nil
	}
	packs := make([]string, 0, len(m))
	for p := range m {
		packs = append(packs, p)
	}
	sort.Strings(packs)
	for _, p := range packs {
		for k, v := range m[p] {
			var s string
			if !bytes.HasPrefix(bytes.TrimSpace(v), []byte(`"`)) || json.Unmarshal(v, &s) != nil {
				continue
			}
			if err := q.AddTicketMeta(ctx, sqldb.AddTicketMetaParams{TicketID: id, Pack: p, Mkey: k, Value: s}); err != nil {
				return err
			}
		}
	}
	return nil
}

// syncResources rewrites ticket_resources from fields.resources (#426),
// so the slot queries join a table instead of reading JSON.
func syncResources(ctx context.Context, q sqldb.Querier, id string, fields json.RawMessage) error {
	if err := q.ClearTicketResources(ctx, id); err != nil {
		return err
	}
	var f struct {
		Resources []string `json:"resources"`
	}
	if len(fields) == 0 || json.Unmarshal(fields, &f) != nil {
		return nil
	}
	for _, k := range f.Resources {
		if err := q.AddTicketResource(ctx, sqldb.AddTicketResourceParams{TicketID: id, Rkey: k, Rname: store.ResourceName(k)}); err != nil {
			return err
		}
	}
	return nil
}
