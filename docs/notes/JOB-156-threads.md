# Thread subscription engine

2026-11-08. Part of JOB-139. No behavior change.

`pkg/control/threads` holds what `subscribe.go` kept after the earlier
extractions: the thread and comment types, implicit thread discovery,
the comment sources (gh, Linear, the mux), the poll plan (`Engine.Poll`),
read positions (`Engine.Save`), the host's own login (`Engine.Self`), the
reply filters (`Ignored`, `LinearMatch`), message text and the delivery
author. It reaches the controller through `threads.Host`: config, clock,
logger, store, `linearsvc.Service.Do` and the identity login.

`threads.Comment` embeds `github.Comment` and adds the host-only and
Linear fields, so there is no third comment type. `threads.Source` is
the one source the poller and the `thread.reply` wait share;
`threads.Replies` implements `waitengine.ReplySource` on it.

The controller keeps delivery in `subscribe.go`: `deliverThread` and
`resumeFollower` (comment, message, reopen, assign), the `subscribe`
RPC, the pump and `inboxComment` (the inbox write the poll hands each
comment to). `pollThreads` runs the ingest pass when the engine wrote.

`watch.go` is the config watcher (fsnotify), not a thread watcher. It
stays with reload.

Tests: the behavior tests in `pkg/control` are unchanged apart from type
names. Pure cases (delivery author, reply line, text, owner) moved to
`pkg/control/threads`.
