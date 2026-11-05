# #625 Sandbox host paths and env from config

2026-11-05.

- The controller no longer adds system mounts or env names of its own.
  It binds only what jobshop needs: its binary's dir, the agent command's
  dir, the session dir, the pool dir and the repo's git common dir.
- `sandbox.ro` holds the system paths (`/nix/store`, `/usr/lib`, ...).
  `sandbox.env_keep` lists controller env names kept inside bwrap; a
  trailing `*` keeps a prefix (`GIT_*`). Both append like `rw`.
- `sandbox_profiles:` at the top level names sandbox fragments. A
  `sandbox.profiles: [a, b]` list under defaults, an agent or a pool
  applies them before that level's own fields, so the level still wins
  for backend and network. Packs contribute `sandbox_profiles` like
  pools; a prefixed pack renames its profiles and its references to them.
  A profile cannot name profiles.
- A sandboxed pool with no ro mounts and no `shared_tools` loads with a
  warning: its agent would fail to exec.
- NixOS: `services.jobshop.sandbox.systemRO` and `sandbox.envKeep` hold
  the values the code used to (#262, #623) and go to `defaults.sandbox`,
  so module hosts see no change. Hosts not using the module must add
  them to their jobshop.yaml.
- Network modes stay host, none and pasta (mechanism). A named network
  setup (pasta plus deny CIDRs) is a profile: `{network: pasta, deny: [...]}`.
