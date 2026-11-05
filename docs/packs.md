# Writing a pack

A pack is a directory with a `pack.yaml` and the prompts it names. It
declares the pools, triggers and workflows for one way of working. Most
roots have a pack written for their own flow; the packs in this repo's
`packs/` are examples of how we run Jobshop on itself, not defaults.

## Loading

`jobshop.yaml` lists packs. A relative `path` resolves against the root,
so `<root>/packs/<name>/` is the normal place:

```yaml
packs:
  - path: packs/main
    vars: {agent: jcode, checkout: /srv/widget}
```

- **Vars.** Every string in the pack and its prompt files is a Go
  `text/template` (#383): `{{.name}}` is filled from the entry's
  `vars`, else from the pack's `defaults:`. A var with neither is an
  error at load naming the field.
  `{{if .team}}...{{end}}`, `{{.cmd | default "make test"}}`,
  `{{range .repos}}` and `join`, `quote`, `lower`, `indent`, `toYAML`
  work. A value that is only `{{.n}}` keeps its type (a number stays a
  number); quote it in YAML, since a bare `{{` starts a mapping. Write a
  literal `{{` as `{{"{{"}}`. Step `title`, `prompt` and `pool` keep
  their templates until run start, where run vars (`--var issue=12`) fill
  them. The older `${name}` form was removed (#566): a pack using it
  fails to load. `${SESSION_DIR}` in agent args is not pack syntax and
  is left for the controller.
- **`prefix:`** on a `packs:` entry renames every pool, agent, trigger
  and workflow the pack declares to `<prefix><name>` and rewrites the
  references between them. Use it to load one pack several times (one
  instance per repo). Prompts see `{{.pack_prefix}}`.
- **`extends:`** in `pack.yaml` lists base packs (`path` relative to the
  pack, plus `vars`). The base loads first; the pack's own keys override
  it. `packs/selfdev` is `packs/github` with Jobshop's values.
- `jobshop config show` prints the resolved config; `jobshop reload`
  re-reads it, keeping the old one if the new one is invalid.

## The parts

**Pools** are where sessions run. Fields: `agent` (an `agents:` entry),
`repo` (a checkout; each session gets a worktree at `ref`) or `dir`,
`prompt` or `prompt_file` (the session's standing instructions),
`max_sessions`, `idle_timeout`, `turn_timeout`, `admin` (admin RPC
grants). `type: chat` makes a pool whose sessions are started per
conversation (by a gateway or `jobshop send`) instead of by tickets.

**Prompts and `prompt_rules`.** A pool prompt is sent on a session's
first turn. `prompt_rules` add text to many pools at once: `text` or
`file`, `at: system|before|after`, and a `match` on `pools`, `agents`,
`workflows`, `steps`, `labels`, `keys` or `chat`.

**Triggers** create tickets or start runs (`ticket:` or `run:`) when
`when:` holds:
- count: `pool` with `ready_below`, `open_above` and so on;
- `cron` (with `timezone`);
- `check`: a shell command; non-empty output fires, `{{.check_output}}`
  holds it;
- `stall`: stall signals (`ready_unclaimed`, `step_stalled`,
  `needs_human`, ...) each with an `after` threshold.

`cooldown` and `max_runs` limit how often a trigger fires.

**Workflows** are how work moves from one pool to the next. A run is
one ticket per step, each blocked on the one before, so step 2 is
claimable only after step 1 closes `done`. Fields: `vars` (required run
vars), `steps` (`id`, `pool`, `title`, `prompt` or `prompt_file`,
`retries`), and `escalate_pool`, where a step that used up its retries
goes for a human decision. `jobshop run resume <run> --step <id>` reopens
an earlier step, for example to send a PR back from review to implement.
An intake to implement to review flow is three steps in three pools.

**Resources** cap how many steps hold something at once (#400). They
are host-wide, so they are declared in `jobshop.yaml`, not in a pack:
`resources: {emulator: {max: 8}, pr: {keyed: true}}`. A pack step lists
`needs_resources: [emulator, "pr:{{.repo}}#{{.issue}}"]`; the config
fails to load if a named resource is not declared. A keyed resource
has `capacity_per_key` slots (default 1) per key. A step whose
resource is full is not ready, so no session starts for it; the slot
frees when the step's claim ends.

**Turning an entry off.** `enabled: false` on a pool, workflow or
trigger keeps it validated but not loaded: nothing fires, starts or
spawns, and stall checks skip it. It takes a template
(`enabled: "{{.flow_on}}"`), so a host can switch it per instance. A
live entry that needs a disabled one fails to load
([notes](notes/JOB-24-enabled-false.md)).

**Text in files.** Long text can live in files, relative to the file
that declares it: pool and step `prompt_file`, trigger ticket
`body_file`, prompt rule `file`. Set the inline field or the file
field, not both ([notes](notes/382-text-files.md)).

## A minimal pack

One `work` pool, one prompt and a `change` workflow with `implement` and
`review` steps. `packs/main/pack.yaml`:

```yaml
# A minimal pack: one pool, one workflow with two steps.
# Vars: agent, checkout.
defaults:
  sessions: 1
pools:
  work:
    agent: "{{.agent}}"
    repo: "{{.checkout}}"
    ref: origin/main
    max_sessions: "{{.sessions}}"
    prompt_file: prompts/work.md
workflows:
  change:
    vars: [task]
    steps:
      - id: implement
        pool: work
        title: "Implement: {{.task}}"
        prompt: "Make this change on a new branch and push it: {{.task}}. Close when pushed."
      - id: review
        pool: work
        title: "Review: {{.task}}"
        prompt: "Review the branch pushed for: {{.task}}. Close done if it is right, failed if not."
```

`packs/main/prompts/work.md`:

```
You work in a git worktree of this repo. Run the tests before you push.
Use `jobshop comment` for progress and `jobshop close` when the step is done.
```

`jobshop.yaml`:

```yaml
default_project: js
projects:
  js: {}
agents:
  jcode: {command: jcode, args: [acp]}
packs:
  - path: packs/main
    vars:
      agent: jcode
      checkout: /srv/widget
```

Try it:

```
jobshop serve &
jobshop config show
jobshop run start change --var task="add a --verbose flag" --key verbose
jobshop run list
```

A copy lives in `pkg/config/testdata/minimal`; `TestMinimalPackDoc`
loads it. Keep the two in sync.

## Tailoring to a flow

- **No review step (trunk-based).** One step whose prompt says to push
  to the main branch once the tests pass.
- **Human approval before merge.** A `hooks:` entry `on: merge` whose
  command answers `{"decision":"ask_human"}`; the merge waits on a
  `needs-human` ticket (`docs/notes/188-guardrail-hooks.md`). Or end the
  review step's prompt with `jobshop report waiting ... --question`.
- **Intake or triage in front.** A `triage` pool and a trigger (`check`
  on new issues, or a count) that starts a `triage` step; its prompt
  runs `jobshop run start change ...` for work worth doing.
- **Linear instead of GitHub issues.** Prompts use `jobshop linear`
  instead of `gh`; a `linear_issue_no_run` stall signal starts runs.
  `packs/github` with `tracker: linear` shows the whole setup.
  Anything the fixed subcommands miss: `jobshop linear api` (raw
  GraphQL, like `gh api`; mutations ask the `linear-api-mutation`
  guard, `docs/notes/546-linear-api.md`).
- **Several repos.** One `packs:` entry per repo, each with its own
  `prefix:` and `vars`.
- **A scheduled report (#508).** A cron trigger starts a one-step run
  whose prompt ends with `jobshop artifact deliver`. The artifact
  carries the trigger name, so Desk's Artifacts view groups the
  reports by it:

  ```yaml
  workflows:
    weekly-report:
      steps:
        - id: report
          pool: work
          title: Weekly report
          prompt: >-
            Summarize this week's merged PRs and open incidents in
            report.md. Then run `jobshop artifact deliver "Weekly report"
            --file report.md` and `jobshop close`.
  triggers:
    weekly-report:
      when: {cron: "0 8 * * 1"}
      run: {workflow: weekly-report}
  ```

## Example packs

- `packs/github`: works one GitHub repo (or Linear team) through issues
  and PRs: planner, `pr-work` (implement, review-merge, post-merge),
  `ci-repair` and `stall-check`.
- `packs/selfdev`: `packs/github` extended with the values we use to
  build Jobshop itself.
- `packs/operator`: a chat pool for the operator to manage tickets,
  issues and runs across instances.

`jobshop docs packs` prints links to this guide and `packs/` for the
binary's commit.
