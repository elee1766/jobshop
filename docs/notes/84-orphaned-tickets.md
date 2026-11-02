# Orphaned tickets (#84)

An orphaned ticket is open, unclaimed (no holder), claimable (`task`), and
has an empty pool or one not in the loaded config. Every supervisor tick
and every successful reload the controller queries them, logs WARN
`ticket has no claimable pool` with `reason=no_pool|unknown_pool` the first
time it sees a ticket with that reason, and sets the gauge
`jobshop_tickets_orphaned{reason}` (dashboard panel "Orphaned tickets").
A ticket that gets a configured pool, a holder, or is closed drops out;
if it is orphaned again later it warns again. `jobshop ticket list
--orphaned [--project p] [-n N]` lists them through the controller.
