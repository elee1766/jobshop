-- +goose Up
-- Timestamps move from unix ms to unix ns (JOB-58), as sqlite 00025.
-- recall.at is text here. A defer_until past 2200 (the old year-9999
-- gate hold) is clamped to 2200-01-01, the new hold.
UPDATE tickets SET
  lease_until = lease_until * 1000000,
  defer_until = CASE WHEN defer_until > 7258118400000 THEN 7258118400000000000 ELSE defer_until * 1000000 END,
  created_at  = created_at * 1000000,
  updated_at  = updated_at * 1000000,
  closed_at   = closed_at * 1000000;
UPDATE edges SET created_at = created_at * 1000000;
UPDATE history SET at = at * 1000000;
UPDATE comments SET created_at = created_at * 1000000, edited_at = edited_at * 1000000;
UPDATE recall SET at = CAST(CAST(at AS BIGINT) * 1000000 AS TEXT);
UPDATE resource_slots SET updated_at = updated_at * 1000000;
UPDATE records SET
  lease_until = lease_until * 1000000,
  created_at  = created_at * 1000000,
  updated_at  = updated_at * 1000000;

-- +goose Down
UPDATE tickets SET
  lease_until = lease_until / 1000000,
  defer_until = defer_until / 1000000,
  created_at  = created_at / 1000000,
  updated_at  = updated_at / 1000000,
  closed_at   = closed_at / 1000000;
UPDATE edges SET created_at = created_at / 1000000;
UPDATE history SET at = at / 1000000;
UPDATE comments SET created_at = created_at / 1000000, edited_at = edited_at / 1000000;
UPDATE recall SET at = CAST(CAST(at AS BIGINT) / 1000000 AS TEXT);
UPDATE resource_slots SET updated_at = updated_at / 1000000;
UPDATE records SET
  lease_until = lease_until / 1000000,
  created_at  = created_at / 1000000,
  updated_at  = updated_at / 1000000;
