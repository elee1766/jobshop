# Desk user from a trusted proxy header (#515)

2026-11-05. Desk has no login; it runs behind an authenticating proxy and
takes the user from a header that proxy sets.

- `jobshop desk --user-header <name> --trusted-proxy <cidr|addr>...`
  (NixOS: `services.jobshop.desk.userHeader`, `trustedProxies`, with
  `desk.enable` running `jobshop-desk.service`). Without `--user-header`
  nothing changes: writes act as `admin`.
- With it, every request must come from a trusted peer (the TCP peer, not
  `X-Forwarded-For`) with exactly one non-empty header value: printable,
  at most 256 bytes with the `desk:` prefix. Anything else is 403.
- Desk sends the user on each admin call as `_meta.actor = desk:<user>`.
  The controller accepts `_meta.actor` only on the admin socket; a session
  socket gets `forbidden`. The actor is what `created_by`, history,
  comments, reports and cordons record.
- `GET /api/whoami` returns `{user}`; the nav shows it.

Tailscale Serve on the same host:

```nix
services.jobshop.desk = {
  enable = true;
  userHeader = "Tailscale-User-Login";
  trustedProxies = [ "127.0.0.1" "::1" ];
};
# tailscale serve --bg http://127.0.0.1:7480
```

oauth2-proxy (`--reverse-proxy --set-xauthrequest` or
`--pass-user-headers`) in front, on 10.0.0.2:

```
jobshop desk --listen 0.0.0.0:7480 --allow-remote \
  --user-header X-Forwarded-User --trusted-proxy 10.0.0.2
```

Do not expose Desk to other hosts in a trusted range: such a range
should hold only the proxy.
