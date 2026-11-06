# 508: Artifacts

2026-11-05.

An artifact is a deliverable a human reads in Desk: a `message` ticket to
`human` with label `artifact`, like a question (#186). It is never
claimed and is not an orphan. Open means unread; `ack` closes it
`acknowledged`, and ticket history records who and when.

- `jobshop artifact deliver "<title>" --file report.md [--attach f ...]
  [-t <ticket>]` (RPC `artifact.deliver`, session or admin). Body is the
  file or stdin. `meta.artifact` holds `ticket` (held ticket or `-t`),
  `run` (the step's run), `trigger` (the run's `meta.trigger.name`),
  `pool`, `session` and `files [{name, size, type, pruned}]`. A `tracks`
  edge goes from the artifact to the source, and the source gets the
  comment `artifact delivered: <id> <title>`.
- Attachments: absolute paths, resolved with symlinks, must be inside
  the session dir (which holds the worktree). Admin may pass any path.
  The controller checks names and the `artifacts.max_bytes` total
  (default 10MB) before writing, copies into a temp dir and renames it
  to `<root>/artifacts/<id>/`, so a refusal or a failed copy writes
  nothing.
- `artifact list [--unread] [--trigger] [--pool] [--json]`, `artifact
  show <id>`, `artifact ack <id> [--note]` (ack is admin).
- Retention: the tick prunes hourly. An artifact acked more than
  `artifacts.retention` (default 90d) ago loses its attachment dir and
  its files are marked `pruned`; the ticket stays.
- Desk (2026-11-06, JOB-80: the built-in view and unread badge were
  removed; the API and the `artifact` plugin slot stay, see
  `docs/notes/JOB-80-remove-artifacts-view.md`): `?view=artifacts`, nav
  shows the unread count. `GET
  /api/artifacts`, `/api/artifacts/{id}`, `/api/artifacts/{id}/files/{name}`
  and `POST /api/artifacts/{id}/ack` (CSRF header, as scale and cordon).
  Files are served only by a name listed in meta, opened with `os.Root`,
  with `nosniff`, `Content-Disposition` and a sandbox CSP. png, jpeg,
  gif and webp are inline; text is served as `text/plain`; svg and the
  rest download. The body is markdown rendered with `marked` and
  sanitized with DOMPurify (`markdown()` in `ui/src/lib/format.js`).
  Keys: `j`/`k`, `a` acks and moves on, `o` opens the source.
- Gateway: `gateway.<name>.artifact_room` gets one line per created
  artifact (title, source, Desk link from `desk_url`).
