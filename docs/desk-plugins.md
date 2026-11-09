# Desk API

Desk is Jobshop's first-party web application. It no longer loads plugins,
custom elements or pack-provided browser code. Remove `desk.plugins` from
configuration. Existing `desk_commands` remain controller-side commands,
not browser extensions.

## JavaScript client

The framework-neutral [client](../packages/jobshop-client/index.js) and its
[types](../packages/jobshop-client/index.d.ts) live in this repository.
The package is private while the API settles. No npm publication is required.

```js
import { createClient } from './packages/jobshop-client/index.js';

const client = createClient({ baseURL: 'https://jobshop.example/desk/' });
const close = client.live(({ kind, data }) => console.log(kind, data));
// Call close() when this consumer is disposed.
```

The client has no UI, framework, DOM, storage or shared connection state.
Supply `fetch` and `EventSource` implementations outside browsers. Requests
stay beneath the explicit base URL, support abort signals, and add Desk's
CSRF header for writes. `ApiError` carries HTTP status and controller code.
Generated RPC parameter and result types come from `rpc.Specs`:

```sh
go run ./pkg/transport/rpc/rpcschema/cmd/deskgen
node --test packages/jobshop-client/index.test.js
```

`GET api/rpc` describes the allowed controller methods. Desk's existing
same-origin, trusted-proxy and `desk.rpc.allow`/`desk.rpc.deny` restrictions
still apply. A client does not bypass controller authorization. Alternative
frontends can use this transport without adopting Desk's layout.

## Live delivery

`client.live` consumes `GET api/live`. The stream sends initial attention,
in-flight runs, system state, named chats and improvements after registering
its controller subscription. Store commits trigger new snapshots, unchanged
sections are omitted, and bursts are coalesced. Scheduler passes deliver their
measurements directly. There is no browser state polling or event-triggered
HTTP refresh. Fifteen-second SSE comments only keep idle connections alive.

A reconnect takes a fresh snapshot, including changes missed while offline.
`client.events` remains the lower-level controller change feed with replay
cursors. `client.session` streams the active conversation and its journal
replay independently. Explicit detail and older-journal reads are still HTTP
requests. External integrations still depend on the events or polling their
upstream service supports.

## View commands

A named chat agent can issue `jobshop desk navigate view=trace ref=<ticket>`
or `jobshop desk filter window=inflight repo=<repo> step=<step>`. The
controller validates the command, records it on that chat, and delivers it
to a connected Desk displaying that chat. Commands navigate, filter,
select, attach context or expand/collapse windows. They cannot change work.
Work-changing operations use their own controller methods and grants.

Named chats support create, rename, archive and restore through
`chat.manage`. Archive stops the runner without deleting the journal.
Restore retains the session ID and ACP conversation and loads it on send.
Attention conversations remain separate from named chats. Attached references
are model context, not extra visible operator messages.
