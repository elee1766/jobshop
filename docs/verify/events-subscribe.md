# events.subscribe (#147)

On the controller host, as the jobshop user:

1. `jobshop events --label needs-human` in one terminal. It prints
   `{"seq":N,"kind":"ready",...}` at once.
2. `jobshop ticket create "events check" --label needs-human`, then
   `jobshop comment -t <id> "x"` and `jobshop close <id>`. Each prints one
   line (`create`, `comment`, `close`) within a second, with rising `seq`.
3. Ctrl-C, then `jobshop events --since <seq of create minus 1>`: the
   same lines replay before `ready`.
4. With a gateway that has `escalation_room`, a new escalation ticket
   gets its thread within a second (no 10s wait), and its log has no
   `gateway: events` warnings.
