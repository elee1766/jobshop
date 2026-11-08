# Shared GitHub and Linear clients

2026-11-08. Each API has one client package that holds its HTTP and auth
code, and every caller goes through it:

- `pkg/linear.Client` and `pkg/github.Client` ask an `apicall.Source` for the
  token on every request. On a 401 they call `Renew(stale)` once and resend
  only if the token changed.
- The controller keeps one client per identity (`pkg/control/apiclients.go`).
  A github-app has one per installation owner, since each installation has its
  own token. Its reads and writes, the ingest poller, activity, identity check,
  reconcile and the `github` and `linear` built-ins all share these instances.
  `identitySource` reads the identity from the live config at each request, so
  a reload that rotates a key applies on the next call.
- Per-call settings ride on the context: `apicall.WithRequestTimeout` (the
  Linear pollers' 15s per request, JOB-103) and `apicall.WithObserver`. The
  built-ins use the observer to redact the token they send and to mark the
  request as started and completed in the call journal.
- The CLI's `linear` commands use a `controllerToken` source. It asks the
  controller once per command, then names the refused token on a 401 (#866).
  `LINEAR_API_KEY` is `apicall.Static`.
- A repo that no identity covers still falls back to the host's gh token
  (`hostToken`) for activity, the poller and decision writes. The `github`
  built-in has no fallback.

Still on the gh CLI: thread comments (`ghSource`) and the wait reads
(`ghWaitSource`, including `gh pr checks --required`).
