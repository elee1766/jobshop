# Desk plugins: author reference

2026-11-05. How a plugin talks to Desk. History and design:
`docs/notes/517-desk-plugins.md` (#517), `docs/notes/535-desk-pwa.md`
and `docs/notes/535-attention-chat.md` (#535).

## Manifest

`<dir>/plugin.json`: `name` (matches the directory), `entry` (an ES
module), `api` (`">=2"` for everything below), `contributes`.

Slots:

- `view`: nav entry and `?view=<id>`. A view whose id is a built-in's
  replaces it (`attention`, `chat`, `deploys` are built-in plugins).
- `drawer`, `action`: ticket and run drawer, matched by `kind`, `label`,
  `workflow`, `pool`. Context `{ticket, run?}`.
- `artifact`: attachment renderer by `type`, `ext`.
- `header`: nav widget.
- `session` (#535): the session panel body. The shell owns the frame
  (side column on desktop, full screen on a phone). The last one wins.
  Context `{session, item}`, where `item` is the attention item
  (`kind:id`) the session resolves, or empty.

## api (version 2)

Version 1: `get`, `post`, `call(method, params)`, `subscribe(fn)`,
`events(filter, fn)`, `navigate`, `toast`, `confirm`, `url`, `user`,
`base`, `lastSeq`.

Added in version 2 (#535):

- `params()`: Desk's URL state (`view`, `ticket`, `session`, `item`, ...).
- `openSession(id, {item?})`, `closeSession()`, `openTicket(id)`.
- `session.stream(id, {tail}, fn)`: the live stream; `fn` gets
  `{type: 'open'|'event'|'end'|'error', event}`. Returns a close function.
  Events are `rpc.SessionEvent`: `turn_start` (text, context), `text`,
  `reasoning`, `tool` (`tool.id`, `title`, `status`, `input`, `output`),
  `turn_end` (`stop_reason` or `error`), `usage` (`usage.used`,
  `usage.size`), `state`.
- `session.older(id, before, limit)`: `{events, first}`, the decoded
  journal before seq `before` (paged through `session.journal`).
- `session.access(id)`, `session.send({session|key, text, interrupt})`,
  `session.cancel(id)` (Stop: cancels the turn, sends nothing).
- `push.status()`, `push.subscribe()`, `push.unsubscribe()`,
  `push.test()`: each resolves to `{state, error}`. `state` is
  `unsupported`, `unavailable` (Desk without `--push`), `denied`, `off`
  or `on`.

## Theme

Read these CSS custom properties; Desk sets them on `:root` and swaps
them for `[data-theme='light']`:

| Property | Use |
| --- | --- |
| `--desk-bg`, `--desk-side`, `--desk-panel`, `--desk-raised`, `--desk-hover` | surfaces, darkest to lightest, and row hover |
| `--desk-border`, `--desk-border-strong` | hairlines |
| `--desk-text`, `--desk-muted`, `--desk-faint` | text, secondary, tertiary |
| `--desk-accent`, `--desk-accent-fg`, `--desk-accent-soft` | the one accent, text on it, tinted background |
| `--desk-ok`, `--desk-warn`, `--desk-bad` | states |
| `--desk-shadow`, `--desk-radius` | overlays, corners |
| `--desk-font`, `--desk-mono` | font stacks |
| `--desk-tabbar` | phone tab bar height |

Global classes a view may use: `.vh` (view header, `h1`, `.sub`,
`.actions`), `.pad`, `.card`, `.pill` (`.ok`, `.warn`, `.bad`,
`.accent`), `.dot`, `.mono`, `.muted`, `.faint`, `table.stack` (stacked
rows on a phone, with `td.c-id`, `c-st`, `c-title`, `c-time`, `full`).
