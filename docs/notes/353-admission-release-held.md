# Admission: release answered holds (#353)

2026-11-04.

Admission moves an item to `held` on `ask_human` and files a needs-human
ticket keyed `admission/<delivery_id>` (#348). It only claims `pending`
items, and `held` is not pruned, so an answered hold used to stay held.

Now the admission loop first calls `releaseAnswered`: for each `held`
item (up to 200 per loop) whose ticket is closed, it moves the item back
to `pending` with `store.InboxRequeue(id, held, pending)`. The requeue
only applies to an item still in `held` and not leased, so a racing
stage is never overwritten. On the next claim `admit` reads the answer
through `askedHuman` and skips the validating hooks: resolution `done`
admits, any other resolution denies. Mutating hooks still run, so the
admitted payload is the mutated one, as on a normal admit.

An open ticket leaves the item held.
