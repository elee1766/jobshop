# Packs and reload (#27, #38)

`packs:` lists `path` (relative to the root) and `vars`. `pack.yaml` holds
`agents`, `pools`, `triggers`, `workflows`, and `defaults`/`extends`
(`107-github-pack-extends.md`); other keys are rejected. Merge: packs in
order, then jobshop.yaml; later sources win per key and per field through
nested maps; lists and scalars replace whole. `${name}` comes from the
pack's vars (a bare `${name}` keeps the var's type, `$$` is a literal `$`,
unset is an error); jobshop.yaml is not substituted. `prompt_file` is read
relative to its declaring file. `jobshop reload`/SIGHUP swaps the config
atomically or keeps the old one on error: pool limits and triggers apply on
the next tick (reload waits for a running tick, #102), prompts and timeouts
on the next turn, agent command and env on the next host. Telemetry changes need a restart. `jobshop config show`
(admin) prints the live merged config with jobshop.yaml keys, and
`sources`: for each `agents.<name>`, `pools.<name>` and `triggers.<name>`,
the files that set it in merge order (packs, then jobshop.yaml). OTLP
header values print unexpanded.
