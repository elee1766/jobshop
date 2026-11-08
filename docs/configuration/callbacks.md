# Callback wiring, events and transitions

2026-11-07. JOB-112. See the [built-in/action reference](automation.md),
[CUE pack reference](packs.md) and [migration guide](migration.md).

## Workflow and trigger registration

A workflow uses either its static `steps` or both callback entry points,
not a mixture. In an instance's CUE definition:

```cue
workflows: hooked: {
    enabled: V.enabled
    project: "js"
    vars: ["task", "pool"]
    on_start: "workflow.star:begin"
    on_step_completed: "workflow.star:completed"
}
```

`project` supplies the workflow's project without inferring it from a
statically declared first step. Entry paths resolve from the instance's
explicit asset directory. A callback workflow must return an explicit
`finish_run` when its work is complete. Having no more static steps does
not complete it automatically.

A trigger declares a handler plus static `args`, rather than a static
`run`, `ticket` or `command` action:

```cue
triggers: "demo-event": {
    enabled: false
    on: [{source: "generic", event: "tasks"}]
    handler: "events.star:handle"
    args: {workflow: "demo-hooked", pool: "demo-dev"}
    cooldown: "1m"
    max_open: 1
}
```

This is the fixture's deliberately disabled custom-event trigger. The
script receives fully qualified target names through `args`. Do not
assume CUE prefixes arbitrary strings inside your payload/arguments. Keep
local/shared scope explicit and let controller validation reject cross-
instance routing.

Existing GitHub and Linear `on` filters retain their nested source forms,
for example `on: [{github: {event: "issue_comment", action: "created",
repo: "acme/widget"}}]`. Custom source matching uses
`{source: "generic", event: "tasks"}`. It does not register a new network
listener or bypass ingest authentication/admission.

## Callback signatures

| Entry | Signature | When |
| --- | --- | --- |
| Workflow `on_start` | `begin(run)` | Initial transition for the persisted run |
| Workflow `on_step_completed` | `completed(run, step)` | A relevant recorded finished-step transition |
| Trigger `handler` | `handle(event, args)` | An admitted source delivery, a timer or fallback wakeup, or a stall |
| Hook or desk command `handler` | `fn(input, args)` | A guard check or a Desk `api.exec`, while the caller waits |

The function name is user-chosen. Its positional argument contract is not.
`config check` rejects an export that cannot accept its callback's
positional arguments, before the controller accepts the configuration.
Return `result(actions=[...], output={...})` for selected actions and arbitrary
object output, or return an action/list as shorthand. A dictionary alone
records output without actions. `[]` and `result()` are no-ops, not
implicit success or commands to finish a run. Unknown/malformed action
references fail before the batch is accepted. A missing payload key or
wrong argument type produces a contextual Starlark traceback.

Envelope objects expose attributes. Their dictionary-valued members use
keys. For example `run.inputs["task"]`, not `run.inputs.task`.

### Run envelope

| Attribute | Meaning |
| --- | --- |
| `id` | Persisted run ticket ID |
| `workflow` | Resolved workflow name, including its instance prefix |
| `key` | Existing logical run dedup key |
| `project` | Resolved project |
| `inputs` | Structured persisted run inputs, a string-keyed dictionary |
| `vars` | Existing string-valued run variables, a dictionary |
| `state` | Recorded run state string |
| `steps` | List of recorded step snapshots |

For a callback workflow, `inputs` contains the resolved pack vars, then
explicit structured run inputs, then supplied scalar run vars in that
precedence order. `run start --var name=value` supplies strings.
`start_run(inputs=...)` preserves structured members; a member also named
in the workflow's required `vars` must be scalar and is also passed through
that string-valued compatibility path. `vars` remains the separate run-var
dictionary, not an alias for every structured input. Declare and validate
the input shape your callback actually consumes.

### Step envelope

| Attribute | Meaning |
| --- | --- |
| `id` | Logical step ID from the action, not a ticket ID |
| `ticket_id` | Persisted step ticket ID |
| `pool` | Resolved selected pool |
| `status`, `resolution` | Recorded ticket completion state |
| `outcome` | Recorded step outcome used for routing |
| `output` | Structured step output dictionary |
| `inputs` | Structured inputs supplied to this step |
| `failures`, `retries` | Controller-owned failure count and retry policy |

Use `step.id` for workflow branching and `step.ticket_id` for ticket
inspection. Outcome and structured output are different fields. A human
completion message is not structured output. Check required output keys
and types explicitly before they select a pool or construct a prompt.

### Event envelope

| Attribute | Meaning |
| --- | --- |
| `source`, `event` | Source name and delivered event kind |
| `delivery` | Source delivery identity used for repeated-delivery handling |
| `repository` | Repository/team metadata from the adapter |
| `payload` | Effective admitted JSON object, or original payload if no admitted replacement exists |
| `fields` | Legacy matched event fields as a dictionary |

The source adapter defines payload shape. Custom payloads do not require a
new Go struct. There is no fabricated complete PR on an `issue_comment`
payload. `repository` is adapter metadata, not a claim that every custom
sender includes a GitHub repository object. There is no user-settable
envelope version in this form. Version a custom producer's own payload
explicitly when it needs schema evolution.

## Executable workflow example

The fixture's `packs/demo/workflow.star` starts one step:

```python
def begin(run):
    return [start_step(
        id = "implement",
        pool = run.inputs["pool"],
        prompt_file = "implement.md",
        inputs = {"task": run.inputs["task"]},
        title = "Implement: " + run.inputs["task"],
        retries = 1,
        output_schema = {
            "type": "object",
            "required": ["summary"],
            "properties": {"summary": {"type": "string"}},
            "additionalProperties": False,
        },
    )]

def completed(run, step):
    if step.id != "implement":
        fail("unexpected completed step: " + step.id)
    outcome = step.outcome if step.outcome else step.resolution
    if not outcome:
        fail("completed step has no recorded outcome or resolution")
    return [finish_run(outcome = outcome, output = step.output)]
```

This is a minimal API example, not a production shipping workflow. In a
disposable copy with a deliberately enabled pool and controller, start it
with the same required vars:

```sh
jobshop run start demo-hooked --var task="Add a verbose flag" --var pool=demo-dev --key demo-widget-1
```

That command writes a run. Do not run it on the live host as a validation
shortcut. The checked-in example keeps pool capacity zero.

To route after triage, branch on a recorded outcome and construct only the
selected step. Use a dictionary of explicit allowed pool names and reject
an unknown route. Preserve shipping, review, post-merge and failure phases
from the source workflow. The [jobcity inventory](jobcity.md) records why
a shorter implementation-only callback is not a complete migration.

## Custom event example

The fixture's `event.json` is:

```json
{"event":"tasks","tasks":[{"key":"widget-1","task":"Add a verbose flag"}]}
```

Its handler validates each task before returning actions:

```python
def handle(event, args):
    if event.source == "timer":
        return []
    tasks = event.payload.get("tasks")
    if type(tasks) != "list":
        fail("event.payload.tasks must be a list")
    actions = []
    for task in tasks:
        if type(task) != "dict":
            fail("each task must be a dictionary")
        key = task.get("key")
        text = task.get("task")
        if type(key) != "string" or not key:
            fail("task.key must be a nonempty string")
        if type(text) != "string" or not text:
            fail("task.task must be a nonempty string")
        actions.append(start_run(
            workflow = args["workflow"],
            key = "demo:" + key,
            inputs = {"task": text, "pool": args["pool"]},
            title = "Implement: " + text,
        ))
    return result(actions=actions, output={"selected": len(actions)})
```

The engine tests load these exact files and invoke their functions. To
exercise ingest in a disposable root, configure a service-owned generic
endpoint and its HMAC secret using [generic ingest](../ingest.md#generic-endpoints),
map `kind` to payload field `event`, then deliberately enable the trigger.
Trusted local ingress uses `jobshop ingest put --endpoint <name>
--delivery-id <stable-id> < event.json`. Reusing an unchanged delivery ID
is a duplicate delivery, not a new business subject. Keep subject keys
stable across both event delivery and reconciliation.

The `demo:` key prefix is example-specific. Do not substitute it for an
existing production key such as `stall:review_requested:<subject>`.
Returning this event batch does not perform a shell command, HTTP request
or model decision. Those separate capability examples live in
`packs/demo/capabilities.star` and remain unregistered by default.

## Guard hooks and desk commands

A guard hook (`on: claim`, `action` or `merge`) and a `desk_commands`
entry may name a Starlark `handler` instead of a `command` or `url`. The
handler is called as `fn(input, args)` while the caller waits, within the
entry's `timeout` (30s default), with the entry's static `args`:

```cue
hooks: [{
    on: "action"
    actions: ["push"]
    handler: "guard.star:push"
    args: {protected: ["master", "main"]}
    timeout: "5s"
}]
desk_commands: "deploy-state": {handler: "deploy.star:state", timeout: "5s"}
```

A hook's `input` is the guard event dict (`on`, `action`, `target`,
`detail`, `session`, `pool`, `project`, `ticket`, `extra`). It returns
`{"decision": "allow"|"deny"|"ask_human", "reason": "..."}`; `{}` is
allow, and an error, a timeout or another decision is deny. A desk
handler gets the plugin's JSON input and returns the dict the plugin
receives as `output`. Neither may return actions. `on: intake` has no
handler form.

Each call is journaled with state `sync_done` or `sync_failed` and a
`sync` field naming `hooks.<on>` or `desk_commands.<name>`, so
`automation list --state sync_done` and `automation show` list its
capability calls. Nothing replays or retries a sync call. Its journal is
deleted after 72 hours.

## Script steps

`start_step(script=fn, pool=...)` starts a step function instead of an
agent step (JOB-124). `fn` is a top-level function of a loaded module,
or the string `file.star:function` relative to the callback's module
directory. It must accept `(run, step)`, the same envelopes a completion
callback gets.

The controller runs the function in-process as a durable step: it claims
the step, holds a lease for `timeout` (default `10m`) plus a minute,
retries up to `retries`, and escalates an exhausted step like any other.
A controller restart reruns the attempt. The step ticket has no pool, so
no session claims it. `pool` names where `shell()` runs:

- In the pool's sandbox, user and cgroup when it has them, else
  directly as the controller user.
- With `worktree=True`, in a worktree of the pool's repo under the run
  directory, shared by every script step of the run. Otherwise in the
  pool's `dir` or repo. A relative `cwd` resolves from there.
- With the step environment of a command step: `JOBSHOP_BIN`,
  `JOBSHOP_TICKET`, `JOBSHOP_STEP_KEY`, plus the call's `env`.

`http` and `jev` run from the controller with the service grants, the
same as in callbacks. Each call is recorded in the automation journal:
`jobshop automation list --run <id>` shows one `step_done`,
`step_failed` or `step_interrupted` record per attempt, and `automation
show` lists its calls. Recorded results are never reused: a retry
reruns the whole function, so make side effects safe to repeat.

Return `result(outcome="word", output={...})`. The outcome routes the
next callback (`step.outcome`); the output is checked against
`output_schema` and becomes `step.output`. A `fail(...)` or other error
is a failed attempt. A step function cannot return actions.

```python
def begin(run):
    return start_step(id = "check", pool = "dev", script = check,
                      worktree = True, timeout = "20m", retries = 0,
                      inputs = {"runbook": run.inputs["runbook"]})

def check(run, step):
    r = shell(["jobshop", "runbook", "run", step.inputs["runbook"]])
    if r.exit_code != 0:
        fail("runbook failed")
    return result(outcome = "pass", output = {"report": r.stdout})
```

## Structured step output and retries

`start_step(output_schema=...)` accepts a JSON Schema object, not a CUE
expression string. The controller validates the schema before creating
the step. The example requires an object with one string member,
`summary`. A completion publishes that object separately from its outcome
word and free-text comment:

```sh
jobshop close <step-ticket-id> --resolution done --outcome implemented --output-file /path/to/result.json
```

For the example, `result.json` contains `{"summary":"Added a verbose
flag"}`. `--output-file -` reads stdin. The CLI requires exactly one JSON
object of at most 1 MiB. It cannot share stdin with `--comment -`. The
controller checks successful output against the recorded schema. Invalid
or host-limit-exceeding output is recorded as a failed attempt, with an
explanation, instead of being accepted as a successful result. It does not
route on unvalidated agent prose. Output without a declared schema still
must be structured data. Do not assume every JSON Schema dialect/keyword
is supported: unsupported constructs fail schema validation rather than
being ignored.

The controller's output map ceiling is `automation.output_bytes`, or
1 MiB when unset. A file accepted by the CLI can still exceed a smaller
host ceiling. The `request_changes` resolution accepts findings through
`--comment`, not structured output or an outcome override.

Retry accounting remains in Go. A failed close, release without closing,
or expired step lease increments the persisted failure count once. A step
reopens while `failures < retries`. Reaching the threshold closes it
failed and moves the run toward `needs_human`, with the existing
escalation policy. The name `retries` is a failure threshold, not a count
of extra attempts after the initial attempt. Omitting it selects 2.
Explicit 0 or 1 exhausts on the first counted failure. `retry_pool`
selects an allowed fallback pool under the existing controller rules.
Starlark does not sleep, poll the step or perform its own retry loop.

A successful finished step closes `done` or controller-owned `skipped`.
Only a recorded finished transition is an `on_step_completed` input.
An exhausted failure does not masquerade as a successful completion
callback. Human recovery, review rounds, resource admission and lease
ownership remain controller concerns.

## Action-batch and lifecycle contract

1. Persist a run or admitted source delivery and its callback input.
   Run inputs and step snapshots are event-time values, not live handles.
2. Evaluate the callback outside the controller's scheduling locks under
   the invocation limits. Immediate capabilities may have effects here.
3. Validate the complete returned batch, including target instance/project,
   step IDs, dependencies, prompt paths and schemas. Compare the recorded
   run/step revisions before accepting its graph.
4. Persist the prepared graph, then apply the batch atomically. A malformed
   later action does not leave earlier action-created tickets behind.
5. Record the accepted transition. A later finished step produces the next
   callback input. `[]` acknowledges a no-op transition, not run completion.

`start_run` is allowed from a trigger or a workflow callback. It uses the
resolved workflow and logical key for **open-run** deduplication, including
duplicate starts in one batch. A closed historic run is not a permanent
business-key tombstone. Reusing its key can deliberately start new work.
Preserve any existing per-subject history cap in the producer/reconciler.

`start_step` and `finish_run` require a workflow invocation. Step IDs must
be unique within the run. `needs` refers to existing steps or earlier steps
in the same batch, not forward references. Pool/retry-pool selection stays
inside the declaring instance and run project. A local reference may
resolve through the declaring instance's prefix. A shared root definition
can be referenced explicitly. Another instance's private pool/workflow
cannot be selected by constructing its prefixed name.

`finish_run` requires every recorded step to be finished. It cannot be
combined with a new `start_step` or a second finish. It closes the run with
the supplied outcome and retains its output dictionary. A callback workflow
does not auto-finish because all current steps are done. If it returns
`[]` instead, the run remains open pending another relevant transition.

The controller journals invocation identity, module digest, inputs,
read-set revisions, action plan, captured output, decisions and errors.
Re-delivery of the same source delivery and resumption of an already
accepted graph do not create a second transition. Changed/stale run or
step revisions invalidate action acceptance. This does **not** roll back
an HTTP request or command already executed while evaluating that input.

Callback identity binds the loaded `.star` bytes, callback selector,
asset and configuration roots, declaring instance scope, and host capability
authority. Accepted last-good metadata preserves that identity and rejects
drift rather than silently accepting current scripts. The module digest does
not cover arbitrary `prompt_file` contents. A dynamic prompt
is confined, rendered and persisted with the accepted action plan instead.
A pending invocation whose active source no longer matches fails rather
than silently running new code against the old snapshot. Preserve the
exact source and module paths for in-flight runs, or drain them before
changing execution shape.
Reload is not an automatic upgrade of persisted runs.

## Capability records and ambiguous outcomes

Each immediate HTTP, shell or Jev call is recorded before execution. The
record starts `UNKNOWN`, then becomes `completed` or `failed` if the
controller records the result. It includes the invocation/call identity,
timing, redacted inputs and bounded result/error. Large observations spill
to protected files under `<root>/automation/artifacts`; retain these with
the store when investigating or backing up audit state. They may still
contain non-secret private business data.

The automation configuration does not set an automatic retention/expiry
period for invocation/call records or artifact files. Include both in the
operator's protected retention and backup plan. Do not remove evidence for
an unresolved effect, a pending invocation or an in-flight run to reduce
disk usage. Store backup alone does not include a filesystem artifact.

An `UNKNOWN` record means the effect might have happened. A recorded
failure, especially a timeout, is also not proof of no remote mutation.
The controller never treats an existing unknown/failed observation as an
unexecuted call. During recovery of the same invocation, it may reuse a
completed saved result only when the call index, capability and input
digest match and redaction has not changed that result. A mismatched,
missing-artifact or redacted result fails closed rather than issuing the
effect again. This is bounded result reuse, not an exactly-once guarantee
or a general workflow replay engine. Reconcile with the external system,
use stable idempotency keys and obtain approval for an irreversible
continuation. Reapplying a prepared graph does not rerun capabilities. See
[rollback and reconciliation](migration.md#8-handle-ambiguous-side-effects-and-roll-back).

## Inspect and resume recorded work

In Desk, open **System**, then **automation audit**. Select an invocation,
filter its current summary window by state/run/trigger/delivery, or enter
an older invocation's complete ID. A resulting ticket also links back to
its automation invocation. Inspection shows public event identity,
ordered attempts and capability calls, timing, sanitized errors,
captured output, Jev choices, planned actions and resulting ticket links.
Private recovery inputs are not exposed through the audit API.
An invocation can be inspected before it creates any ticket. Parent
invocation links preserve causality across callback-created work.

The summary window is the latest 50 invocations. Selected detail displays
at most 50 attempts and 200 calls and reports truncation. Large bodies
remain artifact references. Filtering that window is not a search across
all historic records. The admin RPC surface is:

| Method | Parameters | Result |
| --- | --- | --- |
| `automation.list` | Optional `run`, `trigger`, `state`, `limit` | Invocation collection records, newest first |
| `automation.show` | `id` equal to the complete invocation key | `invocation`, `attempts`, `calls` |

The CLI exposes the same read-only admin operations:

```sh readonly
jobshop --root "$JOBSHOP_ROOT" --socket "$JOBSHOP_ROOT/admin.sock" automation list --run "$RUN_ID" --limit 50
jobshop --root "$JOBSHOP_ROOT" --socket "$JOBSHOP_ROOT/admin.sock" automation show "$INVOCATION_ID"
```

Set `RUN_ID` to the run ticket ID. Choose `INVOCATION_ID` from a list
record's `Key`, not from its run or step ID. `list` also accepts `--trigger`
and `--state`. Both commands print JSON without a `--json` flag. `show`
returns `invocation`, `attempts` and `calls`; collection records contain
`Key`, `State` and a `Value` object. Inspect `Value.actions` and
`Value.targets` on the invocation and each attempt to distinguish a planned
action from its resulting ticket IDs. A callback that only finishes a run
can have no new ticket targets. An empty call list is expected when the
callback makes no immediate capability calls.

These require the admin socket. Do not grant worker sessions broad audit
access just to debug a callback. The public projection omits private recovery
snapshots, redacts known host credentials and sensitive fields, and bounds
large values with artifact references. Other business data can remain private.

After inspecting the failure and correcting a recoverable host/dependency
problem, a permitted operator can deliberately resume the same run:

```sh
jobshop run show <run-id> --json
jobshop run resume <run-id>
```

`run resume` is a write, not an audit command. It can requeue a failed hook
even before that hook created a step. It preserves the invocation identity,
requires the pinned code digest, refreshes the run/step read set, and reuses
only safe completed calls. Unknown, failed or redacted call observations
refuse retry. There is no force option that turns them into unexecuted
calls. Reconciling an external outcome does not itself rewrite its journal
record or make it reusable. Obtain an operator-reviewed recovery decision
for those cases. Do not edit SQL, erase the call record or invent a new key
to bypass the safety check.

For an ordinary failed agent step, the existing resume command resets its
failure count according to the run engine. Distinguish that from retrying
callback evaluation and from retrying an external effect. A code bug in a
pinned callback is not fixed by silently swapping in a new script for the
old invocation. Preserve the source or use the migration/drain procedure.

## Event wakeups and reconciliation

Scheduled/fallback callbacks use `event.source == "timer"` and
`event.event == "timer"`, empty `payload`/`fields`, and a scheduler-derived
delivery key. A webhook handler must not assume every call contains a
source payload. The demo explicitly returns `[]` for timer wakeups because
it has no external authoritative task source. A production reconciliation
handler must instead list/reread that source and apply the same stable
subject keys used by its webhooks.

Keep source-specific `on` filters, `fallback` intervals and `when.cron`
schedules distinct. Timers are wakeups, not evidence of review/delegation
authority. A trigger with `on:` and a handler polls every `fallback` with
no `when` conditions: the handler reads its source and decides.
`event.trigger` names the trigger the handler runs for.

A handler may combine with `when.stall`. Go keeps detection, the per-stall
`max_runs`, the cooldown, `max_open` (default 3) and the #704 subject hold,
and calls the handler once per fireable stall with `event.source ==
"stall"`, `event.event` the signal, `event.delivery` the stall id
`<signal>:<subject>` and `payload` holding `signal`, `subject`, `detail`,
`key` (the default run key) and `runs` (earlier runs for this stall). The
run a handler starts carries `meta.trigger.stall`, so it holds the stall
and counts toward `max_runs` whatever key it picks. `args.workflow` names
the workflow whose own runs and steps the signals skip. A handler that
returns `[]` is offered the stall again after a cooldown (30 minutes
without one) while the signal persists.

A signal that is not a store signal may name a `script` instead of a shell
`check`: `{signal: "pr_unmerged", script: "signals.star:prs", args: {...},
poll: "10m"}`. The controller calls `fn(args)` outside the tick at most
once per `poll` (5 minutes by default) and keeps the answer in between.
The function returns a dict of subject to detail (a leading `#` is
dropped), may use `github`, `http` and store reads, and is journaled like
a desk handler under `stall.<signal>`. An `issue_no_run` subject with an
open run in the signal's `repo` is skipped as for a check.

For handler triggers, a terminal invocation consumes its cron/cooldown slot
even when it returns `[]`. Handler cooldown runs from the later of the
accepted fire and the close of the newest run or ticket the trigger filed.
`max_open` counts the trigger's open runs and `create_ticket` tasks, not
their steps, plus the proposed new batch. A started run's
`meta.trigger.key` is `<trigger>/<run key>`. An unset
or zero handler `max_open` does not impose that cap. A batch that would
exceed a configured cap fails as a whole, rather than partially creating the
first few runs.
