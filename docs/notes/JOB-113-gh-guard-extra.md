# JOB-113: jobshop gh guard extra

2026-11-06. Behavior as built.

`detail` on a `jobshop gh` guard event is still the body. The event's
`extra` now holds the parsed write, so a hook can refuse what is not in
a body (a review request to a given person, a label, an api call):

```json
{"argv": ["pr","edit","5","--add-reviewer","elee1766"],
 "command": "pr edit",
 "title": "...",
 "reviewers_added": ["elee1766"], "reviewers_removed": [],
 "labels_added": [], "labels_removed": [],
 "assignees_added": [], "assignees_removed": [],
 "method": "POST", "path": "repos/o/n/pulls/5/requested_reviewers",
 "fields": [{"key": "reviewers[]", "value": "elee1766"}],
 "input": "..."}
```

- `argv` is the gh arguments as run (after `--key` and
  `--allow-secret-findings` are cut). `command` is the subcommand words.
- Empty fields are left out. Comma lists (`--add-reviewer a,b`) are split.
- `title`: pr and issue create and edit, release create.
- Added reviewers, labels and assignees: `pr create` (`-r/--reviewer`,
  `-l/--label`, `-a/--assignee`), `issue create`, and `pr|issue edit`
  (`--add-*`, `--remove-*`). Short flags are read only in the commands
  where gh gives them that meaning (`-r` is `--rebase` in `pr merge`).
- `gh api`: `method`, `path`, `fields` in argument order (`-f`, `-F`,
  `--raw-field`, `--field`; a `-F k=@file` or `@-` value is the content
  read), and `input` (the `--input` file or stdin).
- Tests: `TestGHExtra`, `TestGHGuardExtra`.
