# batch start

On gasserver, with the dev pool running. A and B are two small open
issues with no open run, B depending on A.

```sh
jobshop batch start --key verify-$(date +%s) --issues A,B --order 'A>B'
jobshop run list -s running   # both runs listed, parent_id = the batch
jobshop ready                 # A implement listed, B implement not
```

Pass: B's `implement` is not claimed (`jobshop run show <B run>`, holder
`-`) until A's `post-merge` closes `done`, and is claimed within a tick
after. Repeating the start command prints `[existing batch for key]`.
