# Workflow config (#88)

`workflows:` merges like `pools:` (per entry, then per field; `steps` is a
list and replaces whole). Fields: `vars`, `issue`, `escalate_pool`,
`steps` (`id`, `pool`, `title`, `prompt` or `prompt_file`, `retries`,
default 2). Pack var substitution skips step `title` and `prompt`, so
both stay templates for run start; the loader records the declaring
pack's vars as `pack_vars` (not user-settable) and `prompt_file` is read
relative to the declaring file. Load fails on: an unknown step pool or
`escalate_pool`, a missing or duplicate step id, no steps, a missing
title, both or neither of `prompt` and `prompt_file`, negative
`retries`, a `${var}` in a title or prompt that is neither a workflow
var nor a pack var, and `issue` not in `vars`. `config show` lists
workflows with `sources["workflows.<name>"]`; `config reload` reports
their count. `jobshop run start` (`89-runs.md`) uses them.
