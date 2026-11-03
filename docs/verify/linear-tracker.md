# Linear tracker (#201)

Needs a Linear API key for a throwaway team (needs-human for the token
only). Put it in a file, add `identities.<name>: {kind: linear,
token_file, workspace, teams: [<KEY>]}` and list it on the project, then
`jobshop reload`. From a session of that project:

- `jobshop linear issue <KEY>-1` prints the issue with comments.
- `jobshop linear comment <KEY>-1 "test from jobshop"` shows up in Linear.
- `jobshop linear state <KEY>-1 Done` moves it; a bad name lists states.
- `jobshop linear issue OTHER-1` fails with `no identity covers OTHER`.
