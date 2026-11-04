# #284 Desk pool scale and session cordon

2026-11-04. Desk's Pools view shows the #281 scale columns per ticket
pool (desired, open, working, idle, cordoned, min/max, target or auto)
with Scale (prompt for a target) and Auto buttons. Sessions shows a
`cordoned: <reason>` badge and Cordon / Uncordon buttons on live
sessions.

The buttons POST to `/api/pool/scale` (pool.scale params) and
`/api/session/cordon` (session.cordon params), behind the same Host,
Origin and `X-Jobshop-Desk` checks as chat writes. Desk rejects a body
with neither or both of target/auto, a negative target, or drain plus
uncordon before calling the controller.
