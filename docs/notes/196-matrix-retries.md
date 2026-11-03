# Matrix posts retry on 429 and 5xx (#196)

2026-11-03. mautrix defaults `DefaultHTTPRetries` to 0, and
`Gateway.post` only logs a failed post, so a rate limit
(`M_LIMIT_EXCEEDED`) or a homeserver restart at a turn end dropped that
reply.

Built: `NewMatrix` sets `DefaultHTTPRetries = 4`. mautrix then retries
429 and 5xx with its own backoff (4s, doubling) and honors
`Retry-After`. Discord is unchanged. `TestMatrixPostRetries` serves one
429 from an httptest homeserver and checks the post lands on the
second try. Live check: `docs/verify/matrix-retries.md`.
