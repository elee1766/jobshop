# Inbox on a collection (#579)

2026-11-05. Last domain PR of #579 step 2 (design:
`docs/notes/579-collections.md`).

- `pkg/ingest` (`inbox.go`) keeps the inbox (#321) in the `inbox`
  collection: `InboxPut`, `InboxPutBatch`, `InboxClaim`, `InboxRenew`,
  `InboxSetState`, `InboxRequeue`, `InboxGet`, `InboxList`,
  `InboxPrune`, `InboxCounts`, `InboxCountOlder`.
- Key is the item id as text: the ingest RPC (`IngestAdmitParams.ID`)
  and logs name items by id. State, lease and attempts are the
  record's. `created_at` is `received_at`, so claims (oldest first),
  prune and the backlog count read it with no extra field.
- Delivery id and natural key are Unique fields; source is indexed. A
  put looks both up first and returns the stored item as a duplicate,
  as the table did.
- Ids come from a counter record in the `inbox_ids` collection, not from
  the newest records: drained and polled items carry their own received
  times, so record order is not id order.
- `InboxPutBatch` is no longer one transaction (collections have no
  cross-call transaction). A failure part way leaves the earlier items;
  the drain keeps the spool rows until the batch is stored, so the
  retry dedupes them.
- `Store` loses the 11 `Inbox*` methods. Migration sqlite 00023 /
  postgres 00008 copies rows with ids, leases and JSON fields, seeds
  the counter past the highest id and drops the table; Down copies them
  back (Postgres moves the identity). `storemigrate` no longer copies an
  inbox table.
- `store.Store` now has 48 methods (from 75 at the start of #579).
  Recall, resources and slots are what is left of the domain methods.
