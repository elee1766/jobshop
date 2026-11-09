# JOB-170: learnings and memory edits under pkg/knowledge

2026-11-09. Child of JOB-164.

`pkg/learning`, `pkg/memedit` and `pkg/memory` are gone. Their paths are not
forwarded.

| Was | Now |
|---|---|
| `pkg/learning` | `pkg/knowledge` (`AddLearning`, `GetLearning`, `ListLearnings`, `LearningsBetween`, `LearningsByPR`, `CountLearnings`, `UpdateLearning`, `SetLearningGate`, `RecallLearnings`) |
| `pkg/memedit` records | `pkg/knowledge` (`AddPREdit`, `AddWriteEdit`, `GetEdit`, `ListEdits`, `SetEditResult`, `EditRef`) |
| `pkg/memedit` gate | `pkg/knowledge` (`EditMetrics`, `DecideEdit`, `GateWorse`, `GateSimilarMin`, `ReviewFindingPrefix`) |
| `pkg/memory` | `pkg/knowledge/memory`, unchanged API |

New in `pkg/knowledge`, moved out of `pkg/control`: `Measure`, `RecordPREdit`,
`EditDue`, `SettleEdit`, and `LocalMemory` (`Write`, `Restore`, with the mutex
that orders local writes and restores). The two collections shared an
id-allocation loop and a compare-and-set loop; both now use `putNext` and
`casUpdate`.

Control keeps the decisions: the `memory-write` grant, the cap and the baseline
and gate thresholds, which pools a learning points at (`editPools`, it reads
tickets), when an evaluation is due, and logging.

Unchanged: collection names and specs, record JSON, the `memory:<id>` value in
a learning's `pr` field, the unique PR on an edit, ids, the recall project and
ticket filters. Existing databases need no migration.

A local write is a file change followed by SQL records, not one transaction.
`LocalMemory.Write` puts the file back when the edit record fails. A failure
marking a cited learning applied leaves the file and the edit in place, as
before. `SettleEdit` stores the result on the edit first, so a failure copying
the gate line to a learning does not make the next evaluation decide again.

`pkg/knowledge/memory` imports nothing of the module (rule `independent-leaves`),
so the prompt rule and `jobshop memory check` do not link learning storage.
