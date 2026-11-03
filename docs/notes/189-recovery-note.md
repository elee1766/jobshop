# Recovery note for resumed sessions (#189)

2026-11-03. When a session's host is spawned again for an existing
conversation (session restart, crash restart, stuck-agent respawn, env
refresh, missing worktree, self-deploy), the runner sets `resumed`. The
next prompt, ticket or reminder, starts with a fixed note telling the
agent to check its worktree, git state, GitHub and `jobshop ticket show`
before repeating any outside action (push, PR, comment, close). It is
sent once per respawn, in memory only, and never on a new session or on
an adopted host after a controller restart (the agent kept running).
With a load fallback the journal summary comes first, then the note.
Test: `TestLoadFallback`.
