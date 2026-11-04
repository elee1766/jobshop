# Step pool from a run var (#384)

2026-11-04. A workflow step's `pool:` may be a template such as
`${review_pool}`, filled at run start like `title` and `prompt`. The var
comes from `run start --var`, a `run:` trigger (including
`${check_output}`), or a pack var default. Every step's pool is resolved
before any ticket is written; an unknown pool fails the start with
`unknown pool "<name>"` and creates nothing. Config load checks a pool
that uses only pack vars; one that uses a workflow var is checked at
run start. `run show` reports each step's resolved pool.
