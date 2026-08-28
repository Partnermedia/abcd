#!/usr/bin/env bash
# Proves check-decisions-append.sh can actually FAIL, and fails for the right
# reasons.
#
# DA001-DA003 are claims about what a commit did to an append-only ledger, and a
# claim is a fact about the diff or it is nothing. The failure mode guarded
# against is the one this repository keeps meeting (iss-2608230847432286): a real
# gate, green because it sees nothing. An append-only rule is especially easy to
# write green, and the first draft of this gate proved it — a position-only rule
# passed a whole-file rewrite, a tail-reaching reword, a truncate-then-restore
# pair, a forged merge resolution and a malformed bullet planted in the header,
# and it REFUSED the honest commit that seeds an empty ledger. Every one of those
# is a case below.
#
# So every rule is asserted in BOTH directions: a violating fixture must be
# refused, and each legitimate shape — a pure append, a BACK-DATED append at the
# tail, a header edit, a genuine multi-region union merge, the seeding of an
# empty ledger — must pass. The report's own bytes are asserted too: the ledger
# and the commit subject are attacker-controlled, so a fixture plants terminal
# escapes in both and the output is checked for raw control bytes.
#
# Fixtures are built in a scratch repository, never against this one: the rules
# are about committed diffs, which are cheap to stage in a throwaway repo and
# impossible to stage honestly in a tree someone is working in.
#
# Usage: check-decisions-append-cases.sh
# Exit 0 all cases behaved, 1 a case did not.
set -euo pipefail

# Hermetic git (iss-28, iss-313): every scratch-repo command below is `git -C
# "$d" …`, but an inherited absolute GIT_DIR overrides -C and redirects these
# fixture commits onto the ambient repository — which then reports all-green while
# its real history is rewritten. An inherited GIT_CONFIG_GLOBAL / core.hooksPath
# also fires the developer's global hooks inside the scratch repos and breaks the
# run for a reason that has nothing to do with the ledger. Neutralise the ambient
# git environment before the first git call, matching check-reviews-cases.sh,
# check-issue-resolution-cases.sh and gitutil.IsolatedEnv. (commit.gpgsign is set
# per-repo below; this closes the rest.)
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY \
	GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_CONFIG GIT_CONFIG_COUNT GIT_CONFIG_PARAMETERS
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0

GATE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-decisions-append.sh"
[ -x "$GATE" ] || {
	echo "cases: gate not executable: $GATE" >&2
	exit 2
}

failures=0
tmproot="$(mktemp -d)"
trap 'rm -rf "$tmproot"' EXIT

LEDGER=".abcd/work/DECISIONS.md"

# expect <want: pass|fail|fault> <repo> <label> -- <gate args...>
#
# fault is exit 2 specifically — the refusal polarity reserved for "the gate
# could not answer", which must never collapse into either a pass or a rule
# violation.
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
			return
		fi
		;;
	fail)
		if [ "$rc" -ne 1 ]; then
			printf 'cases: FAIL %s — expected a rule refusal (exit 1), got exit %d:\n%s\n' "$label" "$rc" "$out" >&2
			failures=$((failures + 1))
			return
		fi
		;;
	fault)
		if [ "$rc" -ne 2 ]; then
			printf 'cases: FAIL %s — expected the environment-fault refusal (exit 2), got exit %d:\n%s\n' "$label" "$rc" "$out" >&2
			failures=$((failures + 1))
			return
		fi
		;;
	esac
	printf 'cases: ok   %s (exit %d, as expected)\n' "$label" "$rc"
}

# expect_rule asserts which RULE refused, not merely that something did. A gate
# whose rules have collapsed into one another still refuses every violating
# fixture, so the exit code alone cannot tell DA002 from DA001.
expect_rule() {
	local rule="$1" repo="$2" label="$3"
	shift 4
	local out rc=0
	out="$(cd "$repo" && bash "$GATE" "$@" 2>&1)" || rc=$?
	if [ "$rc" -ne 1 ]; then
		printf 'cases: FAIL %s — expected a %s refusal (exit 1), got exit %d:\n%s\n' "$label" "$rule" "$rc" "$out" >&2
		failures=$((failures + 1))
		return
	fi
	if ! printf '%s' "$out" | grep -q "$rule"; then
		printf 'cases: FAIL %s — refused, but not by %s:\n%s\n' "$label" "$rule" "$out" >&2
		failures=$((failures + 1))
		return
	fi
	printf 'cases: ok   %s (%s, as expected)\n' "$label" "$rule"
}

# newrepo makes a scratch repo whose baseline commit holds a miniature ledger of
# the real shape: a header paragraph, then dated bullets, the last of which spans
# a continuation line. The change under test goes on a branch — with everything
# on main, main..HEAD is empty and the gate correctly reports "nothing to check",
# which a naive fixture reads as a pass.
newrepo() {
	local d="$tmproot/$1"
	mkdir -p "$d/$(dirname "$LEDGER")"
	git -C "$d" init -q -b main
	git -C "$d" config user.name t
	git -C "$d" config user.email t@example.invalid
	git -C "$d" config commit.gpgsign false
	printf '%s\n' "$LEDGER merge=union" >"$d/.gitattributes"
	cat >"$d/$LEDGER" <<'EOF'
# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed.

- 2026-01-01 — The first decision.
- 2026-01-02 — The second decision.
- 2026-01-03 — The third decision, which runs to
  a continuation line of its own.
EOF
	git -C "$d" add -A
	git -C "$d" commit -qm "baseline: the ledger"
	# `base` pins the baseline commit, so a fixture whose final commit lands on
	# main (a pull-request merge does) still has a stable ref to range from.
	git -C "$d" branch -q base
	git -C "$d" checkout -q -b work
	echo "$d"
}

# --- DA001: position ---------------------------------------------------------

# A mid-file insertion is the shape the position rule exists to refuse: an entry
# written above existing ones silently reorders a log other records cite by
# position.
d="$(newrepo insert-middle)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines.insert(6, "- 2026-01-04 — Inserted above an existing entry.")
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "insert an entry mid-file"
expect_rule DA001 "$d" "an entry inserted mid-file is refused" -- commits main work

# The top of the entry region — directly below the header, above the first
# bullet — is still an insertion. It is the boundary the header exemption sits
# against, so it is asserted rather than assumed.
d="$(newrepo insert-top)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines.insert(4, "- 2026-01-04 — Inserted at the very top of the log.")
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "insert an entry above the first bullet"
expect_rule DA001 "$d" "an entry inserted above the first bullet is refused" -- commits main work

# The header exemption keyed on the CANONICAL `- YYYY-MM-DD` bullet made
# malformity the bypass: a list item one space out of shape reads as header prose
# and plants an entry above the whole log. Each near-miss is asserted, because a
# shape test that covers only the ones someone thought of is the same hole again.
for variant in \
	'-  2026-01-04 — Two spaces after the dash.' \
	'* 2026-01-04 — An asterisk bullet.' \
	'+ 2026-01-04 — A plus bullet.' \
	'- **2026-01-04** — A bolded date.' \
	'1. 2026-01-04 — An ordered item.' \
	'2026-01-04 — A bare date-led line.'; do
	d="$(newrepo "header-forge-$(printf '%s' "$variant" | cksum | cut -d' ' -f1)")"
	python3 - "$d/$LEDGER" "$variant" <<'PY'
import sys
p, variant = sys.argv[1], sys.argv[2]
lines = open(p).read().split("\n")
lines.insert(4, variant)
open(p, "w").write("\n".join(lines))
PY
	git -C "$d" commit -qam "plant a non-canonical entry in the header"
	expect_rule DA001 "$d" "list-shaped '${variant:0:16}…' in the header is refused" -- commits main work
done

# The rule is per-commit, so a clean append followed by an insertion must still
# be refused — a range check that only compared endpoints would see one net
# addition at the tail and pass.
d="$(newrepo append-then-insert)"
printf -- '- 2026-01-04 — An honest append.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "append an entry"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines.insert(5, "- 2026-01-05 — And then one inserted above it.")
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "insert an entry mid-file"
expect_rule DA001 "$d" "an insertion later in the range is refused" -- commits main work

# The last entry's interior is deliberately NOT exempt: a continuation appended
# after the file's last line already extends that entry, so writing above the
# last line buys nothing and is indistinguishable from any other insertion.
d="$(newrepo insert-last-entry-interior)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines.insert(7, "  an amendment slipped inside the last entry,")
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "amend the last entry from the inside"
expect_rule DA001 "$d" "a line inserted inside the last entry is refused" -- commits main work

# --- DA002: preservation -----------------------------------------------------
#
# Position alone only fires while pre-existing content SURVIVES below the
# addition. Every fixture here rewrites or erases committed decisions with
# nothing beneath them to trip it, and every one of them passed the
# position-only draft.

# A reword of a historical entry mid-file. Position happens to catch this one too
# (content survives below it), which is why the assertion names DA002
# specifically — the two rules must not be collapsing into each other.
d="$(newrepo rewrite-history)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
s = open(p).read().replace("The second decision.", "The second decision, silently reworded.")
open(p, "w").write(s)
PY
git -C "$d" commit -qam "reword a historical entry"
expect_rule DA002 "$d" "rewording a historical entry is refused" -- commits main work

# The last bullet AND its continuation line reworded together: the hunk runs to
# end-of-file, so nothing survives below it and the position rule sees no
# insertion at all.
d="$(newrepo reword-through-eof)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines[6] = "- 2026-01-03 — The third decision, quietly restated"
lines[7] = "  with a continuation that says something else."
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "restate the last entry through EOF"
expect_rule DA002 "$d" "a reword of the last entry through EOF is refused" -- commits main work

# Amending the file's LAST LINE in place — the single edit position could never
# see. The ledger's own header says to correct an entry by appending a new one,
# and this is the rule that makes that true.
d="$(newrepo amend-last-line)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines[7] = "  a continuation line, quietly changed."
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "amend the last line"
expect_rule DA002 "$d" "amending the file's last line is refused" -- commits main work

# The whole entry region replaced in one hunk: every dated decision swapped for
# new ones, same line count, nothing surviving below.
d="$(newrepo replace-entry-region)"
cat >"$d/$LEDGER" <<'EOF'
# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed.

- 2026-01-01 — A fabricated first decision.
- 2026-01-02 — A fabricated second decision.
- 2026-01-03 — A fabricated third decision, which runs to
  a fabricated continuation line.
EOF
git -C "$d" commit -qam "replace every entry in the log"
expect_rule DA002 "$d" "replacing the whole entry region is refused" -- commits main work

# A whole-file rewrite, header and all.
d="$(newrepo whole-file-rewrite)"
cat >"$d/$LEDGER" <<'EOF'
# LEDGER

Totally different prose.

- 2026-01-09 — A fabricated decision never taken.
- 2026-01-01 — The first decision, silently reworded.
EOF
git -C "$d" commit -qam "rewrite the ledger wholesale"
expect_rule DA002 "$d" "a whole-file rewrite is refused" -- commits main work

# Truncate the log in one commit, restore a forged history in the next. Each half
# is positionally innocent — the first adds nothing, the second appends to an
# empty log — so only preservation can refuse the pair.
d="$(newrepo truncate-restore)"
cat >"$d/$LEDGER" <<'EOF'
# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed.
EOF
git -C "$d" commit -qam "housekeeping: trim the ledger"
cat >>"$d/$LEDGER" <<'EOF'

- 2026-01-01 — The first decision, rewritten.
- 2026-01-03 — Reordered.
- 2026-01-02 — Reordered.
EOF
git -C "$d" commit -qam "restore the ledger"
expect_rule DA002 "$d" "truncate-then-restore across two commits is refused" -- commits main work

# A pure deletion adds nothing at all, so the position rule is blind to it by
# construction.
d="$(newrepo pure-delete)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = [l for l in open(p).read().split("\n") if "The second decision" not in l]
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "drop an inconvenient decision"
expect_rule DA002 "$d" "deleting a committed decision is refused" -- commits main work

# --- merges: DA002 per parent, DA003 against all of them ---------------------

# union_merge_repo stages the routine shape this repository actually produces:
# a branch appends, merges main in mid-flight, appends again, and is then merged
# BACK into main by a pull request. Each side's entries end up above the other's
# in different regions, because the union driver resolves each conflicting region
# on its own — so the final merge's ledger is a tail extension of NEITHER parent.
# HEAD ends on main, which is why newrepo pins `base`.
union_merge_repo() {
	local d
	d="$(newrepo "$1")"
	git -C "$d" checkout -q main
	printf -- '- 2026-02-01 — Main, first batch.\n- 2026-02-02 — Main, still the first batch.\n' >>"$d/$LEDGER"
	git -C "$d" commit -qam "main appends"
	git -C "$d" checkout -q work
	printf -- '- 2026-03-01 — Branch, first batch.\n' >>"$d/$LEDGER"
	git -C "$d" commit -qam "branch appends"
	git -C "$d" merge -q --no-ff -m "merge main into the branch" main >/dev/null 2>&1
	git -C "$d" checkout -q main
	printf -- '- 2026-02-03 — Main, second batch.\n' >>"$d/$LEDGER"
	git -C "$d" commit -qam "main appends again"
	git -C "$d" checkout -q work
	printf -- '- 2026-03-02 — Branch, second batch.\n' >>"$d/$LEDGER"
	git -C "$d" commit -qam "branch appends again"
	git -C "$d" checkout -q main
	git -C "$d" merge -q --no-ff -m "merge the branch into main" work >/dev/null 2>&1
	echo "$d"
}

# amend_merge rewrites the merge commit's tree in place, keeping both parents,
# and REFUSES to continue if the result is not still a merge. Without that
# assertion a failed amend would leave an ordinary commit on top, the fixture
# would quietly stop exercising the merge path, and the case would pass for the
# wrong reason — the precise defect this script exists to prevent.
amend_merge() {
	local d="$1"
	git -C "$d" commit -q --amend --no-edit -a
	local parents
	parents="$(git -C "$d" rev-list --parents -n 1 HEAD | wc -w)"
	if [ "$parents" -lt 3 ]; then
		printf 'cases: FAIL fixture setup — HEAD of %s is not a merge after the amend (%s parent field(s)); the merge cases would test nothing.\n' "$d" "$parents" >&2
		failures=$((failures + 1))
		return 1
	fi
}

# assert_interleaved proves the fixture is the shape it claims to be: for BOTH
# parents, the merge introduced content ABOVE that parent's last line, so the
# result is a tail extension of neither. Without this the clean-merge case could
# degrade into an ordinary fast-forward-shaped merge and would then pass under a
# position rule too — pinning nothing.
assert_interleaved() {
	local d="$1" p first_hunk parent_lines
	for p in $(git -C "$d" rev-list --parents -n 1 HEAD | cut -d' ' -f2-); do
		parent_lines="$(git -C "$d" show "$p:$LEDGER" | awk 'END{print NR}')"
		first_hunk="$(git -C "$d" diff --unified=0 --no-renames "$p" HEAD -- "$LEDGER" |
			awk '/^@@ /{h=$2; sub(/^-/,"",h); sub(/,.*/,"",h); print h+0; exit}')"
		if [ -z "$first_hunk" ] || [ "$first_hunk" -ge "$parent_lines" ]; then
			printf 'cases: FAIL fixture setup — the merge in %s is a tail extension of parent %s (first hunk at %s of %s lines); it does not pin the interleave.\n' \
				"$d" "${p:0:12}" "${first_hunk:-none}" "$parent_lines" >&2
			failures=$((failures + 1))
			return 1
		fi
	done
}

# The clean polarity, and the one that constrains every merge rule: a genuine
# multi-region union merge must pass. This fixture is why DA001 is not applied to
# merges in any form — the union driver concatenates each conflicting region
# independently, so both sides end up with lines above the other's and the result
# extends neither parent's tail. A "must extend one parent" merge rule refused
# eight real merges in this repository's last eight hundred commits.
d="$(union_merge_repo union-merge-clean)"
assert_interleaved "$d"
expect pass "$d" "a multi-region union merge passes" -- commits base HEAD

# A merge that DROPS a parent's decision in its resolution. Skipping merges — the
# first draft's choice — left this write path entirely unchecked.
d="$(union_merge_repo union-merge-drop)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = [l for l in open(p).read().split("\n") if "Main, first batch" not in l]
open(p, "w").write("\n".join(lines))
PY
amend_merge "$d"
expect_rule DA002 "$d" "a merge that drops a parent's decision is refused" -- commits base HEAD

# A merge that INVENTS a decision without deleting anything. Nothing about its
# position is anomalous — a union merge interleaves anyway — so only "the merge
# authored a line no parent has" can refuse it.
d="$(union_merge_repo union-merge-invent)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines.insert(4, "- 2026-01-09 — FORGED in the merge, present in no parent.")
open(p, "w").write("\n".join(lines))
PY
amend_merge "$d"
expect_rule DA003 "$d" "a merge that invents a decision is refused" -- commits base HEAD

# A merge that DUPLICATES a decision only ONE parent contributed. Every line is
# present in a parent and nothing is removed, so plain set membership — DA003's
# first form — passed it. Only counting can see it.
d="$(union_merge_repo union-merge-duplicate)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
victim = next(l for l in lines if "Main, second batch" in l)
lines.insert(4, victim)
open(p, "w").write("\n".join(lines))
PY
amend_merge "$d"
expect_rule DA003 "$d" "a merge that duplicates a one-sided decision is refused" -- commits base HEAD

# A merge that duplicates a line of COMMON history — one both parents inherited
# unchanged from the base. This is the case summing the parents could never
# refuse: a shared line counts once per parent, so the sum bound licensed exactly
# two copies of it. Against the base-relative bound it is 1 + 0 + 0 = 1, and the
# second copy is refused.
d="$(union_merge_repo union-merge-duplicate-shared)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
victim = next(l for l in lines if "The first decision" in l)   # from the baseline
lines.insert(4, victim)
open(p, "w").write("\n".join(lines))
PY
amend_merge "$d"
expect_rule DA003 "$d" "a merge that duplicates a line of common history is refused" -- commits base HEAD

# The full exploit the double-budget permitted: a complete reordered rendition of
# the ENTIRE log spliced above the real one. Nothing deleted, nothing invented,
# every line present in a parent and — under the sum bound — every line within
# budget, because each side was credited separately for history they shared. On
# the real 2,155-line ledger this added 2,150 lines and the gate stayed green.
d="$(union_merge_repo union-merge-splice-rendition)"
python3 - "$d/$LEDGER" <<'PY'
import sys, re
p = sys.argv[1]
lines = open(p).read().split("\n")
trailing = lines and lines[-1] == ""
if trailing:
    lines = lines[:-1]
first = next(i for i, l in enumerate(lines) if re.match(r"^- \d\d\d\d-\d\d-\d\d", l))
header, body = lines[:first], lines[first:]
# Only lines the two sides SHARE — the batches each parent added on its own are
# left out, so no line exceeds the sum bound and nothing is invented.
side = ("Main, first batch", "Main, still the first batch", "Main, second batch",
        "Branch, first batch", "Branch, second batch")
shared = [l for l in body if not any(s in l for s in side)]
out = header + list(reversed(shared)) + body
open(p, "w").write("\n".join(out) + ("\n" if trailing else ""))
PY
amend_merge "$d"
expect_rule DA003 "$d" "a merge splicing a reordered rendition of the log is refused" -- commits base HEAD

# --- the base term is not the merge author's to choose -----------------------
#
# DA003's bound rests entirely on the merge base, so whoever picks the base picks
# the rule. Anchoring on `git merge-base --octopus` — a common ancestor of ALL
# parents — handed that choice to the author: ONE extra parent older than the
# ledger drags the ancestor below the point the ledger existed, count_base
# collapses to 0, the bound reverts to the plain sum, and the splice above is
# green again with no message of any kind. Max over the PAIRWISE bases of the
# ledger-carrying parents is monotone the other way: an extra parent only adds
# pairs, and a maximum over a superset can only rise, so a bolted-on parent
# tightens the rule against its author instead of loosening it.

# splice_octopus builds the forged tree and commits it with the given extra
# parent appended, so the only variable between these two cases is which commit
# the author bolted on.
splice_octopus() {
	local d="$1" extra="$2" work main tree m
	work="$(git -C "$d" rev-parse work)"
	main="$(git -C "$d" rev-parse main)"
	git -C "$d" checkout -q work
	git -C "$d" checkout -q "$main" -- "$LEDGER"
	printf -- '- 2026-03-01 — Branch appends.\n' >>"$d/$LEDGER"
	python3 - "$d/$LEDGER" <<'PY'
import sys, re
p = sys.argv[1]
lines = open(p).read().split("\n")
trailing = lines and lines[-1] == ""
if trailing:
    lines = lines[:-1]
first = next(i for i, l in enumerate(lines) if re.match(r"^- \d\d\d\d-\d\d-\d\d", l))
header, body = lines[:first], lines[first:]
side = ("Branch appends.", "Main appends.")
shared = [l for l in body if not any(s in l for s in side)]
open(p, "w").write("\n".join(header + list(reversed(shared)) + body) + ("\n" if trailing else ""))
PY
	git -C "$d" add -A
	tree="$(git -C "$d" write-tree)"
	m="$(git -C "$d" commit-tree "$tree" -p "$work" -p "$main" -p "$extra" -m "merge main into work, and re-attach an old parent")"
	git -C "$d" update-ref refs/heads/work "$m"
	git -C "$d" checkout -q -f work
	local n
	n="$(git -C "$d" rev-list --parents -n 1 work | wc -w)"
	if [ "$n" -lt 4 ]; then
		printf 'cases: FAIL fixture setup — %s is not a 3-parent octopus (%s parent field(s)).\n' "$d" "$n" >&2
		failures=$((failures + 1))
		return 1
	fi
}

# preledger_repo: a root with NO ledger, then the ledger seeded on top, then the
# usual two-sided appends. `base` pins the seeding commit; `root` pins the
# pre-ledger commit an author can bolt on.
preledger_repo() {
	local d="$tmproot/$1"
	mkdir -p "$d/$(dirname "$LEDGER")"
	git -C "$d" init -q -b main
	git -C "$d" config user.name t
	git -C "$d" config user.email t@example.invalid
	git -C "$d" config commit.gpgsign false
	printf '# repo\n' >"$d/README.md"
	git -C "$d" add -A
	git -C "$d" commit -qm "initial commit"
	git -C "$d" branch -q root
	printf '%s merge=union\n' "$LEDGER" >"$d/.gitattributes"
	cat >"$d/$LEDGER" <<'EOF'
# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed.

- 2026-01-01 — The first decision.
- 2026-01-02 — The second decision.
- 2026-01-03 — The third decision, which runs to
  a continuation line of its own.
EOF
	git -C "$d" add -A
	git -C "$d" commit -qm "seed the ledger"
	git -C "$d" branch -q base
	git -C "$d" checkout -q -b work
	# The branch's append is the SAME text splice_octopus re-adds, so the forged
	# tree invents nothing: every line in it is present in some parent, and only
	# the base term can decide the verdict. An earlier draft of this fixture
	# appended different text, which made the forgery contain an invented line —
	# refused by every version of the rule, and so proving nothing about the base.
	printf -- '- 2026-03-01 — Branch appends.\n' >>"$d/$LEDGER"
	git -C "$d" commit -qam "branch appends"
	git -C "$d" checkout -q main
	printf -- '- 2026-02-01 — Main appends.\n' >>"$d/$LEDGER"
	git -C "$d" commit -qam "main appends"
	echo "$d"
}

# The repository's own pre-ledger commit is enough. It is an ordinary ancestor of
# this same history, so nothing about the merge looks irregular.
d="$(preledger_repo octopus-preledger-parent)"
splice_octopus "$d" "$(git -C "$d" rev-parse root)"
expect_rule DA003 "$d" "a splice under a bolted-on pre-ledger parent is refused" -- commits base work

# An unrelated orphan root does it differently and just as well: the octopus base
# then fails to resolve AT ALL, which the old code read as "no base" and treated
# as licence rather than as a fault.
d="$(preledger_repo octopus-orphan-parent)"
git -C "$d" checkout -q --orphan orphan
git -C "$d" rm -rq --cached . >/dev/null 2>&1 || true
rm -f "$d/$LEDGER" "$d/.gitattributes"
printf 'notes\n' >"$d/notes.txt"
git -C "$d" add -A
git -C "$d" commit -qm "notes: an unrelated import"
git -C "$d" checkout -q -f work
splice_octopus "$d" "$(git -C "$d" rev-parse orphan)"
expect_rule DA003 "$d" "a splice under a bolted-on orphan parent is refused" -- commits base work

# The refusing direction is not the only one that matters: an HONEST octopus —
# every parent sharing the ledger's history, each side appending — must still
# pass, or the fix would have replaced a bypass with a wall.
#
# Built with commit-tree rather than `git merge`, because git's octopus STRATEGY
# does not consult the union merge driver at all: it reports "content conflict …
# Should not be doing an octopus" and refuses. The tree here is exactly what a
# union resolution yields — the common history once, then each side's append —
# so the fixture is honest by construction rather than by git's say-so.
d="$(preledger_repo octopus-honest)"
git -C "$d" checkout -q -b third base
printf -- '- 2026-04-01 — A third line of work appends.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "third branch appends"
git -C "$d" checkout -q work
git -C "$d" show base:"$LEDGER" >"$d/$LEDGER"
printf -- '- 2026-03-01 — Branch appends.\n- 2026-02-01 — Main appends.\n- 2026-04-01 — A third line of work appends.\n' >>"$d/$LEDGER"
git -C "$d" add -A
honest_tree="$(git -C "$d" write-tree)"
honest_merge="$(git -C "$d" commit-tree "$honest_tree" \
	-p "$(git -C "$d" rev-parse work)" \
	-p "$(git -C "$d" rev-parse main)" \
	-p "$(git -C "$d" rev-parse third)" \
	-m "merge main and third into work")"
git -C "$d" update-ref refs/heads/work "$honest_merge"
git -C "$d" checkout -q -f work
expect pass "$d" "an honest three-parent union merge passes" -- commits base work

# LOUD DEGRADATION. Two parents carry the ledger and no pairwise base carries it,
# so the bound cannot be anchored. The old code silently fell back to the weaker
# sum bound — and a silent fallback IS the exploit, since the attack never has to
# beat the bound, only to reach the branch where the bound stops applying. Exit 2,
# the same environment-fault polarity every git probe in this gate already uses.
d="$tmproot/unanchorable-base"
mkdir -p "$d/$(dirname "$LEDGER")"
git -C "$d" init -q -b main
git -C "$d" config user.name t
git -C "$d" config user.email t@example.invalid
git -C "$d" config commit.gpgsign false
printf '%s merge=union\n' "$LEDGER" >"$d/.gitattributes"
printf '# DECISIONS\n\nAppend-only, newest last.\n\n- 2026-01-01 — Main seeded it.\n' >"$d/$LEDGER"
git -C "$d" add -A
git -C "$d" commit -qm "main seeds the ledger"
git -C "$d" branch -q base
# A second, entirely unrelated history that ALSO carries a ledger.
git -C "$d" checkout -q --orphan other
git -C "$d" rm -rq --cached . >/dev/null 2>&1 || true
printf '%s merge=union\n' "$LEDGER" >"$d/.gitattributes"
printf '# DECISIONS\n\nAppend-only, newest last.\n\n- 2026-01-02 — The other side seeded its own.\n' >"$d/$LEDGER"
git -C "$d" add -A
git -C "$d" commit -qm "an unrelated history with its own ledger"
git -C "$d" checkout -q main
git -C "$d" merge -q --no-ff --allow-unrelated-histories -m "merge two unrelated ledgers" other >/dev/null 2>&1 || true
git -C "$d" add -A >/dev/null 2>&1 || true
git -C "$d" commit -qm "merge two unrelated ledgers" >/dev/null 2>&1 || true
expect fault "$d" "an unanchorable merge base refuses loudly instead of weakening" -- commits base main

# --- a parent listed twice is one parent -------------------------------------
#
# DA003 credits each parent with what it ADDED over the base, so a parent counted
# twice contributes its additions twice. That reopens plant-a-line-above-the-log
# for any line that side legitimately appended: two copies, nothing deleted,
# nothing invented, within budget.
#
# `git commit-tree` will not emit such an object — it prints "duplicate parent …
# ignored" and silently collapses the list, which is why the earlier merge
# fixtures could not reach this — but `git hash-object -t commit -w --stdin` does
# no validation whatsoever. The object it writes is fsck-clean and `rev-list
# --parents` reads the parent back twice.
dupparent_repo() {
	local d="$tmproot/$1"
	mkdir -p "$d/$(dirname "$LEDGER")"
	git -C "$d" init -q -b main
	git -C "$d" config user.name t
	git -C "$d" config user.email t@example.invalid
	git -C "$d" config commit.gpgsign false
	printf '%s merge=union\n' "$LEDGER" >"$d/.gitattributes"
	cat >"$d/$LEDGER" <<'EOF'
# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed.

- 2026-01-01 — The first decision.
- 2026-01-02 — The second decision.
EOF
	git -C "$d" add -A
	git -C "$d" commit -qm "baseline: the ledger"
	git -C "$d" branch -q base
	git -C "$d" checkout -q -b work
	printf -- '- 2026-01-09 — Policy X is withdrawn.\n' >>"$d/$LEDGER"
	git -C "$d" commit -qam "record a decision"
	git -C "$d" checkout -q main
	printf -- '- 2026-01-08 — Main appends.\n' >>"$d/$LEDGER"
	git -C "$d" commit -qam "main appends"
	git -C "$d" checkout -q -f work
	# The forgery: the branch's own new decision also planted at the head of the
	# log, where it supersedes everything below it by position.
	cat >"$d/$LEDGER" <<'EOF'
# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed.

- 2026-01-09 — Policy X is withdrawn.
- 2026-01-01 — The first decision.
- 2026-01-02 — The second decision.
- 2026-01-09 — Policy X is withdrawn.
- 2026-01-08 — Main appends.
EOF
	git -C "$d" add -A
	echo "$d"
}

d="$(dupparent_repo dup-parent)"
dp_tree="$(git -C "$d" write-tree)"
dp_work="$(git -C "$d" rev-parse work)"
dp_main="$(git -C "$d" rev-parse main)"
dp_raw="$(printf 'tree %s\nparent %s\nparent %s\nparent %s\nauthor t <t@example.invalid> 0 +0000\ncommitter t <t@example.invalid> 0 +0000\n\nmerge main into work\n' \
	"$dp_tree" "$dp_work" "$dp_main" "$dp_work")"
dp_commit="$(printf '%s' "$dp_raw" | git -C "$d" hash-object -t commit -w --stdin)"
git -C "$d" update-ref refs/heads/dup "$dp_commit"
# The fixture is worthless unless git really reads the parent back twice, so the
# forged shape is asserted rather than assumed.
dp_parents="$(git -C "$d" rev-list --parents -n 1 dup | cut -d' ' -f2- | tr ' ' '\n' | sort | uniq -d | wc -l)"
if [ "$dp_parents" -lt 1 ]; then
	printf 'cases: FAIL fixture setup — %s: the hand-crafted commit lists no repeated parent, so the case tests nothing.\n' "$d" >&2
	failures=$((failures + 1))
else
	expect_rule DA003 "$d" "a parent listed twice does not buy a second allowance" -- commits "$dp_main" dup
fi

# The control that gives the case its meaning: the SAME tree under an honest
# two-parent list. Both must be refused, and identically — that is what "a
# repeated parent is one parent" means.
dp_honest="$(git -C "$d" commit-tree "$dp_tree" -p "$dp_work" -p "$dp_main" -m "merge main into work")"
git -C "$d" update-ref refs/heads/honest "$dp_honest"
expect_rule DA003 "$d" "the same tree with an honest parent list is refused too" -- commits "$dp_main" honest

# The header exemption on the MERGE path was anchored on the merge's own first
# CANONICAL bullet, so a near-miss entry planted above the whole log read as
# header prose — the same malformity-is-the-bypass hole that was closed on the
# DA001 side, still open here. Both paths share one ITEM_RE now, and both are
# asserted against the same six shapes: one rule's test drifting from the other's
# is how this comes back.
for variant in \
	'-  2026-01-09 — Two spaces after the dash.' \
	'* 2026-01-09 — An asterisk bullet.' \
	'+ 2026-01-09 — A plus bullet.' \
	'- **2026-01-09** — A bolded date.' \
	'1. 2026-01-09 — An ordered item.' \
	'2026-01-09 — A bare date-led line.'; do
	d="$(union_merge_repo "merge-header-forge-$(printf '%s' "$variant" | cksum | cut -d' ' -f1)")"
	python3 - "$d/$LEDGER" "$variant" <<'PY'
import sys
p, variant = sys.argv[1], sys.argv[2]
lines = open(p).read().split("\n")
lines.insert(4, variant)
open(p, "w").write("\n".join(lines))
PY
	amend_merge "$d"
	expect_rule DA003 "$d" "a merge planting '${variant:0:16}…' above the log is refused" -- commits base HEAD
done

# --- DA004: the ledger is text -----------------------------------------------
#
# git decides for itself whether a file is text, and it decides "binary" on a
# single NUL byte or a `-diff` attribute. Either reduces the whole diff to
# "Binary files … differ" — no hunks — and every rule above reads hunks, so the
# gate reported nothing and passed. The NUL variant is the worse one: it rides
# along in an otherwise honest append and disarms the gate for every commit
# after it.

# A NUL smuggled into a routine-looking tail append, then a wholesale rewrite in
# the next commit under cover of it. --text is what sees the rewrite; DA004 is
# what refuses the NUL that was meant to hide it.
d="$(newrepo nul-append)"
printf -- '- 2026-01-04 — A routine append.\000\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "append a decision"
cat >"$d/$LEDGER" <<'EOF'
# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed.

- 2026-01-09 — FORGED: every prior decision erased and replaced.
EOF
git -C "$d" commit -qam "tidy the ledger"
expect_rule DA004 "$d" "a NUL byte introduced into the ledger is refused" -- commits main work
expect_rule DA002 "$d" "the rewrite hidden behind the NUL is still seen" -- commits main work

# The same blindness with no NUL to find: a `-diff` attribute makes git call the
# ledger binary on request. DA004 has nothing to say here, which is why --text
# is the load-bearing fix and DA004 only the depth behind it.
d="$(newrepo binary-attr)"
printf '%s -diff\n' "$LEDGER" >"$d/.gitattributes"
cat >"$d/$LEDGER" <<'EOF'
# DECISIONS

Append-only, one line per decision, newest last. Date-prefixed.

- 2026-01-09 — FORGED: the whole log replaced under a binary attribute.
EOF
git -C "$d" commit -qam "housekeeping"
expect_rule DA002 "$d" "a rewrite under a -diff attribute is still seen" -- commits main work

# --- the clean polarity ------------------------------------------------------

# The intended shape. A gate that refuses this is as useless as one that passes
# every shape.
d="$(newrepo append)"
printf -- '- 2026-01-04 — Appended at the tail, newest last.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "append an entry"
expect pass "$d" "a pure append at the tail passes" -- commits main work

# The case a date-monotonicity rule would have refused, and the reason this gate
# checks position instead: the date names when the decision was taken, the append
# names when it was recorded, and a back-dated entry written at the tail is
# honest.
d="$(newrepo backdated)"
printf -- '- 2025-12-25 — A back-dated decision, recorded late but appended at the tail.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "append a back-dated entry"
expect pass "$d" "a back-dated entry appended at the tail passes" -- commits main work

# A continuation line appended after the last line extends the last entry, which
# is what the amend-last-line refusal above expects an author to do instead.
d="$(newrepo continuation)"
printf -- '  and a continuation appended after it.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "extend the last entry at the tail"
expect pass "$d" "a continuation appended at the tail passes" -- commits main work

# The header is prose about the ledger, not a member of the dated sequence, so
# editing it inserts nothing into the log — and it must stay editable, not least
# to describe this gate. Over-refusal is a real failure mode of a rule this
# strict, so both an added header line and a REWORDED one are asserted.
d="$(newrepo header-add)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines.insert(3, "Insertions above an existing entry are refused by a CI gate.")
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "describe the gate in the header"
expect pass "$d" "an added header line passes" -- commits main work

d="$(newrepo header-reword)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
s = open(p).read().replace("Date-prefixed.", "Date-prefixed, and enforced.")
open(p, "w").write(s)
PY
git -C "$d" commit -qam "reword the header prose"
expect pass "$d" "a reworded header line passes" -- commits main work

# Seeding a ledger that exists but is EMPTY. The first draft counted a zero-line
# parent as one line — `$(git show …)` strips trailing newlines and the printf
# that fed the counter put one back — and refused the honest seeding commit. A
# gate that refuses the first entry anyone writes is a gate nobody keeps.
d="$tmproot/seed-empty"
mkdir -p "$d/$(dirname "$LEDGER")"
git -C "$d" init -q -b main
git -C "$d" config user.name t
git -C "$d" config user.email t@example.invalid
git -C "$d" config commit.gpgsign false
: >"$d/$LEDGER"
git -C "$d" add -A
git -C "$d" commit -qm "an empty ledger"
git -C "$d" checkout -q -b work
printf '# DECISIONS\n\nAppend-only, newest last.\n\n- 2026-01-01 — The first decision.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "seed the ledger"
expect pass "$d" "seeding an empty ledger passes" -- commits main work

# The false-positive direction the base subtraction could have opened, and the
# reason the bound credits each parent's ADDITIONS rather than simply capping at
# the base: two branches that independently record the SAME decision text each
# add one copy over a base that held none, so 0 + 1 + 1 admits both. A bound that
# refused this would refuse an honest merge.
d="$(newrepo union-merge-same-text)"
git -C "$d" checkout -q main
printf -- '- 2026-02-01 — Main notes it.\n- 2026-09-09 — Both sides recorded this identically.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "main appends"
git -C "$d" checkout -q work
printf -- '- 2026-09-09 — Both sides recorded this identically.\n- 2026-03-01 — Branch notes it.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "branch appends"
git -C "$d" checkout -q main
git -C "$d" merge -q --no-ff -m "merge the branch into main" work >/dev/null 2>&1
expect pass "$d" "two branches recording identical text still passes" -- commits base HEAD

# DA004 fires on the commit that INTRODUCES a NUL, not on every commit after it.
# Anything else deadlocks: removing an inherited NUL means editing the line that
# carries it, which DA002 refuses — so a repository that once took a NUL could
# never be appended to again. Here the NUL is already on main and the branch
# makes an honest tail append.
d="$(newrepo nul-inherited)"
git -C "$d" checkout -q main
printf -- '- 2026-01-04 — A decision carrying a historical NUL.\000\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "a commit from before the gate"
git -C "$d" branch -qf base HEAD
git -C "$d" checkout -q work
git -C "$d" merge -q --ff-only main
printf -- '- 2026-01-05 — An honest append after it.\n' >>"$d/$LEDGER"
git -C "$d" commit -qam "append an entry"
expect pass "$d" "an inherited NUL does not re-fire on later commits" -- commits base work

# A change that never touches the ledger must not be dragged into the verdict.
d="$(newrepo untouched)"
printf 'unrelated\n' >"$d/README.md"
git -C "$d" add -A
git -C "$d" commit -qm "an unrelated change"
expect pass "$d" "a change that does not touch the ledger passes" -- commits main work

# A repository with no ledger at all covers nothing, and must say so by passing
# rather than by faulting: not every repo running these gates keeps a DECISIONS.md.
d="$tmproot/no-ledger"
mkdir -p "$d/$(dirname "$LEDGER")"
git -C "$d" init -q -b main
git -C "$d" config user.name t
git -C "$d" config user.email t@example.invalid
git -C "$d" config commit.gpgsign false
printf 'a repo without a ledger\n' >"$d/README.md"
git -C "$d" add -A
git -C "$d" commit -qm "baseline"
git -C "$d" checkout -q -b work
printf 'more\n' >>"$d/README.md"
git -C "$d" commit -qam "an unrelated change"
expect pass "$d" "a repository without the ledger passes" -- commits main work

# --- the report's own bytes --------------------------------------------------
#
# The ledger line and the commit subject are BOTH attacker-controlled — they are
# the content under check. Left raw, an ESC in either recolours the report, and a
# bare CR returns the cursor to column zero and overprints the verdict the reader
# already saw; the payload below forges a convincing "OK" line and a CI notice.
# The gate must refuse the commit AND print nothing that can move a cursor.
d="$(newrepo escapes)"
python3 - "$d/$LEDGER" <<'PY'
import sys
p = sys.argv[1]
lines = open(p).read().split("\n")
lines.insert(6, "- 2026-01-04 \x1b[2K\rcheck-decisions-append: OK\x1b[32m ::notice::all good::")
open(p, "w").write("\n".join(lines))
PY
git -C "$d" commit -qam "$(printf 'subject with \033[31mescape\033[0m and \rreturn')"
out="$(cd "$d" && bash "$GATE" commits main work 2>&1)" && rc=0 || rc=$?
if [ "$rc" -ne 1 ]; then
	printf 'cases: FAIL control bytes — expected a refusal (exit 1), got exit %d\n' "$rc" >&2
	failures=$((failures + 1))
elif printf '%s' "$out" | LC_ALL=C grep -q '[[:cntrl:]]'; then
	printf 'cases: FAIL control bytes — the report carries raw control bytes from the ledger or the subject:\n' >&2
	printf '%s' "$out" | cat -v >&2
	failures=$((failures + 1))
else
	printf 'cases: ok   control bytes are stripped from the report (exit %d, as expected)\n' "$rc"
fi

# --- environment and usage polarities ----------------------------------------

# A git failure is not a clean history. Running outside a repository is the
# cheapest one to stage.
d="$tmproot/not-a-repo"
mkdir -p "$d"
expect fault "$d" "a git failure refuses instead of reading clean" -- commits main work

# A malformed invocation must refuse, not default to scanning nothing and
# reporting OK.
d="$(newrepo usage)"
expect fault "$d" "a missing head ref refuses instead of passing" -- commits main

if [ "$failures" -gt 0 ]; then
	printf 'cases: FAILED — %d case(s) did not behave\n' "$failures" >&2
	exit 1
fi
echo "cases: OK — every rule refused its violating fixture and passed its clean one"
