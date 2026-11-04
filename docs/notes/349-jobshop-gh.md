# #349: jobshop gh

2026-11-04. Behavior as built.

`jobshop gh <args>` runs the real gh and replaces the shell gh wrapper
from #216 in `nix/module.nix`. Main handles it before kong, so gh's
flags (including `--help`) reach gh unparsed.

- The real gh: `JOBSHOP_GH`, else the first `gh` on PATH that does not
  resolve (EvalSymlinks) to the jobshop binary itself.
- Token: unless `GH_TOKEN` is set, `identity.token` for the repo
  (`-R`/`--repo`, a github.com URL argument, `api repos/o/n/...`, else the
  cwd's git origin). No token found runs gh with its own auth.
- Writes: `pr create|merge|comment|review|edit|close|reopen|ready`,
  `issue create|edit|comment|close|reopen`, `release create`, `repo`
  writes, and `api` with a method other than GET (or with `-f`/`-F`/
  `--input`, which make gh POST). Each calls the `guard` RPC with the
  action, target `owner/name[#n]` and the body as detail. No controller
  fails closed (exit 1, gh not run); a deny or ask_human exits 1 with
  the reason.
- Each write, run or refused, goes to the activity feed through
  `activity.record` (kind `gh-<action>`, the command words without flag
  values). A recording failure only warns.
- Reads run with no guard and no record.
