-- +goose Up
-- Provider-owned resource slots (#407). A resource with a provider has
-- concrete slots (an emulator serial, a device) instead of a count. A
-- step needing it is ready only while a slot is ready, and the claim
-- takes one in the claim transaction. ticket and fence name the claim
-- holding a slot, so any claim end shows as a status or fence mismatch.
CREATE TABLE resource_providers (
  name TEXT PRIMARY KEY
);

CREATE TABLE resource_slots (
  pool       TEXT NOT NULL,
  id         TEXT NOT NULL,
  state      TEXT NOT NULL CHECK (state IN ('ready','held','releasing','gone')),
  env        TEXT NOT NULL DEFAULT '{}',
  ticket     TEXT NOT NULL DEFAULT '',
  fence      INTEGER NOT NULL DEFAULT 0,
  -- removing: the provider removed a held slot; it is deleted once reset.
  removing   INTEGER NOT NULL DEFAULT 0,
  updated_at INTEGER NOT NULL,
  PRIMARY KEY (pool, id)
);

CREATE INDEX resource_slots_state ON resource_slots(pool, state);
CREATE INDEX resource_slots_ticket ON resource_slots(ticket) WHERE ticket <> '';

-- +goose Down
DROP TABLE resource_slots;
DROP TABLE resource_providers;
