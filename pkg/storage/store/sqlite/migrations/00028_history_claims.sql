-- +goose Up
-- The trace finds the tickets a session held when it created another
-- (JOB-190) from its claim rows. Without this index that read walks all
-- of history; only claim rows are indexed, so it stays small.
CREATE INDEX history_claims ON history(actor, at) WHERE op = 'claim';

-- +goose Down
DROP INDEX history_claims;
