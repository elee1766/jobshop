# Secret commands

2026-11-04. `pkg/config/secret.go`.

- Every `*_file` secret (identity token, App key, Linear client id and
  secret, webhook secrets, ingest endpoint `secret_file`, gateway
  `token_file`) may be `!<command>`, as in git credential helpers. The
  rest runs with `sh -c` as the jobshop user, stdout is the secret.
  Lets people use pass, 1Password, vault, systemd-creds or a custom
  credential holder without writing plaintext files.
- Run on every read, same as files are re-read today, so rotation needs
  no restart. Commands that are slow or rate limited should cache
  themselves. Timeout 30s.
- Errors name "secret command" and stderr's last line only, so neither
  the command line nor the secret reaches logs.
- Root-relative values (ingest, gateway) are not joined to the root
  when they are commands.
- The config is trusted: whoever can edit jobshop.yaml can already run
  commands through hooks and triggers.
