# JOB-53: mount check without host parents

2026-11-06

acphost v0.15.0 (#162) maps the controller as root in the user
namespace when a sandboxed session runs as a mapped uid. bwrap binds
with the controller's own access and only then switches to the agent's
uid, so the agent never opens a mount source's host-side parents.

- `checkMounts`: sandboxed, each mount source and the cwd are checked
  only for their own mode and ACL against the mapped uid. Unsandboxed,
  the agent opens host paths, so every parent is still checked.
- The search-only ACL grants on the root, `sessions` and `homes` dirs
  (JOB-19) are made only for unsandboxed sessions.
- A home-manager host needs no 711 home or ACL activation for
  isolation; gasserver can drop `home.activation.isolationSearch`.
- Also moves the remaining pkg/control uses of the removed acphost root
  aliases (v0.14.0) to the `sandbox`, `cgroup` and `userns` packages.
