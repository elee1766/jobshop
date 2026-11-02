# Forced rate-limit backoff

The controller treats a prompt error containing `429`, `rate_limit`,
`rate limit` or `is unavailable` as a rate limit. To force one without
burning the real account, give a scratch pool an agent whose provider URL
points at a stub that always answers 429:

```sh
python3 -c '
import http.server
class H(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        self.send_response(429); self.send_header("content-type","application/json"); self.end_headers()
        self.wfile.write(b"{\"type\":\"error\",\"error\":{\"type\":\"rate_limit_error\",\"message\":\"429\"}}")
http.server.HTTPServer(("127.0.0.1", 8429), H).serve_forever()' &
```

In `jobshop.yaml`, add a pool `verify-rl` using the normal agent with
`env: {ANTHROPIC_BASE_URL: "http://127.0.0.1:8429"}` (use the variable your
agent's provider reads), then `sudo systemctl reload jobshop` and:

```sh
jobshop ticket create "rate limit probe" --pool verify-rl
journalctl -u jobshop -f -o cat | grep -E 'provider rate limited|session ended'
jobshop session list                       # backoff_until is set
```

Pass:
- `provider rate limited, backing off wait=1m0s`, then `2m0s`, `4m0s` ...
  capped at `15m0s`.
- `session list` shows `backoff_until` for the session; its ticket stays
  claimed; no second session starts for `verify-rl`.
- Kill the stub. On the next retry the prompt goes through or fails as a
  normal error, `backoff_until` clears.
- No `resolution=crashed` or `resolution=stuck`.

Clean up: close the probe ticket, remove the pool, reload.
