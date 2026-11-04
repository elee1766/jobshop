# Linear communication

Check that questions and escalations reach the Linear issue and replies
come back (#293). Needs a Linear identity for the team, `tracker:
linear`, `linear_human` set, and subscriptions on.

1. Start `pr-work` for a test issue. In its implement session run
   `jobshop report waiting "scope" --question "Keep the old page?"`.
2. Within a tick, the Linear issue has one comment mentioning
   `linear_human` with the question, and the `Blocked` label.
   `jobshop ticket show <question>` has `meta.linear_ask` and an
   `asked on Linear` comment. A controller restart does not post again.
3. Reply under that comment in Linear as the human. Within a
   subscription interval the question closes `answered`, the session
   gets `Answer to your question ...`, and `Blocked` is gone.
4. A top-level comment on the issue reaches the session as `New on
   Linear issue ...`, not as an answer.
5. Fail the run to its retry limit: the `Escalate run` task is posted on
   the issue (summary and ticket id, no `Blocked`).
6. A GitHub-tracker run's question does not post to Linear.
