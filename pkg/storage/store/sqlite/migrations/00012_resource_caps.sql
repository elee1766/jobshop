-- +goose Up
-- Resource pools (#400): a step names resource keys in fields.resources
-- ("emulator", "pr:acme/w#1"). Occupancy is the in-progress tickets that
-- name a key, so every claim end frees it. resource_caps holds each
-- resource's capacity, synced from config by the controller; a key
-- "name:..." takes the cap of "name". A key with no cap row is unlimited.
CREATE TABLE resource_caps (
  name     TEXT PRIMARY KEY,
  capacity INTEGER NOT NULL CHECK (capacity > 0)
);

-- +goose Down
DROP TABLE resource_caps;
