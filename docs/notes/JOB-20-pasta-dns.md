# Pasta sandbox DNS (JOB-20)

2026-11-05.

With `sandbox.network: pasta` the agent has its own network namespace.
On a host whose /etc/resolv.conf names a loopback stub (systemd-resolved's
127.0.0.53), every lookup in the sandbox fails. Jobshop passes three
generic settings through to acphost (v0.8.5, acphost#124). It does not
detect resolvers or pick defaults, and the NixOS module has no DNS
option: set them in jobshop.yaml or the module's `settings`.

- `sandbox.dns: {forward, host}`: pasta's `--dns-forward` and
  `--dns-host`. Both are IP addresses. Needs network pasta. The forward
  address must not fall in a `deny` CIDR (acphost rejects that).
- `sandbox.resolv_conf: <path>`: a host file bound read-only over
  /etc/resolv.conf in the sandbox. `~/` is the controller's HOME.
- `sandbox.dns_check: <name>`: before a pasta pool's first spawn, and
  again after its DNS settings change, the controller resolves this name
  in a pasta namespace set up like the agent's (acphost
  `CheckSandboxDNS`). A failure is a permanent spawn error, so the pool is
  quarantined with "no DNS in sandbox: ..." naming `sandbox.dns`, instead
  of every agent turn failing. Unset: no check.

All three merge like `network`: the pool's beats the agent's beats the
default's. `dns` and `dns_check` are dropped when the merged network is
not pasta (`resolv_conf` still binds). They feed the sandbox env hash,
so a reload respawns idle sessions.

## Example: a systemd-resolved host

Host config, not a default:

```
# /etc/jobshop/resolv.conf
nameserver 169.254.1.1
search tailnet.example
```

```yaml
defaults:
  sandbox:
    network: pasta
    deny: [10.0.0.0/8, 100.64.0.0/10, 192.168.0.0/16]
    dns: {forward: 169.254.1.1, host: 127.0.0.53}
    resolv_conf: /etc/jobshop/resolv.conf
    dns_check: github.com
```

169.254.1.1 is outside those deny ranges. The host side (127.0.0.53, and
MagicDNS behind it) is not filtered, since pasta makes that connection
outside the namespace.
