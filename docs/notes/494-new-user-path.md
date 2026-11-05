# New-user path test (#494)

2026-11-05. `TestNewUserPath` in `pkg/control/newuser_test.go` follows
`docs/packs.md` the way a new user would: `jobshop init` (through
`cli.Main`), the guide's minimal pack written to `<root>/packs/main`,
the guide's `jobshop.yaml`, a controller, the guide's `jobshop run start`
line through `cli.Main`, then it waits for the run's close event and
checks it finished `done`.

The four fenced blocks of the "A minimal pack" section are read from
the doc at test time, in order: `pack.yaml`, `prompts/work.md`,
`jobshop.yaml`, the try-it lines. Only the agent command (the test's
fake agent) and the `checkout` var (a temp git repo with `origin/main`)
are replaced. Editing the guide's pack into something invalid, for
example a step pool `wrok`, fails the test at config load. Adding or
removing a fenced block in that section fails it too: keep the four.

It runs in the normal `go test ./...` CI job with in-process hosts, with no
network or GitHub.
