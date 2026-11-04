# jcode built-in memory off (#320)

2026-11-04.

Why: packs/learn curates learnings into session prompts. jcode's own
memory (`~/.jcode/memory`) injected a second, uncurated set that
duplicated and contradicted them. Agents now run jcode with
`JCODE_MEMORY_ENABLED=0` (packs examples; gasserver sets it in
`agents.jcode.env`).

Import: the gasserver store held ~1200 active entries over 102 project
files. Most were planner and dev process rules already in the pack
prompts, repeated many times, or one-off incidents (broken session
environment, section renumbering, issue counts). Six still-applicable
entries became learnings 35-40: `gh pr merge --delete-branch`, watching
the merge commit's CI run by SHA, `ctx.Err()` before `stop()`,
`reflect.DeepEqual` for config reload, ACP `SetLogger` race, gofumpt in
lint. The memory files are left in place, unused.

Check after deploy: in a new session's journal, the first prompt has no
jcode memory block and still has the learnings section.
