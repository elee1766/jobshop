-- +goose Up
-- Ingest endpoints (#339): the webhook endpoint a delivery came in on,
-- so admission hooks can match on it. Empty for poller finds.
ALTER TABLE inbox ADD COLUMN endpoint TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE inbox DROP COLUMN endpoint;
