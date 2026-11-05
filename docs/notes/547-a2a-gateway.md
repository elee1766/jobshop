# A2A gateway: escalations to another agent (#547)

2026-11-05.

`network: a2a` is a gateway kind that mirrors escalation chats to another
agent over A2A v1.0 (JSON-RPC binding) instead of a chat room. It is
outbound only: no server, no push webhook, no agent card served.

```yaml
gateway:
  hermes:
    network: a2a
    url: https://hermes.example/a2a     # card: <url>/.well-known/agent-card.json
    token_file: /run/secrets/hermes-a2a # bearer token, never in yaml
    pool: operator                      # chat pool (fallback when meta.escalation has none)
    escalations: [all]                  # or causes: run, pool, wait, guard, label; default none
    poll: 30s
```

- **Routing:** an open needs-human ticket with an escalation chat
  (`meta.escalation`) whose `cause` is listed. The peer decides whether
  its own human is involved.
- **Conversation:** `contextId` is the ticket id. Each escalation session
  turn is a `SendMessage`; the first carries the ticket id, cause, title
  and Desk link. A peer `Message` reply, or a `Task` status message or
  artifact text, goes to `session.send` on the escalation key.
- **Tasks:** `working`/`submitted` are polled with `GetTask` every `poll`
  on the gateway clock. `input-required` sends the next session turn on
  the same task id. `failed`, `rejected`, `canceled`: one ticket comment
  with the peer's text, mirroring stops, the escalation stays open.
- **Restart:** `meta.gateway.a2a.<name>` holds context, last journal seq
  sent, open task and state, so a restarted gateway resumes polling and
  does not resend turns.
- **Close:** one last message saying who resolved it (resolution, actor).
- **Failures:** retry with backoff (poll, doubling, 10m cap). The first
  error of a streak is one ticket comment. The token is only in the
  `Authorization` header; errors carry method and HTTP status, never bodies.
- **Loop guard:** at most 20 messages per escalation, then a comment and stop.
- Stdlib JSON-RPC, no SDK: two methods and an agent card did not justify a
  dependency. Card selection prefers a v1.x `JSONRPC` interface, then any
  `JSONRPC` one, then a v0.3 `url`. Replies in v0.3 shape are also read.
  Streaming is not used yet: polling covers it.
- Resolution is unchanged: the escalation session has the tools; Desk
  and humans can answer at any time, and the first resolution wins.
