This ticket is for {{.repo}}, not the repo your pool prompt names. Your
worktree is a checkout of {{.repo}}; run gh there or pass `-R {{.repo}}`.
{{.guide}} Check before pushing: `{{.check_cmd}}`. {{.repo}}'s
workflows start with `{{.pack_prefix}}` (`jobshop run start
{{.pack_prefix}}pr-work ...`). A ticket you file for its code work needs
`--meta '{"github":{"repo":"{{.repo}}"}}'` to reach its sessions.
