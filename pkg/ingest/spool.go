// Package ingest receives webhook deliveries and keeps them in a spool
// database until the controller drains them into the inbox (#318).
package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // database/sql driver "sqlite"
)

// Sources.
const (
	SourceGitHub = "github"
	SourceLinear = "linear"
)

// Delivery is one spooled webhook delivery.
type Delivery struct {
	ID         int64     `json:"id"`
	Endpoint   string    `json:"endpoint"`
	Source     string    `json:"source"`
	DeliveryID string    `json:"delivery_id"`
	Kind       string    `json:"kind"`
	ReceivedAt time.Time `json:"received_at"`
	// Headers are the signature, event and delivery headers only.
	Headers map[string]string `json:"headers"`
	Payload json.RawMessage   `json:"payload"`
}

// ErrBadDelivery is a delivery the spool will not store.
var ErrBadDelivery = errors.New("ingest: bad delivery")

// schemaVersion is PRAGMA user_version once schema is applied. One
// table needs no migration tool; a later change bumps this and adds a
// step in OpenSpool.
const schemaVersion = 2

const schema = `
CREATE TABLE IF NOT EXISTS spool (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  endpoint    TEXT NOT NULL,
  source      TEXT NOT NULL CHECK (source IN ('github', 'linear')),
  delivery_id TEXT NOT NULL,
  kind        TEXT NOT NULL DEFAULT '',
  received_at INTEGER NOT NULL,
  headers     TEXT NOT NULL DEFAULT '{}',
  payload     TEXT NOT NULL,
  UNIQUE (source, delivery_id)
);
PRAGMA user_version = 1;
`

// schemaV2 adds spool_bad (#335): rows the drain could not parse, kept
// with the error for a person to look at instead of being dropped.
const schemaV2 = `
CREATE TABLE IF NOT EXISTS spool_bad (
  id          INTEGER PRIMARY KEY,
  endpoint    TEXT NOT NULL,
  source      TEXT NOT NULL,
  delivery_id TEXT NOT NULL,
  kind        TEXT NOT NULL DEFAULT '',
  received_at INTEGER NOT NULL,
  headers     TEXT NOT NULL,
  payload     TEXT NOT NULL,
  error       TEXT NOT NULL,
  moved_at    INTEGER NOT NULL
);
PRAGMA user_version = 2;
`

// Spool is the receiver's own SQLite database (#339). Only the receiver
// writes rows and only the drain deletes them, so the main jobshop.db
// keeps one writer and the receiver works while the controller is down.
type Spool struct{ db *sql.DB }

// OpenSpool opens or creates the spool at path. synchronous=FULL makes a
// committed Write durable before the handler answers 202.
func OpenSpool(ctx context.Context, path string) (*Spool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	dsn := "file:" + path + "?_txlock=immediate" +
		"&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	var v int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&v); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ingest spool %s: %w", path, err)
	}
	if v > schemaVersion {
		_ = db.Close()
		return nil, fmt.Errorf("ingest spool %s: schema version %d is newer than this binary (%d)", path, v, schemaVersion)
	}
	if v < 1 {
		if _, err := db.ExecContext(ctx, schema); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("ingest spool %s: %w", path, err)
		}
	}
	if v < 2 {
		if _, err := db.ExecContext(ctx, schemaV2); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("ingest spool %s: %w", path, err)
		}
	}
	return &Spool{db: db}, nil
}

// Close closes the database.
func (s *Spool) Close() error { return s.db.Close() }

// Write stores d. It returns false when a delivery with the same source
// and delivery id is already spooled; nothing changes then.
func (s *Spool) Write(ctx context.Context, d Delivery) (bool, error) {
	if (d.Source != SourceGitHub && d.Source != SourceLinear) || d.DeliveryID == "" || d.Endpoint == "" {
		return false, fmt.Errorf("%w: endpoint %q source %q id %q", ErrBadDelivery, d.Endpoint, d.Source, d.DeliveryID)
	}
	hdr, err := json.Marshal(d.Headers)
	if err != nil {
		return false, err
	}
	payload := string(d.Payload)
	if payload == "" {
		payload = "{}"
	}
	if d.ReceivedAt.IsZero() {
		d.ReceivedAt = time.Now()
	}
	res, err := s.db.ExecContext(ctx, `INSERT INTO spool (endpoint, source, delivery_id, kind, received_at, headers, payload)
VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT (source, delivery_id) DO NOTHING`,
		d.Endpoint, d.Source, d.DeliveryID, d.Kind, d.ReceivedAt.UnixMilli(), string(hdr), payload)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

// Rows returns up to limit rows, oldest first, as stored: parsing is
// the drain's job so it can move a bad row aside. limit <= 0 is 100.
func (s *Spool) Rows(ctx context.Context, limit int) ([]Row, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, endpoint, source, delivery_id, kind, received_at, headers, payload
FROM spool ORDER BY id LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.ID, &r.Endpoint, &r.Source, &r.DeliveryID, &r.Kind, &r.ReceivedAt, &r.Headers, &r.Payload); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Row is one spool row as stored.
type Row struct {
	ID                                 int64
	Endpoint, Source, DeliveryID, Kind string
	ReceivedAt                         int64 // Unix ms
	Headers, Payload                   string
}

// Delivery parses r.
func (r Row) Delivery() (Delivery, error) {
	d := Delivery{
		ID: r.ID, Endpoint: r.Endpoint, Source: r.Source, DeliveryID: r.DeliveryID, Kind: r.Kind,
		ReceivedAt: time.UnixMilli(r.ReceivedAt).UTC(),
	}
	if err := json.Unmarshal([]byte(r.Headers), &d.Headers); err != nil {
		return d, fmt.Errorf("headers: %w", err)
	}
	if !json.Valid([]byte(r.Payload)) {
		return d, errors.New("payload is not JSON")
	}
	d.Payload = json.RawMessage(r.Payload)
	return d, nil
}

// MarkBad moves row id to spool_bad with reason, in one transaction.
func (s *Spool) MarkBad(ctx context.Context, id int64, reason string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `INSERT OR REPLACE INTO spool_bad (id, endpoint, source, delivery_id, kind, received_at, headers, payload, error, moved_at)
SELECT id, endpoint, source, delivery_id, kind, received_at, headers, payload, ?, ? FROM spool WHERE id = ?`, reason, now.UnixMilli(), id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM spool WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// BadCount is the number of rows in spool_bad.
func (s *Spool) BadCount(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM spool_bad`).Scan(&n)
	return n, err
}

// List returns up to limit deliveries, oldest first. limit <= 0 is 100.
func (s *Spool) List(ctx context.Context, limit int) ([]Delivery, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, endpoint, source, delivery_id, kind, received_at, headers, payload
FROM spool ORDER BY id LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []Delivery
	for rows.Next() {
		var d Delivery
		var recv int64
		var hdr, payload string
		if err := rows.Scan(&d.ID, &d.Endpoint, &d.Source, &d.DeliveryID, &d.Kind, &recv, &hdr, &payload); err != nil {
			return nil, err
		}
		d.ReceivedAt = time.UnixMilli(recv).UTC()
		if err := json.Unmarshal([]byte(hdr), &d.Headers); err != nil {
			return nil, fmt.Errorf("ingest spool row %d headers: %w", d.ID, err)
		}
		d.Payload = json.RawMessage(payload)
		out = append(out, d)
	}
	return out, rows.Err()
}

// Delete removes the rows with these ids. Missing ids are not an error.
func (s *Spool) Delete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	q := "DELETE FROM spool WHERE id IN (?" + strings.Repeat(",?", len(ids)-1) + ")"
	_, err := s.db.ExecContext(ctx, q, args...)
	return err
}
