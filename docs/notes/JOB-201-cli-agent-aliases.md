# CLI names agents use (JOB-201)

2026-11-10.

Agent session logs on selfdev and jobcity showed calls rejected as parse
errors that were another name for something that exists. When the call sat
in an `||` chain, the text was lost. These forms now do what the agent meant.

| Agents call | Maps to |
| -- | -- |
| `close -m "<text>"`, `close --message "<text>"` | `close --comment` (also `ticket close`). `-c` and `-n` are not added: one short flag per option. |
| `comment <ticket> "<body>"` | `comment -t <ticket> "<body>"`. With two positionals the first is the ticket. `-t` plus a positional ticket, or three positionals, is a parse error. `comment <body>` and `comment -t <ticket> -` are unchanged. |
| `linear show <KEY>` | `linear issue <KEY>` |
| `linear list` / `linear search "<q>"` without `--team` | `--team` is the host's Linear team when there is exactly one. None or several is an error that lists them. |
| `ticket show <id> --json`, `run list --json` | Accepted and ignored: both always print JSON. |
| `handoff --to <pool> -r "<why>"`, `--reason` | `handoff --note` |
| `run cancel <id> -r "<why>"` | New: adds `Cancelled: <why>` as a comment on the run, then cancels it. |
| `jobshop --version` | `jobshop version` (hidden flag, same line, stops before any command). |
| `wait rm <id>` | `wait clear <id>` |

## As built

* The CLI does not read the config. `linear.teams` (new RPC, no params,
  returns `{teams}` from `Config.ActivityTeams`) is how `--team` gets its
  default. It returns team keys only, no credentials, and is open to every
  caller and to read-only sessions. The token call still decides whether the
  caller may use the team.
* `run.cancel` takes `RunCancelParams{id, reason}` instead of `IDParams`. The
  reason is redacted like other writes and added before the steps close, so a
  failed cancel keeps it and the cancelled run is not asked to take a comment.
* Aliases are kong `aliases` on flags and commands, so `--help` shows the
  short flag (`-m`, `-r`) and the command alias (`wait clear (rm)`). The long
  flag aliases (`--message`, `--reason` on handoff) are not listed.

## Not included

Seen three times or fewer, each a new command or filter: top-level `status`
and `show`, `session show`, `learn show`/`learning show`, `ticket list --all`,
`events --ticket`, `wait list -t`.
