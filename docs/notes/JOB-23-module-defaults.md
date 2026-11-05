# JOB-23: NixOS module defaults (2026-11-05)

The module's defaults are what the controller itself needs, not what
Jobshop's own agents on gasserver use.

- Controller PATH: git, bash, coreutils, diffutils, openssh, plus the
  `gh` wrapper (identities), bwrap (`sandbox.enable`), pasta and nft
  (`isolation.enable`). From an audit of `exec.Command`/`sh -c` in pkg/
  (git, gh, sh, bwrap/pasta/nft through acphost) and the module's units
  (the sync script uses diff, sed, head; the clone unit may use ssh).
- `sandbox.rw` defaults to `[ ]`. `sandbox.envKeep` keeps HOME, USER,
  LOGNAME, LANG, LC_ALL, TZ, TERM, SSL_CERT_FILE, NIX_SSL_CERT_FILE and
  `GIT_*`.
- The mise and nix-ld env and the `jobshop-mise` unit/timer are gone.
  docs/nixos.md shows them as host config.
- Compatibility for one release: `defaultPath` stays, default false,
  and warns when set. `sandbox.mise.*` stays as ignored options that warn
  when `tools` is set. A later PR removes both
  (`lib.mkRemovedOptionModule`).

gasserver must set its tools in its host config before this deploys.
