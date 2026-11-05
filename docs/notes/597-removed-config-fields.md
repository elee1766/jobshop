# Removed config fields warn for one release (#597)

2026-11-05. Self-deploy's pre-switch check (#261) runs the new binary
against the jobshop.yaml the old module generated. #577 made
`attention.rollback_file` a load error in the same commit that stopped
the module writing it, so the check failed, nothing switched, the file
was never regenerated, and every later deploy was skipped.

Rule (option b on the issue): a removed or renamed field that the NixOS
module sets becomes a load warning and is ignored for one release. A
later PR, after the host has switched, makes it an error.
`attention.rollback_file` and `attention.config_repo_file` warn since #596.
The selfdev pack guide states the rule; nix/module-test.nix checks a
warning-only config check still switches.
