# Verify: resource slots (#400)

On gasserver, with `resources: {pr: {keyed: true}}` in `jobshop.yaml` and
a test workflow whose step has `needs_resources: ["pr:${repo}#${issue}"]`:

1. `jobshop config reload`; it succeeds.
2. Start two runs on the same issue:
   `jobshop run start <wf> --var issue=1 --key a`, then `--key b`.
3. `jobshop ready`: only one of the two first steps is listed.
   `jobshop run show <run>` lists `resources pr:<repo>#1` on the step.
4. Close the claimed step (`jobshop close` from its session, or
   `jobshop ticket close <id>` as admin). The other run's step now shows in
   `jobshop ready`, and the supervisor starts a session for it on the
   next pass.
5. Misspell the resource in the step's `needs_resources`:
   `jobshop config reload` fails with `unknown resource` and the old
   config stays active.
