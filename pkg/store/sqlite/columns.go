package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"sort"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store/sqlite/sqldb"
	"github.com/elee1766/jobshop/pkg/store/storerules"
)

// fieldCols are storerules.Columns as the SQL parameter types.
type fieldCols struct {
	key, to, state, workflow sql.NullString
	index                    sql.NullInt64
	waiting                  int64
}

func fieldColumns(fields json.RawMessage) fieldCols {
	r := storerules.FieldColumns(fields)
	str := func(p *string) sql.NullString {
		if p == nil {
			return sql.NullString{}
		}
		return sql.NullString{String: *p, Valid: true}
	}
	c := fieldCols{key: str(r.Key), to: str(r.To), state: str(r.State), workflow: str(r.Workflow)}
	if r.Index != nil {
		c.index = sql.NullInt64{Int64: *r.Index, Valid: true}
	}
	if r.Waiting {
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
		if err := q.AddTicketResource(ctx, sqldb.AddTicketResourceParams{TicketID: id, Rkey: k, Rname: domain.ResourceName(k)}); err != nil {
			return err
		}
	}
	return nil
}
