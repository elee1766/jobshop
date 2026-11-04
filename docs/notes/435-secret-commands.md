# Secret commands (#435)

2026-11-04. `pkg/config/secret.go`. Reworks PR #366.

- Every secret setting (identity `token_file`, `private_key_file`,
  `client_id_file`, `client_secret_file`, `webhook_secret_file`, ingest
  endpoint `secret_file`, admission hook `secret_file`, gateway
  `token_file`) is a `config.Secret`: a path string as before, or a
  mapping with an explicit type:

  ```yaml
  token_file: /run/secrets/gh            # a file, as before
  token_file: {type: file, path: /run/secrets/gh}
  token_file: {type: command, command: "pass show gh/jobshop"}
  ```

  No prefix makes a string a command (the operator asked on #366 to drop
  `!`), so a path starting with `!` stays a path.
- A command runs with `sh -c` as the jobshop user and its stdout is the
  secret. It lets people use pass, 1Password, vault, systemd-creds or a
  custom holder without plaintext files.
- Run on every read, as files are re-read, so rotation needs no
  restart. Commands that are slow or rate limited should cache. Timeout
  30s, reported as "timed out".
- Errors say "secret command" and stderr's last line only, so neither
  the command line nor the secret reaches logs.
- Root-relative file paths (ingest, admission, gateway) are joined to
  the root; commands are not.
- A file secret marshals back as its plain path, so last-good and
  reload output keep their form.
- The config is trusted: whoever can edit jobshop.yaml can already run
  commands through hooks and triggers.
- NixOS module: `identities.*` take any value already; gateway
  `tokenFile` takes a string or the mapping.
