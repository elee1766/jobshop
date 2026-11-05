# #623 Sandbox on FHS hosts (2026-11-05)

The bubblewrap sandbox bound `/lib64` but not `/lib` or `/usr/lib`. On
Ubuntu, `/lib64/ld-linux-x86-64.so.2` points into
`/usr/lib/x86_64-linux-gnu`, so every dynamically linked program in the
sandbox failed to exec: the race-built test agent and `/bin/sh` for
command steps. CI hit it once ubuntu-latest became 26.04, which ships
bwrap, so the sandbox tests stopped skipping (#619). NixOS hosts were
not affected, since everything there lives under `/nix/store`.

`sandboxSystemRO` now also binds `/lib`, `/usr/lib` and `/usr/lib64`
read-only. Missing ones are skipped. TestSandboxedSession prints the
session's host.log when the session ends early.
