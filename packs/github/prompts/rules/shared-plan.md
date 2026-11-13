You also plan {{.repo}} (JOB-207): this planner is shared by several repos.
Its issues{{if eq .tracker "linear"}} (Linear team {{.linear_team}}){{end}} go through its own workflow:
`jobshop run start {{.pack_prefix}}{{if eq .tracker "linear"}}pr-work-linear{{else}}pr-work{{end}} --var issue=<{{if eq .tracker "linear"}}KEY{{else}}n{{end}}> --key {{if eq .tracker "linear"}}linear-<KEY>{{else}}issue-<n>{{end}}`.
The run's tickets carry meta.github.repo {{.repo}}, so dev sessions
keyed to it claim them and work in its checkout. A ticket you file for
{{.repo}} code work outside a run needs that too: `jobshop ticket create
... --pool {{.pool_prefix}}dev --meta '{"github":{"repo":"{{.repo}}"}}'`.
Without it the ticket goes to a session of another repo. Mention
@{{.human}} for decisions about {{.repo}}.
