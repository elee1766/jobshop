#!/bin/sh
# Usage: changed.sh [go list flags] <package>...
# Prints run=true to $GITHUB_OUTPUT when CI must run tests for the given
# packages: a push (not a PR), a change to go.mod, go.sum or .github, or a
# changed file of any jobshop package they depend on, test deps included.
set -eu

out=${GITHUB_OUTPUT:-/dev/stdout}
if [ "${GITHUB_EVENT_NAME:-}" != pull_request ]; then
	echo run=true >>"$out"
	exit 0
fi

# HEAD is the PR merged into its base, so this is the PR's diff.
git cat-file -e "$BASE_SHA^{commit}" 2>/dev/null || git fetch -q --depth=1 origin "$BASE_SHA"
changed=$(git diff --name-only "$BASE_SHA" HEAD)

# Package directories (catches deleted files) plus every embedded file,
# which can sit in a subdirectory such as migrations/.
files=$(go list -test -deps -f '{{if not .Standard}}{{$d := .Dir}}{{$d}}
{{range .EmbedFiles}}{{$d}}/{{.}}
{{end}}{{end}}' "$@" |
	grep -F "$PWD/" | sed "s|^$PWD/||" | sort -u)

for f in $changed; do
	case $f in
	go.mod | go.sum | .github/*)
		echo "run: $f" >&2
		echo run=true >>"$out"
		exit 0
		;;
	esac
	if printf '%s\n' "$files" | grep -qxF -e "$f" -e "$(dirname "$f")"; then
		echo "run: $f" >&2
		echo run=true >>"$out"
		exit 0
	fi
done
echo "skip: no changed package in the dependency set" >&2
echo run=false >>"$out"
