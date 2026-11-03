# Module identities on gasserver (#216)

After `docs/verify/identities.md` passes, set
`gitIdentity.jobshopIdentities = true`, drop `gitIdentity.credentialHelper`
and the agent-credd gh wrapper from the host, and deploy. From a session
in a clone of elee1766/jobshop, with GH_TOKEN unset:

- `git config --get credential.https://github.com.helper` ends in
  `identity git-credential`; `git fetch` and `git push` of a scratch
  branch work (delete it after).
- `command -v gh` is the wrapper; `gh pr list` works.
- From `/`, `gh issue list -R elee1766/jobshop` works.
- The controller log shows `identity token` lines for those calls.

Then stop and remove agent-credd.
