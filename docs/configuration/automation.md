# Starlark automation reference

2026-11-07. JOB-112. See [callback wiring and lifecycle](callbacks.md),
[pack composition](packs.md), [migration](migration.md), and the
[verification runbook](../verify/cue-starlark.md).

## Execution and trust boundary

A callback runs once with recorded input, returns actions, and exits. It
cannot suspend until a step completes. The controller invokes a later
callback for the next relevant event. Ordinary Starlark `if`, loops,
functions, lists and dictionaries perform routing and transformation.
There is no expression-template interpreter inside callback arguments.

Treat `.star` files as trusted executable code. Starlark restricts the
language, but `shell()` can exercise the controller user's OS permissions
and `http` can reach services available to that user. Review a third-party
pack like a program, not inert YAML. Agent-step sandbox and isolation
profiles do not sandbox a controller-side hook. A callback cannot grant
itself a capability or expand host limits.

Action constructors are data, not effects. Immediate calls such as
`http.get`, `http.post`, `shell` and `jev` run during callback evaluation.
Returning no actions does not undo those calls. Do not use a hook to bypass
admission, review, exact-head evidence, approval, merge or human sign-off.

## Module loading

An entry is `file.star:exported_function`. Files must end in `.star`.
Entry selection is trusted filesystem configuration: relative entries
resolve from the pack asset directory, while absolute paths and `../`
entries can select a file elsewhere. Functions beginning with `_` are
not entry points. `load("helpers.star", "helper")` loads a local module
relative to the importing module. Absolute `load` paths, escapes from the
resolved module root, cycles and remote imports fail, including symlink escapes.
Loaded globals are frozen. Each invocation has separate callback state,
not durable module-global mutation.

Resolve an entry relative to its pack asset directory. The resolved entry's
containing directory then becomes its module and `prompt_file` root. For
an entry under `scripts/`, place its helpers and prompt files under that
directory too. `load` cannot escape to a parent pack directory. These
module boundaries do not sandbox shell filesystem access.

Loading parses and evaluates modules and verifies named exports. Capability
calls and `print` are unavailable while loading, even when the host grants
them during invocation. Put calls inside functions, not top-level module
initializers. `config check` compiles modules, checks exports and verifies
that each function accepts its callback's positional arguments without
resolving host credentials or making shell, network or model requests.
Static validation cannot prove every branch's dynamic argument types or
selected references. Built-ins and the controller validate those at the
call boundary and before applying actions.

## Data and action values

Payload maps use dictionary access, for example
`event.payload["repository"]["full_name"]`. Envelope objects supplied as Go
structs use attributes. Do not assume an arbitrary dictionary supports
attribute access. Structured data contains strings, booleans, finite
numbers, null/`None`, lists and string-keyed maps. Cycles, unsupported values,
non-string keys and excess depth/size fail. Do not parse agent prose to
recover structured output.

Constructors accept named arguments only:

| Constructor | Required | Optional |
| --- | --- | --- |
| `start_run(...)` | `workflow`, `key` | `inputs={}`, `title=""`, `reason=""` |
| `start_step(...)` | `id`, `pool` (or `placement`), exactly one of `script`, `prompt` or `prompt_file`, or `type="command"` with `run` | `inputs={}`, `title=""`, `retries`, `retry_pool=""`, `needs=[]`, `output_schema={...}`, `needs_resources=[]`, `reason=""`; agent steps: `placement={...}`, `max_review_rounds`, `on_request_changes=""`; with `script`: `worktree=False`, `timeout="10m"`, `retry_after="1m"`; with `type="command"`: `worktree=False`, `timeout="10m"`, `retry_after="1m"`, `on_restart="retry"` (see [Command steps](callbacks.md#command-steps)) |
| `finish_run(...)` | `outcome` | `output={}`, `reason=""` |
| `create_ticket(...)` | `title`, `key` | `pool=""`, `body=""`, `labels=[]`, `priority` (0-4, default 2), `parent=""`, `reason=""` |
| `comment(...)` | `ticket`, `body` | `reason=""` |

`create_ticket` files a task. Its `key` dedups per trigger or workflow: while
a ticket it filed with that key is open, the action is a no-op. The project
is the pool's, else the parent's, else the run's. A trigger's tickets carry
`meta.trigger` and count toward its `max_open`, like its runs. `comment`
adds a comment to any existing ticket. Both apply in the same transaction as
the rest of the batch.

Names and keys must be nonempty strings. `inputs` and `output` are
string-keyed dictionaries. An explicit `retries` is an integer from 0
through 100. Omitting it uses the controller's default failure threshold
of 2. Explicit zero does not select that default. See the
[retry and output contract](callbacks.md#structured-step-output-and-retries).
Constructors return immutable action values. Return `result(actions=[...],
output={...})` to combine selected actions with arbitrary object data:

```python
return result(
    actions=[start_run(workflow="review", key="pr-42")],
    output={"summary": "review requested", "pr": 42},
)
```

Both arguments are optional. `actions` accepts an action or a list/tuple of
only action constructors; `output` must be a string-keyed dictionary.
One action, an action list/tuple, and `[]` remain shorthand returns. A bare
dictionary is output data with no actions. Scalars and non-action lists
are rejected. No-action results do not finish a workflow automatically.
The controller still validates outcome words, step IDs, references and scope.

The Go return type is `automation.Result`: typed `metadata` (entry, loaded
source digest, capability call count), typed `actions` and `decisions`, captured
`logs`, and an `output` object (`map[string]any`). Metadata is host-owned and
is present even on errors. Scripts cannot override it with constructor
arguments or output dictionary keys. Output retains exact integers and is
independent of `finish_run(output=...)`, which publishes workflow output.
Durable invocation and attempt evidence stores `result_metadata` and
`result_output`. Public evidence applies credential redaction and bounded
artifact handling to both, including after interrupted-plan recovery.

Use `reason` for a short policy rationale when selecting a branch, such as
`reason="small change needs one review"`. It is recorded with the returned
action on the invocation attempt. The host does not infer a reason from
Starlark control flow. An omitted reason stays empty, not a fabricated
explanation. Jev's validation/fallback reason is separate from this
author-supplied routing rationale.

A `prompt_file` is a file reference, not shell source. Prompts retain their
existing Go-template rules and must receive explicit input values. The
controller rejects routing a workflow to another instance's private
pool. `retry_pool` follows the same scope rule. `needs` is a list of
nonempty dependency step IDs and cannot contain the step's own ID.
Constructing a name as a string does not bypass reference checks.

## Host capability configuration

This belongs in `jobshop.yaml`, never in a pack:

```yaml
automation:
  http: true
  shell: false
  timeout: 30s
  max_steps: 1000000
  output_bytes: 1048576
  credentials:
    github:
      origins: [https://api.github.com]
      header: Authorization
      prefix: "Bearer "
      secret: {type: file, path: /run/secrets/github-token}
jev:
  enabled: false
```

`http`, `shell`, `github` and `linear` default to false. Jev defaults to disabled and is
configured separately. A credential declaration does not enable HTTP.
Secret values are resolved by Go, not exposed as script input or copied
into the pack. Secret references use a path string,
`{type: file, path: ...}`, or `{type: command, command: ...}` as described
in [running a root](../running.md#secret-files). The illustrative
`{file: ...}` shorthand from early proposals is not valid.

An HTTP grant permits requests to local and private-network destinations
as well as public hosts. Credential origin restrictions limit where a
named secret attaches, not where an unauthenticated request can go. The
service schema does not make `credentials.origins` a global egress
allowlist. Restrict controller networking at the host when required.

### Execution ceilings

Unset settings select bounded engine defaults. Numeric zero for
`max_steps` or `output_bytes` also selects its default. An explicitly set
`timeout` must be a positive Go duration, not `0s`:

| Limit | Default |
| --- | --- |
| Invocation/load duration | 30 seconds |
| Starlark execution steps | 1,000,000 |
| Captured print, HTTP response, shell output limit | 1 MiB |
| Loaded source bytes | 1 MiB |
| Local modules | 64 |
| Structured argument/result/action data | 1 MiB |
| Returned actions | 1,000 |
| Immediate capability calls per invocation | 100 |
| Structured nesting depth | 64 |

The service exposes `timeout`, `max_steps`, and `output_bytes`. Other
ceilings are engine defaults, not pack-tunable fields. `output_bytes`
bounds captured print and each HTTP/shell result separately, not total
heap allocation across all calls. A call's shorter `timeout` can reduce
but cannot extend the invocation deadline. Cancellation
propagates to requests and child commands. A budget failure is an error,
not a silently truncated successful result. These budgets are not an
in-process heap sandbox: native Starlark operations can allocate memory
before structured-result limits are checked. This is another reason packs
must be trusted code.

## HTTP

```python
response = http.get(
    "https://api.github.com/repos/acme/widget/pulls/42",
    auth = "github",
    headers = {"Accept": "application/vnd.github+json"},
    timeout = "10s",
)
if response.status_code != 200:
    fail("pull request lookup failed")
pr = response.json()
head = pr["head"]["sha"]
```

Available methods are `http.get` and `http.post`:

```text
http.get(url, headers={}, auth="", timeout="", body="", json=None)
http.post(url, headers={}, auth="", timeout="", body="", json=None)
```

- `url` is an explicit absolute HTTP(S) URL with a host, no embedded
  credentials and no fragment. There is no opaque service ID destination.
- `headers` maps header names to strings. `auth` selects a host credential
  handle. Do not pass secret header values directly in pack data.
- POST accepts a string `body` or JSON-like `json`, not both. `json` sets
  `Content-Type: application/json`. GET rejects a nonempty body.
- The response has `status_code`, `body` and `text` (the same bounded
  string), lowercase-keyed `headers`, and `json()` for strict JSON decoding.
  Invalid JSON raises an error when `json()` is called.
- Non-2xx status codes are returned to the script. Test them explicitly.
  Transport, cancellation and oversize responses raise errors. There is
  no automatic HTTP retry policy hidden in the built-in.

Credential handles bind to exact scheme, hostname and effective port:
`https://api.github.com` equals port 443, not HTTP or another port.
Origins have no path/query. Host-provided credential headers override
caller headers. Authenticated cross-origin redirects fail rather than
forwarding a token. Unauthenticated cross-origin redirects drop all
caller headers, since those too could contain credentials. Redirects are
bounded to ten hops and no cookie jar supplies ambient credentials.
Never rely on a redirect to select a permitted credential destination.

Use endpoint idempotency keys for POST mutations and reconcile after an
unknown outcome. A timed-out request may already have changed the server.

## GitHub

`automation.github: true` enables the `github` built-in. Each call takes a
token from the GitHub identity (`identities:`, kind `github-app` or
`github-token`) covering the call's repo, minted or read per call through
the controller's identity cache, so an App's hourly tokens and a rotated
token file need no reload. The token never becomes a script value and is
redacted from logs, errors and the call journal. There is no fallback to
the `gh` CLI or `GH_TOKEN`.

```python
pulls = github.get("/repos/acme/widget/pulls", params = {"state": "open"}, pages = 3).json()
github.post("/repos/acme/widget/issues/7/comments", json = {"body": "done"})
data = github.graphql("query($o: String!) { organization(login: $o) { login } }", variables = {"o": "acme"}, repo = "acme/widget")
```

```text
github.get(path, repo="", params={}, headers={}, json=None, pages=1, timeout="")
github.post / patch / put / delete(path, repo="", params={}, headers={}, json=None, timeout="")
github.graphql(query, variables={}, repo="", timeout="")
```

- `path` is an API path without query; `params` become the query string.
  `repo` (owner/repo) selects the identity and defaults to the repo in a
  `/repos/<owner>/<repo>/` path. With neither, a `github-token` identity
  answers; an App needs a repo for its installation.
- REST calls return the same response value as `http` (`status_code`,
  `body`, `headers`, `json()`). Non-2xx is returned, not raised.
- `pages` (GET, up to 10) follows `rel="next"` links and returns one JSON
  array of all pages. The combined body counts against `output_bytes`.
- `graphql` returns the `data` object and raises on `errors` or a non-200
  status, with GitHub's messages.
- Redirects off the identity's API origin fail. The script cannot set
  `Authorization`.

## Linear

`automation.linear: true` enables the `linear` built-in, the same way:
each call takes a token from the Linear identity (kind `linear-app` or
`linear`) covering the call's team, minted or read per call, redacted and
never a script value.

```python
data = linear.graphql(
    "query($k: String!) { issue(id: $k) { identifier state { name } } }",
    variables = {"k": "ENG-12"},
    team = "ENG",
)
```

```text
linear.graphql(query, variables={}, team="", timeout="")
```

- `team` (a key such as `ENG`) selects the identity. Without it the only
  Linear identity answers; with several, the call raises.
- It returns the `data` object and raises on `errors` or a non-200
  status, with Linear's messages. Mutations go through like queries: the
  `linear-api-mutation` guard covers `jobshop linear api`, not callbacks,
  so a mutation belongs in reviewed pack code.

## Store reads, json and time

```text
tickets(kind="", status=None, project="", pool="", label="", workflow="", key="", state="", parent="", holder="", meta={}, order="newest", limit=100, body=False)
runs(workflow="", key="", status=None, state="", meta={}, order="newest", limit=100)
sessions(pool="", status=["open", "in_progress"], limit=100)
```

Each returns a list of ticket dictionaries as `jobshop ticket show --json`
prints them, with `fields` and `meta` decoded and `body` dropped unless
`body=True`. `status` is a string or list of `open`, `in_progress`,
`closed`. `meta` maps up to three `"pack.key"` names to string values,
for example `{"trigger.name": "scan"}`. `limit` is 1 to 500. Reads see the
whole store, are journaled and replayed like other calls, and count toward
the call limit. They do not fence: a decision made on a read can race a
write, and the controller still checks every returned action.

`json.encode`, `json.decode` and `json.indent` are Starlark's `json`
module. `time` is Starlark's `time` module; `time.now()` is the time the
invocation was accepted, so a resumed attempt sees the same value.

## Shell

```python
result = shell(
    ["/srv/tools/classify-change", head],
    cwd = "/srv/widget",
    env = {"LANG": "C.UTF-8"},
    timeout = "30s",
)
if result.exit_code != 0:
    fail("classification failed")
classification = result.stdout.strip()
```

Signature: `shell(argv, cwd="", env={}, timeout="")`.

`argv` is a nonempty list of strings. Values from events belong in separate
arguments, not a generated `sh -c` string. There is no implicit shell.
Choosing `["sh", "-c", script]` deliberately runs trusted shell code and
restores its injection risks. `env` maps valid variable names to strings.
Use absolute executable paths for reproducible behavior.

The result has `stdout`, `stderr`, and integer `exit_code`. A nonzero
command exit is data for the callback to handle. Spawn failure,
cancellation or output overflow is an error. Stdout and stderr share one
output budget. Linux cancellation terminates the command process group.
A callback's command inherits the controller's environment, plus
explicit `env` values, so user-manager tools such as `systemctl --user`
work. A step function's `shell()` runs in its pool's sandbox and gets the
sandbox environment instead. Use the `github` built-in rather than `gh`.

The controller supplies the Jobshop root as the default working directory.
Set `cwd` explicitly when a tool depends on a checkout or configuration
directory.

Shell runs as the controller OS user, outside the agent sandbox. A timeout
cannot undo filesystem writes or external actions the command completed.
Do not automatically retry a command with ambiguous side effects.

## Jev

Jev is an optional decision capability, not permission enforcement. Host
service configuration supplies the full Decisions endpoint and secret:

```yaml
jev:
  enabled: true
  endpoint: https://jev.example.com/v1/decisions
  api_key: {type: file, path: /run/secrets/jev-api-key}
```

The endpoint above is illustrative. Use the full endpoint provided by your
Jev service, not a guessed path. Packs cannot enable Jev or override its
endpoint/key. Neither the key nor a secret reference becomes a Starlark
value. With Jev disabled, a call fails as an unavailable capability with a
callback traceback.

Signature:

```text
jev(instructions, options, state={}, default="", timeout="10s")
```

```python
route = jev(
    instructions = "Choose an implementation route from the supplied options.",
    options = {
        "mechanical": "Specified mechanical edit",
        "scoped": "Scoped implementation requiring local judgment",
        "primary": "Investigation or higher-risk change",
    },
    state = {"classification": classification},
    default = "primary",
    timeout = "10s",
)
selected_pool = route.choice
```

`options` maps 2 through 255 distinct IDs to nonblank descriptions. IDs
are 1 through 64 characters without surrounding whitespace. Descriptions
and `instructions` are nonblank and at most 4,000 characters. `state` is a
string-keyed dictionary of bounded JSON-like values. A nonempty `default`
must name a supplied option. The encoded request is capped at the smaller
of the structured-data limit and 80 KiB. The response is capped at the
smaller of the output limit and 256 KiB.
The result has `choice`, `reason`, `defaulted` and `confidence` attributes.
`confidence` is `None` when absent or a number from 0 through 1.
`choice` always belongs to the supplied options. `reason` reports local
validation or fallback context, not private model reasoning.

A valid default is used only after transport failure, a call timeout or a
non-2xx response. A malformed response, unoffered choice or invalid type is
an error, not a default. Parent invocation cancellation never becomes a
fallback. Disabled Jev is an error even if `default` was supplied. Calls
do not silently retry. The engine separately returns decision records so
the controller can retain the selected route even when the callback's
return value contains only actions.

Use only bounded options supplied by the pack. Do not use a decision to
grant a capability, waive a review, authorize funds or replace human
sign-off. Keep existing agent triage during a format migration. Changing
the routing policy to Jev is a separate policy change that needs measured
acceptance.

## Diagnostics

CUE errors include constraint/source positions and instance context.
Starlark parse, load and invocation errors include the module, function,
line and traceback. Built-in argument errors name the offending argument
and required type or constraint. Do not suppress these errors by returning
an arbitrary fallback action.

For a callback error, inspect the controller's run/trigger context, then
read the earliest failing `.star` line and built-in parameter. Reproduce
with the same input and host grants, not with broadened permissions.
`print` is bounded captured diagnostic output, not a durable workflow
state channel. Never print payload credentials or full HTTP headers.
Static config checks do not execute every branch, and successful loading
does not establish that an external endpoint or model service is healthy.

| Diagnostic | Safe response |
| --- | --- |
| CUE conflict/incomplete value | Correct the named source constraint or supply a concrete required var, then recheck before reload |
| Module export/load error | Check the entry function, local module root and traceback. Do not enable capabilities for module initialization |
| Missing dictionary key/wrong type | Inspect the admitted payload or recorded output contract. Do not infer a richer source event |
| Disabled HTTP/shell/Jev | Review whether the host should grant that specific capability. A pack cannot grant it |
| Credential origin/redirect error | Use the intended exact origin. Do not broaden origins to follow an unexpected redirect |
| Invalid action/schema/reference | Repair the whole batch. No earlier action is accepted as a partial success |
| Stale/superseded invocation | Inspect the newer recorded run/step revisions before proposing a continuation |
| Pinned hook source changed | Restore the pinned revision/paths for existing work or use the approved drain/migration procedure |
| Unknown external outcome | Reconcile the remote operation or command effects before any retry |
| Execution/output ceiling exceeded | Reduce work and bounded data. Do not silently return a successful fallback route |

Known host credential values are redacted from captured output and error
diagnostics as well as capability observations. Redaction is not an
arbitrary secret detector, a payload-admission rule or a guarantee that
derived/encoded secrets cannot be printed. Keep secrets out of script
inputs, action data, completion output and external response bodies.
