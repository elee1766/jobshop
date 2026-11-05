# Runbook format

Each file is one live check. A step is a fenced code block.

A block whose info string has the word `readonly` (```` ```sh readonly ````)
only reads state: it does not write files outside /tmp, restart,
deploy, create tickets or runs, or need an operator. `jobshop runbook
run <file>` and the `live-check` workflow (#635) run those blocks with
`sh -e` and record each one's output and pass/fail; a block passes when
it exits 0, so write `grep ... || true` where no match is fine. Every
other block is listed as left for the operator. Mark a block readonly
only when all of it is safe to run unattended on the host.
