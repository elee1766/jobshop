# Deploy-check concurrency (#174)

2026-11-03. With `deploy_sessions: 1`, one post-merge step sleeping up
to an hour on a deploy left every other post-merge step ready and
unclaimed, and stall-check fired on them.

Built: `packs/github` defaults `deploy_sessions` to 3, so a slow deploy
wait holds one session while others finish steps whose commits are
already live. `prompts/post-merge.md` now says to rerun the deploy check
fresh each loop instead of reusing an earlier version. Releasing with
`--until` was not changed. `TestGithubPackLoads` checks the default.
