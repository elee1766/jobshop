# Yaml-only switch reloads the controller (#210)

On gasserver, change only a `livePackVars` value (or another setting
that ends up in jobshop.yaml) and switch. `systemctl show jobshop -p
ExecMainPID` is unchanged (no restart), the journal has the controller's
reload line at the switch time, and `jobshop config show` has the new
value without a manual `jobshop reload`.
