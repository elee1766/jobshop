# JOB-55: sandbox overlay

2026-11-06

`sandbox.overlay` (defaults, agents, pools, `sandbox_profiles`) merges
several host dirs onto one target with bwrap's overlay mounts
(acphost#163, acphost v0.11.1):

```yaml
sandbox:
  overlay:
    ~/.agents:
      layers: [/nix/store/...-agents-base, /nix/store/...-agents-go]
      mode: ro   # ro (default) | tmp | persist
```

- Layers for a target append in sandbox order (defaults, agent, pool,
  each level's profiles before its own fields); the last layer is on
  top and wins file by file. `mode` takes the most specific level that
  sets it. Each target is one overlay mount.
- `~/` targets expand to the controller's HOME, the same in-sandbox path
  rw binds use.
- Load checks: backend bubblewrap, layers absolute and existing dirs, at
  most 500 merged layers per target, at least one. A layer gone at spawn
  is a permanent error and quarantines the pool.
- `tmp`: writes go to a tmpfs and end with the session. `persist`:
  upper dir is `<session dir>/overlay/<target>/upper`; the work dir next
  to it is emptied on each spawn (overlayfs needs it empty).
- `jobshop session sandbox <id>` prints `overlay <target> (<mode>): <layers>`.
- The bump to acphost v0.11 also adds `proc.MaybeChild()` to `cli.Main`
  and the control TestMain (acphost#157).
