# Prompt parts go out once per conversation

2026-11-10.

A session got the pool prompt at the top of every ticket turn, and every
`at: system` rule again after each host respawn, even when the respawn
reloaded the conversation with `session/load`. In a jobcity delivery
session that was about 4 KB per ticket and about 20 KB per restart.

The pool prompt is now deduped like prompt rules (#397), under the key
`@pool`: sent on the conversation's first ticket turn, again only when its
text changes ("Updated pool instructions:"). The record of what the
conversation holds (`SessionFields.Injected`) is stored with the session,
so a controller restart or a respawn whose `session/load` works sends
nothing again. It resets with a new conversation: `session/new`, a failed
load, or a replaced conversation. `at: system` rules follow the same
record, so a reloaded conversation does not get them twice.

Desk: `raw model prompt` drops the paragraphs already shown under
`controller context`, so on a message turn it holds the envelope only, and
it uses the same font size as the context box.

Check: `TestPromptRules` (second ticket has no pool prompt; restarted
controller shows the pool and rules unchanged; a failed load resends them).
