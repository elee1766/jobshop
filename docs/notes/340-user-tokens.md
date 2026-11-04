# Running as your own user (#340)

2026-11-04.

- Identity kind `github-token`: `token_file` (fine-grained or classic
  PAT), `repos` globs, optional `api_url`. No `installations`,
  `app_id` or Linear fields.
- The minter reads the file on each request (no expiry, no cache), so
  `jobshop identity token`, the git credential helper, the gh wrapper
  and the controller's activity reads use it like an App token.
- Self filter: `Minter.Login` reads `GET /user` once per identity
  (cleared on reload). Subscriptions mark comments by that login as
  bot, and activity items by it get `bot: true`. A failed lookup logs
  a warning and filters nothing.
- `identity check` on a `github-token` shows the login as `viewer`.
- Kind `linear` (API key) now accepts `webhook_secret_file` for a
  workspace webhook. It gets an implicit `/linear/webhook` endpoint or
  can be named by an `ingest.endpoints` entry (#339), like a
  linear-app's. Intake, asks, subscriptions and the
  planner already went through `Minter.Token`, which works for both.
- Not for `github-token`: reconcile of failed App deliveries (#335) and
  App webhooks. Use a repo webhook secret (#339).
