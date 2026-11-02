# Matrix gateway round trip (#123)

Against the local tuwunel on gasserver (`http://127.0.0.1:6167`,
`jobshop.localhost`). Passed 2026-11-02 with a throwaway root and the
control test binary's `fake-agent` (echo).

1. Register a bot and a test user with the token in
   `/var/lib/jobshop-matrix/registration-token` (`m.login.registration_token`
   UIA on `/_matrix/client/v3/register`), log both in, and write the bot's
   access token to a file in the root (mode 600).
2. As the test user, create a room inviting the bot.
3. Config: a `type: chat` pool, and `gateway.local` with `network: matrix`,
   `homeserver`, `token_file`, `pool`, `rooms: [<room id>]`,
   `senders: [<test user id>]`.
4. `jobshop serve`, then `jobshop gateway local` (log: `gateway: running`).
5. As the test user, post a message, then a reply in its thread
   (`m.relates_to.rel_type: m.thread`).

Pass:
- `jobshop session list` shows one session with key
  `matrix:<room id>:<root event id>`.
- Room `/messages` shows a 👀 reaction on each message and bot
  `m.notice` replies (`echo: ...`) with `rel_type: m.thread` and the
  root's event id.
- A message from an unlisted user gets no reaction and no session.
