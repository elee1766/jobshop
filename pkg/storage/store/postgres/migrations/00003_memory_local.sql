-- +goose Up
-- Local memory writes (#592), as sqlite 00018: pr may be empty (unique
-- when set), plus the written path and the text it replaced.
ALTER TABLE memory_edits DROP CONSTRAINT memory_edits_pr_key;
ALTER TABLE memory_edits ALTER COLUMN pr SET DEFAULT '';
ALTER TABLE memory_edits ADD COLUMN path TEXT NOT NULL DEFAULT '';
ALTER TABLE memory_edits ADD COLUMN before TEXT NOT NULL DEFAULT '';
ALTER TABLE memory_edits ADD COLUMN existed BIGINT NOT NULL DEFAULT 0;
CREATE UNIQUE INDEX memory_edits_pr ON memory_edits(pr) WHERE pr != '';

-- +goose Down
DELETE FROM memory_edits WHERE pr = '';
DROP INDEX memory_edits_pr;
ALTER TABLE memory_edits DROP COLUMN existed;
ALTER TABLE memory_edits DROP COLUMN before;
ALTER TABLE memory_edits DROP COLUMN path;
ALTER TABLE memory_edits ALTER COLUMN pr DROP DEFAULT;
ALTER TABLE memory_edits ADD CONSTRAINT memory_edits_pr_key UNIQUE (pr);
