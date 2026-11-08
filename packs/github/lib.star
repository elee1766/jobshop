# Helpers shared by the trigger handlers and scan signals.

# listed splits a comma-separated var into lowercase names.
def listed(value):
    return [v.strip().lower() for v in value.split(",") if v.strip()]

# green: the PR has checks and every one passed, was skipped or neutral.
def green(pr):
    if not pr["checks"]:
        return False
    for c in pr["checks"]:
        if c not in ["SUCCESS", "SKIPPED", "NEUTRAL"]:
            return False
    return True

# older_than: an RFC 3339 timestamp is more than a Go duration before the
# invocation's time.
def older_than(stamp, duration):
    return time.now() - time.parse_time(stamp) > time.parse_duration(duration)
