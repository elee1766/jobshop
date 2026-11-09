-- +goose Up
-- Session journal summaries move from the shared recall table onto the
-- journals store collection (#579), as sqlite 00024. The partition and
-- indexes are what registration creates. recall then holds only
-- comments.
CREATE TABLE IF NOT EXISTS records_journals PARTITION OF records FOR VALUES IN ('journals');
CREATE INDEX IF NOT EXISTS rf_ticket ON records ((value #>> '{ticket}'), created_at);
CREATE INDEX IF NOT EXISTS rf_project ON records ((value #>> '{project}'), created_at);
CREATE INDEX IF NOT EXISTS rs_journals ON records_journals USING gin (to_tsvector('english', coalesce((value #>> '{text}'), '')));

INSERT INTO records (collection, key, state, value, created_at, updated_at)
SELECT DISTINCT ON (ref) 'journals', ref, '',
  jsonb_build_object('ticket', ticket_id, 'project', project, 'at', CAST(at AS BIGINT), 'text', body),
  CAST(at AS BIGINT), CAST(at AS BIGINT)
FROM recall WHERE kind = 'journal'
ORDER BY ref, id DESC
ON CONFLICT DO NOTHING;
DELETE FROM recall WHERE kind = 'journal';

-- +goose Down
INSERT INTO recall (body, kind, ref, ticket_id, project, at)
SELECT value->>'text', 'journal', key, value->>'ticket', value->>'project', value->>'at'
FROM records WHERE collection = 'journals';
DELETE FROM records WHERE collection = 'journals';
DROP INDEX IF EXISTS rs_journals;
