#!/bin/sh
# Download release assets of a GitHub repo: fetch-release.sh <owner/repo> <tag> <asset>...
# Files land in the current directory.
#
# github.com/<repo>/releases/download returned 504 on every try for hours
# on 2026-11-08 and failed master CI (#1040). The API asset endpoint was up,
# so try it first (with the job's token), then the download host with
# retries. Callers check a sha256 from the release's checksums file.
set -eu
repo=$1 tag=$2
shift 2
auth="Authorization: Bearer ${GH_TOKEN:?GH_TOKEN is not set}"

api_get() {
	id=$(curl -fsS --retry 3 --retry-all-errors -H "$auth" "https://api.github.com/repos/$repo/releases/tags/$tag" |
		jq -r --arg n "$1" '.assets[] | select(.name == $n) | .id') &&
		[ -n "$id" ] &&
		curl -fsSL --retry 3 --retry-all-errors -H "$auth" -H "Accept: application/octet-stream" -o "$1" "https://api.github.com/repos/$repo/releases/assets/$id"
}

for f in "$@"; do
	api_get "$f" || curl -fsSLO --retry 20 --retry-all-errors --retry-delay 10 --retry-max-time 600 "https://github.com/$repo/releases/download/$tag/$f"
done
