Routing pr-work runs (JOB-59). Every `run start {{.pack_prefix}}pr-work`
or `{{.pack_prefix}}pr-work-linear` you run passes
`--var dev_pool=<pool>`, picked from the issue (`batch start` takes no
vars: its runs stay on `{{.pack_prefix}}dev`):

- A `model:sonnet` or `model:opus` label on the issue wins:
  `{{.pack_prefix}}dev-sonnet` or `{{.pack_prefix}}dev`.
- `{{.pack_prefix}}dev-sonnet`: clear spec and acceptance checks, one
  area, known patterns, tests, docs or config, mechanical renames, a bug
  fix with a known cause.
- `{{.pack_prefix}}dev`: needs investigation or design, cross-cutting
  (store schema plus controller plus RPC), concurrency, leases or
  isolation, unclear root cause, security.
- Unsure: `{{.pack_prefix}}dev`.

Then comment the choice and a one-line reason on the issue
(`dev_pool=<pool>: <reason>`); `run show` lists the var. A failed
attempt or a review resume moves the step to `{{.pack_prefix}}dev` on
its own; never route a run back down.
