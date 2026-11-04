# Ingest pipeline end to end (#444)

2026-11-04. Covers #318: receiver, spool, drain, admission, dispatch,
dedupe, gap fill and split mode. Run on gasserver as the jobshop user
with the admin socket. Step checks for single stages are in
`ingest-pollers.md`, `ingest-dispatch.md` and `ingest-workers.md`.

Setup (the host may lack these tools):

```sh
R=/var/lib/jobshop
q() { nix shell nixpkgs#sqlite -c sqlite3 -readonly "$R/jobshop.db" "$@"; }
sp() { nix shell nixpkgs#sqlite -c sqlite3 -readonly "$R/ingest/spool.db" "$@"; }
hmac() { nix shell nixpkgs#openssl -c openssl dgst -sha256 -hmac "$(tr -d '\n' <"$1")" -r | cut -d' ' -f1; }
GH=/run/secrets/jobshop-github-webhook-secret
URL=http://127.0.0.1:8740
```

## 1. Receiver up and signature checked

```sh
curl -s -o /dev/null -w '%{http_code}\n' -X POST -H 'X-GitHub-Event: ping' \
  -H 'X-Hub-Signature-256: sha256=00' -d '{}' $URL/github/webhook
B='{"zen":"verify"}'
curl -s -o /dev/null -w '%{http_code}\n' -X POST -H 'X-GitHub-Event: ping' \
  -H 'X-GitHub-Delivery: verify-ping' \
  -H "X-Hub-Signature-256: sha256=$(printf %s "$B" | hmac $GH)" -d "$B" $URL/github/webhook
```

Pass: `401`, then `204` (good signature, ping filtered, nothing spooled).

## 2. Signed GitHub delivery spools, then drains

Stop the controller's drain from racing the check only if you want to see
the spool row; otherwise skip the `sp` line.

```sh
B=$(jq -nc '{action:"created",repository:{full_name:"elee1766/jobshop"},
  issue:{number:444},comment:{id:1,body:"verify ingest",user:{login:"elee1766"},
  updated_at:"2026-11-04T00:00:00Z"},sender:{login:"elee1766"}}')
curl -s -o /dev/null -w '%{http_code}\n' -X POST -H 'X-GitHub-Event: issue_comment' \
  -H 'X-GitHub-Delivery: verify-gh-1' \
  -H "X-Hub-Signature-256: sha256=$(printf %s "$B" | hmac $GH)" -d "$B" $URL/github/webhook
sp "select delivery_id from spool where delivery_id='verify-gh-1'"
jobshop ingest drain --once
q "select endpoint, kind, number, state from inbox where delivery_id='github:verify-gh-1'"
```

Pass: `202`; the spool row (if seen) then `moved N (new N), bad 0`
(or `moved 0` if the controller drained first); the inbox row is
`jobshop  issue_comment  444  dispatched` (or `admitted` for a moment).
Sending the same request again gives `202` and no second inbox row.

## 3. Generic HMAC endpoint

Add and restart `jobshop-ingest.service` (serve reads config once):

```yaml
ingest:
  endpoints:
  - {name: verify, source: generic, secret_file: secrets/verify-hmac,
     signature_header: X-Sig, delivery_id_header: X-Id,
     map: {kind: kind, repo_or_team: repo, number: n}}
```

```sh
B='{"kind":"verify","repo":"x/y","n":7}'
curl -s -o /dev/null -w '%{http_code}\n' -X POST -H 'X-Id: verify-gen-1' \
  -H "X-Sig: $(printf %s "$B" | hmac $R/secrets/verify-hmac)" -d "$B" $URL/generic/verify
jobshop ingest drain --once
q "select source, kind, number, state from inbox where delivery_id like '%verify-gen-1'"
echo "$B" | jobshop ingest put --endpoint verify --delivery-id verify-gen-2
```

Pass: `202`; row `generic  verify  7  ...`; `ingest put` prints `spooled`,
and `duplicate` when run twice.

## 4. Hook results stored on the row

Add, reload (`jobshop config check` passes first):

```yaml
ingest:
  admission:
  - {name: tag, type: mutating, match: {sources: [generic]},
     command: 'jq -c "{patch:[{op:\"add\",path:\"/payload/tagged\",value:true}]}"'}
  - {name: deny-test, type: validating, match: {sources: [generic]},
     command: 'jq -c "if (.payload|tostring|test(\"jobshop-deny-test\")) then {decision:\"deny\",reason:\"test\"} else {decision:\"allow\"} end"'}
```

Put `{"kind":"verify","repo":"x/y","n":8}` as `verify-gen-3`, then:

```sh
q "select state, hook_results, admitted from inbox where delivery_id like '%verify-gen-3'"
```

Pass: `hook_results` has `tag` with `decision":"mutated"` and a `diff`
adding `/payload/tagged`, and `deny-test` `allow`; `admitted` contains
`"tagged":true`.

## 5. Denied never reaches a session, admitted does

Run `docs/verify/ingest-dispatch.md` (comment `jobshop-deny-test` on an
issue a working ticket follows). Pass: that row is `denied` with the
hook's reason and the session journal has no message for it; a normal
comment goes `dispatched` and the session gets exactly one message.

## 6. Webhook and poller give one row

With `ingest.pollers` covering elee1766/jobshop (`ingest-pollers.md`
step 1) and the webhook live, comment once on an issue, wait one poll
interval:

```sh
q "select delivery_id, state from inbox where natural_key like 'github:comment:elee1766/jobshop:<comment id>:%'"
```

Pass: exactly one row; its delivery id is `github:<guid>` or
`poll:...`, whichever came first. The session got one message.

## 7. Gap fill after a receiver outage

```sh
systemctl stop jobshop-ingest.service
# comment on a followed issue
systemctl start jobshop-ingest.service
```

Pass: within one poll interval (poller) or one `ingest.github_reconcile`
(15m, github-app identity only) the comment has one inbox row and is
`dispatched`. With neither configured, this step fails by design: note
it as a gap, not a bug. GitHub's own retry does not redeliver failed
webhooks.

## 8. External worker mode

Set `ingest.workers: external`, `ingest.worker_missing: 1m`,
`ingest.backlog_after: 1m`, reload, keep `jobshop-ingest-work.service`
stopped.

```sh
jobshop attention          # after 1m: ingest-worker
# comment on a followed issue, wait 1m
jobshop attention          # also ingest-backlog
jobshop ingest workers     # pending count > 0
systemctl start jobshop-ingest-work.service
jobshop ingest workers     # worker listed, recent seen, pending 0
jobshop attention          # neither ingest item
```

Pass as commented; the row's controller log line is
`inbox admission ... by=worker:<id>`. Revert to in-process workers and
remove the step 3 and 4 config afterwards.
