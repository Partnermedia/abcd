#!/usr/bin/env bash
# iss-232: the `ahoy install` staleness refusal keys on the binary's BuildInfo
# vcs.revision (internal/core/vintage, current-from-BuildInfo). Release
# binaries are stamped today only by default (-buildvcs=auto from a git
# checkout); a build with -buildvcs=false, or from a .git-less source
# archive, ships binaries whose vintage is undeterminable — Known flips
# false for EVERY pinned user and `ahoy install` starts refusing until
# --allow-stale-binary. This guard makes that coupling impossible to break
# silently: every binary about to be checksummed and published must carry a
# vcs.revision, or the release aborts before anything is hashed or uploaded.
set -euo pipefail
shopt -s nullglob

bins=(bin/abcd-*)
if [ ${#bins[@]} -eq 0 ]; then
  echo "check-vcs-stamp: no binaries under bin/ — run \`make build\` first" >&2
  exit 1
fi

fail=0
for b in "${bins[@]}"; do
  info="$(go version -m "$b")"
  if ! printf '%s\n' "$info" | grep -q 'vcs\.revision='; then
    echo "check-vcs-stamp: FAIL $b carries no vcs.revision — pinned users' \`ahoy install\` would refuse (iss-232)" >&2
    fail=1
  elif printf '%s\n' "$info" | grep -q 'vcs\.modified=true'; then
    echo "check-vcs-stamp: FAIL $b was built from a dirty tree (vcs.modified=true) — its vintage is undeterminable and pinned users' \`ahoy install\` would refuse (iss-232)" >&2
    fail=1
  else
    echo "check-vcs-stamp: OK   $b"
  fi
done
exit "$fail"
