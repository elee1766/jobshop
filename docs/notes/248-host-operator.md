# Operator pack at host scope (#248)

2026-11-04. `packs/operator` is included once in jobshop.yaml, without a
prefix, next to the repo packs. `packs/selfdev` no longer extends it,
so a host with selfdev alone has no `operator` pool. Vars:
`jobshop_repo` (where prompt fixes go), `checkout` (Jobshop's checkout,
the pool's repo), `agent`, `operator_sessions`. The single `repo` and
`guide` vars are gone.

`config show` returns `instances`: one entry per `packs:` item with its
path, prefix, vars after defaults (a base pack's vars win, so selfdev
shows `repo: elee1766/jobshop` and `checkout` from its github base) and
the pools, workflows and triggers it declares, prefix applied. The
prompt tells the session to read it and to use each instance's pools,
`<prefix>pr-work` workflow and tracker.

Admin grant: `config-reload`, `config-show`, `session-list`,
`pool-list`, `trigger-list`, `attention`. Run list/show and tickets need
none.

Host change: add the operator entry to `settings.packs` (gasserver has
it through selfdev until then; the pool disappears on reload after this
merges unless the host adds it).

Tests: `TestOperatorPackLoads`, `TestSelfdevPackLoads`,
`TestOperatorGrant`.
