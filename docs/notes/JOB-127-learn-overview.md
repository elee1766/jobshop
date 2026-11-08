# Learn overview (JOB-127, 2026-11-08)

The learn view (`?view=improvements`) was two flat lists: memory changes,
then every learning. It now groups by where each problem stands, newest
first, with details behind a disclosure. No server change: it reads the
same `improvements` live section (`learnings` and `edits`).

## Units, not records

A memory edit and the learnings it cites are one item, shown once. The
edit is the change that was tried and the gate's state is its outcome. A
learning no edit cites stands alone. A cited id missing from the loaded
list (the list is capped at 500) is named in details, not dropped.

## Stages

Open by default: `not helped` (gate said no), `fix proposed` (a curator PR
is open), `evaluating` (applied, gate undecided), `awaiting curator` (open,
no PR). Collapsed: `helped`, `applied, not evaluated` (applied with no edit
or gate result, for example a `memory:<id>` write), `rejected`. A filter or
a stage chip opens what it finds. Each section shows 10 rows, then `more`.

## Problem and attempted fix

Each row shows the problem (learning summary) and the fix text the
learning recorded. The label is `attempted fix` only once a change landed
(an edit exists or status is applied), `proposed fix` while a PR is open or
was turned down, and `suggested fix` otherwise. Missing text reads `not
recorded`. Nothing is inferred: the outcome line is the gate's own reason
(`helped`, `not helped`, or `not evaluated yet`), and the before and after
table shows the gate's measures, or `not measured` before it ran.

Details hold each cited learning (kind, source, status, ticket link,
session), the edit's pools and times, the measurements and the text a
memory-file edit replaced (clipped at 8000 characters).

Checks: `docs/verify/desk-learn.md`. Spec: `ui/e2e/learn.spec.js`, which
mocks the live stream, so it needs no seeded controller.
