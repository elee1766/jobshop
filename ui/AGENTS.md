# Desk UI

Read [the Desk design guide](../docs/desk-design.md) before changing UI behavior,
layout, styling or copy. It defines the Plan 9/Acme visual direction and the
simple, accessible Unix approach to interactions.

- Reuse `src/tokens.css`, `src/app.css` and existing components before adding
  local styles or another component abstraction. Fix shared inconsistencies at
  their source, not with per-page overrides.
- Use lowercase, short labels for navigation and controls. Preserve the case of
  user content, names, identifiers and code.
- Keep actions discoverable, keyboard accessible and usable on touch screens.
  Compact does not mean tiny targets or hover-only functionality.
- Verify populated desktop and mobile views, light and dark themes, focus,
  overflow and live state. Inspect rendered alignment, not just CSS values.
- Run `npm --prefix ui run check` and `npm --prefix ui run build` for UI changes.
  The resulting `pkg/transport/desk/dist` is ignored output needed before Go builds and
  tests, not source to commit. Nix and CI generate it automatically. Follow the
  root checks and relevant browser acceptance runbooks in `docs/verify/`.
- Run mutating browser tests only against an isolated controller, never live
  Jobcity. See `docs/verify/desk-redesign.md`.
- Before pushing a UI change, run the e2e specs it touches:
  `JOBSHOP_BIN=... FAKE_AGENT_BIN=... npm --prefix ui run e2e:affected`
  (binaries as in the header of `e2e/run-isolated.mjs`; `node e2e/affected.mjs`
  alone lists the specs). It maps your diff to specs with the table in
  `e2e/affected.mjs`: add a rule there for a new source file, and list a new
  spec in `e2e/groups.mjs`, or CI never runs it. See
  `docs/notes/JOB-213-e2e-affected.md`.
