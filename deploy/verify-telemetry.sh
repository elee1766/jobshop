#!/bin/sh
# Telemetry check for a deployed Jobshop (docs/verify/telemetry.md).
# Exits non-zero when the scrape target is down, a required series is
# missing, or VictoriaTraces has no jobshop session trace in the window.
set -u

metrics_url=${JOBSHOP_METRICS_URL:-http://127.0.0.1:9464/metrics}
vm_url=${VM_URL:-http://127.0.0.1:18428}
vt_url=${VT_URL:-http://127.0.0.1:20428}
job=${VM_JOB:-jobshop}
service=${VT_SERVICE:-jobshop}
window=${WINDOW:-3600}
curl=${CURL:-curl}

usage() {
	cat <<EOF
Usage: verify-telemetry.sh [--metrics URL] [--vm URL] [--vt URL]
                           [--job NAME] [--service NAME] [--window SECONDS]

Checks, in order:
  1. the controller /metrics endpoint serves jobshop_* series
  2. the VictoriaMetrics scrape target with job=NAME is up
  3. jobshop_sessions and jobshop_tickets_ready exist for every configured
     pool (pools are read from jobshop_tickets_ready)
     (counters such as restarts and turns are listed, not required:
     they appear after their first increment)
  4. VictoriaTraces has a "session" trace for service NAME in the
     last SECONDS

Options (environment variable in brackets):
  --metrics URL     controller /metrics [JOBSHOP_METRICS_URL]
                    default $metrics_url
  --vm URL          VictoriaMetrics base URL [VM_URL], default $vm_url
  --vt URL          VictoriaTraces base URL [VT_URL], default $vt_url
  --job NAME        scrape job name [VM_JOB], default $job
  --service NAME    trace service name [VT_SERVICE], default $service
  --window SECONDS  trace look-back [WINDOW], default $window
  -h, --help        show this help

CURL sets the curl binary. Needs curl and jq. Exit 0 when every check
passes, 1 when one fails, 2 on a usage error.
EOF
}

while [ $# -gt 0 ]; do
	case $1 in
	-h | --help)
		usage
		exit 0
		;;
	--metrics | --vm | --vt | --job | --service | --window)
		if [ $# -lt 2 ]; then
			echo "verify-telemetry: $1 needs a value" >&2
			exit 2
		fi
		case $1 in
		--metrics) metrics_url=$2 ;;
		--vm) vm_url=$2 ;;
		--vt) vt_url=$2 ;;
		--job) job=$2 ;;
		--service) service=$2 ;;
		--window) window=$2 ;;
		esac
		shift 2
		;;
	*)
		echo "verify-telemetry: unknown argument: $1" >&2
		usage >&2
		exit 2
		;;
	esac
done

case $window in
'' | *[!0-9]*)
	echo "verify-telemetry: --window must be whole seconds, got: $window" >&2
	exit 2
	;;
esac
for tool in "$curl" jq; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		echo "verify-telemetry: $tool not found" >&2
		exit 2
	fi
done

failed=0
pass() { echo "ok    $*"; }
fail() {
	echo "FAIL  $*"
	failed=1
}
get() { "$curl" -fsS -m 10 "$@"; }

# PromQL query against VictoriaMetrics; prints the JSON result list.
vm_query() {
	get -G "$vm_url/api/v1/query" --data-urlencode "query=$1" | jq -c '.data.result'
}

# 1. Controller endpoint.
if body=$(get "$metrics_url"); then
	n=$(printf '%s\n' "$body" | grep -c '^jobshop_')
	if [ "$n" -gt 0 ]; then
		pass "controller $metrics_url: $n jobshop_* samples"
	else
		fail "controller $metrics_url: no jobshop_* samples"
	fi
else
	fail "controller $metrics_url: unreachable"
fi

# 2. Scrape target.
if targets=$(get "$vm_url/api/v1/targets"); then
	health=$(printf '%s' "$targets" |
		jq -r --arg job "$job" '[.data.activeTargets[] | select(.labels.job == $job) | .health] | join(",")')
	case $health in
	up) pass "scrape target job=$job: up" ;;
	'') fail "scrape target job=$job: not configured" ;;
	*) fail "scrape target job=$job: $health" ;;
	esac
else
	fail "VictoriaMetrics $vm_url: unreachable"
fi

# 3. Series per pool. Pools come from jobshop_tickets_ready, which the
# supervisor records every tick for configured pools only. jobshop_sessions
# also carries pools of leftover open sessions after a pool is removed from
# config, so it would make a removed pool look broken (#80). last_over_time keeps a
# controller restart (one missed scrape) from looking like a missing series.
if pools=$(vm_query 'count by (jobshop_pool) (last_over_time(jobshop_tickets_ready[5m]))' | jq -r '.[].metric.jobshop_pool' | sort); then
	if [ -z "$pools" ]; then
		fail "series: jobshop_tickets_ready has no pools"
	fi
	for pool in $pools; do
		for name in jobshop_sessions jobshop_tickets_ready; do
			n=$(vm_query "count(last_over_time($name{jobshop_pool=\"$pool\"}[5m]))" | jq -r '.[0].value[1] // "0"')
			if [ "$n" != 0 ]; then
				pass "series $name pool=$pool"
			else
				fail "series $name pool=$pool: missing"
			fi
		done
	done
	for name in jobshop_claims_total jobshop_closes_total jobshop_turns_total \
		jobshop_session_restarts_total jobshop_lease_expiries_total jobshop_sessions_ended_total; do
		have=$(vm_query "count by (jobshop_pool) ($name)" |
			jq -r 'map(.metric.jobshop_pool // "all") | join(",")')
		echo "info  $name: ${have:-not yet incremented}"
	done
	# Failed session ends in the window (#71); the runbook wants none.
	bad=$(vm_query "sum by (jobshop_pool, jobshop_resolution) (increase(jobshop_sessions_ended_total{jobshop_resolution=~\"crashed|stuck\"}[${window}s])) > 0" |
		jq -r 'map(.metric.jobshop_pool + "/" + .metric.jobshop_resolution) | join(",")')
	echo "info  crashed or stuck sessions in ${window}s: ${bad:-none}"
else
	fail "VictoriaMetrics $vm_url: query failed"
fi

# 4. Traces. The Jaeger API takes microseconds.
now=$(date +%s)
start=$(((now - window) * 1000000))
end=$((now * 1000000))
if traces=$(get -G "$vt_url/select/jaeger/api/traces" \
	--data-urlencode "service=$service" --data-urlencode operation=session \
	--data-urlencode "start=$start" --data-urlencode "end=$end" --data-urlencode limit=20); then
	n=$(printf '%s' "$traces" | jq '.data | length')
	if [ "$n" -gt 0 ]; then
		pass "traces: $n session trace(s) for $service in the last ${window}s"
	else
		fail "traces: no session trace for $service in the last ${window}s"
	fi
else
	fail "VictoriaTraces $vt_url: unreachable"
fi

if [ "$failed" -ne 0 ]; then
	echo "telemetry check FAILED"
	exit 1
fi
echo "telemetry check passed"
