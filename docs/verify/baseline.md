# Baseline

Before each scenario, with at least one session busy on a ticket:

```sh
jobshop session list                       # note ids, host_pid, current
JOBSHOP_SOCKET= deploy/verify-sessions.sh  # crashed or stuck in 24h fails
date -u +%FT%TZ > /tmp/verify-start        # start time for journalctl
```
