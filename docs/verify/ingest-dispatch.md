# Ingest dispatch (#442)

1. Configure a validating admission hook that denies bodies containing
   `jobshop-deny-test`.
2. Comment `jobshop-deny-test` on an issue a working ticket follows.
3. After the next poll, `jobshop inbox list` shows the row `denied` and
   the session received no message for it.
4. Comment something else: the row goes `admitted` then `dispatched`,
   and the session gets exactly one message.
