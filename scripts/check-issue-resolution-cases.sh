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

# Hermetic git (iss-28, iss-313): every scratch-repo command below is `git -C
# "$d" …`, but an inherited absolute GIT_DIR overrides -C and redirects these
# fixture commits onto the ambient repository — which then reports all-green while
# its real history is rewritten. An inherited GIT_CONFIG_GLOBAL / core.hooksPath
# also fires the developer's global hooks inside the scratch repos and breaks the
# run for a reason that has nothing to do with the ledger. Neutralise the ambient
# git environment before the first git call, matching check-attribution-cases.sh
# and gitutil.IsolatedEnv. (commit.gpgsign is set per-repo below; this closes the
# rest.)
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY \
	GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_CONFIG GIT_CONFIG_COUNT GIT_CONFIG_PARAMETERS
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0

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

# A bare `git rm` of the open record with a Resolves: trailer is NOT a
# resolution: the record leaves open/ but ENTERS neither resolved/ nor wontfix/,
# so it vanishes from the ledger and its changelog line is lost. RS001 keys on
# the id ENTERING a terminal folder, so a delete-only "resolution" is refused.
d="$(newrepo rs001-delete-only)"
git -C "$d" rm -q "$ISS_DIR/open/iss-999-a-fixture.md"
git -C "$d" add -A
git -C "$d" commit -qm "fix: something

Resolves: iss-999"
expect fail "$d" "RS001 trailer with a bare delete of the record" -- commits main HEAD

# A rewrite-move that stays inside open/ is not a resolution either, however
# git reports it: rewritten heavily enough it is a D plus an A, and the D half
# must not read as a terminal landing.
d="$(newrepo rs001-renamed-within-open)"
git -C "$d" mv "$ISS_DIR/open/iss-999-a-fixture.md" "$ISS_DIR/open/iss-999-a-fixture-renamed.md"
for i in 1 2 3 4 5 6 7 8 9 10; do
	echo "Rewritten line $i so git reports the move as a delete plus an add." >>"$d/$ISS_DIR/open/iss-999-a-fixture-renamed.md"
done
git -C "$d" add -A
git -C "$d" commit -qm "fix: something

Resolves: iss-999"
expect fail "$d" "RS001 trailer, record renamed within open/" -- commits main HEAD

# The mirror of the bare delete, and the load-bearing case: a record CAPTURED and
# resolved in the same change lands as a plain add into resolved/ — a two-dot
# base..head diff shows no departure from open/, because the record was never in
# open/ at the base. It MUST pass: this is how the bug-hunt loop resolves its own
# findings. RS001 keys on entering a terminal folder, not on leaving open/, so it
# holds. (A blanket intersection of leaves-open and enters-closed would fail this
# and break every such PR at the push gate.)
d="$(newrepo rs001-fresh-capture)"
cat >"$d/$ISS_DIR/resolved/iss-777-born-resolved.md" <<'EOF'
---
schema_version: 1
id: "iss-777"
---
A finding captured and resolved in one change.
EOF
git -C "$d" add -A
git -C "$d" commit -qm "fix: address a fresh finding

Resolves: iss-777"
expect pass "$d" "RS001 trailer with a same-change capture-and-resolve" -- commits main HEAD

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

# --- boundary/robustness regressions ----------------------------------------

# RS002 reads the frontmatter only: a `commit:` line in a record's PROSE BODY is
# narrative, not a stamp, and must not be reachability-checked. The frontmatter
# here carries a real reachable stamp; the body documents the stamp shape with a
# nonexistent sha. Before the boundary fix the raw-diff scan flagged the body sha.
d="$(newrepo rs002-body-example)"
base="$(git -C "$d" rev-parse HEAD)"
resolve_record "$d" "$base"
cat >>"$d/$ISS_DIR/resolved/iss-999-a-fixture.md" <<'EOF'

Example of the stamp this rule checks:

    resolved_by:
      commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
EOF
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve and document the stamp shape"
expect pass "$d" "RS002 ignores a commit: example in the prose body" -- commits main HEAD

# A non-iss file entering resolved/ or wontfix/ (a README, a policy note) is not a
# record, and it must not abort the gate. The id extractors skip a non-iss
# basename; before the fix that skip exited non-zero and, when it was the last
# diff entry under errexit, killed the whole run silently with RS001/RS002 never
# reached. zzz- sorts last, so it is that final entry.
d="$(newrepo rs001-nonrecord-last)"
resolve_record "$d"
echo "policy notes" >"$d/$ISS_DIR/wontfix/zzz-policy.md"
git -C "$d" add -A
git -C "$d" commit -qm "fix: something

Resolves: iss-999"
expect pass "$d" "a non-iss file last in the diff does not abort the gate" -- commits main HEAD

# The gate resolves its paths from the repository root, so it scans the same
# records from any working directory. Before the cd fix, a run from a subdirectory
# matched nothing and reported a clean pass over an unreachable-stamp ledger.
d="$(newrepo rs003-subdir)"
resolve_record "$d" "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve"
mkdir -p "$d/sub/deep"
expect fail "$d/sub/deep" "RS003 still scans when run from a subdirectory" -- ledger HEAD

# A git probe that fails must exit 2 as an environment fault, never read as an
# empty ledger. Before the rc-check fix, `cd "$(git rev-parse --show-toplevel)"`
# collapsed to a successful `cd ""` and the ls-tree || true turned the failure
# into "no ledger records — nothing to check", exit 0 — a vacuous pass from the
# one gate that notices rewritten resolution stamps. Two shapes: no repository
# at all, and a repository git refuses to read (dubious ownership, the form a
# container/devcontainer uid split produces on a real checkout).
d="$tmproot/not-a-repo"
mkdir -p "$d"
out="$(cd "$d" && bash "$GATE" ledger HEAD 2>&1)" && rc=0 || rc=$?
if [ "$rc" -ne 2 ]; then
	printf 'cases: FAIL a non-repository cwd must exit 2, got exit %d:\n%s\n' "$rc" "$out" >&2
	failures=$((failures + 1))
else
	printf 'cases: ok   a failing git probe is an environment fault, not an empty ledger\n'
fi

d="$(newrepo git-refused)"
resolve_record "$d" "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
git -C "$d" add -A
git -C "$d" commit -qm "chore: resolve"
out="$(cd "$d" && GIT_TEST_ASSUME_DIFFERENT_OWNER=1 bash "$GATE" ledger HEAD 2>&1)" && rc=0 || rc=$?
if [ "$rc" -ne 2 ]; then
	printf 'cases: FAIL a git-refused repository must exit 2, got exit %d:\n%s\n' "$rc" "$out" >&2
	failures=$((failures + 1))
else
	printf 'cases: ok   a git-refused repository (dubious ownership) is an environment fault\n'
fi

# The legitimate empty case survives the hardening: git succeeds, zero records —
# a loud "nothing to check" pass, not a fault.
d="$tmproot/empty-ledger"
mkdir -p "$d"
git -C "$d" init -q -b main
git -C "$d" config user.name t
git -C "$d" config user.email t@example.invalid
git -C "$d" config commit.gpgsign false
echo x >"$d/README.md"
git -C "$d" add -A
git -C "$d" commit -qm "baseline"
expect pass "$d" "an actually-empty ledger still passes loudly" -- ledger HEAD

# --- the sibling record families are outside the gate's scope ----------------
#
# The ledger root now holds readings/<run-id>/ and dispositions/<item-id>/, whose
# files are NOT issue records. RS002 and RS003 read the frontmatter of every .md
# under the pathspec and treat an indented `commit:` key as a resolution stamp, so
# an unscoped gate would reachability-check a value in one of these and refuse a
# clean tree. The gate scopes to the status directories; these two fixtures are
# what prove it, in the only direction that can regress — a record the gate must
# ignore.
#
# The frontmatter below deliberately carries that indented `commit:` shape. The
# record schemas do not mint one today, and that is exactly why the fixture spells
# it out: the gate reads BYTES under a pathspec, not schemas, so what protects
# these families is the scope and nothing else. A fixture that relied on today's
# field list would pass unscoped and prove nothing.

d="$(newrepo reading-record-ignored)"
mkdir -p "$d/$ISS_DIR/readings/rdg-1"
cat >"$d/$ISS_DIR/readings/rdg-1/rdi-2.md" <<'EOF'
---
schema_version: 1
id: "rdi-2"
run: "rdg-1"
manifest: "sha256:beef"
position: "detection"
regime: "registrative"
pattern: "a stated constraint"
tension: "the two sides disagree"
constraint_in_play: "the stated invariant"
why_a_tension: "one of them must give"
cited_by:
  commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
---
EOF
git -C "$d" add -A
git -C "$d" commit -qm "chore: ingest a reading run"
expect pass "$d" "a reading record is outside the gate's scope" -- ledger HEAD
expect pass "$d" "a reading record is outside the commits scan too" -- commits main HEAD

d="$(newrepo disposition-ignored)"
mkdir -p "$d/$ISS_DIR/dispositions/rdi-2"
cat >"$d/$ISS_DIR/dispositions/rdi-2/dsp-3.md" <<'EOF'
---
schema_version: 1
id: "dsp-3"
item: "rdi-2"
state: "accepted"
disposition_grounds: "worth acting on"
cited_by:
  commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
---
EOF
git -C "$d" add -A
git -C "$d" commit -qm "chore: answer a reading item"
expect pass "$d" "a disposition is outside the gate's scope" -- ledger HEAD
expect pass "$d" "a disposition is outside the commits scan too" -- commits main HEAD

d="$(newrepo admission-ignored)"
mkdir -p "$d/$ISS_DIR/admissions/rdg-1" "$d/$ISS_DIR/surprises"
cat >"$d/$ISS_DIR/admissions/rdg-1/adm-4.md" <<'EOF'
---
schema_version: 1
id: "adm-4"
run: "rdg-1"
proposal: "rdi-2"
grounds: "the configuration it admits is one the frame does not hold"
cited_by:
  commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
---
EOF
cat >"$d/$ISS_DIR/surprises/srp-5.md" <<'EOF'
---
schema_version: 1
id: "srp-5"
occasioned_by: "rdi-2"
cited_by:
  commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
---
EOF
git -C "$d" add -A
git -C "$d" commit -qm "chore: admit a proposal and record a surprise"
expect pass "$d" "the step-2 records are outside the gate's scope" -- ledger HEAD
expect pass "$d" "the step-2 records are outside the commits scan too" -- commits main HEAD

if [ "$failures" -gt 0 ]; then
	printf 'cases: FAILED — %d case(s) did not behave\n' "$failures" >&2
	exit 1
fi
echo "cases: OK — every rule refused its violating fixture and passed its clean one"
