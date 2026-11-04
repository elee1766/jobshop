package ingest

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

var hintEndpoint = config.IngestEndpoint{
	Name: "gh-hints", Source: config.IngestGeneric,
	SignatureHeader: "X-Hint-Signature", SignaturePrefix: "sha256=", DeliveryIDHeader: "X-Hint-Delivery",
	Map: config.IngestMap{
		Kind: "event", RepoOrTeam: "repo", Number: "prs.0", Actor: "sender.login",
		NaturalKey: "{repo}:{head}",
	},
}

const hint = `{"delivery":"d-1","event":"check_suite","action":"completed","repo":"acme/app","prs":[42],"head":"abc123","sender":{"login":"ci"}}`

func genericHandler(t *testing.T) (*Handler, *Spool) {
	t.Helper()
	sp := openSpool(t, filepath.Join(t.TempDir(), "spool.db"))
	return &Handler{
		Spool: sp, Endpoints: []config.IngestEndpoint{hintEndpoint},
		Secret: func(config.IngestEndpoint) (Secret, error) { return Secret{Key: []byte("hint-secret")}, nil },
		Now:    func() time.Time { return time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC) },
	}, sp
}

// TestGenericEndpoint: a signed body is spooled (202), a bad signature
// is 401, a repeat delivery id dedupes, and the drain maps the payload
// to inbox fields (#399).
func TestGenericEndpoint(t *testing.T) {
	h, sp := genericHandler(t)
	body := []byte(hint)
	hdr := map[string]string{"X-Hint-Signature": "sha256=" + mac("hint-secret", body), "X-Hint-Delivery": "d-1"}
	assert.Equal(t, 202, post(h, "/generic/gh-hints", body, hdr))
	assert.Equal(t, 202, post(h, "/generic/gh-hints", body, hdr), "repeat is accepted")
	assert.Equal(t, 401, post(h, "/generic/gh-hints", body, map[string]string{"X-Hint-Signature": "sha256=" + mac("wrong", body)}))
	assert.Equal(t, 401, post(h, "/generic/gh-hints", body, nil))
	assert.Equal(t, 400, post(h, "/generic/gh-hints", []byte("nope"), map[string]string{"X-Hint-Signature": "sha256=" + mac("hint-secret", []byte("nope"))}))
	ds := rows(t, sp)
	require.Len(t, ds, 1, "deduped by delivery id")
	assert.Equal(t, "d-1", ds[0].DeliveryID)
	assert.Equal(t, "check_suite", ds[0].Kind)

	rs, err := sp.Rows(t.Context(), 0)
	require.NoError(t, err)
	it, err := InboxItem(rs[0])
	require.NoError(t, err)
	assert.Equal(t, store.InboxItem{
		Source: SourceGeneric, DeliveryID: "generic:d-1", Endpoint: "gh-hints", Kind: "check_suite",
		RepoOrTeam: "acme/app", Number: 42, Actor: "ci", NaturalKey: "generic:gh-hints:acme/app:abc123",
		ReceivedAt: it.ReceivedAt, Payload: it.Payload,
	}, it)
}

// TestGenericDrain: generic rows reach the inbox; a missing delivery id
// header falls back to the body hash.
func TestGenericDrain(t *testing.T) {
	sp, st := drainFixture(t)
	ctx := t.Context()
	ep := hintEndpoint
	ep.Map = config.IngestMap{Kind: "event"}
	fresh, err := Put(ctx, sp, ep, "", "", []byte(`{"event":"ping"}`), time.Now())
	require.NoError(t, err)
	assert.True(t, fresh)
	res, err := Drain(ctx, sp, st, time.Now, quiet)
	require.NoError(t, err)
	assert.Equal(t, DrainResult{Moved: 1, Fresh: 1}, res)
	got, err := st.InboxList(ctx, store.InboxFilter{})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, SourceGeneric, got[0].Source)
	assert.Equal(t, "ping", got[0].Kind)
	assert.Contains(t, got[0].DeliveryID, "generic:sha256-")
}

// TestPut: idempotent per delivery id, the kind flag wins over the map,
// only generic endpoints, and a spool error fails.
func TestPut(t *testing.T) {
	ctx := t.Context()
	sp := openSpool(t, filepath.Join(t.TempDir(), "s.db"))
	for i, want := range []bool{true, false} {
		fresh, err := Put(ctx, sp, hintEndpoint, "x-1", "custom", []byte(hint), time.Now())
		require.NoError(t, err, i)
		assert.Equal(t, want, fresh)
	}
	ds := rows(t, sp)
	require.Len(t, ds, 1)
	assert.Equal(t, "custom", ds[0].Kind)
	_, err := Put(ctx, sp, config.IngestEndpoint{Name: "gh", Source: config.IngestGitHub}, "a", "", []byte(`{}`), time.Now())
	require.ErrorContains(t, err, "generic endpoints")
	_, err = Put(ctx, sp, hintEndpoint, "a", "", []byte(`not json`), time.Now())
	require.ErrorIs(t, err, ErrBadDelivery)
	require.NoError(t, sp.Close())
	_, err = Put(ctx, sp, hintEndpoint, "y", "", []byte(hint), time.Now())
	require.Error(t, err, "a spool write failure is an error, so the caller does not ack")
}

func TestLookup(t *testing.T) {
	p := map[string]any{"a": map[string]any{"b": []any{"x", 3.0, true, nil, map[string]any{"c": 1.5}}}}
	for path, want := range map[string]string{
		"a.b.0": "x", "a.b.1": "3", "a.b.2": "true", "a.b.3": "", "a.b.4.c": "1.5",
		"a.b.9": "", "a.z": "", "": "", "a.b.x": "", "a.b.0.q": "", "a.b.4": `{"c":1.5}`,
	} {
		assert.Equal(t, want, lookup(p, path), path)
	}
}

// TestSpoolV2Upgrade: a v2 spool with rows gains the generic source
// and keeps its rows.
func TestSpoolV2Upgrade(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.db")
	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), schema)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), schemaV2)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), `INSERT INTO spool (endpoint, source, delivery_id, received_at, payload) VALUES ('e', 'github', 'old', 1, '{}')`)
	require.NoError(t, err)
	require.NoError(t, db.Close())
	sp := openSpool(t, path)
	fresh, err := Put(t.Context(), sp, hintEndpoint, "new", "", []byte(hint), time.Now())
	require.NoError(t, err)
	assert.True(t, fresh)
	assert.Len(t, rows(t, sp), 2)
}
