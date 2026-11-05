# home-manager module (JOB-22)

2026-11-05.

- `nix/module.nix` takes a mode (`"nixos"` or `"home-manager"`). The
  option tree and the yaml generation are shared, not copied.
- Units are defined once in the internal `services.jobshop.units` option
  in NixOS form. NixOS mode passes them to `systemd.*`. home-manager mode
  renders them to `systemd.user.*`: `wantedBy multi-user.target` becomes
  `default.target`, `network-online.target` is dropped, `User`/`Group`
  are dropped, `script` becomes a `writeShellScript` ExecStart,
  `restartIfChanged = false` becomes `X-SwitchMethod=keep-old`, and the
  PATH is built as NixOS builds it.
- Root-only steps (`runuser`, `nixos-rebuild`, the system profile) are
  replaced in user mode by the user itself, `home-manager switch` and the
  home-manager profile. The rollback runs the previous generation's
  `activate`.
- `restrictEgress` and a local VictoriaLogs are refused in user mode.
  Both need the system manager.
- `checks.module` evaluates a home-manager configuration (new flake input
  `home-manager`, release-26.05, used only by the check) and asserts the
  generated `jobshop.yaml` is the same store path as the NixOS module's
  for the same settings.
