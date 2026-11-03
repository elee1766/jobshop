# batches: state, show, list

On gasserver, with the dev, refinery and deploy-check pools running.
A and B are two small open issues with no open run, B depending on A.

```sh
key=verify-batch-$(date +%s)
jobshop batch start --key $key --issues A,B --order 'A>B'
jobshop batch list            # the batch is listed
jobshop batch show $key       # state open or running, B blocked by A
```

Repeat `jobshop batch show $key` as the runs progress:

- Once A `implement` is claimed: state `running`, A on `implement`,
  B `blocked by A`.
- After A `post-merge` closes: A `done`, B no longer blocked and on
  `implement`.
- After B `post-merge` closes: state `done`, the batch ticket closed
  `done`, and `jobshop batch list` no longer lists it.

Pass: the batch reaches `done` with no manual edits. If a member goes
`needs_human`, the batch shows `needs_human` and returns to `running`
after `jobshop run resume <run>`. Link the output on #145.
