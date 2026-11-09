-- +goose Up
-- Learning kinds are config (#524), as sqlite 00017: the kind CHECK
-- becomes a name pattern check.
ALTER TABLE learnings DROP CONSTRAINT learnings_kind_check;
ALTER TABLE learnings ADD CONSTRAINT learnings_kind_check CHECK (kind ~ '^[a-z][a-z0-9-]*$');

-- +goose Down
UPDATE learnings SET kind = 'workaround' WHERE kind NOT IN ('correction','failure','workaround');
ALTER TABLE learnings DROP CONSTRAINT learnings_kind_check;
ALTER TABLE learnings ADD CONSTRAINT learnings_kind_check CHECK (kind IN ('correction','failure','workaround'));
