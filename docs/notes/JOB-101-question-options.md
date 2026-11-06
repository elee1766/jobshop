# Questions with context and options (JOB-101)

2026-11-06.

**Ask.** `jobshop report <state> -q "<question>" --context "<markdown>"`
(`-` reads stdin). The CLI rejects `-q` without a context of at least two
non-empty lines. Options are optional: `--option "Title|Description"`
(repeatable) or `--options-json` (`[{title, description, recommended}]`),
at most 5; `--recommend N` marks the asker's pick. The controller
(RPC `report`) validates the same limits (one-line title of at most 120
bytes, description at most 1000) but does not require a context, so a
session on an older binary can still ask after a deploy.

**Stored.** `meta.question.context` is the asker's context followed by
links the controller fills in (ticket, run and step, GitHub issue or PR,
Linear issue, asking session) and the asked ticket's last 3 comments.
`meta.question.options` is `[{id, title, description, recommended}]`
with ids 1..n. The question body repeats both, so Linear asks and
`ticket show` carry them.

**Shown.** `rpc.AttentionItem` (and Desk's `/api/attention`) has
`context` and `options`, so `jobshop attention` prints them. Desk's
Attention row expands the context (markdown, sanitized), shows each
option as a button (the recommended one highlighted) and a **Discuss**
button that opens the item's existing attention chat. The question chat
agent's first prompt includes the context and options.

**Answer.** `jobshop answer <id> --option N [note]` (RPC
`question.answer` with `option`). The asking session gets `Chose option
N: <title>`, the option's description, and `Note: <note>` when given.
Free text without `--option` is still an answer. A Desk option button
calls `question.answer` with the option and no note.

Out of scope: fixed options on controller escalations (run
needs_human, pool quarantine, wait timeout).
