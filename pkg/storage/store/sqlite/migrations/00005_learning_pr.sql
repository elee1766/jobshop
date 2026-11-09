-- +goose Up
-- Curator (#273): the PR whose edit addresses a learning. Set while the
-- learning is still open (PR pending) and kept when it becomes applied or
-- rejected, so a rejected learning names the PR that was turned down.
ALTER TABLE learnings ADD COLUMN pr TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE learnings DROP COLUMN pr;
