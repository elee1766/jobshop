# Ingest drain, poller rows and GitHub reconcile (#335)

2026-11-04. Item 3 of #318.

- **Drain.** The controller drains the spool (`ingest.spool`) into the
  inbox every 5s, and right away when its own Linear listener writes a
  row. `jobshop ingest drain [--once] [--spool path]` runs the same pass
  through the admin socket. Rows are read in id order, 100 at a time;
  each batch is one inbox transaction (`InboxPutBatch`), and the spool
  rows are deleted only after that commit. A crash in between re-drains
  the batch and the inbox dedupes it.
- **Bad rows.** A row whose headers or payload do not parse moves to
  `spool_bad` in the spool database (spool schema v2) with the error and
  a WARN log. It is never dropped silently.
- **Dedupe keys.** Inbox items dedupe on `delivery_id` (drained rows are
  `<source>:<delivery id>`, poller rows `poll:...`) and on `natural_key`
  (migration 00010), which names the comment version itself:
  `github:<comment|review|review_comment>:<repo lowercased>:<id>:<updated_at>`
  and `linear:comment:<id>:<createdAt>`. So a webhook and a poll of the
  same comment give one row, and an edit is a new row.
- **Pollers.** Subscriptions write each new comment with its natural
  key. Activity writes each new item as `poll:activity:<id>` with no
  natural key (the events API gives no comment version). Pollers keep
  dispatching directly; dispatch from inbox rows is item 5.
- **Reconcile.** Every `ingest.github_reconcile` (default 15m, `off` to
  disable) each github-app identity lists `GET /app/hook/deliveries`
  since the last good pass and redelivers each failed GUID (no 2xx
  attempt), oldest first, at most 20 per pass. Off without a github-app
  identity: a PAT has no App deliveries API.
