-- +goose Up
-- Portable resource slots (#426): one row per key in fields.resources,
-- kept in step with fields by the store, so the ready, claim and
-- occupancy queries need no json_each. rname is the configured resource
-- a key takes its capacity from (the part before the first ':').
CREATE TABLE ticket_resources (
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  rkey      TEXT NOT NULL,
  rname     TEXT NOT NULL,
  PRIMARY KEY (ticket_id, rkey)
);
CREATE INDEX ticket_resources_key ON ticket_resources(rkey);

INSERT OR IGNORE INTO ticket_resources (ticket_id, rkey, rname)
SELECT t.id, r.value,
  CASE WHEN instr(r.value, ':') > 0 THEN substr(r.value, 1, instr(r.value, ':') - 1) ELSE r.value END
FROM tickets t, json_each(t.fields, '$.resources') r
WHERE r.type = 'text';

-- +goose Down
DROP TABLE ticket_resources;
