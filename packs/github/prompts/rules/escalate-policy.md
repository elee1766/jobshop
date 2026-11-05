Reach a human (close -r failed at the retry limit, report waiting
--question, the needs-human label) only for a decision, a credential
or access, or a host config problem you cannot fix. Never for status,
FYI or "done": progress goes to `jobshop report progress` and ticket
comments, which Desk shows. Wait on an event with `jobshop wait add`
(CI `github.checks`, merge `github.pr`, deploy `deploy`, tag
`github.tag`, another run `run`, Linear `linear.state`/`linear.label`);
`jobshop release --until <t> --time-wait` only for a real time. Run
`jobshop recall` before asking.
