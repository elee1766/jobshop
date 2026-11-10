# Turn indicator in Desk chat (JOB-192)

2026-11-10. While a turn runs, the transcript shows one muted line under its
last row: `thinking…`, `replying…` or `running: <tool>`. It is gone when the
turn ends, fails or is cancelled.

- Derived in the browser by `turnPhase` (`ui/src/lib/session.ts`) from the
  frames the session stream already delivers: no new endpoint, SSE connection
  or polling. The newest frame of the open turn decides the phase, and a tool
  that has not completed wins over later text. The tool label is its intent,
  else its title.
- A turn ends at `turn_end` (a stop reason, a cancel or an error). A reload
  reads the tail of the journal, so a turn that began earlier still shows its
  phase from the frames in the tail. If the tail holds no turn frame and the
  host reports the chat busy, it shows `thinking…`.
- Chat and the read-only session viewer share `Transcript`, which reserves the
  line's height under the last row (`paddingEnd`), so it never moves rows.
- Browser check: `ui/e2e/turn-phase.spec.js`, run by `test:e2e:isolated`. The
  ACP test agent's `phases <tag>` turn advances when `<tag>.<n>` appears in the
  agent's directory (`DESK_AGENT_DIR`).
