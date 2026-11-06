# Stall trigger key prefix (JOB-95)

2026-11-06.

- `run.key_prefix` on a stall trigger replaces `stall` in its run keys:
  `key_prefix: review` keys runs `review:<signal>:<subject>`, and the
  run title shows that key. Packs use stall triggers for intake (PR
  review requests), where `stall:` read as stuck work. Unset, keys stay
  `stall:<signal>:<subject>`, so existing dedup is unchanged.
- It must match the trigger name pattern and cannot be set with
  `run.key` (which replaces the whole key) or on a trigger without
  `when.stall`.
- Every run a stall trigger starts with a stall key now carries
  `meta.trigger.stall` = `<signal>:<subject>`. The #704 cross-trigger
  hold looks runs up by it, so two triggers watching one subject with
  different prefixes still start one run. It also reads runs keyed
  `stall:<signal>:<subject>`, for runs started before this field.
- `isStallRun` is true for `meta.trigger.stall`, or (old runs and
  `run.key` runs) a `meta.trigger.key` of `<trigger>/stall:...`.
- `max_runs` and `cooldown` count the trigger's runs for its own key
  (`meta.trigger.key` = `<trigger>/<prefix>:<signal>:<subject>`).
  Changing a trigger's prefix starts those counts over.
