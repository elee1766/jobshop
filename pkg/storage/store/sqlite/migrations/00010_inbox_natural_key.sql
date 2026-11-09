-- +goose Up
-- Ingest drain (#335): the same comment can arrive as a webhook delivery
-- and as a poller find, under different delivery ids. natural_key names
-- the thing itself (source, subject, comment id and version), so the
-- second copy dedupes. Empty for items with no natural key.
ALTER TABLE inbox ADD COLUMN natural_key TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX inbox_natural_key ON inbox (natural_key) WHERE natural_key != '';

-- +goose Down
DROP INDEX inbox_natural_key;
ALTER TABLE inbox DROP COLUMN natural_key;
