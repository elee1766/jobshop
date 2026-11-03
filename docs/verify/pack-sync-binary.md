# Pack sync follows the unit's binary (#223)

After a deploy that includes #223, on gasserver:

- `systemctl stop jobshop`, then `systemctl start
  jobshop-workspace-default-sync`: the journal shows no `/proc/.../exe`
  error and the checkout HEAD equals the commit from `$(systemctl cat
  jobshop | grep ExecStart)` version.
- `systemctl start jobshop` succeeds (no `var ... is not set`).
