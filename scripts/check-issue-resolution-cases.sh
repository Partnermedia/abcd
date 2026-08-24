#!/usr/bin/env bash
# Proves check-issue-resolution.sh can actually FAIL.
#
# A gate is an enforcement claim, and an enforcement claim is a fact about the
# code or it is nothing. The failure mode this guards against is the one
# iss-2608230847432286 records as its sharpest shape: a real gate, defended by a
# test that cannot fail, so the hole it leaves is invisible precisely because
# something green is watching it. Every rule here is asserted in BOTH directions
# — a violating fixture must be refused, and a clean one must pass — because a
# check that only ever sees clean input proves nothing about refusal.
#
# Fixtures are built in a scratch repository, never against this one: the rules
# are about ledger moves and commit reachability, and both are cheap to stage
# and impossible to stage honestly in a tree someone is working in.
#
# Usage: check-issue-resolution-cases.sh
# Exit 0 all cases behaved, 1 a case did not.
set -euo pipefail

GATE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-issue-resolution.sh"
[ -x "$GATE" ] || {
	echo "cases: gate not executable: $GATE" >&2
	exit 2
}

failures=0
tmproot="$(mktemp -d)"
trap 'rm -rf "$tmproot"' EXIT

ISS_DIR=".abcd/work/issues"

# expect <want: pass|fail> <repo> <label> -- <gate args...>
#
# The repo is a parameter rather than something the caller cds into, because a
# caller wrapping this in a subshell would discard the failure counter and leave
# THIS script unable to fail — the very defect it exists to catch. The cd lives
# inside the command substitution, where it cannot outlive the capture and the
# exit status still reaches the parent shell.
expect() {
	local want="$1" repo="$2" label="$3"
	shift 4
	local out rc=0
	out="$(cd "$repo" && bash "$GATE" "$@" 2>&1)" || rc=$?
	case "$want" in
	pass)
		if [ "$rc" -ne 0 ]; then
			printf 'cases: FAIL %s — expected clean, got exit %d:\n%s\n' "$label" "$rc" "$out" >&2
			failures=$((failures + 1))
		else
			printf 'cases: ok   %s (clean, as expected)\n' "$label"
		fi
		;;
	fail)
		if [ "$rc" -eq 0 ]; then
			printf 'cases: FAIL %s — the gate PASSED a violating fixture. This rule cannot fail.\n%s\n' "$label" "$out" >&2
			failures=$((failures + 1))
		else
			printf 'cases: ok   %s (refused, as expected)\n' "$label"
		fi
		;;
	esac
}

# newrepo makes a scratch repo with one baseline commit and an open issue.
newrepo() {
	local d="$tmproot/$1"
	mkdir -p "$d/$ISS_DIR/open" "$d/$ISS_DIR/resolved" "$d/$ISS_DIR/wontfix"
	git -C "$d" init -q -b main
	git -C "$d" config user.name t
	git -C "$d" config user.email t@example.invalid
	git -C "$d" config commit.gpgsign false
	cat >"$d/$ISS_DIR/open/iss-999-a-fixture.md" <<'EOF'
---
schema_version: 1
id: "iss-999"
---
A fixture issue.
EOF
	# .gitkeep so the empty destination directories survive the baseline commit.
	touch "$d/$ISS_DIR/resolved/.gitkeep" "$d/$ISS_DIR/wontfix/.gitkeep"
	git -C "$d" add -A
	git -C "$d" commit -qm "baseline"
	# The change under test goes on a branch: with everything on main, main..HEAD
	# is empty and the gate correctly reports "nothing to check" — which a naive
	# fixture reads as a pass.
	git -C "$d" checkout -q -b work
	echo "$d"
}

resolve_record() {
	local d="$1" sha="${2:-}"
	git -C "$d" mv "$ISS_DIR/open/iss-999-a-fixture.md" "$ISS_DIR/resolved/iss-999-a-fixture.md"
	if [ -n "$sha" ]; then
		python3 - "$d/$ISS_DIR/resolved/iss-999-a-fixture.md" "$sha" <<'PY'
import sys
p, sha = sys.argv[1], sys.argv[2]
s = open(p).read()
s = s.replace('id: "iss-999"\n', 'id: "iss-999"\nresolved_by:\n  commit: "%s"\n' % sha)
open(p, "w").write(s)
PY
	fi
}

# --- RS001: a declared resolution must move the record -----------------------

d="$(newrepo rs001-bad)"
echo "touched" >>"$d/README.md"
git -C "$d" add -A
git -C "$d" commit -qm "fix: something

Resolves: iss-999"
# The gate reads the CWD's repo, so run it there.
expect fail "$d" "RS001 trailer, record left in open/" -- commits main HEAD

d="$(newrepo rs001-good)"
resolve_record "$d"
git -C "$d" add -A
git -C "$d" commit -qm "fix: something

Resolves: iss-999"
expect pass "$d" "RS001 trailer with the ledger move" -- commits main HEAD

# A resolution with no trailer is legitimate — a stale issue resolved on its own
# merits has no fixing commit to name — so the gate must NOT demand the reverse.
d="$(newrepo rs001-noclaim)"
resolve_record "$d"
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve a stale issue"
expect pass "$d" "ledger move with no trailer is allowed" -- commits main HEAD

# --- RS002: a stamp added here must name a reachable commit ------------------

d="$(newrepo rs002-bad)"
resolve_record "$d" "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve with a sha that does not exist"
expect fail "$d" "RS002 stamp naming a nonexistent commit" -- commits main HEAD

d="$(newrepo rs002-good)"
base="$(git -C "$d" rev-parse HEAD)"
resolve_record "$d" "$base"
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve citing a real ancestor"
expect pass "$d" "RS002 stamp naming a reachable commit" -- commits main HEAD

# The unreachable-but-real case: a sha that exists in the repo on an unrelated
# branch the head cannot see. This is the squash/rebase shape, staged honestly.
d="$(newrepo rs002-unreachable)"
git -C "$d" checkout -q -b sidebranch main
echo side >"$d/side.txt"
git -C "$d" add -A
git -C "$d" commit -qm "side commit"
side="$(git -C "$d" rev-parse HEAD)"
git -C "$d" checkout -q work
resolve_record "$d" "$side"
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve citing an unreachable commit"
expect fail "$d" "RS002 stamp naming a real but unreachable commit" -- commits main HEAD

# --- RS003: the ledger's existing stamps stay reachable ----------------------

d="$(newrepo rs003-good)"
base="$(git -C "$d" rev-parse HEAD)"
resolve_record "$d" "$base"
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve"
expect pass "$d" "RS003 all stamps reachable" -- ledger HEAD

d="$(newrepo rs003-bad)"
resolve_record "$d" "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve"
expect fail "$d" "RS003 stamp naming a nonexistent commit" -- ledger HEAD

if [ "$failures" -gt 0 ]; then
	printf 'cases: FAILED — %d case(s) did not behave\n' "$failures" >&2
	exit 1
fi
echo "cases: OK — every rule refused its violating fixture and passed its clean one"
