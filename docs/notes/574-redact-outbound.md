# Redaction on outbound paths (#574)

Added 2026-11-05. One detector, `pkg/redact` (gitleaks v8 default rules,
built once; plus values of secret-named env vars where the env is known).

- `session.subscribe` events (Desk chat, gateways, A2A): text, context,
  error and tool title are masked with the session's redactor, built
  once per subscription from the agent's env.
- Stored on write, masked: ticket comments, `report` note and question,
  learnings (summary, fix), artifact title and body, and artifact text
  attachments up to 1MB (binary ones, with a NUL byte, are left as is).
  A masked write adds an activity item `kind: redact`, `redacted: true`,
  naming the ticket, never the secret.
- `jobshop gh` writes: `-b/--body`, `-t/--title`, `-F/--body-file`
  (pr, issue), `api` `-f/-F/--field/--raw-field` values (`@file` read)
  and `--input` are scanned before the guard and gh. A finding exits 1
  and prints the rule and line, not the secret; gh is not run.
  `--allow-secret-findings` sends anyway, for admin callers only.
  Read-only calls never build the detector.
- `jobshop linear comment|ask|create|update` and `api` mutations: the
  same check (`--allow-secret-findings` on `jobshop linear`).
- The last 100 PR and 100 issue bodies of elee1766/jobshop (title and
  body, 2026-11-05) have 0 findings, so there is no allowlist option.
