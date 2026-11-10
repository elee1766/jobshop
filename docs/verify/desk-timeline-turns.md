# Desk trace timeline: turns and live holder state

2026-11-10. Extends `desk-timeline.md`. Run on an isolated controller (see
`desk-redesign.md`). Never seed a live instance.

1. Controller with a fake agent pool (`isolated-controller` steps 1 to 6).
   Create a task for the pool and let the session claim it.
2. `jobshop ticket show <id>`: the history has `turn_start` and `turn_end`
   rows by the session, and the ticket's `revision` is the same on each. A
   session with no ticket (a chat) adds none.
3. `curl -s '<desk>/api/trace/<id>?timeline=1' -H 'X-Jobshop-Desk: 1' | jq
   '.timeline.rows[].segments[] | select(.state=="working") | .turns'` lists
   the turns with `stop`, and `.sessions` holds the holders of open bars.
4. Open `?view=trace&ref=<id>`. The working bar has solid blocks (turns) on a
   vertical-line pattern (held). The legend says `in a turn` and `held, idle`.
   The detail under the chart says `active <t> in <n> turns · held <t>` and its
   `turns` disclosure lists each turn with its stop reason.
5. Live states on the open bar: while the agent runs a turn (hold it with the
   fake agent's `block` file) the bar has a ring and pulses, and its name says
   `turn running` with the tool. Between turns it is the plain held bar,
   `held, idle between turns`. Kill the host (`kill <host pid>` from
   `session list`) while the ticket stays held: the bar turns orange and
   hatched, `host not alive, nobody is working this ticket`.
6. Emulate `prefers-reduced-motion: reduce`: the running bar does not pulse and
   keeps its ring.
7. A ticket claimed before this change, or by an external session, has no
   `turns`: its bar is the plain working bar, and not drawn as idle or lost.
8. The ticket's history list shows one line per claim with `<n> turns`, not two
   lines per turn.
9. Check 1000, 790 and 390 px, light and dark: no page overflow.
10. `DESK_URL=<desk> npm --prefix ui run test:e2e -- timeline.spec.js` passes on
    both projects, also with `DESK_PREFIX=1` (Desk with `--trusted-proxy
    127.0.0.1`). `node ui/e2e/timeline-shots.mjs <dir>` writes the screenshots.
