# Secret scanning with betterleaks (#578)

2026-11-05. `pkg/redact` uses betterleaks v2 (`config` + `scan`, Go
regexp engine) instead of gitleaks. Pinned to v2.0.0-rc.7; bump to
v2.0.0 when it is released.

- Rule order: betterleaks defaults, then `pkg/redact/rules.toml`
  (embedded, `[extend] useDefault = true`: Anthropic OAuth
  `sk-ant-oat01-`, Matrix `syt_`, Linear OAuth `lin_oauth_`, Bitwarden
  SM access tokens), then the host's `redact.rules`, minus
  `redact.disabled_rules`. Config rules need keywords.
- Masking replaces the rule's value group (`Match.Value`), so a URI rule
  masks the password and leaves the host; with no group, the match.
- Config: `redact` is host-only (packs are rejected). Load compiles the
  scanner; a bad regex, duplicate id or unknown disabled id fails
  `config check` naming the rule. Reload swaps the scanner with the
  config; a failed reload keeps the old one.
- The controller's journal, shipping, subscribe, command step and stored
  write paths use the config's scanner. The CLI's `gh`/`linear` write
  check asks the controller (`secret.scan`, open to every caller; it
  returns rule ids and lines only), so config rules apply there too.

Tests: `pkg/redact` table test with values generated at run time (four
built-ins, a Postgres DSN password, ghs_, sk-proj, api03,
`WEBHOOK_SECRET=`; benign shapes not caught), `pkg/config` rule and
error cases, and a watch-reload test in `pkg/control`.
