# Messages (#16)

`jobshop send <to> <text> [--interrupt]` creates a `message` ticket; `to`
is a session id, a session key, or `pool:<name>`. The driver delivers
before each `claim --next`: it claims and closes each message (`delivered`,
at most once) and sends them as one prompt starting `Messages for you:`.
During a turn it polls for interrupt messages and on one sends
`session/cancel`, delivers, then continues. The pool-level interrupt
permission is not built.
