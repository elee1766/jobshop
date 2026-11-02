# Users and security

Controller, hosts and agents all run as `jobshop`. Accepted risk: an agent
can do anything the controller can; socket scoping guards against mistakes,
not hostile agents. Later options if needed: agents as a separate user via
a narrow sudoers rule, or per-session bubblewrap or Landlock applied by the
host (read-only system, writable worktree, own CLI socket).
