# Operator pack (#124)

Added 2026-11-02. `packs/operator` has one pool, `operator`, with
`prompts/operator.md`: manage tickets, issues, messages, runs and reload
for the operator, short replies with links, and turn a correction into
an issue plus a prompt PR worked by a `pr-work` run. `packs/selfdev`
extends it, so gasserver gets the pool through its existing live pack.
Until chat pools (#122) land it has `max_sessions: 0`; a session is
started by hand with a key and a chat ticket, because message prompts
do not carry the pool prompt. `reload` and `session list` need the
admin socket, so from a session socket the prompt says to hand the
command to the operator. Tests: `TestOperatorPackLoads`,
`TestSelfdevPackLoads`.
