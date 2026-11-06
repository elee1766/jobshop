# sandbox.enable writes no sandbox config (2026-11-06)

- `services.jobshop.sandbox.enable` only puts bubblewrap on the
  controller's PATH. It no longer writes `defaults.sandbox` (backend,
  ro, rw, env_keep) or `shared_tools`, so enabling it sandboxes nothing.
- Removed options: `sandbox.systemRO`, `envKeep`, `rw`, `sharedTools`,
  `mise`. Setting one fails evaluation with a message pointing at
  `services.jobshop.settings` (`defaults.sandbox`, `shared_tools` or a
  `sandbox_profiles` entry a pool opts into).
- Why: a host turned it on for one opt-in pool and got default mounts
  and env for every pool's config. Sandbox config lives in one place,
  the jobshop config.
- Test: `nix/module-test.nix` "sandbox.enable writes no sandbox config"
  and "removed sandbox options fail with a message".
