# #495: stale ad-hoc tickets in attention

2026-11-05.

- An open, unclaimed ticket that no pool will claim (no pool, or a pool
  not in the config; the `orphaned` set behind the `no_pool` and
  `unknown_pool` warnings) is listed in `jobshop attention` as kind
  `stale` once it is older than `attention.orphan_age`.
- `attention.orphan_age`: Go duration, default `72h`; `off` disables it.
  Age is from the ticket's creation, on the controller clock.
- The title gives the reason, the age, and the command that closes it:
  `jobshop close <id> -r cancelled`.
- Listed once: a ticket already listed as needs-human or question is not
  listed again as stale. Auto steps (command, http) are never orphans.
- The github pack planner (`prompts/github/plan.md` step 13) closes each
  stale ticket it can verify is fixed, with a comment naming the PR, and
  leaves the rest open.
