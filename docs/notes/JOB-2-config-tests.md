# Config tests on fixtures, not in-repo packs

2026-11-05, JOB-2.

- `TestPacksGolden` and `testdata/packs.golden` are gone, as are the
  `pack_test.go` tests that asserted what packs/selfdev, github and
  operator say (pool counts, caps, prompt phrases, admin lists, vars).
- Loader features are tested on fixture packs written in the test:
  `pkg/config/loader_test.go` (prefix renaming and instances, tracker
  prompt_file selection, typed `{{.n}}` vars and run-time render,
  escalate blocks, host strings left alone) next to the existing
  `TestPackMerge`, `TestPackExtends`, `TestPackRenamedVars`,
  `TestGoTemplatePack`, `TestPromptRules*` and `TestWorkflow*`.
- `TestPacksLoad` is the only test that reads packs/: each
  `packs/*/pack.yaml` loads with minimal vars, asserting only no error.
  A new pack needs its minimal vars added there.
- pkg/control tests that used packs/ for setup (`TestOperatorGrant`,
  `TestLearnNudge`, `TestImprovementSession`, `selfdevRoot`) use inline
  fixture packs. `TestAgentPoolPack` still runs packs/agent end to end
  but asserts only that its run finishes.
- Editing a file under packs/ needs no test change unless it breaks
  loading.
