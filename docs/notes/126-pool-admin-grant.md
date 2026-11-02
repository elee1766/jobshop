# Pool admin grants (#126)

Added 2026-11-02. A pool's `admin` list widens its session sockets to
named admin methods: `config-reload`, `config-show`, `session-list`,
`session-restart`, `drain`, `trigger-list`. Other admin methods (reopen,
session new/send/subscribe/reap, shutdown) cannot be granted; an unknown
name fails config load. The check reads the live config on each call, so
a reload that drops a grant applies at once. `packs/operator` grants
`[config-reload, session-list]`, and its prompt no longer hands those
commands back to the operator. Tests: `TestSessionList`,
`TestParsePoolAdmin`, `TestOperatorPackLoads`.
