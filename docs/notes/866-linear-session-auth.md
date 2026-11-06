# Linear "Authentication required, not authenticated" (#866, 2026-11-06)

## Root cause

The controller's cached Linear app token is revoked by other mints for
the same OAuth app, long before its `expires_in` (30 days). It stays
cached, and every call with it fails 401 `AUTHENTICATION_ERROR` until
something replaces it. Sessions get that same cached token over
`identity.token` and call Linear themselves, so `jobshop linear ...`
failed the same way as the controller's own reads.

Measured on gasserver 2026-11-06 with the app's own client id and secret
(identity `jobshop-linear`):

- A token minted and then left alone answered 200 for 15 to 60 seconds,
  then 401 (three runs, 08:22, 08:25, 08:29), with `expires_in` 2591999.
  It died by an event, not by age.
- Two tokens minted with the same scope set (and one with the same set in
  another order) stayed valid together.
- A mint with a different scope set revoked every earlier token at once,
  and a later mint with the first set revoked that one in turn.
- `agent-credd` on the host mints a Linear app token with
  `read,write,issues:create,comments:create,app:assignable,app:mentionable`
  (the jobshop default lacks the last two). Its journal has
  `{"kind": "linear", "refresh": true}` at 08:23:16 and 08:25:29, inside
  two of the three windows where a probe token died. The third death
  (08:30:29 to 08:30:34) has no credd Linear call in it, so at least one
  more process mints for the app. I could not read the other instance's
  config (jobcity runs a second controller on this host), so I did not
  confirm which app it uses.
- A controller restart mints a fresh token (08:08:23Z), which fits the
  one planner pass that worked at 08:08Z.

## Fix

#871 gave the controller's own Linear clients `Renew`: on a 401 the minter
drops the cache, mints, and the request is sent once more. This change
covers the session side, which #871 left:

- `identity.token` takes `stale`, the token Linear refused. For a Linear
  identity the controller then calls `Minter.Renew`: when its cache
  already holds a different token it returns that (another caller
  renewed), else it mints. Sessions holding the same refused token cause
  one mint between them.
- `jobshop linear ...` and `linear api` (with an identity token, not the
  user's own `LINEAR_API_KEY`) set `Client.Renew` to ask the controller
  that way and resend once.
- A real failure stays an error: after the one retry the 401 body comes
  back as `linear: Authentication required, not authenticated`, and the
  command prints nothing as if the read had worked
  (`TestLinearSessionAuthFailureIsAnError`).

## What this does not fix

A renew revokes the other minter's token when its scope set differs, so
while another process keeps minting for the app, each side is refused once
per foreign mint and recovers with one extra request. The lasting fix is
host config: give the controller its own Linear app, or make every minter
use the same scope set. That is a host decision, not part of this change.
