-- +goose Up
-- As sqlite 00028: claim rows by actor and time (JOB-190).
CREATE INDEX history_claims ON history(actor, at) WHERE op = 'claim';

-- +goose Down
DROP INDEX history_claims;
