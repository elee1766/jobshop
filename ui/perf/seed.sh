#!/bin/sh
# Seed an isolated controller for the Desk baseline (JOB-150).
# Usage: sh ui/perf/seed.sh <jobshop binary> <root> <runs> <waiting>
# Creates <runs> two-step runs and <waiting> needs-human tickets.
set -eu
bin=$1
root=$2
runs=$3
waiting=$4
j() { env -u JOBSHOP_SOCKET -u JOBSHOP_SESSION -u JOBSHOP_ROOT "$bin" --root "$root" "$@"; }
i=2
while [ "$i" -le "$runs" ]; do
  j run start work --var issue="$i" --key "seed-$i" >/dev/null
  i=$((i + 1))
done
i=1
while [ "$i" -le "$waiting" ]; do
  j ticket create "Needs a decision $i: pick a rollout order for the migration" -l needs-human -b "Seeded waiting item $i." >/dev/null
  i=$((i + 1))
done
echo "seeded $runs runs, $waiting waiting"
