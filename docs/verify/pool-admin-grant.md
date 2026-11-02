# Pool admin grant (#126)

On gasserver, with the operator pack live:

1. Start an operator session (`jobshop session new --pool operator`) and
   from its worktree run `jobshop session list` and `jobshop reload`.
   Both succeed.
2. From the same session run `jobshop drain`. It fails with
   `needs the admin socket or pool admin grant "drain"`.
3. From a `dev` session, `jobshop session list` still fails.
