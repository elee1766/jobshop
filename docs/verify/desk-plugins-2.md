# Verify: Desk plugins, part 2 (#541)

On the host running Desk (default `127.0.0.1:7480`):

1. `curl -s 127.0.0.1:7480/api/plugins | jq -r '.[].name'` lists
   activity, artifacts, attention, chat, pools, sessions, ticket, work,
   all with `"source": "builtin"`.
2. Every nav entry renders, `g w` and the other keys still jump, and
   `?view=nope` shows the missing-view note. Open a run from Work: the
   drawer shows details, steps and trace from the ticket plugin.
3. Filters on Work, Activity and Artifacts change the URL; back and
   forward restore them.
4. `curl -s -XPOST -H 'X-Jobshop-Desk: 1' -H 'Content-Type: application/json' -d '{"meta":{"github.repo":"owner/name"},"order":"newest","limit":5}' 127.0.0.1:7480/api/rpc/ticket.list`
   returns only that repo's tickets, newest first. Passing 4 meta keys
   fails with an invalid-argument error.
5. Playwright: `docs/verify/desk.md` step 3 and
   `docs/verify/desk-plugins.md` step 6.
