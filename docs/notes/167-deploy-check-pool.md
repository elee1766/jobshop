# post-merge in its own pool (#167)

2026-11-02.

`post-merge` waits for master CI and the deploy, up to an hour. It ran
in the single-session `refinery` pool, so every `review-merge` queued
behind it (the #147 review sat ready 1h50m behind #146's deploy wait).

Built: `packs/github` adds a `deploy-check` pool (`deploy_sessions`,
default 1, prompt `prompts/deploy-check.md`) and `pr-work`'s `post-merge`
step runs there. The refinery prompt now covers `review-merge` and
`merge` only. Merges stay serialized, and deploy checks run alongside.

Hosts that override pool settings (for example `max_sessions: 0` in
tests) name the new pool `deploy-check`.

Test: `TestPostMergeOwnPool` holds one run's `post-merge` in
`deploy-check` and claims another run's `review-merge` in `refinery`.
