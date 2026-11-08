# Linear pack prompts (#232)

Needs the Linear identity from `docs/verify/linear-tracker.md` on a
throwaway team (key `<KEY>`) and a throwaway GitHub repo.

1. Add a packs/github entry with `prefix: lin-` and vars `tracker:
   linear`, `linear_workspace`, `linear_team: <KEY>`, `linear_target:
   <KEY>-1`, plus the usual repo vars; `jobshop reload`.
2. In Linear, file `<KEY>-2` (a one-line README change) and name it in
   the description of `<KEY>-1`.
3. Wait for a `lin-plan` ticket (or create one in `lin-plan`). Expect a
   `lin-pr-work` run with key `linear-<KEY>-2` in `jobshop run list`,
   and only one, even if the intake triage run also fired (#292).
   Also file `<KEY>-3` in a `linear_intake_states` state without naming
   it in the target: the planner lists it and starts its run too.
4. Expect on `<KEY>-2`: state In Progress, then a `PR <url>` comment;
   the PR title ends `(<KEY>-2)`.
5. After merge and post-merge: a `Shipped in <sha>` comment and state
   Done.
6. Add a line to `<KEY>-1` asking for work no issue covers. The next
   planner pass files it with `linear create` and it shows a related
   link to `<KEY>-1`; a second pass does not file it again (#292).
7. Set `linear_human` to your Linear name. Make a `<KEY>-2` worker ask
   (`jobshop report waiting "q" --question "..." --context "..."` from its session, or
   wait for one): the issue gets a comment mentioning you and the
   `Blocked` label (#293). Reply under that comment: the question
   closes `answered` and `Blocked` is removed. Covered offline by
   `TestLinearEndToEnd` (#298).

## With a Linear App identity (#278)

Instead of a personal key the identity may be `kind: linear-app`
(`docs/notes/278-linear-app.md`): an OAuth app in the workspace with
client credentials enabled, its id and secret in host secret files.

1. Set `identities.<name> = {kind: linear-app, client_id_file,
   client_secret_file, workspace, teams: [<KEY>]}`; `jobshop reload`.
2. `jobshop identity check <KEY>` prints `viewer` (the app's actor) and
   an `expires` about a day out. A wrong secret fails here with
   `invalid_client`.
3. Run steps 2 to 7 above. Comments, state changes, created issues and
   the question on `<KEY>-2` show the app as author, not a person.
4. The token does not appear in `journalctl -u jobshop` (grep for its
   first 12 characters from `jobshop identity token <KEY>`).

## Webhook (#300)

Needs the App identity above and a public HTTPS URL forwarded to the
controller (reverse proxy or tunnel, e.g. `cloudflared`).

1. Set `linear.webhook_listen: 127.0.0.1:9470` and the identity's
   `webhook_secret_file` (the App's webhook signing secret); restart.
   The log shows `serving linear webhook`.
2. In the App settings, point the webhook at `<url>/linear/webhook`
   with Issue and Comment events.
3. Raise `linear_poll` to `1h`. File `<KEY>-4` in an intake state: a
   triage run starts within seconds, not at the next poll.
4. Repeat step 7 above: the reply answers the question within seconds.
5. `curl -X POST -d '{}' <url>/linear/webhook` returns 401.
