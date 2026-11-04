-- +goose Up
-- Portable SQL (#405): the JSON fields queries filter and sort on become
-- plain columns, written by the store from fields on every insert and
-- write, so queries and indexes need no json_extract. fields stays the
-- source of record. A column is NULL when fields lacks the key, as
-- json_extract was.
ALTER TABLE tickets ADD COLUMN f_key TEXT;
ALTER TABLE tickets ADD COLUMN f_to TEXT;
ALTER TABLE tickets ADD COLUMN f_state TEXT;
ALTER TABLE tickets ADD COLUMN f_workflow TEXT;
ALTER TABLE tickets ADD COLUMN f_index INTEGER;
ALTER TABLE tickets ADD COLUMN f_waiting INTEGER NOT NULL DEFAULT 0;

UPDATE tickets SET
  f_key      = CASE WHEN json_type(fields, '$.key') = 'text' THEN json_extract(fields, '$.key') END,
  f_to       = CASE WHEN json_type(fields, '$.to') = 'text' THEN json_extract(fields, '$.to') END,
  f_state    = CASE WHEN json_type(fields, '$.state') = 'text' THEN json_extract(fields, '$.state') END,
  f_workflow = CASE WHEN json_type(fields, '$.workflow') = 'text' THEN json_extract(fields, '$.workflow') END,
  f_index    = CASE WHEN json_type(fields, '$.index') IN ('integer', 'real') THEN CAST(json_extract(fields, '$.index') AS INTEGER) END,
  f_waiting  = CASE WHEN json_type(fields, '$.wait') IS NOT NULL AND json_type(fields, '$.wait') <> 'null' THEN 1 ELSE 0 END;

DROP INDEX tickets_message_to;
DROP INDEX tickets_session_key;
DROP INDEX tickets_run_key;
DROP INDEX tickets_run_state;
CREATE INDEX tickets_message_to ON tickets(f_to) WHERE kind = 'message';
CREATE UNIQUE INDEX tickets_session_key ON tickets(f_key)
  WHERE kind = 'session' AND status <> 'closed' AND f_key IS NOT NULL;
CREATE INDEX tickets_run_key ON tickets(f_key, f_workflow)
  WHERE kind = 'run' AND status <> 'closed';
CREATE INDEX tickets_run_state ON tickets(f_state) WHERE kind = 'run';
CREATE INDEX tickets_children ON tickets(parent_id, f_index);

-- Meta lookups (List with MetaPack/MetaKey): one row per string value
-- meta[pack][key], kept in step with meta by the store.
CREATE TABLE ticket_meta (
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  pack      TEXT NOT NULL,
  mkey      TEXT NOT NULL,
  value     TEXT NOT NULL,
  PRIMARY KEY (ticket_id, pack, mkey)
);
CREATE INDEX ticket_meta_lookup ON ticket_meta(pack, mkey, value);

INSERT INTO ticket_meta (ticket_id, pack, mkey, value)
SELECT t.id, p.key, k.key, k.value
FROM tickets t, json_each(t.meta) p, json_each(p.value) k
WHERE p.type = 'object' AND k.type = 'text';

-- +goose Down
DROP TABLE ticket_meta;
DROP INDEX tickets_children;
DROP INDEX tickets_run_state;
DROP INDEX tickets_run_key;
DROP INDEX tickets_session_key;
DROP INDEX tickets_message_to;
ALTER TABLE tickets DROP COLUMN f_waiting;
ALTER TABLE tickets DROP COLUMN f_index;
ALTER TABLE tickets DROP COLUMN f_workflow;
ALTER TABLE tickets DROP COLUMN f_state;
ALTER TABLE tickets DROP COLUMN f_to;
ALTER TABLE tickets DROP COLUMN f_key;
CREATE INDEX tickets_message_to ON tickets(json_extract(fields, '$.to'))
  WHERE kind = 'message';
CREATE UNIQUE INDEX tickets_session_key ON tickets(json_extract(fields, '$.key'))
  WHERE kind = 'session' AND status <> 'closed' AND json_extract(fields, '$.key') IS NOT NULL;
CREATE INDEX tickets_run_key ON tickets(json_extract(fields, '$.key'))
  WHERE kind = 'run' AND status <> 'closed';
CREATE INDEX tickets_run_state ON tickets(json_extract(fields, '$.state'))
  WHERE kind = 'run';
