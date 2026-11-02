#!/bin/sh
# Crashed-session check for a deployed Jobshop (docs/verify/baseline.md).
# Exits non-zero when any session closed crashed or stuck in
# the window. Stuck counts as a failure (#71).
set -u

since=${SINCE:-24h}
jobshop=${JOBSHOP:-jobshop}

usage() {
	cat <<EOF
Usage: verify-sessions.sh [--since DURATION]

Lists sessions closed with resolution crashed or stuck within DURATION
and fails if there are any, printing id, pool and end reason.

Options (environment variable in brackets):
  --since DURATION  look-back window [SINCE], default $since
  -h, --help        show this help

JOBSHOP sets the jobshop binary. Needs jq. Run with JOBSHOP_SOCKET= (empty)
from a session so the CLI uses the admin socket. Exit 0 when none are
found, 1 when some are or a query fails, 2 on a usage error.
EOF
}

while [ $# -gt 0 ]; do
	case $1 in
	--since)
		[ $# -ge 2 ] || { usage >&2; exit 2; }
		since=$2
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		usage >&2
		exit 2
		;;
	esac
done

found=0
for state in crashed stuck; do
	if ! out=$("$jobshop" session list --state "$state" --since "$since"); then
		echo "FAIL: jobshop session list --state $state failed" >&2
		exit 1
	fi
	if ! lines=$(printf '%s\n' "$out" | jq -r --arg s "$state" \
		'.[] | [.id, (.ticket.pool // "-"), $s, (.exit // "-")] | @tsv'); then
		echo "FAIL: could not parse session list output" >&2
		exit 1
	fi
	if [ -n "$lines" ]; then
		printf '%s\n' "$lines"
		found=1
	fi
done

if [ "$found" -ne 0 ]; then
	echo "FAIL: sessions ended crashed or stuck in the last $since (id pool resolution reason)" >&2
	exit 1
fi
echo "ok: no crashed or stuck sessions in the last $since"
