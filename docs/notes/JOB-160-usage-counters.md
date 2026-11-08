# JOB-160: ACP usage counters in redacted JSON (2026-11-08)

`redact.JSON` masked every value under a key matching `secretKey`, which
includes `token`, so the ACP prompt-response counters (`inputTokens` and
the other four) shipped as `[REDACTED]`.

Now `cachedReadTokens`, `cachedWriteTokens`, `inputTokens`,
`outputTokens` and `totalTokens` keep their value only when the key sits
directly under a `usage` object and the value is a whole number >= 0.
A string, object, array, negative or fractional value, the same key
elsewhere, and any other token-named key are still masked. Env-value and
scanner redaction run on the encoded text as before.

Every `Redactor.JSON` caller shares this rule (journal shipping, session
journal pages, events, stored writes). Only these numeric fields under
`usage` change.

Shipped rows from before the fix keep `[REDACTED]`; cursors are not reset.

Check: `docs/verify/usage-counters.md`.
