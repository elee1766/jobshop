#!/bin/sh
# Closed-loop check for one planner cycle (docs/gasserver-verify.md
# section 7). Reads `jobshop ticket trace <plan-ticket>` and exits non-zero
# naming the first missing link.
# shellcheck disable=SC2016 # $h, $c and $t are jq variables, not shell
set -u

gh=${GH:-gh}
jobshop=${JOBSHOP:-jobshop}
trace_file=
id=

usage() {
	cat <<EOF
Usage: verify-closed-loop.sh [--trace FILE] <plan-ticket-id>
       verify-closed-loop.sh --trace FILE

Checks, in order, on the trace of a plan ticket:
  1. the plan ticket was created by a trigger
  2. the plan ticket is closed
  3. a ticket its planner session created while holding it (dev pointer)
  4. one of those tickets reached the refinery pool
  5. that ticket references a PR that GitHub reports as merged

Options:
  --trace FILE  read trace JSON from FILE (- for stdin) instead of
                running "\$JOBSHOP ticket trace <id>"
  -h, --help    show this help

JOBSHOP sets the jobshop binary, GH the gh binary. Needs jq. Exit 0 when
every check passes, 1 when one fails, 2 on a usage error.
EOF
}

while [ $# -gt 0 ]; do
	case $1 in
	-h | --help)
		usage
		exit 0
		;;
	--trace)
		if [ $# -lt 2 ]; then
			echo "verify-closed-loop: --trace needs a value" >&2
			exit 2
		fi
		trace_file=$2
		shift 2
		;;
	-*)
		echo "verify-closed-loop: unknown argument: $1" >&2
		usage >&2
		exit 2
		;;
	*)
		if [ -n "$id" ]; then
			echo "verify-closed-loop: one ticket id only" >&2
			exit 2
		fi
		id=$1
		shift
		;;
	esac
done

if [ -z "$id" ] && [ -z "$trace_file" ]; then
	usage >&2
	exit 2
fi
if ! command -v jq >/dev/null 2>&1; then
	echo "verify-closed-loop: jq not found" >&2
	exit 2
fi

if [ -n "$trace_file" ]; then
	if ! trace=$(cat -- "$trace_file"); then
		echo "verify-closed-loop: cannot read $trace_file" >&2
		exit 2
	fi
elif ! trace=$("$jobshop" ticket trace "$id"); then
	echo "verify-closed-loop: jobshop ticket trace $id failed" >&2
	exit 1
fi
if ! printf '%s' "$trace" | jq -e 'type == "object" and has("id")' >/dev/null 2>&1; then
	echo "verify-closed-loop: input is not a ticket trace" >&2
	exit 2
fi

q() { printf '%s' "$trace" | jq -r "$@"; }
pass() { echo "ok    $*"; }
fail() {
	echo "FAIL  $*"
	echo "closed-loop check FAILED"
	exit 1
}

plan=$(q .id)
echo "plan ticket $plan: $(q '.title // ""')"

trigger=$(q '.trigger // ""')
[ -n "$trigger" ] || fail "plan ticket $plan was not created by a trigger (no meta.trigger.name)"
pass "created by trigger $trigger"

status=$(q .status)
[ "$status" = closed ] || fail "plan ticket $plan is $status, not closed"
pass "plan ticket closed ($(q '.resolution // "done"'))"

# Dev pointers: children the plan's holder sessions created. Children with
# only a parent_id link do not count; the loop is planner -> dev.
devs=$(q '(.holders // []) as $h | [.children // [] | .[] | select(.created_by as $c | $h | index($c))] | map(.id) | join(" ")')
[ -n "$devs" ] || fail "no dev ticket created by the planner session ($(q '(.holders // []) | join(",")'))"
pass "dev tickets from the planner: $devs"

refinery=$(q '(.holders // []) as $h | [.children // [] | .[] | select(.created_by as $c | $h | index($c)) | select(.pool == "refinery")] | map(.id) | join(" ")')
[ -n "$refinery" ] || fail "no dev ticket reached the refinery pool (checked: $devs)"
pass "reached refinery: $refinery"

merged=
for t in $refinery; do
	prs=$(q --arg t "$t" '.children[] | select(.id == $t) | (.links // [])[] | select(test("/pull/[0-9]+$"))')
	for pr in $prs; do
		state=$("$gh" pr view "$pr" --json state -q .state 2>/dev/null) || state=unknown
		echo "info  $t $pr: $state"
		if [ "$state" = MERGED ]; then
			merged="$t $pr"
			break 2
		fi
	done
done
[ -n "$merged" ] || fail "no merged PR referenced by a refinery ticket ($refinery)"
pass "merged PR: $merged"

echo "closed-loop check passed"
