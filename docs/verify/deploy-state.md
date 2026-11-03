# Self-deploy state matches the unit (#225)

On gasserver, after a deploy with #225:

```sh
cat /var/lib/jobshop-deploy/deployed
"$(readlink -f /proc/$(systemctl show -p MainPID --value jobshop.service)/exe)" version
journalctl -u jobshop-self-deploy --since -2d | grep 'deployed rewritten'
```

The commit in `deployed` is the prefix of the one `version` prints in
parentheses. A `deployed rewritten` WARN, if any, shows when it was rewritten.
