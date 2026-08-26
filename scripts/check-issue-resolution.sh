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
#          issue ENTERING resolved/ or wontfix/ in the same range — whether it
#          moved out of open/ or was captured and resolved in the same change. A
#          commit that says it resolves an issue and lands the record in no
#          terminal folder is the exact drift this gate exists to stop.
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

# Resolve every path from the repository root, like the sibling gate
# check-reviews.sh:11. ISSUES_DIR and the git pathspecs below are relative, and a
# git pathspec is matched against the current directory — so run from a
# subdirectory the diff and ls-tree match nothing and the gate reports a clean
# pass having scanned zero records. cd first, so cwd cannot disarm it.
cd "$(git rev-parse --show-toplevel)"

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

# ids_entering_closed prints every iss-N whose record ENTERS resolved/ or wontfix/
# across the range — the destination half of a resolution. A record moved from
# open/ shows as a rename (or, without rename detection, as an add into the
# terminal folder); a record captured and resolved in the same change shows as a
# plain add into the terminal folder. Both are honest resolutions and both are
# caught here. A record that only LEAVES open/ (a bare delete) enters nothing and
# is deliberately absent, so RS001 refuses a trailer that merely deletes.
ids_entering_closed() {
	local base="$1" head="$2"
	git diff --name-status --find-renames "$base".."$head" -- "$ISSUES_DIR" |
		while IFS=$'\t' read -r status path dest; do
			case "$status" in
			R*)
				case "$dest" in
				"$ISSUES_DIR/resolved/"* | "$ISSUES_DIR/wontfix/"*) basename "$dest" | grep -oE '^iss-[0-9]+' || true ;;
				esac
				;;
			A)
				case "$path" in
				"$ISSUES_DIR/resolved/"* | "$ISSUES_DIR/wontfix/"*) basename "$path" | grep -oE '^iss-[0-9]+' || true ;;
				esac
				;;
			esac
		done
}

# frontmatter_commit prints a record's resolved_by.commit sha at ref, reading the
# frontmatter ONLY: a sha quoted in the prose body is narrative, not provenance,
# so the scan stops at the closing delimiter. RS002 and RS003 share this so the
# two rules cannot drift apart on where a stamp lives.
frontmatter_commit() {
	local ref="$1" path="$2"
	git show "$ref:$path" 2>/dev/null |
		awk 'NR>1 && /^---$/{exit} /^[[:space:]]+commit:/{print}' |
		grep -oE '[0-9a-f]{7,64}' | head -1 || true
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

	# A declared resolution must land the record in a terminal folder: the id must
	# ENTER resolved/ or wontfix/ in this range. That is the honest test of the
	# trailer — the changelog derives from records reaching a terminal folder — and
	# it holds both for a record moved out of open/ AND for one captured and
	# resolved in the same change (a two-dot diff shows the latter only as an add
	# into resolved/, never as a departure from open/). A bare `git rm` of the open
	# record enters nothing, so it does NOT satisfy the trailer: the record would
	# otherwise vanish from the ledger, its changelog line lost, with no other gate
	# to catch it. (A record that enters resolved/ while a copy stays in open/ is a
	# duplicate id, which record-lint's issue_id_unique refuses.)
	local closed
	closed="$(ids_entering_closed "$base" "$head" | sort -u)"

	# RS001 — a declared resolution must move the record.
	local declared=""
	while IFS= read -r sha; do
		[ -n "$sha" ] || continue
		while IFS= read -r line; do
			[[ "$line" =~ $TRAILER_RE ]] || continue
			local id="${BASH_REMATCH[1]}"
			declared="$declared $id"
			if ! printf '%s\n' "$closed" | grep -qx "$id"; then
				fail "RS001 commit ${sha:0:12} declares 'Resolves: $id', but $id does not enter $ISSUES_DIR/resolved/ or $ISSUES_DIR/wontfix/ in $base..$head. Resolve it in this change (abcd capture resolve $id ...) or drop the trailer."
			fi
		done <<<"$(git show -s --format='%B' "$sha")"
	done <<<"$range"

	# RS002 — a stamp added here must name a commit this head can see. Read the
	# frontmatter of each record the range touched and check only a resolved_by.commit
	# this range introduced or changed. Scanning the raw diff for `+  commit:` instead
	# would reachability-check a `commit:` example in a record's prose body — a false
	# violation — which is exactly the boundary RS003 already draws.
	local changed
	changed="$(git diff --name-only "$base".."$head" -- "$ISSUES_DIR" | grep -E '\.md$' || true)"
	while IFS= read -r f; do
		[ -n "$f" ] || continue
		local head_sha base_sha
		head_sha="$(frontmatter_commit "$head" "$f")"
		[ -n "$head_sha" ] || continue
		base_sha="$(frontmatter_commit "$base" "$f")"
		[ "$head_sha" = "$base_sha" ] && continue
		local rc=0
		reachable "$head_sha" "$head" || rc=$?
		case "$rc" in
		2) fail "RS002 resolved_by.commit '$head_sha' is added in this range but names no commit in this repository. --commit is shape-checked only, so a wrong sha is accepted silently." ;;
		1) fail "RS002 resolved_by.commit '$head_sha' is not reachable from $head. Cite a commit on this branch or already on main." ;;
		esac
	done <<<"$changed"

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
		# closing delimiter (frontmatter_commit).
		local sha
		sha="$(frontmatter_commit "$ref" "$f")"
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

# A shallow checkout cannot run these checks honestly: reachable() cannot tell
# an absent commit from an unfetched one, so RS002/RS003 would refuse every
# stamp whose commit lies past the graft — 85 false violations on a clean tree,
# each carrying a diagnosis about a repository fault that did not happen. The
# spec that ruled git resolution out of --commit (spc-25) names shallow states
# in-envelope, so the environment fault must be reported as itself: exit 2, the
# code the contract reserves for it, never a violation.
if [ "$(git rev-parse --is-shallow-repository 2>/dev/null)" = "true" ]; then
	echo "check-issue-resolution: shallow checkout — RS002/RS003 cannot tell an absent commit from an unfetched one; run 'git fetch --unshallow' first (CI checks out with fetch-depth: 0)." >&2
	exit 2
fi

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
