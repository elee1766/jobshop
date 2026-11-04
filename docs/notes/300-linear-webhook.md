# Linear webhook intake (#300)

2026-11-04. Push path next to the polling of #287 and #293.

- Config: `linear.webhook_listen` (empty: off) and, per `linear-app`
  identity, `webhook_secret_file` with the App's webhook signing secret.
  The controller serves `POST /linear/webhook` on its own `http.Server`.
- Checks, in order: `Linear-Signature` is the hex HMAC-SHA256 of the raw
  body under some identity's secret (constant-time, else 401);
  `webhookTimestamp` within 60s (else 401); then 200 and the rest runs
  async. Events are dropped unless the actor is a user (not an app,
  bot or integration), not the App itself (viewer id) and
  `organizationId` is the identity's org (viewer organization).
- Issue events: the issue is re-read with `GetIssue`, the team's
  `linear_issue_no_run` cache is dropped and triggers run, so the intake
  stall fires now instead of after `poll`.
- Comment events: one subscription pass over that issue's thread, the
  same code as the pump, so a reply under an ask comment answers the
  question (`linearAnswers`, `unblockLinear`). Passes are serialized so
  the pump and a webhook do not deliver a comment twice.
- Linear must reach the address: put a reverse proxy or tunnel in front;
  gasserver has none. Polling is unchanged, with the webhook on or off.
- Tests: `TestLinearWebhookIssue`, `TestLinearWebhookComment`.
