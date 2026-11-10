# Activity feed live check (#255)

1. Add `activity: {}` to the host config (module `settings.activity`),
   deploy or `jobshop reload`. No restart is needed: the poller picks it
   up on its next minute.
2. Note `gh api rate_limit --jq .resources.core.used` (as the identity
   or host token the controller uses).
3. After two minutes, `jobshop activity --repo gfx-labs/acphost --since
   1h` from a session in a pool with the `activity` grant (or the admin
   socket) lists pushes, PRs, comments and CI. `<root>/activity.json`
   has an `events_etag` per repo.
4. Desk `?view=activity`: today's group shows both gfx-labs/jobshop and
   gfx-labs/acphost items; "humans" (default) hides elee-agent; items
   about an issue with a pr-work run show the run id, and it opens the
   run drawer, whose "External activity" lists the same items.
5. Leave it 30 minutes, read the rate limit again and record
   `(used after - used before) / 30` per minute in
   `docs/notes/255-activity-feed.md`. Expected near 0 on a quiet repo
   (304s are free) and a few per minute on a busy one.
6. Restart the controller: the feed is still there right away and the
   first polls are 304s.
7. With no Linear identity, `journalctl -u jobshop | grep activity`
   shows nothing per poll.
