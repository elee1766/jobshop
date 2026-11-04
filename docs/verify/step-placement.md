# Verify: step placement (#409)

On gasserver, with `resources: {flagship: {max: 1}, workhorse: {max: 4}}`
and a test workflow with vars `[issue, risk]` whose step has
`placement:` with a flagship option (weight 100) and a workhorse option
(weight 10, `when: "${risk} != 'high'"`), `prefer_for: 2m`:

1. `jobshop config reload` succeeds.
2. Start run A (`--var risk=low`) and claim its step, so flagship is
   full. Start run B (`risk=low`) and run C (`risk=high`).
3. `jobshop run show <B>`: `placement option 1/2 pool <flagship pool>`.
   B is not in `jobshop ready`.
4. After 2 minutes, `run show <B>` reads `option 2/2 ... (fallback)` with
   a "Placement: moved" comment, and a workhorse session picks it up.
   `run show <C>` still shows option 1/1: high risk never falls back.
5. Close A's step: C is claimed on the flagship pool.
