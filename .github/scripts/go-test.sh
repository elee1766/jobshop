#!/bin/sh
# The CI go test step (JOB-187). Every package but pkg/orchestration/control
# runs in one go test. Control's test binary runs as several processes, each
# a slice of its tests by name: one process takes 60-76s on the talos-metal
# runner whatever -test.cpu is (2 to 32), because its tests contend on
# modernc's process-wide allocator lock and SQLite, so more cores do not
# help one process but more processes do. The slices together are every
# test: the last one skips what the others run. Writes the combined
# go test -json stream to $1 and exits non-zero if any part failed.
set -u
out=${1:-test.json}
ctl=github.com/gfx-labs/jobshop/pkg/orchestration/control
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
root=$(pwd)

gotest() {
	go test -race -tags victorialogs \
		-gcflags='modernc.org/...=-d=checkptr=0' \
		-gcflags='modernc.org/sqlite/lib=-race=false' "$@"
}

echo "nproc=$(nproc) load=$(cut -d' ' -f1-3 /proc/loadavg)"
start=$(date +%s)

pkgs=$(go list -tags victorialogs ./... | grep -vx "$ctl")
# shellcheck disable=SC2086 # one argument per package
(gotest -json $pkgs >"$tmp/rest.json" 2>&1; echo $? >"$tmp/rest.rc") &

# One build and link of the control binary, then the slices run it.
# GOMAXPROCS 6 was the fastest single process in the runner sweep.
(
	if ! gotest -c -o "$tmp/control.test" "$ctl" >"$tmp/build.txt" 2>&1; then
		echo 1 >"$tmp/build.rc"
		exit
	fi
	cd "$root/pkg/orchestration/control" || exit
	i=0
	for sel in '-test.run=^Test[A-D]' '-test.run=^Test[E-L]' '-test.run=^Test[M-Q]' '-test.run=^Test[RS]' '-test.skip=^Test[A-S]'; do
		i=$((i + 1))
		(
			s=$(date +%s)
			GOMAXPROCS=6 go tool test2json -t -p "$ctl" "$tmp/control.test" \
				-test.v=test2json -test.timeout=10m "$sel" >"$tmp/control$i.json" 2>&1
			rc=$?
			echo "$rc" >"$tmp/control$i.rc"
			echo "control slice $i ($sel): $(($(date +%s) - s))s rc=$rc" >>"$tmp/slices.txt"
		) &
	done
	wait
) &
wait

cat "$tmp/rest.json" "$tmp/build.txt" "$tmp"/control*.json >"$out"
sort "$tmp/slices.txt" 2>/dev/null
echo "go test step: $(($(date +%s) - start))s"
rc=0
for f in "$tmp"/*.rc; do
	[ "$(cat "$f")" = 0 ] || rc=1
done
exit "$rc"
