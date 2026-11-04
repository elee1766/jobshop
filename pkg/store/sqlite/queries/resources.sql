-- name: DeleteResourceCaps :exec
DELETE FROM resource_caps;

-- name: InsertResourceCap :exec
INSERT INTO resource_caps (name, capacity) VALUES (?, ?);

-- ResourceFull reports whether ticket id names a resource key with no
-- free slot (#400), for Claim by id. Same rule as ReadyTickets.
-- name: ResourceFull :one
SELECT CAST(EXISTS (
  SELECT 1 FROM tickets t, json_each(t.fields, '$.resources') r
  JOIN resource_caps c ON c.name = r.value
    OR (instr(r.value, ':') > 0 AND c.name = substr(r.value, 1, instr(r.value, ':') - 1))
  WHERE t.id = sqlc.arg(id) AND c.capacity <= (
    SELECT count(*) FROM tickets h, json_each(h.fields, '$.resources') hr
    WHERE h.status = 'in_progress' AND h.id <> t.id AND hr.value = r.value)
) AS INTEGER) AS is_full;

-- name: ResourceCaps :many
SELECT name, capacity FROM resource_caps ORDER BY name;

-- ResourceUse counts in-progress tickets per resource key (#400).
-- name: ResourceUse :many
SELECT CAST(r.value AS TEXT) AS rkey, count(*) AS held
FROM tickets t, json_each(t.fields, '$.resources') r
WHERE t.status = 'in_progress'
GROUP BY r.value
ORDER BY r.value;

-- ResourceWaiting counts open tickets per resource key (#400).
-- name: ResourceWaiting :many
SELECT CAST(r.value AS TEXT) AS rkey, count(*) AS waiting
FROM tickets t, json_each(t.fields, '$.resources') r
WHERE t.status = 'open'
GROUP BY r.value
ORDER BY r.value;
