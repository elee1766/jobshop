# Users and security

Controller, hosts and agents all run as `jobshop`. Accepted risk: an agent
can do anything the controller can; socket scoping guards against mistakes,
not hostile agents. Per-session bubblewrap is built (#262, see
`262-sandbox.md`). A later option if needed: agents as a separate user via
a narrow sudoers rule.
