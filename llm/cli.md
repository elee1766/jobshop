# jobshop CLI

Generated from the kong help by `go generate ./pkg/cli`. Do not edit.

## jobshop

```
Usage: jobshop <command> [flags]

Multi-agent work orchestration.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  init [<dir>] [flags]
    Create a Jobshop root (jobshop.yaml) in a directory.

  serve [flags]
    Run the controller for the root.

  shutdown [flags]
    Stop the controller. Sessions keep running unless --all.

  drain [flags]
    Start no new sessions or tickets; running turns finish (admin).

  undrain [flags]
    Undo drain (admin).

  reload [flags]
    Re-read jobshop.yaml and packs; keeps the old config if invalid (admin).

  gateway <name> [flags]
    Run an A2A gateway from jobshop.yaml (admin socket client).

  desk [flags]
    Serve Jobshop Desk, the read-only dashboard (admin socket client).

  ingest serve [flags]
    Serve the webhook endpoints: verify, filter and spool deliveries. Needs no
    controller or main database.

  ingest drain [flags]
    Move spooled deliveries into the inbox through the controller (admin).
    The controller also drains on its own.

  ingest put --endpoint=STRING [flags]
    Spool one JSON delivery from stdin for a generic endpoint, with no HTTP or
    signature (trusted local senders). Exits 0 for new or duplicate deliveries.

  ingest work [flags]
    Admit pending inbox items through the controller (admin) with ingest.workers
    external (#443): claim, run the admission hooks here, report.

  ingest workers
    List external ingest workers' heartbeats and inbox counts by state.

  admin store migrate [flags]
    Copy a SQLite store into Postgres (#632). The controller must be stopped.

  ticket create <title> [flags]
    Create a ticket.

  ticket show [<id>]
    Show a ticket with edges, comments and history.

  ticket trace [<id>] [flags]
    Show the chain from a ticket: its trigger, holders, GitHub links and the
    tickets it led to.

  ticket list [flags]
    List tickets.

  ticket update <id> [flags]
    Change a ticket.

  ticket close [<id>] [flags]
    Close a ticket.

  ticket reopen <id>
    Reopen a closed ticket (admin).

  ticket undefer <id>
    Clear a ticket's deferral so it is ready now (admin, or pool grant
    ticket-undefer in the session's project).

  ticket upsert --key=STRING [<title>] [flags]
    Create or update the open ticket with a key, or close it (admin). For
    outside tools that report a problem and clear it.

  close [<id>] [flags]
    Close a ticket (default: the session's current one).

  comment <body> [flags]
    Comment on a ticket.

  guard <action> [flags]
    Ask the guardrail hooks before an outside action; exits 1 unless allowed.

  subscribe <thread> [flags]
    Follow a GitHub issue or PR thread from a ticket; new human comments come
    back to it.

  wait add --reason=STRING <cond> ... [flags]
    Park a ticket until conditions hold; a held ticket is released (not a failed
    attempt). End the turn after.

  wait list [flags]
    List waiting tickets.

  wait clear <id> [flags]
    Remove a ticket's wait; it becomes ready.

  decision <action> <id> [flags]
    Approve, reject or dismiss a planner proposal (#603).

  send <to> <text> [flags]
    Send a message to a session or pool.

  assign <id> [<holder>] [flags]
    Pre-assign a ticket to a session (admin).

  claim [<id>] [flags]
    Claim a ticket, or the next one with --next.

  heartbeat [<id>] [flags]
    Renew the lease on a held ticket.

  release [<id>] [flags]
    Hand a held ticket back to the queue.

  handoff --to=STRING [<id>] [flags]
    Move a held ticket to another pool, release it and leave a note.

  ready [flags]
    List claimable tickets.

  edge add <from> <to> [flags]
    Add an edge: FROM is blocked by / waits for / tracks TO.

  edge rm <from> <to> [flags]
    Remove an edge.

  label add <id> <label>
    Add a label.

  label rm <id> <label>
    Remove a label.

  session new --agent=STRING [flags]
    Register a session and open its CLI socket (admin). Phase 2 starts agents
    automatically.

  session list [flags]
    List sessions with live host state, plus orphaned hosts (admin).

  session reap [flags]
    Stop hosts with no open session; with --closed also delete old closed
    session dirs (admin).

  session restart <id> [flags]
    Restart a session's host with a fresh spec, keeping its conversation and
    claims (admin).

  session attach <id> [flags]
    Follow a session's conversation; --rw also sends stdin lines as messages.
    Ctrl-C detaches.

  session journal <id> [flags]
    Print a session's journal frames as JSON lines, redacted, through the
    controller (own session, admin or session-journal grant).

  session prompt <id> [flags]
    Print the prompt a session's next turn would get, each part labeled by
    source (admin).

  session sandbox <id> [flags]
    Print the sandbox mounts and bwrap argv a session's next spawn would use
    (admin).

  session cordon <id>
    Take a session out of rotation: it finishes its ticket and claims nothing
    new, but stays open (admin or session-cordon grant).

  session uncordon <id>
    Return a cordoned session to rotation.

  session drain <id>
    Cordon a session and end it once it holds nothing.

  session config <id> [<options> ...] [flags]
    Set ACP config options (model=<route> effort=<level>) on one session,
    applied between turns; --reset goes back to the pool's (admin or
    session-config grant).

  session rename <name> [flags]
    Set this session's display name, a non-unique label for lists; --id renames
    another (admin or session-rename grant). An empty name restores the default.

  run start <workflow> [flags]
    Start a workflow run: a run ticket and one step ticket per workflow step.

  run show <id> [flags]
    Show a run's state, vars and steps.

  run list [flags]
    List runs, newest first.

  run cancel <id> [flags]
    Close a run and its open steps as cancelled.

  run resume <id> [flags]
    Reopen a run's step (default: the first unfinished) with failures reset.

  batch start --key=STRING --issues=ISSUES,... [flags]
    Start one run per issue under a batch, with ordered cross-run edges.

  batch show <id> [flags]
    Show a batch's state and one line per member run.

  batch list [flags]
    List open batches.

  resource add <resource> <id> [flags]
    Add a ready slot to a provider resource, or update its env (admin or pool
    grant).

  resource remove <resource> <id>
    Remove a slot; a held slot goes when its claim ends (admin or pool grant).

  resource unhealthy <resource> <id> [flags]
    Mark a slot gone; its claim is requeued without counting a failure (admin or
    pool grant).

  resource list
    List slots and resource occupancy (admin or pool grant).

  trigger list
    Show each trigger: whether it holds now, last fired, open ticket, next cron
    time.

  pool list [flags]
    Show each pool's labels, min, max, desired, open, working, idle and cordoned
    sessions, plus quarantined and paused pools. --label filters.

  pool scale <pool> <target>
    Set a pool's desired session count, or 'auto' to follow ready work again.
    Persisted (admin or pool-scale grant).

  pool unquarantine <pool>
    Let a quarantined or paused pool start sessions again.

  attention [flags]
    List what waits on a human, oldest first; empty when healthy (admin).

  activity [flags]
    Recent GitHub and Linear activity in the repos and teams jobshop works,
    newest first (admin or pool grant activity).

  report incidents [flags]
    Group controller WARN/ERROR lines, crashed or stuck sessions and failed run
    steps.

  report ci-failures --repo=STRING [flags]
    Group failing tests from the failed-tests artifact of recent failed CI runs,
    with flaky/broken counts (runs gh).

  report progress [<note>] [flags]
    Report progress on your ticket.

  report waiting [<note>] [flags]
    Report waiting (on a question, a time, another change).

  report ready [<note>] [flags]
    Report the work ready for the next step.

  report blocked [<note>] [flags]
    Report blocked; labels the ticket blocked for stall-check.

  report failed [<note>] [flags]
    Report failure and why (then close -r failed).

  report dropped [<note>] [flags]
    Report the work dropped and why (then close).

  answer <question> <text> [flags]
    Answer a worker's question; the answer goes to the asking session (admin,
    or pool grant question-answer).

  artifact deliver <title> [flags]
    Deliver a report for a human to read in Desk, with optional attachments.

  artifact list [flags]
    List artifacts, unread first.

  artifact show <id>
    Show an artifact with its comments and history.

  artifact ack <id> [flags]
    Acknowledge (close) an artifact as read (admin).

  runbook run <file> [flags]
    Run a runbook's read-only steps (fenced blocks tagged readonly) and print
    the results.

  learn add <kind> <summary> [flags]
    Record a learning of any kind the config allows (learnings.kinds).

  learn correction <summary> [flags]
    Record a correction: a human or reviewer told you to do it differently.

  learn failure <summary> [flags]
    Record a failure and its cause.

  learn workaround <summary> [flags]
    Record a workaround for a tool or environment problem.

  learn list [flags]
    List learnings, newest first.

  learn update <id> ... [flags]
    Set the status or PR of learnings (curator, #273).

  learn evaluate [<id>] [flags]
    Curator gate: decide whether merged memory edits helped (#274).

  learn edits [flags]
    List merged memory edits with their baseline and evaluation (#274).

  config show
    Print the merged config the controller runs with, and which pack or file set
    each agent, pool and trigger.

  config check [flags]
    Load and validate a config file with its packs, without a controller.
    Exit 1 on error.

  memory write <scope> [flags]
    Write a local memory file under <root>/memory and record the edit (#592).

  memory show [<scope>]
    Print the local memory files.

  memory log [flags]
    List memory edits with their evaluation, newest first.

  memory check [<dir>] [flags]
    Report memory files over the character cap (exits 1 if any).

  events [flags]
    Print the change stream as JSON lines (admin, for debugging).

  identity list
    List identities and who may act as them (admin).

  identity check <target>
    Show which identity covers a repo or team and obtain a credential to prove
    it works (admin).

  identity token <target>
    Print a token for a repo (owner/name) or Linear team key, as the identity
    that covers it.

  identity git-credential <op>
    git credential helper: get returns an installation token for the repo in the
    request. Set credential.useHttpPath=true.

  linear issue <key> [flags]
    Print a Linear issue with its comments as JSON.

  linear comment <key> <body> [flags]
    Comment on a Linear issue (e.g. the PR link).

  linear state <key> <state> [flags]
    Move a Linear issue to a state of its team by name (e.g. Done on merge).

  linear context <key> [flags]
    Print an issue with labels, blockers, relations, children, assignee,
    state and recent comments as JSON.

  linear list --team=STRING [flags]
    List a team's issues as JSON, one page at a time.

  linear search --team=STRING <query> [flags]
    Search a team's issues as JSON, one page at a time.

  linear catalog --team=STRING <kind> [flags]
    List a team's projects, labels, states or users as JSON.

  linear eligible <key> [flags]
    Intake gate: print whether an issue may start a run and why not. Exits 1
    when not eligible.

  linear create --team=STRING --title=STRING [flags]
    Create an issue. --key makes a retry return the first result.

  linear update <issue> [flags]
    Change title, state, assignee or labels; reads the issue back to verify.

  linear assign <issue> <user> [flags]
    Take an issue: assign it to a user, or to yourself with me (an app token
    becomes the issue's delegate). linear assign <KEY> me.

  linear label <issue> <op> <label> ... [flags]
    Add or remove labels: linear label <KEY> add|remove <label>... (same as
    update --add-label/--remove-label).

  linear ask <issue> <body> [flags]
    Comment a question, mentioning users so they are notified.

  linear relate --blocks=STRING --related=STRING <issue> [flags]
    Mark an issue as blocking or related to another.

  linear api [<query>] [flags]
    Raw GraphQL call: LINEAR_API_KEY if set, else the team's identity. Mutations
    ask the guard (linear-api-mutation).

  recall <query> ... [flags]
    Search past comments, session journals and learnings. Run it on a failure
    and before asking a human.

  journal search <query> [flags]
    Search shipped session journals with LogsQL, e.g. 'status:failed | stats by
    (session) count()' (admin, or journal-search grant for your project).

  journal backfill [flags]
    Ship existing journals to VictoriaLogs from their cursors now (admin).

  whoami [flags]
    Show the caller's scope.

  version [flags]
    Print version information.

  help [<command> ...] [flags]
    Show help for a command, from the same text as --help: jobshop help linear
    assign.

  docs [<topic>] [flags]
    Print links to the docs for this binary (packs, readme, llm). Needs no root.

Run "jobshop <command> --help" for more information on a command.
```

## jobshop init

```
Usage: jobshop init [<dir>] [flags]

Create a Jobshop root (jobshop.yaml) in a directory.

Arguments:
  [<dir>]    Directory to make a Jobshop root.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --project="js"     First project name.
```

## jobshop serve

```
Usage: jobshop serve [flags]

Run the controller for the root.

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

      --log-level="info"     Log level: debug, info, warn or error
                             ($JOBSHOP_LOG_LEVEL).
      --log-format="text"    Log format: text or json ($JOBSHOP_LOG_FORMAT).
      --tick=30s             Lease sweep interval.
```

## jobshop shutdown

```
Usage: jobshop shutdown [flags]

Stop the controller. Sessions keep running unless --all.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --all              Also end every session (stops agents and hosts).
      --wait=2m          How long to wait for the controller to exit.
```

## jobshop drain

```
Usage: jobshop drain [flags]

Start no new sessions or tickets; running turns finish (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --pool=STRING      Drain only this pool (default: every pool and session).
```

## jobshop undrain

```
Usage: jobshop undrain [flags]

Undo drain (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --pool=STRING      Undrain only this pool (default: clear every drain).
```

## jobshop reload

```
Usage: jobshop reload [flags]

Re-read jobshop.yaml and packs; keeps the old config if invalid (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop gateway

```
Usage: jobshop gateway <name> [flags]

Run an A2A gateway from jobshop.yaml (admin socket client).

Arguments:
  <name>    A2A gateway name under gateway: in jobshop.yaml.

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

      --log-level="info"     Log level: debug, info, warn or error
                             ($JOBSHOP_LOG_LEVEL).
      --log-format="text"    Log format: text or json ($JOBSHOP_LOG_FORMAT).
```

## jobshop desk

```
Usage: jobshop desk [flags]

Serve Jobshop Desk, the read-only dashboard (admin socket client).

Flags:
  -h, --help                       Show context-sensitive help.
      --root=STRING                Jobshop root directory (holds jobshop.yaml,
                                   the database, sockets and sessions). Default:
                                   nearest ancestor of the cwd containing
                                   jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING              Socket to talk to. Defaults to the session
                                   CLI socket, else the root's admin socket
                                   ($JOBSHOP_SOCKET).

      --log-level="info"           Log level: debug, info, warn or error
                                   ($JOBSHOP_LOG_LEVEL).
      --log-format="text"          Log format: text or json
                                   ($JOBSHOP_LOG_FORMAT).
      --listen="127.0.0.1:7480"    Address to serve on. Non-loopback needs
                                   --allow-remote.
      --host=HOST,...              Extra Host header values to accept (e.g.
                                   a tailnet name); loopback is always accepted.
      --allow-remote               Allow a non-loopback listen address. Desk
                                   has no login; expose it only on a trusted
                                   network.
      --chat-pool="operator"       type: chat pool for Desk chat (escalation
                                   threads, new sessions). Empty disables
                                   sending.
      --user-header=STRING         Header an authenticating proxy sets to the
                                   user (e.g. Tailscale-User-Login). Writes
                                   record desk:<user>. Empty: no user.
      --base-path="/"              URL prefix Desk is served under behind a
                                   proxy, e.g. /desk/. The proxy may keep or
                                   strip it.
      --trusted-proxy=TRUSTED-PROXY,...
                                   Peer address or CIDR (repeatable) whose
                                   --user-header and X-Forwarded-Prefix
                                   Desk trusts. Required with --user-header.
                                   --base-path wins over X-Forwarded-Prefix.
      --push                       Send Web Push notifications for new attention
                                   items. Browsers allow push only over HTTPS
                                   (or localhost).
      --push-subject="mailto:jobshop@localhost"
                                   VAPID contact push services see (mailto:
                                   or https: URL).
```

## jobshop ingest

```
Usage: jobshop ingest <command> [flags]

Receive GitHub and Linear webhooks into the ingest spool.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  ingest serve [flags]
    Serve the webhook endpoints: verify, filter and spool deliveries. Needs no
    controller or main database.

  ingest drain [flags]
    Move spooled deliveries into the inbox through the controller (admin).
    The controller also drains on its own.

  ingest put --endpoint=STRING [flags]
    Spool one JSON delivery from stdin for a generic endpoint, with no HTTP or
    signature (trusted local senders). Exits 0 for new or duplicate deliveries.

  ingest work [flags]
    Admit pending inbox items through the controller (admin) with ingest.workers
    external (#443): claim, run the admission hooks here, report.

  ingest workers
    List external ingest workers' heartbeats and inbox counts by state.
```

## jobshop ingest serve

```
Usage: jobshop ingest serve [flags]

Serve the webhook endpoints: verify, filter and spool deliveries. Needs no
controller or main database.

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

      --log-level="info"     Log level: debug, info, warn or error
                             ($JOBSHOP_LOG_LEVEL).
      --log-format="text"    Log format: text or json ($JOBSHOP_LOG_FORMAT).
      --listen=STRING        TCP address (default ingest.listen, else
                             127.0.0.1:8740).
      --spool=STRING         Spool database (default ingest.spool, else
                             <root>/ingest/spool.db).
```

## jobshop ingest drain

```
Usage: jobshop ingest drain [flags]

Move spooled deliveries into the inbox through the controller (admin). The
controller also drains on its own.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --once             Drain once and exit.
      --spool=STRING     Spool database to drain (default the controller's
                         ingest.spool).
      --every=5s         Pause between passes without --once.
```

## jobshop ingest put

```
Usage: jobshop ingest put --endpoint=STRING [flags]

Spool one JSON delivery from stdin for a generic endpoint, with no HTTP or
signature (trusted local senders). Exits 0 for new or duplicate deliveries.

Flags:
  -h, --help                  Show context-sensitive help.
      --root=STRING           Jobshop root directory (holds jobshop.yaml, the
                              database, sockets and sessions). Default: nearest
                              ancestor of the cwd containing jobshop.yaml
                              ($JOBSHOP_ROOT).
      --socket=STRING         Socket to talk to. Defaults to the session
                              CLI socket, else the root's admin socket
                              ($JOBSHOP_SOCKET).

      --endpoint=STRING       Generic endpoint name (ingest.endpoints).
      --delivery-id=STRING    Delivery id for dedupe (default the endpoint's
                              map.delivery_id, else a hash of the body).
      --kind=STRING           Event kind (default the endpoint's map.kind).
      --spool=STRING          Spool database (default ingest.spool, else
                              <root>/ingest/spool.db).
```

## jobshop ingest work

```
Usage: jobshop ingest work [flags]

Admit pending inbox items through the controller (admin) with ingest.workers
external (#443): claim, run the admission hooks here, report.

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

      --log-level="info"     Log level: debug, info, warn or error
                             ($JOBSHOP_LOG_LEVEL).
      --log-format="text"    Log format: text or json ($JOBSHOP_LOG_FORMAT).
      --once                 Admit until nothing is pending and exit.
      --worker=STRING        Worker id (default <host>-<pid>).
      --every=1s             Pause when nothing is pending, without --once.
                             Each claim is the worker's heartbeat.
      --dispatch             Also dispatch admitted items through the controller
                             (#633). Safe alongside other dispatchers; set
                             ingest.dispatch external to stop the controller's
                             own.
      --no-admit             Only dispatch (with --dispatch), for ingest.workers
                             internal.
```

## jobshop ingest workers

```
Usage: jobshop ingest workers

List external ingest workers' heartbeats and inbox counts by state.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop admin

```
Usage: jobshop admin <command> [flags]

Offline maintenance with the controller stopped (store migrate).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  admin store migrate [flags]
    Copy a SQLite store into Postgres (#632). The controller must be stopped.
```

## jobshop admin store

```
Usage: jobshop admin store <command>

Ticket store maintenance.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  admin store migrate [flags]
    Copy a SQLite store into Postgres (#632). The controller must be stopped.
```

## jobshop admin store migrate

```
Usage: jobshop admin store migrate [flags]

Copy a SQLite store into Postgres (#632). The controller must be stopped.

Flags:
  -h, --help                Show context-sensitive help.
      --root=STRING         Jobshop root directory (holds jobshop.yaml,
                            the database, sockets and sessions). Default:
                            nearest ancestor of the cwd containing jobshop.yaml
                            ($JOBSHOP_ROOT).
      --socket=STRING       Socket to talk to. Defaults to the session
                            CLI socket, else the root's admin socket
                            ($JOBSHOP_SOCKET).

      --from=sqlite:PATH    Source store, sqlite:<path>. Default: the root's
                            jobshop.db.
      --to="postgres"       Target driver.
      --dsn-file=STRING     File holding the Postgres DSN (URL or key=value).
                            Default: the root config's store.dsn_file.
      --force               Replace the target's rows instead of refusing a
                            non-empty target.
```

## jobshop ticket

```
Usage: jobshop ticket <command> [flags]

Create, show, list, update and close tickets.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  ticket create <title> [flags]
    Create a ticket.

  ticket show [<id>]
    Show a ticket with edges, comments and history.

  ticket trace [<id>] [flags]
    Show the chain from a ticket: its trigger, holders, GitHub links and the
    tickets it led to.

  ticket list [flags]
    List tickets.

  ticket update <id> [flags]
    Change a ticket.

  ticket close [<id>] [flags]
    Close a ticket.

  ticket reopen <id>
    Reopen a closed ticket (admin).

  ticket undefer <id>
    Clear a ticket's deferral so it is ready now (admin, or pool grant
    ticket-undefer in the session's project).

  ticket upsert --key=STRING [<title>] [flags]
    Create or update the open ticket with a key, or close it (admin). For
    outside tools that report a problem and clear it.
```

## jobshop ticket create

```
Usage: jobshop ticket create <title> [flags]

Create a ticket.

Arguments:
  <title>    Title.

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

  -b, --body=STRING          Body. Use - to read stdin.
      --kind="task"          Kind.
      --project=STRING       Project (admin only; sessions use their own).
  -p, --priority=PRIORITY    Priority, lower runs first (default 2).
      --pool=STRING          Pool to route to.
      --assign=STRING        Pre-assign to a session (admin).
      --parent=STRING        Parent ticket id.
  -l, --label=LABEL,...      Labels.
      --fields=STRING        Kind fields as JSON, e.g. '{"to":"pool:dev"}' for a
                             message.
      --meta=STRING          Pack data as JSON.
```

## jobshop ticket show

```
Usage: jobshop ticket show [<id>]

Show a ticket with edges, comments and history.

Arguments:
  [<id>]    Ticket id (default: the session's current ticket).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop ticket trace

```
Usage: jobshop ticket trace [<id>] [flags]

Show the chain from a ticket: its trigger, holders, GitHub links and the tickets
it led to.

Arguments:
  [<id>]    Ticket id (default: the session's current ticket).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

  -d, --depth=INT        Levels of children to follow (default 5).
```

## jobshop ticket list

```
Usage: jobshop ticket list [flags]

List tickets.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --project=STRING    Project.
      --kind=STRING       Kind.
  -s, --status=""         Status.
      --pool=STRING       Pool.
      --holder=STRING     Holder session id.
      --parent=STRING     Parent ticket id.
  -l, --label=STRING      Label.
      --report=STRING     Last worker report state (progress, waiting, ready,
                          blocked, failed, dropped).
  -n, --limit=INT         Max results (default 500).
      --orphaned          Only open, unclaimed tickets with no pool or a pool
                          not in the config.
```

## jobshop ticket update

```
Usage: jobshop ticket update <id> [flags]

Change a ticket.

Arguments:
  <id>    Ticket id.

Flags:
  -h, --help                  Show context-sensitive help.
      --root=STRING           Jobshop root directory (holds jobshop.yaml, the
                              database, sockets and sessions). Default: nearest
                              ancestor of the cwd containing jobshop.yaml
                              ($JOBSHOP_ROOT).
      --socket=STRING         Socket to talk to. Defaults to the session
                              CLI socket, else the root's admin socket
                              ($JOBSHOP_SOCKET).

      --title=TITLE           New title.
  -b, --body=BODY             New body. Use - to read stdin.
  -p, --priority=PRIORITY     New priority.
      --pool=POOL             New pool.
      --defer=DEFER           Hide from ready until now+duration (0 clears).
      --fields=STRING         Replace kind fields (JSON).
      --meta=STRING           Replace pack data (JSON).
      --if-revision=INT-64    Fail unless the ticket is at this revision.
```

## jobshop ticket close

```
Usage: jobshop ticket close [<id>] [flags]

Close a ticket.

Arguments:
  [<id>]    Ticket id (default: the session's current ticket).

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

  -r, --resolution="done"    Resolution: done, failed, rebase (reopen a step
                             without counting a failure), cancelled, ...
      --outcome=STRING       Workflow step outcome word (approve, reject, ...);
                             later steps' when: conditions read it.
```

## jobshop ticket reopen

```
Usage: jobshop ticket reopen <id>

Reopen a closed ticket (admin).

Arguments:
  <id>    Ticket id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop ticket undefer

```
Usage: jobshop ticket undefer <id>

Clear a ticket's deferral so it is ready now (admin, or pool grant
ticket-undefer in the session's project).

Arguments:
  <id>    Ticket id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop ticket upsert

```
Usage: jobshop ticket upsert --key=STRING [<title>] [flags]

Create or update the open ticket with a key, or close it (admin). For outside
tools that report a problem and clear it.

Arguments:
  [<title>]    Title (required unless --close).

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

      --key=STRING           Stable key, e.g. deploy:rollback. At most one open
                             ticket per key and project.
  -b, --body=STRING          Body. Use - to read stdin.
      --project=STRING       Project (default: the config's default project).
      --pool=STRING          Pool to route a new ticket to.
  -l, --label=LABEL,...      Labels to add, e.g. needs-human.
      --close                Close the open ticket with the key, if any;
                             title and body become a closing comment.
  -r, --resolution=STRING    Resolution with --close (default done).
```

## jobshop close

```
Usage: jobshop close [<id>] [flags]

Close a ticket (default: the session's current one).

Arguments:
  [<id>]    Ticket id (default: the session's current ticket).

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

  -r, --resolution="done"    Resolution: done, failed, rebase (reopen a step
                             without counting a failure), cancelled, ...
      --outcome=STRING       Workflow step outcome word (approve, reject, ...);
                             later steps' when: conditions read it.
```

## jobshop comment

```
Usage: jobshop comment <body> [flags]

Comment on a ticket.

Arguments:
  <body>    Comment text. Use - to read stdin.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

  -t, --ticket=STRING    Ticket id (default: the session's current ticket).
```

## jobshop guard

```
Usage: jobshop guard <action> [flags]

Ask the guardrail hooks before an outside action; exits 1 unless allowed.

Arguments:
  <action>    Action about to happen: push, pr-create, merge, comment, message,
              or any name a hook knows.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --target=STRING    What it acts on, such as a PR url or branch.
      --detail=STRING    Free text for the hook, such as the message body.
```

## jobshop subscribe

```
Usage: jobshop subscribe <thread> [flags]

Follow a GitHub issue or PR thread from a ticket; new human comments come back
to it.

Arguments:
  <thread>    owner/name#<n>, or an issue or pull request URL.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --pr               The thread is a pull request (also read reviews).
                         Implied by a /pull/ URL.
  -t, --ticket=STRING    Ticket id (default: the session's current ticket).
```

## jobshop wait

```
Usage: jobshop wait <command> [flags]

Park a ticket until a PR review, checks, PR state, a reply or a time holds
(#381).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  wait add --reason=STRING <cond> ... [flags]
    Park a ticket until conditions hold; a held ticket is released (not a failed
    attempt). End the turn after.

  wait list [flags]
    List waiting tickets.

  wait clear <id> [flags]
    Remove a ticket's wait; it becomes ready.
```

## jobshop wait add

```
Usage: jobshop wait add --reason=STRING <cond> ... [flags]

Park a ticket until conditions hold; a held ticket is released (not a failed
attempt). End the turn after.

Arguments:
  <cond> ...    TYPE:key=val,... Types: github.review (pr, login, head,
                state), github.checks (pr or repo+sha, state=success|failure,
                required, workflow=<name or file>: follow that workflow's
                newest run on the commit instead of every check;
                a cancelled run ends the wait as invalidated), github.pr (pr,
                state=merged|closed|head_changed), github.label (issue, label;
                met when the issue has the label or is closed), github.tag
                (repo, tag), deploy (repo, sha; met when the controller's build
                commit contains sha), run (id, state=done|closed), linear.state
                (issue=KEY, state name or type), linear.label (issue=KEY, label,
                present=false waits for removal), thread.reply (thread, from),
                time (at). pr is owner/name#n or a PR URL; at is a duration or
                RFC3339.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

  -r, --reason=STRING     Why the ticket waits.
      --owner=STRING      Who must act (a login, team or person).
      --any               Clear when any condition holds (default: all).
      --resume            Give the ticket back to this session when the wait
                          clears.
      --timeout=STRING    Duration or RFC3339 time; past it the wait is labeled
                          needs-human and listed in attention.
  -t, --ticket=STRING     Ticket id (default: the session's current ticket).
```

## jobshop wait list

```
Usage: jobshop wait list [flags]

List waiting tickets.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

  -p, --project=STRING    Project (admin; default all).
```

## jobshop wait clear

```
Usage: jobshop wait clear <id> [flags]

Remove a ticket's wait; it becomes ready.

Arguments:
  <id>    Ticket id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

  -r, --reason=STRING    Why the wait is cleared.
```

## jobshop decision

```
Usage: jobshop decision <action> <id> [flags]

Approve, reject or dismiss a planner proposal (#603).

Arguments:
  <action>    approve (relabel the issue proposed -> target), reject (close the
              issue as not planned) or dismiss (close only the ticket).
  <id>        Decision ticket id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

  -n, --note=STRING      Note for the issue comment and the ticket.
```

## jobshop send

```
Usage: jobshop send <to> <text> [flags]

Send a message to a session or pool.

Arguments:
  <to>      Session id, session key, or pool:<name>.
  <text>    Message text. Use - to read stdin.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --interrupt         Cancel the recipient's running turn and deliver now.
      --project=STRING    Project (admin only; sessions use their own).
      --key=STRING        With pool:<chat pool>: the session key; starts the
                          session if none is open (admin).
```

## jobshop assign

```
Usage: jobshop assign <id> [<holder>] [flags]

Pre-assign a ticket to a session (admin).

Arguments:
  <id>          Ticket id.
  [<holder>]    Session id to pre-assign; omit to unassign.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --pool=STRING      Also route to this pool.
      --force            Assign even if the session is cordoned.
```

## jobshop claim

```
Usage: jobshop claim [<id>] [flags]

Claim a ticket, or the next one with --next.

Arguments:
  [<id>]    Ticket id. Omit with --next.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --next              Claim the next ticket: own in-progress work, then
                          assigned, then the pool queue.
      --project=STRING    Project (admin).
      --pool=STRING       Pool (default: the session's pool).
      --holder=STRING     Holder (admin only).
```

## jobshop heartbeat

```
Usage: jobshop heartbeat [<id>] [flags]

Renew the lease on a held ticket.

Arguments:
  [<id>]    Ticket id (default: the session's current ticket).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop release

```
Usage: jobshop release [<id>] [flags]

Hand a held ticket back to the queue.

Arguments:
  [<id>]    Ticket id (default: the session's current ticket).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

  -r, --reason=STRING    Why the ticket is handed back.
      --until=STRING     Defer the ticket: a duration (2h) or RFC3339 time.
                         Not counted as a step failure.
      --time-wait        The deferral is a real timer (a window, a backoff).
                         Without it, --until with a reason about CI, checks,
                         a PR, deploy, tag or merge is refused: use jobshop wait
                         add.
      --force            Release even though the current git branch has commits
                         no remote has.
```

## jobshop handoff

```
Usage: jobshop handoff --to=STRING [<id>] [flags]

Move a held ticket to another pool, release it and leave a note.

Arguments:
  [<id>]    Ticket id (default: the session's current ticket).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --to=STRING        Pool to move the ticket to.
      --note=STRING      Comment to add for the next holder.
      --force            Hand off even though the current git branch has commits
                         no remote has.
```

## jobshop ready

```
Usage: jobshop ready [flags]

List claimable tickets.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --project=STRING    Project.
      --pool=STRING       Pool.
  -n, --limit=INT         Max results.
```

## jobshop edge

```
Usage: jobshop edge <command> [flags]

Manage dependencies between tickets.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  edge add <from> <to> [flags]
    Add an edge: FROM is blocked by / waits for / tracks TO.

  edge rm <from> <to> [flags]
    Remove an edge.
```

## jobshop edge add

```
Usage: jobshop edge add <from> <to> [flags]

Add an edge: FROM is blocked by / waits for / tracks TO.

Arguments:
  <from>    Dependent ticket.
  <to>      Ticket it depends on.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --kind="blocks"    Edge kind.
```

## jobshop edge rm

```
Usage: jobshop edge rm <from> <to> [flags]

Remove an edge.

Arguments:
  <from>    Dependent ticket.
  <to>      Ticket it depends on.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --kind="blocks"    Edge kind.
```

## jobshop label

```
Usage: jobshop label <command> [flags]

Manage labels.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  label add <id> <label>
    Add a label.

  label rm <id> <label>
    Remove a label.
```

## jobshop label add

```
Usage: jobshop label add <id> <label>

Add a label.

Arguments:
  <id>       Ticket id.
  <label>    Label.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop label rm

```
Usage: jobshop label rm <id> <label>

Remove a label.

Arguments:
  <id>       Ticket id.
  <label>    Label.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop session

```
Usage: jobshop session <command> [flags]

Manage sessions.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  session new --agent=STRING [flags]
    Register a session and open its CLI socket (admin). Phase 2 starts agents
    automatically.

  session list [flags]
    List sessions with live host state, plus orphaned hosts (admin).

  session reap [flags]
    Stop hosts with no open session; with --closed also delete old closed
    session dirs (admin).

  session restart <id> [flags]
    Restart a session's host with a fresh spec, keeping its conversation and
    claims (admin).

  session attach <id> [flags]
    Follow a session's conversation; --rw also sends stdin lines as messages.
    Ctrl-C detaches.

  session journal <id> [flags]
    Print a session's journal frames as JSON lines, redacted, through the
    controller (own session, admin or session-journal grant).

  session prompt <id> [flags]
    Print the prompt a session's next turn would get, each part labeled by
    source (admin).

  session sandbox <id> [flags]
    Print the sandbox mounts and bwrap argv a session's next spawn would use
    (admin).

  session cordon <id>
    Take a session out of rotation: it finishes its ticket and claims nothing
    new, but stays open (admin or session-cordon grant).

  session uncordon <id>
    Return a cordoned session to rotation.

  session drain <id>
    Cordon a session and end it once it holds nothing.

  session config <id> [<options> ...] [flags]
    Set ACP config options (model=<route> effort=<level>) on one session,
    applied between turns; --reset goes back to the pool's (admin or
    session-config grant).

  session rename <name> [flags]
    Set this session's display name, a non-unique label for lists; --id renames
    another (admin or session-rename grant). An empty name restores the default.
```

## jobshop session new

```
Usage: jobshop session new --agent=STRING [flags]

Register a session and open its CLI socket (admin). Phase 2 starts agents
automatically.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --agent=STRING      Agent name.
      --pool=STRING       Pool the session works.
      --project=STRING    Project.
      --key=STRING        External key, e.g. matrix:<room>; unique while open.
      --title=STRING      Title.
      --name=STRING       Display name, a non-unique label for lists. Default:
                          the pool's display_name, else pool and ticket title.
```

## jobshop session list

```
Usage: jobshop session list [flags]

List sessions with live host state, plus orphaned hosts (admin).

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --all               Include closed sessions.
      --state=STRING      Only closed sessions with this resolution (e.g.
                          crashed).
      --since=DURATION    Only sessions closed within this long (e.g. 24h).
```

## jobshop session reap

```
Usage: jobshop session reap [flags]

Stop hosts with no open session; with --closed also delete old closed session
dirs (admin).

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --closed            Also delete directories (and worktrees) of closed
                          sessions.
      --older-than=24h    With --closed, only sessions closed longer ago than
                          this.
      --dry-run           Report what would be done without doing it.
```

## jobshop session restart

```
Usage: jobshop session restart <id> [flags]

Restart a session's host with a fresh spec, keeping its conversation and claims
(admin).

Arguments:
  <id>    Session id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --now              Cancel the running turn instead of waiting for it to
                         end.
```

## jobshop session attach

```
Usage: jobshop session attach <id> [flags]

Follow a session's conversation; --rw also sends stdin lines as messages.
Ctrl-C detaches.

Arguments:
  <id>    Session id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --since=SINCE      Start after this journal seq (default: the last 50
                         frames).
      --json             Print raw journal frames.
      --[no-]follow      Keep following new frames.
      --rw               Send each stdin line as a message (admin; the pool must
                         set attach: rw). A line starting with ! interrupts the
                         turn.
```

## jobshop session journal

```
Usage: jobshop session journal <id> [flags]

Print a session's journal frames as JSON lines, redacted, through the controller
(own session, admin or session-journal grant).

Arguments:
  <id>    Session id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --since=INT-64     Start after this journal seq.
      --limit=INT        Frames per page (default 500, max 5000).
      --kinds=STRING     Comma-separated msg.method or sessionUpdate values,
                         e.g. tool_call_update,agent_message_chunk.
      --from=""          Only frames from agent or controller.
      --follow           Keep polling for new frames until interrupted.
```

## jobshop session prompt

```
Usage: jobshop session prompt <id> [flags]

Print the prompt a session's next turn would get, each part labeled by source
(admin).

Arguments:
  <id>    Session id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --json             Print the parts as JSON.
```

## jobshop session sandbox

```
Usage: jobshop session sandbox <id> [flags]

Print the sandbox mounts and bwrap argv a session's next spawn would use
(admin).

Arguments:
  <id>    Session id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --json             Print the result as JSON.
```

## jobshop session cordon

```
Usage: jobshop session cordon <id>

Take a session out of rotation: it finishes its ticket and claims nothing new,
but stays open (admin or session-cordon grant).

Arguments:
  <id>    Session id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop session uncordon

```
Usage: jobshop session uncordon <id>

Return a cordoned session to rotation.

Arguments:
  <id>    Session id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop session drain

```
Usage: jobshop session drain <id>

Cordon a session and end it once it holds nothing.

Arguments:
  <id>    Session id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop session config

```
Usage: jobshop session config <id> [<options> ...] [flags]

Set ACP config options (model=<route> effort=<level>) on one session, applied
between turns; --reset goes back to the pool's (admin or session-config grant).

Arguments:
  <id>               Session id.
  [<options> ...]    <id>=<value> pairs, e.g. model=openai-oauth:gpt-5
                     effort=high.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --reset            Drop the session's override first, back to the pool's
                         values.
```

## jobshop session rename

```
Usage: jobshop session rename <name> [flags]

Set this session's display name, a non-unique label for lists; --id renames
another (admin or session-rename grant). An empty name restores the default.

Arguments:
  <name>    Display name. Not unique, never used for lookup.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --id=STRING        Session id. Default: the calling session.
```

## jobshop run

```
Usage: jobshop run <command> [flags]

Start and inspect workflow runs.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  run start <workflow> [flags]
    Start a workflow run: a run ticket and one step ticket per workflow step.

  run show <id> [flags]
    Show a run's state, vars and steps.

  run list [flags]
    List runs, newest first.

  run cancel <id> [flags]
    Close a run and its open steps as cancelled.

  run resume <id> [flags]
    Reopen a run's step (default: the first unfinished) with failures reset.
```

## jobshop run start

```
Usage: jobshop run start <workflow> [flags]

Start a workflow run: a run ticket and one step ticket per workflow step.

Arguments:
  <workflow>    Workflow name from the config.

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

  -v, --var=KEY=VALUE;...    Workflow var, k=v (repeatable).
  -k, --key=STRING           Dedup key: returns the open run with this workflow
                             and key instead of starting another.
      --project=STRING       Project (default: the first step's pool project).
      --after=AFTER,...      Ticket or run id the first step waits on
                             (repeatable); the edge is created with the run.
      --force                Skip the intake gate: start a Linear issue's run
                             even when linear eligible refuses it.
      --json                 Print JSON.
```

## jobshop run show

```
Usage: jobshop run show <id> [flags]

Show a run's state, vars and steps.

Arguments:
  <id>    Run id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --json             Print JSON.
```

## jobshop run list

```
Usage: jobshop run list [flags]

List runs, newest first.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --project=STRING    Project.
  -s, --state=""          Run state.
  -n, --limit=INT         Max results (default 500).
```

## jobshop run cancel

```
Usage: jobshop run cancel <id> [flags]

Close a run and its open steps as cancelled.

Arguments:
  <id>    Run id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --json             Print JSON.
```

## jobshop run resume

```
Usage: jobshop run resume <id> [flags]

Reopen a run's step (default: the first unfinished) with failures reset.

Arguments:
  <id>    Run id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --step=STRING      Step id to reopen (default: the first step not closed
                         done).
      --json             Print JSON.
```

## jobshop batch

```
Usage: jobshop batch <command> [flags]

Start ordered groups of runs.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  batch start --key=STRING --issues=ISSUES,... [flags]
    Start one run per issue under a batch, with ordered cross-run edges.

  batch show <id> [flags]
    Show a batch's state and one line per member run.

  batch list [flags]
    List open batches.
```

## jobshop batch start

```
Usage: jobshop batch start --key=STRING --issues=ISSUES,... [flags]

Start one run per issue under a batch, with ordered cross-run edges.

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

  -k, --key=STRING           Batch key: returns the open batch with this key
                             instead of starting another.
      --issues=ISSUES,...    Issue numbers, comma separated (122,123,125).
      --order=ORDER,...      Order chain a>b>c: b's first step waits on a's last
                             (repeatable).
      --workflow=STRING      Workflow each member runs (default pr-work).
      --project=STRING       Project (default: the workflow's).
      --json                 Print JSON.
```

## jobshop batch show

```
Usage: jobshop batch show <id> [flags]

Show a batch's state and one line per member run.

Arguments:
  <id>    Batch id or key.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --json             Print JSON.
```

## jobshop batch list

```
Usage: jobshop batch list [flags]

List open batches.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --project=STRING    Project.
  -n, --limit=INT         Max results (default 500).
```

## jobshop resource

```
Usage: jobshop resource <command> [flags]

Provider-owned resource slots: add, remove, unhealthy, list.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  resource add <resource> <id> [flags]
    Add a ready slot to a provider resource, or update its env (admin or pool
    grant).

  resource remove <resource> <id>
    Remove a slot; a held slot goes when its claim ends (admin or pool grant).

  resource unhealthy <resource> <id> [flags]
    Mark a slot gone; its claim is requeued without counting a failure (admin or
    pool grant).

  resource list
    List slots and resource occupancy (admin or pool grant).
```

## jobshop resource add

```
Usage: jobshop resource add <resource> <id> [flags]

Add a ready slot to a provider resource, or update its env (admin or pool
grant).

Arguments:
  <resource>    Resource name.
  <id>          Slot id.

Flags:
  -h, --help                 Show context-sensitive help.
      --root=STRING          Jobshop root directory (holds jobshop.yaml,
                             the database, sockets and sessions). Default:
                             nearest ancestor of the cwd containing jobshop.yaml
                             ($JOBSHOP_ROOT).
      --socket=STRING        Socket to talk to. Defaults to the session
                             CLI socket, else the root's admin socket
                             ($JOBSHOP_SOCKET).

      --env=KEY=VALUE;...    Env the claim gets, k=v (repeatable).
```

## jobshop resource remove

```
Usage: jobshop resource remove <resource> <id>

Remove a slot; a held slot goes when its claim ends (admin or pool grant).

Arguments:
  <resource>    Resource name.
  <id>          Slot id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop resource unhealthy

```
Usage: jobshop resource unhealthy <resource> <id> [flags]

Mark a slot gone; its claim is requeued without counting a failure (admin or
pool grant).

Arguments:
  <resource>    Resource name.
  <id>          Slot id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

  -r, --reason=STRING    Why, for the requeued ticket's comment.
```

## jobshop resource list

```
Usage: jobshop resource list

List slots and resource occupancy (admin or pool grant).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop trigger

```
Usage: jobshop trigger <command> [flags]

Inspect ticket triggers (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  trigger list
    Show each trigger: whether it holds now, last fired, open ticket, next cron
    time.
```

## jobshop trigger list

```
Usage: jobshop trigger list

Show each trigger: whether it holds now, last fired, open ticket, next cron
time.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop pool

```
Usage: jobshop pool <command> [flags]

Inspect and unquarantine pools (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  pool list [flags]
    Show each pool's labels, min, max, desired, open, working, idle and cordoned
    sessions, plus quarantined and paused pools. --label filters.

  pool scale <pool> <target>
    Set a pool's desired session count, or 'auto' to follow ready work again.
    Persisted (admin or pool-scale grant).

  pool unquarantine <pool>
    Let a quarantined or paused pool start sessions again.
```

## jobshop pool list

```
Usage: jobshop pool list [flags]

Show each pool's labels, min, max, desired, open, working, idle and cordoned
sessions, plus quarantined and paused pools. --label filters.

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

  -l, --label=LABEL,...    Only pools with this label: a label, or k=v for k:v.
                           Repeat to require several.
```

## jobshop pool scale

```
Usage: jobshop pool scale <pool> <target>

Set a pool's desired session count, or 'auto' to follow ready work again.
Persisted (admin or pool-scale grant).

Arguments:
  <pool>      Pool name.
  <target>    Session count, or auto.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop pool unquarantine

```
Usage: jobshop pool unquarantine <pool>

Let a quarantined or paused pool start sessions again.

Arguments:
  <pool>    Pool name.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop attention

```
Usage: jobshop attention [flags]

List what waits on a human, oldest first; empty when healthy (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop activity

```
Usage: jobshop activity [flags]

Recent GitHub and Linear activity in the repos and teams jobshop works, newest
first (admin or pool grant activity).

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --since=DURATION    Only items newer than this, e.g. 1h.
      --source=""         github or linear.
      --repo=STRING       owner/name, or a Linear team key.
      --kind=STRING       push, pr, merge, review, review_comment, comment,
                          issue, branch, release, ci, state, assignee.
      --actor=STRING      Login or Linear name.
      --bots=""           humans, bots or all (default).
  -t, --ticket=STRING     Only items linked to this run or ticket.
      --limit=50          Most items to print, newest first.
```

## jobshop report

```
Usage: jobshop report <command> [flags]

Report your ticket's state (progress, waiting, ready, blocked, failed, dropped),
or reliability reports (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  report incidents [flags]
    Group controller WARN/ERROR lines, crashed or stuck sessions and failed run
    steps.

  report ci-failures --repo=STRING [flags]
    Group failing tests from the failed-tests artifact of recent failed CI runs,
    with flaky/broken counts (runs gh).

  report progress [<note>] [flags]
    Report progress on your ticket.

  report waiting [<note>] [flags]
    Report waiting (on a question, a time, another change).

  report ready [<note>] [flags]
    Report the work ready for the next step.

  report blocked [<note>] [flags]
    Report blocked; labels the ticket blocked for stall-check.

  report failed [<note>] [flags]
    Report failure and why (then close -r failed).

  report dropped [<note>] [flags]
    Report the work dropped and why (then close).
```

## jobshop report incidents

```
Usage: jobshop report incidents [flags]

Group controller WARN/ERROR lines, crashed or stuck sessions and failed run
steps.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --since=24h        How far back to look. Log lines go back to the
                         controller start at most.
```

## jobshop report ci-failures

```
Usage: jobshop report ci-failures --repo=STRING [flags]

Group failing tests from the failed-tests artifact of recent failed CI runs,
with flaky/broken counts (runs gh).

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

      --repo=STRING        GitHub repo, owner/name.
      --branch="master"    Branch whose runs to read.
      --workflow="ci"      Workflow name or file; empty for all.
      --limit=5            How many recent failed runs to read.
```

## jobshop report progress

```
Usage: jobshop report progress [<note>] [flags]

Report progress on your ticket.

Arguments:
  [<note>]    Short note: what is done, what you wait for, why it is blocked or
              dropped.

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

  -q, --question=STRING    Ask a human. Files a needs-human question; the answer
                           is sent to this session.
  -t, --ticket=STRING      Ticket id (default: the session's current ticket).
```

## jobshop report waiting

```
Usage: jobshop report waiting [<note>] [flags]

Report waiting (on a question, a time, another change).

Arguments:
  [<note>]    Short note: what is done, what you wait for, why it is blocked or
              dropped.

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

  -q, --question=STRING    Ask a human. Files a needs-human question; the answer
                           is sent to this session.
  -t, --ticket=STRING      Ticket id (default: the session's current ticket).
      --until=STRING       Release the ticket deferred until then: a duration
                           (2h) or RFC3339 time.
```

## jobshop report ready

```
Usage: jobshop report ready [<note>] [flags]

Report the work ready for the next step.

Arguments:
  [<note>]    Short note: what is done, what you wait for, why it is blocked or
              dropped.

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

  -q, --question=STRING    Ask a human. Files a needs-human question; the answer
                           is sent to this session.
  -t, --ticket=STRING      Ticket id (default: the session's current ticket).
```

## jobshop report blocked

```
Usage: jobshop report blocked [<note>] [flags]

Report blocked; labels the ticket blocked for stall-check.

Arguments:
  [<note>]    Short note: what is done, what you wait for, why it is blocked or
              dropped.

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

  -q, --question=STRING    Ask a human. Files a needs-human question; the answer
                           is sent to this session.
  -t, --ticket=STRING      Ticket id (default: the session's current ticket).
```

## jobshop report failed

```
Usage: jobshop report failed [<note>] [flags]

Report failure and why (then close -r failed).

Arguments:
  [<note>]    Short note: what is done, what you wait for, why it is blocked or
              dropped.

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

  -q, --question=STRING    Ask a human. Files a needs-human question; the answer
                           is sent to this session.
  -t, --ticket=STRING      Ticket id (default: the session's current ticket).
```

## jobshop report dropped

```
Usage: jobshop report dropped [<note>] [flags]

Report the work dropped and why (then close).

Arguments:
  [<note>]    Short note: what is done, what you wait for, why it is blocked or
              dropped.

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

  -q, --question=STRING    Ask a human. Files a needs-human question; the answer
                           is sent to this session.
  -t, --ticket=STRING      Ticket id (default: the session's current ticket).
```

## jobshop answer

```
Usage: jobshop answer <question> <text> [flags]

Answer a worker's question; the answer goes to the asking session (admin,
or pool grant question-answer).

Arguments:
  <question>    Question ticket id.
  <text>        Answer text. Use - to read stdin.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop artifact

```
Usage: jobshop artifact <command> [flags]

Deliver reports for a human to read in Desk; list, show and ack them.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  artifact deliver <title> [flags]
    Deliver a report for a human to read in Desk, with optional attachments.

  artifact list [flags]
    List artifacts, unread first.

  artifact show <id>
    Show an artifact with its comments and history.

  artifact ack <id> [flags]
    Acknowledge (close) an artifact as read (admin).
```

## jobshop artifact deliver

```
Usage: jobshop artifact deliver <title> [flags]

Deliver a report for a human to read in Desk, with optional attachments.

Arguments:
  <title>    Artifact title.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

  -f, --file=STRING      Markdown body file; - or empty reads stdin.
  -a, --attach=ATTACH    File to attach (repeatable); must be inside the session
                         dir or worktree.
  -t, --ticket=STRING    Source ticket (default: the session's current ticket).
```

## jobshop artifact list

```
Usage: jobshop artifact list [flags]

List artifacts, unread first.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --unread            Only unread (open) artifacts.
      --trigger=STRING    Only artifacts from this trigger.
      --pool=STRING       Only artifacts from this pool.
      --json              Print JSON.
```

## jobshop artifact show

```
Usage: jobshop artifact show <id>

Show an artifact with its comments and history.

Arguments:
  <id>    Artifact ticket id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop artifact ack

```
Usage: jobshop artifact ack <id> [flags]

Acknowledge (close) an artifact as read (admin).

Arguments:
  <id>    Artifact ticket id.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

  -n, --note=STRING      Note, added as a comment.
```

## jobshop runbook

```
Usage: jobshop runbook <command> [flags]

Run a docs/verify runbook's read-only steps (#635).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  runbook run <file> [flags]
    Run a runbook's read-only steps (fenced blocks tagged readonly) and print
    the results.
```

## jobshop runbook run

```
Usage: jobshop runbook run <file> [flags]

Run a runbook's read-only steps (fenced blocks tagged readonly) and print the
results.

Arguments:
  <file>    Runbook markdown file, e.g. docs/verify/desk.md.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --dir=STRING       Working directory for the steps (default: the cwd).
      --timeout=2m       Timeout per step.
      --json             Print the report as JSON instead of markdown.
```

## jobshop learn

```
Usage: jobshop learn <command> [flags]

Record a correction, failure or workaround for the learning loop, or list
learnings.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  learn add <kind> <summary> [flags]
    Record a learning of any kind the config allows (learnings.kinds).

  learn correction <summary> [flags]
    Record a correction: a human or reviewer told you to do it differently.

  learn failure <summary> [flags]
    Record a failure and its cause.

  learn workaround <summary> [flags]
    Record a workaround for a tool or environment problem.

  learn list [flags]
    List learnings, newest first.

  learn update <id> ... [flags]
    Set the status or PR of learnings (curator, #273).

  learn evaluate [<id>] [flags]
    Curator gate: decide whether merged memory edits helped (#274).

  learn edits [flags]
    List merged memory edits with their baseline and evaluation (#274).
```

## jobshop learn add

```
Usage: jobshop learn add <kind> <summary> [flags]

Record a learning of any kind the config allows (learnings.kinds).

Arguments:
  <kind>       Learning kind, e.g. friction or blocker. Must match
               learnings.kinds.
  <summary>    One line: what happened and what to do instead. Use - to read
               stdin.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --fix=STRING       The fix or the rule to follow next time.
      --scope=STRING     global, repo:<owner/name> or pool:<name> (default:
                         the ticket's repo or pool).
  -t, --ticket=STRING    Ticket it is about (default: the session's current
                         ticket).
```

## jobshop learn correction

```
Usage: jobshop learn correction <summary> [flags]

Record a correction: a human or reviewer told you to do it differently.

Arguments:
  <summary>    One line: what happened and what to do instead. Use - to read
               stdin.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --fix=STRING       The fix or the rule to follow next time.
      --scope=STRING     global, repo:<owner/name> or pool:<name> (default:
                         the ticket's repo or pool).
  -t, --ticket=STRING    Ticket it is about (default: the session's current
                         ticket).
```

## jobshop learn failure

```
Usage: jobshop learn failure <summary> [flags]

Record a failure and its cause.

Arguments:
  <summary>    One line: what happened and what to do instead. Use - to read
               stdin.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --fix=STRING       The fix or the rule to follow next time.
      --scope=STRING     global, repo:<owner/name> or pool:<name> (default:
                         the ticket's repo or pool).
  -t, --ticket=STRING    Ticket it is about (default: the session's current
                         ticket).
```

## jobshop learn workaround

```
Usage: jobshop learn workaround <summary> [flags]

Record a workaround for a tool or environment problem.

Arguments:
  <summary>    One line: what happened and what to do instead. Use - to read
               stdin.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --fix=STRING       The fix or the rule to follow next time.
      --scope=STRING     global, repo:<owner/name> or pool:<name> (default:
                         the ticket's repo or pool).
  -t, --ticket=STRING    Ticket it is about (default: the session's current
                         ticket).
```

## jobshop learn list

```
Usage: jobshop learn list [flags]

List learnings, newest first.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --since=DURATION    Only learnings newer than this, e.g. 24h.
      --scope=STRING      global, repo:<owner/name> or pool:<name>.
      --kind=STRING       Kind or glob, e.g. failure or review-*.
      --status=""         open, applied or rejected.
  -t, --ticket=STRING     Only learnings about this ticket.
      --limit=50          Most learnings to print.
```

## jobshop learn update

```
Usage: jobshop learn update <id> ... [flags]

Set the status or PR of learnings (curator, #273).

Arguments:
  <id> ...    Learning ids.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --status=""         open, applied or rejected.
      --pr=STRING         PR URL that addresses them.
      --merged-at=TIME    With --status applied: when the PR merged (RFC3339,
                          default now). Starts the curator gate's window.
```

## jobshop learn evaluate

```
Usage: jobshop learn evaluate [<id>] [flags]

Curator gate: decide whether merged memory edits helped (#274).

Arguments:
  [<id>]    Memory edit id: evaluate it even when not due.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --runs=20          An edit is due after this many runs in its pools since
                         the merge.
      --age=72h          Or after this long since the merge.
```

## jobshop learn edits

```
Usage: jobshop learn edits [flags]

List merged memory edits with their baseline and evaluation (#274).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --state=""         pending, helped or not_helped.
      --limit=50         Most edits to print.
```

## jobshop config

```
Usage: jobshop config <command> [flags]

Inspect the active config (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  config show
    Print the merged config the controller runs with, and which pack or file set
    each agent, pool and trigger.

  config check [flags]
    Load and validate a config file with its packs, without a controller.
    Exit 1 on error.
```

## jobshop config show

```
Usage: jobshop config show

Print the merged config the controller runs with, and which pack or file set
each agent, pool and trigger.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop config check

```
Usage: jobshop config check [flags]

Load and validate a config file with its packs, without a controller. Exit 1 on
error.

Flags:
  -h, --help                   Show context-sensitive help.
      --root=STRING            Jobshop root directory (holds jobshop.yaml, the
                               database, sockets and sessions). Default: nearest
                               ancestor of the cwd containing jobshop.yaml
                               ($JOBSHOP_ROOT).
      --socket=STRING          Socket to talk to. Defaults to the session
                               CLI socket, else the root's admin socket
                               ($JOBSHOP_SOCKET).

      --file=STRING            Config file. Default: the root's jobshop.yaml.
      --replace=OLD=NEW,...    OLD=NEW: read packs whose path starts with OLD
                               from NEW instead, e.g. the new commit's checkout.
```

## jobshop memory

```
Usage: jobshop memory <command> [flags]

Local memory files and skills under <root>/memory (#266, #592).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  memory write <scope> [flags]
    Write a local memory file under <root>/memory and record the edit (#592).

  memory show [<scope>]
    Print the local memory files.

  memory log [flags]
    List memory edits with their evaluation, newest first.

  memory check [<dir>] [flags]
    Report memory files over the character cap (exits 1 if any).
```

## jobshop memory write

```
Usage: jobshop memory write <scope> [flags]

Write a local memory file under <root>/memory and record the edit (#592).

Arguments:
  <scope>    global, repo/<owner>__<name>, pool/<name> or skills/<name>.

Flags:
  -h, --help                       Show context-sensitive help.
      --root=STRING                Jobshop root directory (holds jobshop.yaml,
                                   the database, sockets and sessions). Default:
                                   nearest ancestor of the cwd containing
                                   jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING              Socket to talk to. Defaults to the session
                                   CLI socket, else the root's admin socket
                                   ($JOBSHOP_SOCKET).

      --learnings=LEARNINGS,...    Learning ids the edit addresses
                                   (comma-separated).
      --from-file=STRING           Read the new file text from this path.
                                   Default: stdin.
```

## jobshop memory show

```
Usage: jobshop memory show [<scope>]

Print the local memory files.

Arguments:
  [<scope>]    Only this scope (global, repo/<owner>__<name>, pool/<name>,
               skills/<name>).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop memory log

```
Usage: jobshop memory log [flags]

List memory edits with their evaluation, newest first.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --state=""         pending, helped or not_helped.
      --limit=50         Most edits to print.
```

## jobshop memory check

```
Usage: jobshop memory check [<dir>] [flags]

Report memory files over the character cap (exits 1 if any).

Arguments:
  [<dir>]    Memory directory (holds memory/ and skills/). Default: every prompt
             rule's memory.dir in jobshop.yaml, e.g. <root>/memory.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --cap=INT          Per-file character cap. Default: the rule's cap,
                         else 2200.
```

## jobshop events

```
Usage: jobshop events [flags]

Print the change stream as JSON lines (admin, for debugging).

Flags:
  -h, --help               Show context-sensitive help.
      --root=STRING        Jobshop root directory (holds jobshop.yaml,
                           the database, sockets and sessions). Default:
                           nearest ancestor of the cwd containing jobshop.yaml
                           ($JOBSHOP_ROOT).
      --socket=STRING      Socket to talk to. Defaults to the session
                           CLI socket, else the root's admin socket
                           ($JOBSHOP_SOCKET).

      --since=INT-64       Replay history after this seq first.
      --kind=KIND,...      Only these kinds (ticket, run, session); repeatable.
      --label=LABEL,...    Only tickets with one of these labels; repeatable.
      --[no-]follow        Reconnect from the last seq after a lagged event or
                           error.
```

## jobshop identity

```
Usage: jobshop identity <command> [flags]

GitHub App and Linear identities: list, check, tokens and the git credential
helper.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  identity list
    List identities and who may act as them (admin).

  identity check <target>
    Show which identity covers a repo or team and obtain a credential to prove
    it works (admin).

  identity token <target>
    Print a token for a repo (owner/name) or Linear team key, as the identity
    that covers it.

  identity git-credential <op>
    git credential helper: get returns an installation token for the repo in the
    request. Set credential.useHttpPath=true.
```

## jobshop identity list

```
Usage: jobshop identity list

List identities and who may act as them (admin).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop identity check

```
Usage: jobshop identity check <target>

Show which identity covers a repo or team and obtain a credential to prove it
works (admin).

Arguments:
  <target>    owner/repo or Linear team key.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop identity token

```
Usage: jobshop identity token <target>

Print a token for a repo (owner/name) or Linear team key, as the identity that
covers it.

Arguments:
  <target>    owner/repo or Linear team key.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop identity git-credential

```
Usage: jobshop identity git-credential <op>

git credential helper: get returns an installation token for the repo in the
request. Set credential.useHttpPath=true.

Arguments:
  <op>    get, store or erase (git passes it).

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop linear

```
Usage: jobshop linear <command> [flags]

Read, comment on and move Linear issues as the identity for their team.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

Commands:
  linear issue <key> [flags]
    Print a Linear issue with its comments as JSON.

  linear comment <key> <body> [flags]
    Comment on a Linear issue (e.g. the PR link).

  linear state <key> <state> [flags]
    Move a Linear issue to a state of its team by name (e.g. Done on merge).

  linear context <key> [flags]
    Print an issue with labels, blockers, relations, children, assignee,
    state and recent comments as JSON.

  linear list --team=STRING [flags]
    List a team's issues as JSON, one page at a time.

  linear search --team=STRING <query> [flags]
    Search a team's issues as JSON, one page at a time.

  linear catalog --team=STRING <kind> [flags]
    List a team's projects, labels, states or users as JSON.

  linear eligible <key> [flags]
    Intake gate: print whether an issue may start a run and why not. Exits 1
    when not eligible.

  linear create --team=STRING --title=STRING [flags]
    Create an issue. --key makes a retry return the first result.

  linear update <issue> [flags]
    Change title, state, assignee or labels; reads the issue back to verify.

  linear assign <issue> <user> [flags]
    Take an issue: assign it to a user, or to yourself with me (an app token
    becomes the issue's delegate). linear assign <KEY> me.

  linear label <issue> <op> <label> ... [flags]
    Add or remove labels: linear label <KEY> add|remove <label>... (same as
    update --add-label/--remove-label).

  linear ask <issue> <body> [flags]
    Comment a question, mentioning users so they are notified.

  linear relate --blocks=STRING --related=STRING <issue> [flags]
    Mark an issue as blocking or related to another.

  linear api [<query>] [flags]
    Raw GraphQL call: LINEAR_API_KEY if set, else the team's identity. Mutations
    ask the guard (linear-api-mutation).
```

## jobshop linear issue

```
Usage: jobshop linear issue <key> [flags]

Print a Linear issue with its comments as JSON.

Arguments:
  <key>    Issue key, e.g. ENG-123.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).
```

## jobshop linear comment

```
Usage: jobshop linear comment <key> <body> [flags]

Comment on a Linear issue (e.g. the PR link).

Arguments:
  <key>     Issue key.
  <body>    Comment text. Use - to read stdin.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).
```

## jobshop linear state

```
Usage: jobshop linear state <key> <state> [flags]

Move a Linear issue to a state of its team by name (e.g. Done on merge).

Arguments:
  <key>      Issue key.
  <state>    State name, e.g. "In Review" or Done.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).
```

## jobshop linear context

```
Usage: jobshop linear context <key> [flags]

Print an issue with labels, blockers, relations, children, assignee, state and
recent comments as JSON.

Arguments:
  <key>    Issue key, e.g. ENG-123.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --all-comments             Every comment (up to 250), not the 10 newest.
```

## jobshop linear list

```
Usage: jobshop linear list --team=STRING [flags]

List a team's issues as JSON, one page at a time.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --team=STRING              Team key, e.g. ENG.
      --state=STRING             State name, e.g. Todo.
      --label=STRING             Label name.
      --assignee=STRING          Assignee name, or me.
      --limit=50                 Page size (max 250).
      --after=STRING             Cursor: end_cursor of the previous page.
```

## jobshop linear search

```
Usage: jobshop linear search --team=STRING <query> [flags]

Search a team's issues as JSON, one page at a time.

Arguments:
  <query>    Search text.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --team=STRING              Team key, e.g. ENG.
      --state=STRING             State name, e.g. Todo.
      --label=STRING             Label name.
      --assignee=STRING          Assignee name, or me.
      --limit=50                 Page size (max 250).
      --after=STRING             Cursor: end_cursor of the previous page.
```

## jobshop linear catalog

```
Usage: jobshop linear catalog --team=STRING <kind> [flags]

List a team's projects, labels, states or users as JSON.

Arguments:
  <kind>    projects, labels, states or users.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --team=STRING              Team key, e.g. ENG.
```

## jobshop linear eligible

```
Usage: jobshop linear eligible <key> [flags]

Intake gate: print whether an issue may start a run and why not. Exits 1 when
not eligible.

Arguments:
  <key>    Issue key.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --skip-labels=SKIP-LABELS,...
                                 Labels that make an issue ineligible
                                 (comma-separated; the pack's
                                 linear_skip_labels).
```

## jobshop linear create

```
Usage: jobshop linear create --team=STRING --title=STRING [flags]

Create an issue. --key makes a retry return the first result.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --team=STRING              Team key, e.g. ENG.
      --title=STRING             Issue title.
      --description=STRING       Issue description (markdown). Use - to read
                                 stdin.
      --state=STRING             State name.
      --label=LABEL,...          Label name (repeatable).
      --assignee=STRING          Assignee name, display name or email; me is the
                                 token's user (an app becomes the delegate).
      --parent=STRING            Parent issue key.
      --key=STRING               Idempotency key: a repeat with the same key
                                 returns the first result instead of writing
                                 again.
      --dry-run                  Print the mutation and send nothing.
```

## jobshop linear update

```
Usage: jobshop linear update <issue> [flags]

Change title, state, assignee or labels; reads the issue back to verify.

Arguments:
  <issue>    Issue key.

Flags:
  -h, --help                       Show context-sensitive help.
      --root=STRING                Jobshop root directory (holds jobshop.yaml,
                                   the database, sockets and sessions). Default:
                                   nearest ancestor of the cwd containing
                                   jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING              Socket to talk to. Defaults to the session
                                   CLI socket, else the root's admin socket
                                   ($JOBSHOP_SOCKET).

      --allow-secret-findings      Send writes that hold a secret finding (admin
                                   callers only).

      --title=STRING               New title.
      --state=STRING               State name.
      --assignee=STRING            Assignee name, display name or email;
                                   me is the token's user (an app becomes the
                                   delegate).
      --add-label=ADD-LABEL,...    Label to add (repeatable).
      --remove-label=REMOVE-LABEL,...
                                   Label to remove (repeatable).
      --if-updated-at=TIME         Fail unless the issue's updated_at is this
                                   (RFC3339), so a concurrent change is not
                                   overwritten.
      --key=STRING                 Idempotency key: a repeat with the same key
                                   returns the first result instead of writing
                                   again.
      --dry-run                    Print the mutation and send nothing.
```

## jobshop linear assign

```
Usage: jobshop linear assign <issue> <user> [flags]

Take an issue: assign it to a user, or to yourself with me (an app token becomes
the issue's delegate). linear assign <KEY> me.

Arguments:
  <issue>    Issue key.
  <user>     me, or a team member's name, display name or email.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --state=STRING             Also move the issue to this state (e.g.
                                 In Progress).
      --key=STRING               Idempotency key: a repeat with the same key
                                 returns the first result instead of writing
                                 again.
      --dry-run                  Print the mutation and send nothing.
```

## jobshop linear label

```
Usage: jobshop linear label <issue> <op> <label> ... [flags]

Add or remove labels: linear label <KEY> add|remove <label>... (same as update
--add-label/--remove-label).

Arguments:
  <issue>        Issue key.
  <op>           add or remove.
  <label> ...    Label names.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --key=STRING               Idempotency key: a repeat with the same key
                                 returns the first result instead of writing
                                 again.
      --dry-run                  Print the mutation and send nothing.
```

## jobshop linear ask

```
Usage: jobshop linear ask <issue> <body> [flags]

Comment a question, mentioning users so they are notified.

Arguments:
  <issue>    Issue key.
  <body>     Question text. Use - to read stdin.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --mention=MENTION,...      User to notify: name, display name or email
                                 (repeatable). Their profile URL goes in the
                                 body, since a plain @name does not notify.
      --reply-to=STRING          Comment id to reply under.
      --key=STRING               Idempotency key: a repeat with the same key
                                 returns the first result instead of writing
                                 again.
      --dry-run                  Print the mutation and send nothing.
```

## jobshop linear relate

```
Usage: jobshop linear relate --blocks=STRING --related=STRING <issue> [flags]

Mark an issue as blocking or related to another.

Arguments:
  <issue>    Issue key.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --blocks=STRING            Issue key this issue blocks.
      --related=STRING           Issue key this issue is related to.
      --key=STRING               Idempotency key: a repeat with the same key
                                 returns the first result instead of writing
                                 again.
      --dry-run                  Print the mutation and send nothing.
```

## jobshop linear api

```
Usage: jobshop linear api [<query>] [flags]

Raw GraphQL call: LINEAR_API_KEY if set, else the team's identity. Mutations ask
the guard (linear-api-mutation).

Arguments:
  [<query>]    GraphQL document.

Flags:
  -h, --help                     Show context-sensitive help.
      --root=STRING              Jobshop root directory (holds jobshop.yaml,
                                 the database, sockets and sessions). Default:
                                 nearest ancestor of the cwd containing
                                 jobshop.yaml ($JOBSHOP_ROOT).
      --socket=STRING            Socket to talk to. Defaults to the session
                                 CLI socket, else the root's admin socket
                                 ($JOBSHOP_SOCKET).

      --allow-secret-findings    Send writes that hold a secret finding (admin
                                 callers only).

      --team=STRING              Team key whose identity makes the call
                                 (default: from an issue key in the variables).
  -f, --field=FIELD              String variable name=value (@file reads the
                                 file). query=... sets the query.
  -F, --raw-field=RAW-FIELD      Typed variable name=json (number, bool, object;
                                 @file reads the file).
      --input=STRING             Request body file, - for stdin: a JSON request
                                 or bare GraphQL.
```

## jobshop recall

```
Usage: jobshop recall <query> ... [flags]

Search past comments, session journals and learnings. Run it on a failure and
before asking a human.

Arguments:
  <query> ...    Words to search for.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

  -n, --limit=INT         Most hits to return (default 10).
      --scope=""          Only one source kind: comment, journal or learning.
  -t, --ticket=STRING     Only hits on this ticket.
      --project=STRING    Project (admin only; sessions search their own).
```

## jobshop journal

```
Usage: jobshop journal <command> [flags]

Search every session's journal with LogsQL on VictoriaLogs
(journal.victorialogs), and backfill it.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

Commands:
  journal search <query> [flags]
    Search shipped session journals with LogsQL, e.g. 'status:failed | stats by
    (session) count()' (admin, or journal-search grant for your project).

  journal backfill [flags]
    Ship existing journals to VictoriaLogs from their cursors now (admin).
```

## jobshop journal search

```
Usage: jobshop journal search <query> [flags]

Search shipped session journals with LogsQL, e.g. 'status:failed | stats by
(session) count()' (admin, or journal-search grant for your project).

Arguments:
  <query>    LogsQL query. Rows are deduped by (session, seq) with '| uniq by
             (session, seq)' if a resend made doubles.

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --since=DURATION    Look back this far, e.g. 2h. Default: all.
      --until=TIME        End of the window (RFC3339).
      --limit=INT         Most rows (default 1000, max 10000).
```

## jobshop journal backfill

```
Usage: jobshop journal backfill [flags]

Ship existing journals to VictoriaLogs from their cursors now (admin).

Flags:
  -h, --help              Show context-sensitive help.
      --root=STRING       Jobshop root directory (holds jobshop.yaml,
                          the database, sockets and sessions). Default:
                          nearest ancestor of the cwd containing jobshop.yaml
                          ($JOBSHOP_ROOT).
      --socket=STRING     Socket to talk to. Defaults to the session CLI socket,
                          else the root's admin socket ($JOBSHOP_SOCKET).

      --since=DURATION    Only sessions closed this recently, e.g. 168h (open
                          sessions always). Default: all.
```

## jobshop whoami

```
Usage: jobshop whoami [flags]

Show the caller's scope.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop version

```
Usage: jobshop version [flags]

Print version information.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).

      --server           Print the running controller's version instead of
                         this binary's. A session's own binary is fixed when it
                         starts, so deploy checks use this (#180).
```

## jobshop help

```
Usage: jobshop help [<command> ...] [flags]

Show help for a command, from the same text as --help: jobshop help linear
assign.

Arguments:
  [<command> ...]    Command path, e.g. linear assign. Empty prints the command
                     list.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```

## jobshop docs

```
Usage: jobshop docs [<topic>] [flags]

Print links to the docs for this binary (packs, readme, llm). Needs no root.

Arguments:
  [<topic>]    Topic (packs, desk-plugins, readme, llm). Empty lists topics.

Flags:
  -h, --help             Show context-sensitive help.
      --root=STRING      Jobshop root directory (holds jobshop.yaml,
                         the database, sockets and sessions). Default:
                         nearest ancestor of the cwd containing jobshop.yaml
                         ($JOBSHOP_ROOT).
      --socket=STRING    Socket to talk to. Defaults to the session CLI socket,
                         else the root's admin socket ($JOBSHOP_SOCKET).
```
