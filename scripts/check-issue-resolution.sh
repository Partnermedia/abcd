#!/usr/bin/env bash
# Deterministic gate for abcd's issue-resolution convention (AGENTS.md, the
# issue ledger under .abcd/work/issues/).
#
# `abcd capture resolve` exists and works. Nothing anywhere could fail when it
# was not run, so an issue that had in fact been fixed stayed in open/ until a
# human remembered, and a forgotten one left no marker to find it by: 328
# issues in resolved/ against 78 carrying resolved_by when this was measured.
# The resolution convention has no denominator of its own, because a
# fixed-but-unresolved issue is indistinguishable from an open one. Prose is not
# the missing piece; a check that fails closed is (iss-2608241347321757).
#
# The mechanic that forced the forgetting is resolved_by.commit. Every other
# field in a resolution is knowable while the record is being edited; the fixing
# commit's sha is not, because the record and the fix are the same change. That
# defers stamping past merge, and a deferred step is the one that gets dropped.
# So the gate's job is to make the resolution land INSIDE the change that fixes
# the issue, where no later step exists to forget.
#
# Three checks:
#
#   RS001  A `Resolves: iss-N` trailer in the range must be accompanied by that
#          issue leaving open/ for resolved/ or wontfix/ in the same range. A
#          commit that says it resolves an issue and does not move the record is
#          the exact drift this gate exists to stop.
#
#   RS002  A resolved_by.commit sha ADDED in the range must name a commit that
#          exists and is reachable from the head being pushed. The --commit flag
#          is shape-checked only (^[0-9a-f]{7,64}$), so a stamp naming a commit
#          that never existed reads exactly like a good one.
#
#   RS003  Every resolved_by.commit already in the ledger must still be
#          reachable. This is the drift detector, and it is not hypothetical:
#          the repository allows merge, squash AND rebase, the method is a
#          per-pull-request choice, and under squash or rebase a cited branch
#          sha is rewritten out of existence. All 76 stamps were reachable when
#          this landed; RS003 is what notices the day one is not.
#
# Usage:
#   check-issue-resolution.sh commits <base-ref> <head-ref>   # RS001 + RS002
#   check-issue-resolution.sh ledger [<ref>]                  # RS003 (default HEAD)
#
# Exit 0 clean, 1 a violation, 2 a usage/environment fault.
set -euo pipefail

ISSUES_DIR=".abcd/work/issues"
TRAILER_RE='^Resolves:[[:space:]]+(iss-[0-9]+)[[:space:]]*$'

violations=0

fail() {
	printf 'check-issue-resolution: %s\n' "$1" >&2
	violations=$((violations + 1))
}

usage() {
	echo "usage: check-issue-resolution.sh commits <base-ref> <head-ref> | ledger [<ref>]" >&2
	exit 2
}

# ids_leaving_open prints every iss-N whose record left open/ for resolved/ or
# wontfix/ across the range. Rename detection is requested but not relied on: a
# move recorded as a delete plus an add is the same event, so the two halves are
# collected independently and intersected by the caller.
ids_leaving_open() {
	local base="$1" head="$2"
	git diff --name-status --find-renames "$base".."$head" -- "$ISSUES_DIR" |
		while IFS=$'\t' read -r status path dest; do
			case "$status" in
			R*)
				case "$path/$dest" in
				"$ISSUES_DIR/open/"*"$ISSUES_DIR/resolved/"* | "$ISSUES_DIR/open/"*"$ISSUES_DIR/wontfix/"*)
					basename "$dest" | grep -oE '^iss-[0-9]+'
					;;
				esac
				;;
			D)
				case "$path" in
				"$ISSUES_DIR/open/"*) basename "$path" | grep -oE '^iss-[0-9]+' ;;
				esac
				;;
			esac
		done
}

ids_entering_closed() {
	local base="$1" head="$2"
	git diff --name-status --find-renames "$base".."$head" -- "$ISSUES_DIR" |
		while IFS=$'\t' read -r status path dest; do
			case "$status" in
			R*)
				case "$dest" in
				"$ISSUES_DIR/resolved/"* | "$ISSUES_DIR/wontfix/"*) basename "$dest" | grep -oE '^iss-[0-9]+' ;;
				esac
				;;
			A)
				case "$path" in
				"$ISSUES_DIR/resolved/"* | "$ISSUES_DIR/wontfix/"*) basename "$path" | grep -oE '^iss-[0-9]+' ;;
				esac
				;;
			esac
		done
}

# reachable reports whether sha names a real commit that ref can see. A sha that
# does not resolve at all and one that resolves but is unreachable are distinct
# faults, so they are reported separately rather than folded into "bad sha".
reachable() {
	local sha="$1" ref="$2"
	git cat-file -e "${sha}^{commit}" 2>/dev/null || return 2
	git merge-base --is-ancestor "$sha" "$ref" 2>/dev/null || return 1
	return 0
}

check_commits() {
	local base="$1" head="$2"
	local range
	range="$(git rev-list --no-merges "$base".."$head")"
	if [ -z "$range" ]; then
		echo "check-issue-resolution: no non-merge commits in $base..$head — nothing to check"
		return 0
	fi

	local closed
	closed="$(
		{
			ids_leaving_open "$base" "$head"
			ids_entering_closed "$base" "$head"
		} | sort -u
	)"

	# RS001 — a declared resolution must move the record.
	local declared=""
	while IFS= read -r sha; do
		[ -n "$sha" ] || continue
		while IFS= read -r line; do
			[[ "$line" =~ $TRAILER_RE ]] || continue
			local id="${BASH_REMATCH[1]}"
			declared="$declared $id"
			if ! printf '%s\n' "$closed" | grep -qx "$id"; then
				fail "RS001 commit ${sha:0:12} declares 'Resolves: $id', but $id does not leave $ISSUES_DIR/open/ in $base..$head. Resolve it in this change (abcd capture resolve $id ...) or drop the trailer."
			fi
		done <<<"$(git show -s --format='%B' "$sha")"
	done <<<"$range"

	# RS002 — a stamp added here must name a commit this head can see.
	local added
	added="$(git diff -U0 "$base".."$head" -- "$ISSUES_DIR" |
		grep -E '^\+[[:space:]]*commit:' |
		grep -oE '[0-9a-f]{7,64}' | sort -u || true)"
	while IFS= read -r sha; do
		[ -n "$sha" ] || continue
		local rc=0
		reachable "$sha" "$head" || rc=$?
		case "$rc" in
		2) fail "RS002 resolved_by.commit '$sha' is added in this range but names no commit in this repository. --commit is shape-checked only, so a wrong sha is accepted silently." ;;
		1) fail "RS002 resolved_by.commit '$sha' is not reachable from $head. Cite a commit on this branch or already on main." ;;
		esac
	done <<<"$added"

	if [ -n "${declared// /}" ]; then
		echo "check-issue-resolution: RS001 checked$declared"
	fi
}

check_ledger() {
	local ref="${1:-HEAD}"
	local checked=0
	local files
	files="$(git ls-tree -r --name-only "$ref" -- "$ISSUES_DIR" | grep -E '\.md$' || true)"
	[ -n "$files" ] || {
		echo "check-issue-resolution: no ledger records at $ref — nothing to check"
		return 0
	}
	while IFS= read -r f; do
		[ -n "$f" ] || continue
		# Only the frontmatter's resolved_by block carries a stamp; a sha quoted in
		# the prose body is narrative, not provenance, so the scan stops at the
		# closing delimiter.
		local sha
		sha="$(git show "$ref:$f" 2>/dev/null |
			awk 'NR>1 && /^---$/{exit} /^[[:space:]]+commit:/{print}' |
			grep -oE '[0-9a-f]{7,64}' | head -1 || true)"
		[ -n "$sha" ] || continue
		checked=$((checked + 1))
		local rc=0
		reachable "$sha" "$ref" || rc=$?
		case "$rc" in
		2) fail "RS003 $f stamps resolved_by.commit '$sha', which names no commit in this repository." ;;
		1) fail "RS003 $f stamps resolved_by.commit '$sha', which is no longer reachable from $ref — a squash or rebase merge rewrote it." ;;
		esac
	done <<<"$files"
	echo "check-issue-resolution: RS003 checked $checked stamped record(s) at $ref"
}

case "${1:-}" in
commits)
	[ $# -eq 3 ] || usage
	check_commits "$2" "$3"
	;;
ledger)
	[ $# -le 2 ] || usage
	check_ledger "${2:-HEAD}"
	;;
*)
	usage
	;;
esac

if [ "$violations" -gt 0 ]; then
	printf 'check-issue-resolution: FAILED — %d violation(s)\n' "$violations" >&2
	exit 1
fi
echo "check-issue-resolution: OK"
