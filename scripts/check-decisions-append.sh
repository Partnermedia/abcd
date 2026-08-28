#!/usr/bin/env bash
# Deterministic append-only gate for .abcd/work/DECISIONS.md
# (iss-2608271804494867).
#
# The ledger's own header declares "Append-only, one line per decision, newest
# last", and nothing enforced it. The file already carries five backwards date
# steps among its dated bullets — historical, committed, and NOT to be repaired:
# reordering a committed append-only log is the one operation append-only
# forbids, and a rewrite to satisfy a gate is worse than the drift it hides.
#
# So this gate does not check dates and does not read the existing order. A
# date-monotonicity rule would refuse an honest back-dated entry appended at the
# tail — a real and legitimate act, since the date names when the decision was
# taken and the append names when it was recorded. What is never legitimate is
# writing INTO the log that is already committed.
#
# THE CONTRACT, in four rules. The first three are scoped BELOW THE HEADER
# REGION (see the boundary section): the header is prose about the ledger, the
# rules govern the dated log underneath it. The fourth is about the file's bytes
# and is scoped to the whole of it.
#
#   DA001  POSITION. In a single-parent commit, a line ADDED to the ledger must
#          land after the last line the parent already had. An addition with
#          surviving pre-existing content below it is an insertion, and is
#          refused.
#
#   DA002  PRESERVATION. In ANY commit, against EVERY parent, the diff must
#          contain no REMOVED line below the header region. A legitimate append
#          deletes nothing. Position alone is not append-only: DA001 only fires
#          while pre-existing content SURVIVES below the addition, so a hunk
#          whose changed region reaches end-of-file rewrites or erases committed
#          decisions with nothing beneath them to trip the position rule. That is
#          the tail-reaching rewrite, and its two-commit cousin — truncate the
#          log in one commit, restore a forged history in the next, each half
#          positionally innocent. DA002 is what refuses both.
#
#          DA002 is therefore stricter than position alone, deliberately and in
#          the direction the ledger's own header asks for: correcting an entry
#          means APPENDING a new dated entry, never editing the old one in place.
#          Amending the file's last line — the one edit position could not see —
#          is a refusal under this rule too. Reword-in-place has no legitimate
#          form in an append-only log.
#
#   DA003  MERGE AUTHORS NOTHING. A merge commit's ledger may contain a line no
#          more times than the MERGE BASE held it plus what each parent ADDED
#          over that base:
#
#            bound(t) = base(t) + Σ_parents max(0, parent(t) − base(t))
#
#          Merges cannot simply be skipped: a hand-forged conflict resolution is
#          authored content that appears in no single-parent commit, so skipping
#          merges leaves an unchecked write path straight into the ledger. But
#          neither can DA001 be applied to a merge, in any form. The ledger is
#          declared `merge=union` in .gitattributes (iss-118), and the union
#          driver concatenates the two sides of each conflicting region
#          independently: where a branch and main have both appended across more
#          than one region, the result genuinely carries each side's lines above
#          the other's in different places, and is a tail extension of NEITHER
#          parent. That shape is the norm here, not the exception — it appears in
#          eight of the merges in the last eight hundred commits, every one of
#          them a routine `Merge branch 'main' into <branch>`.
#
#          What a union merge never does is INVENT a line, or emit one more times
#          than the base held it plus what the sides added. That is the invariant
#          which survives the interleaving, and it took two corrections to state.
#          Set membership was the first form, blind to multiplicity: a merge could
#          copy a committed decision to the head of the log and pass, every line
#          present in a parent and nothing removed. Summing the parents was the
#          second, and it DOUBLE-BUDGETED common history — a line both sides
#          inherited unchanged counts once per parent, licensing two copies of
#          every line the repository ever committed. That is not a rounding error:
#          it admits splicing a complete reordered rendition of the whole log
#          above the real one, and it was demonstrated against the live 2,155-line
#          ledger with 2,150 lines added, zero removed, gate green. Subtracting
#          what the base already held is the correction, and the legitimate shape
#          still passes: two branches that independently record the same decision
#          text have base 0 and one addition each, so 0 + 1 + 1 admits both copies.
#
#          Together with DA002 against every parent — no side's decisions
#          disappear — a merge can neither drop, fabricate, nor multiply a
#          decision, which is the whole of what the merge write path can do.
#
#   DA004  THE LEDGER IS TEXT. No commit may introduce a NUL byte into it. This
#          is not hygiene: git decides on its own whether a file is text, and one
#          NUL makes it binary, which reduces the entire diff to "Binary files …
#          differ" with no hunks. Every rule above reads hunks, so a NUL smuggled
#          into an otherwise honest append silently disarms all of them, for that
#          commit and for every commit after it. The diffs are read with --text so
#          that cannot happen; DA004 refuses the NUL as well, because --text
#          rescues this gate and nothing else — a NUL breaks grep, the site
#          renderer and any editor that opens the file. It fires only on the
#          commit that INTRODUCES one: a NUL already in a parent is history, and
#          demanding its removal would deadlock against DA002.
#
#          The `-diff` attribute produces the same hunkless diff with no NUL to
#          find, which is why --text is the load-bearing fix and DA004 is depth
#          behind it, not the other way round.
#
# Boundaries, decided from the file's actual structure:
#
#   * The HEADER REGION — everything above the first `- <YYYY-MM-DD>` bullet in
#     the PARENT's copy (today lines 1-7: the title and the paragraph describing
#     the convention) — is exempt from the first three rules. It is prose about the
#     ledger, not a member of the dated sequence, so an edit there inserts
#     nothing into the log and deletes no decision; the paragraph must stay
#     editable, not least to describe this gate. The exemption is anchored on the
#     PARENT's first bullet, so it cannot be widened by the same commit that uses
#     it.
#
#     The exemption is withdrawn from any hunk that adds a LIST-SHAPED line,
#     because the header's last line and the log's first entry share one
#     insertion point: a paragraph appended to the prose and an entry slipped
#     above the oldest decision land at the same line number, and only the added
#     line's own shape tells them apart. The shape test is deliberately wider
#     than the canonical `- YYYY-MM-DD` bullet — `-  ` with two spaces, `* `,
#     `+ `, `- **2026-01-04**`, a dash-led or date-led line, an ordered item —
#     because a forged entry does not have to be well-formed to be read as one,
#     and matching only the canonical shape makes malformity the bypass. A header
#     note must therefore be prose, which is what a header paragraph is.
#
#   * The LAST ENTRY'S interior is NOT exempt. A dated bullet spans continuation
#     lines, and extending the final entry is legitimate — but a continuation
#     appended after the file's last line already extends it, because the last
#     entry ends at the tail. Writing above that last line buys nothing and is
#     mechanically indistinguishable from any other mid-file insertion, so
#     allowing it would reopen the whole rule for no gain.
#
# REDACTION AND GRADUATION — the two legitimate operations this gate refuses.
#
# DA002 refuses every deletion below the header, and two honest changes are
# deletions: redacting a line that leaked a private name or other content that
# must not stay in the working tree, and the graduation the ledger's own header
# plans for — migrating to per-file `decisions/<date>--<slug>.md` when size or
# merge contention bites, which empties this file by construction.
#
# Neither gets an escape hatch, and that is the design. A bypass trailer, an
# allow-marker, a skip path keyed on a commit message: each one is a lock whose
# key is written on the door, because the commit that wants to bypass the gate is
# the commit that writes the trailer. An append-only gate with a
# craftable exemption is not append-only; it is a gate that asks nicely.
#
# So both operations are DELIBERATE GATE-EDIT-AND-REVIEW changes. The pull
# request that redacts, or that graduates the ledger to per-file records, ALSO
# adjusts or retires this gate in the same diff — and the two halves are reviewed
# together, by a human who can see that the deletion and the permission for it
# arrived at once. That is slower than a trailer by exactly the amount of scrutiny
# the operation deserves: a redaction is a rewrite of a durable record, and a
# graduation retires the record's format. Neither should be reachable by a commit
# nobody had to approve. Note also that redacting the working tree does not
# redact the history — the leaked line stays in every prior commit — so the
# redaction PR is where that separate remedy gets decided too.
#
# A pure deletion adds nothing, so DA001 cannot see it; DA002 is what refuses it.
# Every untrusted string this gate prints — a commit subject, a line of the
# ledger — is passed through the shell analogue of internal/termsafe before it
# reaches the report: the content under check is exactly the content an attacker
# controls, and a raw ESC or CR in it can overprint or recolour the verdict the
# reader is relying on.
#
# Usage:
#   check-decisions-append.sh commits <base-ref> <head-ref>
#
# Exit 0 clean, 1 a violation, 2 a usage/environment fault.
set -euo pipefail

# Resolve every path from the repository root, like the sibling gates
# check-reviews.sh and check-issue-resolution.sh: LEDGER is relative and a git
# pathspec is matched against the current directory, so from a subdirectory the
# diff would match nothing and the gate would report a clean pass having scanned
# nothing. cd first, so cwd cannot disarm it.
#
# Every git probe below is fail-closed: `cd "$(...)"` collapses to a successful
# `cd ""` when the substitution fails under errexit, and a swallowed git error
# reads exactly like an untouched ledger — a gate that cannot tell them apart is
# a false green.
rc=0
toplevel="$(git rev-parse --show-toplevel 2>&1)" || rc=$?
if [ "$rc" -ne 0 ]; then
	echo "check-decisions-append: not a readable git repository (git rev-parse --show-toplevel exit $rc) — refusing rather than reporting a vacuous pass:" >&2
	echo "$toplevel" >&2
	exit 2
fi
cd "$toplevel"

LEDGER=".abcd/work/DECISIONS.md"
# A dated bullet opens an entry. Written out longhand rather than with {4}:
# interval expressions are not portable across every awk this repository runs on.
BULLET_RE='^- [0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]'
# A LIST-SHAPED line: what a forged entry looks like when it is not well formed.
# Declared ONCE and handed to every awk that needs it. Two rules test this shape
# — DA001's header exemption and DA003's — and a second copy of it is how
# "malformity is the bypass" comes back: the copies drift, and the shape one rule
# refuses becomes the shape the other lets through.
#
# Wider than BULLET_RE on purpose: a bullet with two spaces, `*`/`+`/`>`, an
# ordered item, a bolded date, a dash-led line (ASCII, en or em), or a bare
# date-led line. `[[:space:]]*` rather than `[ \t]*` because awk's -v processing
# would resolve the escape before the regex ever sees it.
ITEM_RE='^[[:space:]]*([-*+>]|[0-9]+[.)]|[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]|—|–)'

violations=0

fail() {
	printf 'check-decisions-append: %s\n' "$1" >&2
	violations=$((violations + 1))
}

usage() {
	echo "usage: check-decisions-append.sh commits <base-ref> <head-ref>" >&2
	exit 2
}

# termsafe is the shell analogue of internal/termsafe.Sanitize for a value
# interpolated into ONE line of the report. Tab becomes a space; every other C0
# control and DEL is dropped — ESC (so an injected ANSI sequence cannot recolour
# or reposition the verdict), bare CR (which returns the cursor to column zero
# and overprints the line the reader already saw), NUL, and an injected LF (which
# would forge an extra report line, including a convincing fake "OK").
#
# The C1 range 0x80-0x9F that termsafe.Sanitize also masks is deliberately NOT
# stripped here: `tr` under LC_ALL=C deletes BYTES, and those byte values are
# continuation bytes of ordinary UTF-8 — the em dash that separates every entry
# in this ledger is E2 80 94. Deleting them would corrupt the honest text of
# every message this gate prints in order to close a path that needs an 8-bit
# terminal to be a path at all.
termsafe() {
	printf '%s' "$1" | LC_ALL=C tr '\011' ' ' | LC_ALL=C tr -d '\000-\037\177'
}

# A shallow checkout blinds this gate the same way it blinds its siblings: past
# the graft boundary a commit's parent is not present, so every append would read
# as an add of the whole file and the check would cover nothing. Report the
# environment fault as itself rather than passing vacuously. The probe is
# rc-checked, so the very fault this arm exists to report cannot disarm it.
rc=0
shallow="$(git rev-parse --is-shallow-repository 2>&1)" || rc=$?
if [ "$rc" -ne 0 ]; then
	echo "check-decisions-append: git rev-parse --is-shallow-repository failed (exit $rc) — refusing rather than reporting a vacuous pass:" >&2
	echo "$shallow" >&2
	exit 2
fi
case "$shallow" in
true)
	echo "check-decisions-append: shallow checkout — the rules compare each commit against its parents, which are not present past the graft; run 'git fetch --unshallow' first (CI checks out with fetch-depth: 0)." >&2
	exit 2
	;;
false) ;;
*)
	echo "check-decisions-append: unexpected git rev-parse --is-shallow-repository output \"$shallow\" — refusing rather than guessing." >&2
	exit 2
	;;
esac

# The empty tree, for diffing a root commit (which has no parent) against
# something. `hash-object` without -w computes the id and writes nothing.
rc=0
EMPTY_TREE="$(git hash-object -t tree /dev/null 2>&1)" || rc=$?
if [ "$rc" -ne 0 ]; then
	echo "check-decisions-append: git hash-object failed (exit $rc) — refusing rather than reporting a vacuous pass:" >&2
	echo "$EMPTY_TREE" >&2
	exit 2
fi

# ledger_shape prints "<total-lines>\t<header-end-line>" for the ledger at a
# tree-ish. header-end is the line above the first dated bullet; with no bullet at
# all (a ledger that is still only a header, or an empty one) the whole file is
# header, which makes the very first entry an append by definition.
#
# The line count comes from awk reading git's own byte stream, never from a
# command substitution: `$(git show …)` strips every trailing newline, so an
# EMPTY ledger would come back as the empty string and `printf '%s\n'` would then
# count it as one line — refusing the commit that seeds a new ledger, and
# miscounting any file that ends in blank lines. pipefail (set above) is what
# makes the pipeline's git half still fail-closed.
ledger_shape() {
	local ref="$1"
	if ! git cat-file -e "$ref:$LEDGER" 2>/dev/null; then
		printf '0\t0\n'
		return 0
	fi
	local rc=0 shape
	shape="$(git show "$ref:$LEDGER" 2>/dev/null | awk -v re="$BULLET_RE" '
		!seen && $0 ~ re { seen = 1; h = NR - 1 }
		END { if (!seen) h = NR; printf "%d\t%d\n", NR, h }
	')" || rc=$?
	if [ "$rc" -ne 0 ]; then
		echo "check-decisions-append: reading $LEDGER at $ref failed (exit $rc) — refusing rather than reporting a vacuous pass." >&2
		exit 2
	fi
	printf '%s\n' "$shape"
}

# analyse prints one tab-separated finding per line — "<rule>\t<old-line>\t
# <count>\t<sample>" — for the ledger diff between a parent and a commit. It
# emits both DA001 and DA002 findings; the caller decides which apply, because a
# merge's parents are judged by DA002 individually and by DA001 collectively.
analyse() {
	local commit="$1" parent="$2"

	local shape old_total header_end
	shape="$(ledger_shape "$parent")"
	old_total="${shape%%$'\t'*}"
	header_end="${shape##*$'\t'}"

	# --unified=0 so each hunk's own line numbers locate the change exactly;
	# --no-renames so a rename into the path is reported as a plain add rather
	# than a similarity header with no hunks to read.
	#
	# --text is LOAD-BEARING, not a tidiness flag. Without it git decides for
	# itself whether the ledger is text, and it decides "binary" on either of two
	# things an author controls: a single NUL byte anywhere in the file, or a
	# `-diff` mark in .gitattributes. Either one reduces the whole diff to
	# "Binary files … differ" — no hunks at all — and every rule here reads hunks,
	# so the gate reports nothing and passes. Worse, it passes PERMANENTLY: the
	# NUL rides along in an otherwise honest-looking append, and every later
	# commit is then free to rewrite the log wholesale. --text forces the textual
	# diff regardless, which restores every rule. (DA004 refuses the NUL as well;
	# neither fix is allowed to be the only one, because each covers a case the
	# other does not — DA004 says nothing about the `-diff` attribute, and --text
	# says nothing about the rest of the toolchain that a NUL breaks.)
	#
	# On a ledger that already carries a NUL, bash prints "ignored null byte in
	# input" for this capture and drops the byte. That warning is left visible
	# rather than silenced: it is true, it is harmless here (a NUL is not a
	# newline, so no hunk header, line number or +/- prefix moves, and the only
	# loss is a byte termsafe would strip from the sample anyway), and it appears
	# only alongside the DA004 refusal that is already telling the reader the
	# ledger is not text. Silencing a warning inside a gate is how a gate stops
	# being one.
	local rc=0 diffout
	diffout="$(git diff --unified=0 --no-color --no-ext-diff --no-renames --text "$parent" "$commit" -- "$LEDGER" 2>&1)" || rc=$?
	if [ "$rc" -ne 0 ]; then
		echo "check-decisions-append: git diff failed for ${commit} against ${parent} (exit $rc) — refusing rather than reporting a vacuous pass:" >&2
		echo "$diffout" >&2
		exit 2
	fi
	[ -n "$diffout" ] || return 0

	# For a hunk `@@ -a,b +c,d @@` under --unified=0: the hunk removes old lines
	# a .. a+b-1 (with no context, every old line it covers is removed) and places
	# its additions among them. With b == 0 nothing is removed and the addition
	# sits after old line a, so the next surviving old line is a + 1; with b > 0
	# the next surviving old line is a + b. An addition is an INSERTION exactly
	# when that next surviving old line exists — i.e. is <= old_total.
	printf '%s\n' "$diffout" | awk -v old_total="$old_total" -v header_end="$header_end" -v item_re="$ITEM_RE" '
		function flush(   first_after, last_old) {
			if (!pending) return
			if (adds > 0) {
				first_after = (b == 0) ? a + 1 : a + b
				last_old    = (b == 0) ? a     : a + b - 1
				# Header prose is exempt — unless the addition is list-shaped, and
				# so an entry smuggled in at the boundary the header shares with
				# the first decision.
				if (!(last_old <= header_end && !has_item) && first_after <= old_total)
					printf "DA001\t%d\t%d\t%s\n", a, adds, add_sample
			}
			# Every removed line below the header region is a committed decision
			# being erased or rewritten. Reported once per hunk, naming the first.
			if (dels_below > 0)
				printf "DA002\t%d\t%d\t%s\n", del_first_line, dels_below, del_sample
		}
		/^@@ / {
			flush()
			hdr = $2                                    # "-a,b"
			sub(/^-/, "", hdr)
			if (index(hdr, ",")) { split(hdr, p, ","); a = p[1] + 0; b = p[2] + 0 }
			else                 { a = hdr + 0; b = 1 }
			pending = 1
			adds = 0; add_sample = ""; has_item = 0
			seen_dels = 0; dels_below = 0; del_first_line = 0; del_sample = ""
			next
		}
		pending && substr($0, 1, 1) == "+" {
			adds++
			content = substr($0, 2)
			if (add_sample == "") add_sample = content
			# List-shaped (ITEM_RE, shared with DA003): an entry, however
			# malformed. Its presence withdraws the header exemption.
			if (content ~ item_re) has_item = 1
			next
		}
		pending && substr($0, 1, 1) == "-" {
			# The k-th removed line of the hunk is old line a + k - 1.
			oldno = a + seen_dels
			seen_dels++
			if (oldno > header_end) {
				dels_below++
				if (del_first_line == 0) { del_first_line = oldno; del_sample = substr($0, 2) }
			}
			next
		}
		END { flush() }
	'
}

# report_findings prints the findings that apply. `want` is DA001+DA002 for a
# single-parent commit and DA002 alone for one parent of a merge.
report_findings() {
	local commit="$1" parent="$2" subject="$3" old_total="$4" want="$5" findings="$6"
	local rule at count sample
	while IFS=$'\t' read -r rule at count sample; do
		[ -n "$rule" ] || continue
		case "$want" in
		DA002) [ "$rule" = "DA002" ] || continue ;;
		esac
		sample="$(termsafe "$sample")"
		case "$rule" in
		DA001)
			fail "DA001 commit ${commit:0:12} (${subject}) inserts ${count} line(s) into $LEDGER above line $((at + 1)) of its parent's copy, which has ${old_total} lines. The ledger is append-only, newest last: append the entry at the end of the file instead. First inserted line: ${sample}"
			;;
		DA002)
			fail "DA002 commit ${commit:0:12} (${subject}) removes ${count} committed line(s) from $LEDGER, starting at line ${at} of ${parent:0:12}'s copy. The ledger is append-only: correct an entry by appending a new dated one, never by editing or deleting the old one. First removed line: ${sample}"
			;;
		esac
	done <<<"$findings"
}

check_commit() {
	local commit="$1"
	local rc=0 line
	line="$(git rev-list --parents -n 1 "$commit" 2>&1)" || rc=$?
	if [ "$rc" -ne 0 ]; then
		echo "check-decisions-append: git rev-list --parents failed for $commit (exit $rc) — refusing rather than reporting a vacuous pass:" >&2
		echo "$line" >&2
		exit 2
	fi
	# "<sha> <parent1> <parent2> …" — drop the commit's own sha, then DE-DUPLICATE.
	#
	# A commit object may list the same parent twice. `git commit-tree` refuses to
	# emit one ("duplicate parent … ignored"), but `git hash-object -t commit -w`
	# does no validation at all, and the resulting object is fsck-clean and read
	# back by `rev-list --parents` with the parent appearing twice. That is a
	# forged allowance: DA003 credits each parent with what it ADDED over the base,
	# so a parent counted twice contributes its additions twice, and any line that
	# side legitimately appended may then be planted a second time above the whole
	# log — within budget, nothing deleted, nothing invented.
	#
	# De-duplicating HERE rather than inside DA003 fixes every consumer at once:
	# the parent count that decides merge-versus-single-parent (so [X, X] is judged
	# as the single-parent commit it actually is, which subjects it to DA001 too),
	# the DA002 loop (which would otherwise report the same violation once per
	# copy), and DA003's own parent list. One dedup, one point of truth.
	local parents="" p
	# shellcheck disable=SC2086  # sha words, deliberately split
	set -- $line
	shift
	for p in "$@"; do
		case " $parents " in
		*" $p "*) continue ;;
		esac
		parents="${parents:+$parents }$p"
	done

	local subject
	subject="$(termsafe "$(git show -s --format='%s' "$commit")")"

	local nparents=0
	for p in $parents; do nparents=$((nparents + 1)); done

	# DA004 first, on every shape of commit: a NUL is what makes git call the
	# ledger binary, and a binary ledger is what makes the rules below see nothing.
	check_nul "$commit" "$subject" "$parents"

	if [ "$nparents" -le 1 ]; then
		# A root commit has no parent; diff it against the empty tree so the whole
		# file reads as an add — and therefore as an append into an empty ledger.
		local parent="${parents:-$EMPTY_TREE}"
		local shape old_total findings
		shape="$(ledger_shape "$parent")"
		old_total="${shape%%$'\t'*}"
		findings="$(analyse "$commit" "$parent")"
		[ -n "$findings" ] || return 0
		report_findings "$commit" "$parent" "$subject" "$old_total" "DA001+DA002" "$findings"
		return 0
	fi

	# A merge. DA002 against every parent — no side's committed decisions may
	# disappear through a conflict resolution. DA001 is NOT applied: a union
	# merge interleaves, so the result is a tail extension of neither parent.
	# DA003 carries the other half — the merge may invent nothing.
	local p
	for p in $parents; do
		local shape old_total findings
		shape="$(ledger_shape "$p")"
		old_total="${shape%%$'\t'*}"
		findings="$(analyse "$commit" "$p")"
		if [ -n "$findings" ]; then
			report_findings "$commit" "$p" "$subject" "$old_total" "DA002" "$findings"
		fi
	done
	check_merge_authored "$commit" "$subject" "$parents"
}

# check_merge_authored implements DA003: for every line, the merge's ledger may
# hold it no more times than the BASE held it plus what each parent ADDED.
#
#     bound(t) = count_base(t) + Σ_parents max(0, count_parent(t) − count_base(t))
#
# Counting rather than a diff, because the union driver's interleaving defeats
# every position-based comparison while leaving this invariant intact.
#
# THE BASE TERM IS THE WHOLE POINT, and it is the correction to a bound that was
# wrong. The first counting form summed the parents outright, which DOUBLE-BUDGETS
# every line of COMMON history: a decision both sides inherited unchanged is
# counted once per parent, so a two-parent merge was licensed to hold two copies
# of every line the repository had ever committed. That is not a rounding error —
# it admits splicing a complete, reordered rendition of the entire log above the
# real one, deleting nothing and inventing nothing. It was demonstrated against
# the live 2,155-line ledger: 2,150 lines added, zero removed, gate green.
# Subtracting the base collapses a shared line to 1 + 0 + 0 = 1 and refuses it.
#
# The bound is still exactly what a union merge can emit, so the legitimate shapes
# survive: two branches that independently record the SAME decision text have
# base 0 and one addition each, 0 + 1 + 1 = 2, and both copies are admitted.
#
# Written as "base plus each parent's additions" rather than the equivalent
# "Σ parents − base" because only this form generalises: with three or more
# parents the subtractive version has no single sensible reading, while additions
# per parent stay well defined however many there are.
#
# count_base(t) IS THE MAXIMUM OVER THE PAIRWISE MERGE-BASES of the parents that
# carry the ledger — never `git merge-base --octopus`, and this is the difference
# between a rule and a suggestion. The base term is the only thing standing
# between a merge and a licence to duplicate all of common history, and the
# octopus base is chosen by data the merge author controls: adding ONE extra
# parent older than the ledger drags the common ancestor of ALL parents below the
# point the ledger existed, count_base collapses to 0 through the "base carries no
# ledger" path, the bound reverts to the plain sum, and the full-log splice is
# green again. The repository's own initial commit is enough; an unrelated orphan
# root does it too, by making the octopus base fail to resolve at all. Both were
# demonstrated against the live ledger at +2,150/−0 with the gate silent.
#
# Max over pairwise bases is the MONOTONE choice, and monotone in the direction
# that matters: every extra parent can only add pairs, and a maximum taken over a
# superset is greater than or equal to the one over the subset. So an added parent
# can only RAISE count_base, which only LOWERS the bound. An author who bolts on
# parents can tighten this rule against themselves; they cannot loosen it. The
# octopus base has exactly the opposite monotonicity, which is why it was
# exploitable: reaching further back shrinks count_base and enlarges the bound.
#
# Degradation is therefore LOUD. When two or more parents carry the ledger and no
# pairwise base resolves, or none of the ones that resolve carries the ledger,
# this gate exits 2 rather than quietly falling back to a weaker bound — the same
# "refusing rather than reporting a vacuous pass" doctrine every git probe here
# already follows. A silent fallback IS the exploit: the attack does not need to
# defeat the bound, only to reach the branch where the bound stops applying. That
# shape does not occur in this repository's honest history — the ~800-commit
# replay produces no such merge — so the refusal costs nothing real and closes the
# path entirely. With fewer than two ledger-carrying parents there is no shared
# history to double-budget, so count_base is 0 legitimately and the bound is
# simply what that one parent holds.
#
# Residual, accepted: if the union driver ever pulls a line the two sides SHARE
# into a conflicting region, it emits that line once per side and the merge holds
# one more copy than this bound allows. It needs the shared line to sit inside the
# conflict rather than in the context around it, which two tail appends do not
# produce; the ~800-commit replay over this repository's real history reports no
# such merge. The alternative — an interleaving/subsequence rule — does not
# actually close this: each parent IS a subsequence of the spliced-rendition
# forgery above, so subsequence alone passes it, and the "every merge line
# consumed" half needed to refuse it is the same multiplicity argument implemented
# less directly.
#
# The header exemption is withdrawn from LIST-SHAPED lines here exactly as it is
# in DA001, and by the same shared ITEM_RE. Anchoring the exemption on the
# merge's own first canonical bullet is what let a near-miss entry — `-  ` with
# two spaces, `* `, a bolded date — sit above the whole log and read as header
# prose. That is the malformity-is-the-bypass class, and it has to be closed on
# both paths or it is not closed.
check_merge_authored() {
	local commit="$1" subject="$2" parents="$3"

	# A merge that does not touch the ledger cannot have authored anything in it.
	git cat-file -e "$commit:$LEDGER" 2>/dev/null || return 0

	local basetmp allowtmp rc=0 p
	basetmp="$(mktemp)" && allowtmp="$(mktemp)" || {
		echo "check-decisions-append: mktemp failed — refusing rather than reporting a vacuous pass." >&2
		exit 2
	}

	# Only the parents that CARRY the ledger take part. A parent without it holds
	# no line, so it can neither share history with another parent nor add
	# anything — and letting it into the base computation is precisely the
	# exploit, since a ledgerless parent is what drags a common ancestor below the
	# point the ledger existed.
	# The same-sha guard is repeated here, not merely inherited from check_commit.
	# This function's whole result is a per-parent SUM, so a repeated parent is a
	# doubled allowance — the one input it must never take on trust. check_commit
	# de-duplicates before calling, and this keeps the property local to the
	# function that depends on it. Both nl and lparents count a sha once, so the
	# ledger-carrying-parent tally that gates the exit-2 refusal below cannot be
	# tricked into seeing one parent as two either.
	local lparents="" nl=0
	for p in $parents; do
		case " $lparents " in
		*" $p "*) continue ;;
		esac
		git cat-file -e "$p:$LEDGER" 2>/dev/null || continue
		lparents="$lparents $p"
		nl=$((nl + 1))
	done

	# count_base(t) = MAX over the pairwise merge-bases of those parents. Each
	# base's per-line counts go into basetmp as "<count><TAB><line>", the maximum
	# resolved as they are read.
	local a b pb bases_found=0 ledger_bases=0 tmpa tmpb
	tmpa="$(mktemp)" && tmpb="$(mktemp)" || {
		echo "check-decisions-append: mktemp failed — refusing rather than reporting a vacuous pass." >&2
		exit 2
	}
	# shellcheck disable=SC2064  # expand the paths now, not at trap time
	trap "rm -f '$basetmp' '$allowtmp' '$tmpa' '$tmpb'" RETURN
	for a in $lparents; do
		for b in $lparents; do
			[ "$a" \< "$b" ] || continue # each unordered pair once
			# A pair with no common ancestor is not a fault on its own; the
			# aggregate verdict below is what decides.
			pb="$(git merge-base "$a" "$b" 2>/dev/null)" || pb=""
			pb="${pb%%$'\n'*}"
			[ -n "$pb" ] || continue
			bases_found=$((bases_found + 1))
			git cat-file -e "$pb:$LEDGER" 2>/dev/null || continue
			ledger_bases=$((ledger_bases + 1))
			git show "$pb:$LEDGER" >"$tmpa" 2>/dev/null || {
				echo "check-decisions-append: reading $LEDGER at the merge base $pb failed — refusing rather than reporting a vacuous pass." >&2
				exit 2
			}
			awk '{ c[$0]++ } END { for (t in c) printf "%d\t%s\n", c[t], t }' "$tmpa" >>"$tmpb"
		done
	done

	# Loud degradation. A silent fallback to a weaker bound IS the exploit: the
	# attack never has to beat the bound, only to reach the branch where the bound
	# stops applying. With two or more ledger-carrying parents there is shared
	# history by construction, so a base that cannot be found or cannot be read is
	# an environment the gate cannot judge — not a licence to judge it leniently.
	if [ "$nl" -ge 2 ] && [ "$ledger_bases" -eq 0 ]; then
		echo "check-decisions-append: merge commit ${commit:0:12} has $nl parents carrying $LEDGER, but $( [ "$bases_found" -eq 0 ] && echo "no pairwise merge-base resolves between them" || echo "no pairwise merge-base carries the ledger" ). DA003's bound is anchored on that base; without it the rule would silently weaken to one a merge can satisfy by duplicating all of common history. Refusing rather than reporting a vacuous pass." >&2
		echo "check-decisions-append: this is reachable honestly — two histories that each seeded their own $LEDGER (two abcd repositories joined with --allow-unrelated-histories, or the ledger created independently on two branches) share no ledger history for a bound to rest on. It is not a dead end, and it is not something to work around by crafting the merge differently. Either seed the ledger once on the default branch and branch from there, so every side descends from one ledger history; or, if the join itself is the intent, treat it as the deliberate gate-edit-and-review change this gate's header describes for redaction and graduation — the pull request that performs the join also adjusts or retires this gate in the same diff, and a human reviews both halves together." >&2
		exit 2
	fi

	# Resolve the per-line maximum across every contributing base.
	if [ "$ledger_bases" -gt 0 ]; then
		awk '
			{
				i = index($0, "\t")
				c = substr($0, 1, i - 1) + 0
				t = substr($0, i + 1)
				if (c > m[t]) m[t] = c
			}
			END { for (t in m) printf "%d\t%s\n", m[t], t }
		' "$tmpb" >"$basetmp"
	fi

	# Each parent contributes only what it ADDED over the base — max(0, p − base)
	# per line, accumulated across parents as "<delta><TAB><line>" records. Split
	# on the FIRST tab when read back, so a ledger line containing tabs survives.
	for p in $lparents; do
		git show "$p:$LEDGER" >"$tmpa" 2>/dev/null || rc=$?
		if [ "$rc" -ne 0 ]; then
			echo "check-decisions-append: reading $LEDGER at $p failed (exit $rc) — refusing rather than reporting a vacuous pass." >&2
			exit 2
		fi
		awk -v bf="$basetmp" '
			BEGIN {
				while ((getline line < bf) > 0) {
					i = index(line, "\t")
					base[substr(line, i + 1)] = substr(line, 1, i - 1) + 0
				}
			}
			{ mine[$0]++ }
			END {
				for (t in mine) {
					d = mine[t] - (base[t] + 0)
					if (d > 0) printf "%d\t%s\n", d, t
				}
			}
		' "$tmpa" >>"$allowtmp"
	done

	# The header region of the MERGE's own copy: prose about the ledger, exempt
	# like everywhere else.
	local shape header_end
	shape="$(ledger_shape "$commit")"
	header_end="${shape##*$'\t'}"

	local authored
	rc=0
	authored="$(git show "$commit:$LEDGER" 2>/dev/null | awk -v bf="$basetmp" -v af="$allowtmp" -v header_end="$header_end" -v item_re="$ITEM_RE" '
		BEGIN {
			# basetmp is "<count><TAB><line>" — the per-line MAXIMUM already
			# resolved across the pairwise bases — not one line per occurrence.
			while ((getline line < bf) > 0) {
				i = index(line, "\t")
				base[substr(line, i + 1)] = substr(line, 1, i - 1) + 0
			}
			while ((getline line < af) > 0) {
				i = index(line, "\t")
				allow[substr(line, i + 1)] += substr(line, 1, i - 1) + 0
			}
		}
		{
			# In scope: anything below the header, plus any list-shaped line
			# wherever it sits — a forged entry does not become header prose by
			# being planted above the first well-formed bullet.
			if (FNR > header_end || $0 ~ item_re) {
				scoped[$0]++
				if (scoped[$0] == 1) first_at[$0] = FNR
			}
		}
		END {
			for (t in scoped) {
				bound = (base[t] + 0) + (allow[t] + 0)
				if (scoped[t] > bound)
					printf "%d\t%d\t%d\t%s\n", first_at[t], scoped[t], bound, t
			}
		}
	' | sort -n)" || rc=$?
	if [ "$rc" -ne 0 ]; then
		echo "check-decisions-append: reading $LEDGER at $commit failed (exit $rc) — refusing rather than reporting a vacuous pass." >&2
		exit 2
	fi
	[ -n "$authored" ] || return 0

	# Read the first record and the count WITHOUT piping $authored into a reader
	# that stops early. A full-log splice produces thousands of findings, and
	# `printf … | head -1` then kills printf with SIGPIPE — which pipefail turns
	# into the script's own exit status, so the gate died at 141 instead of
	# reporting the very forgery it had just detected. Parameter expansion for the
	# first record; awk (which consumes all of its input) for the count.
	local distinct first_rec first_line got want first_sample
	first_rec="${authored%%$'\n'*}"
	distinct="$(awk 'END { print NR }' <<<"$authored")"
	IFS=$'\t' read -r first_line got want first_sample <<<"$first_rec" || true
	first_sample="$(termsafe "$first_sample")"
	fail "DA003 merge commit ${commit:0:12} (${subject}) holds ${distinct} distinct line(s) in $LEDGER more times than the merge base held them plus what the parents added — a union merge concatenates both sides and neither invents nor duplicates, so this content was authored in the conflict resolution. First such line appears ${got} time(s) in the merge against a bound of ${want} (line ${first_line} of the merge's copy): ${first_sample}"
}

# has_nul reports whether the ledger at a tree-ish contains a NUL byte. Compares
# the blob's real size against its size with NULs removed, because there is no
# portable grep for a NUL and a shell variable cannot hold one.
has_nul() {
	local ref="$1" size stripped
	git cat-file -e "$ref:$LEDGER" 2>/dev/null || return 1
	size="$(git cat-file -s "$ref:$LEDGER" 2>/dev/null)" || return 1
	stripped="$(git show "$ref:$LEDGER" 2>/dev/null | LC_ALL=C tr -d '\000' | wc -c | tr -d '[:space:]')" || return 1
	[ "$stripped" -lt "$size" ]
}

# check_nul implements DA004: the ledger is prose, and a NUL byte in it is never
# honest. Fires only on the commit that INTRODUCES one — a NUL already in a
# parent is history, and demanding its removal would deadlock against DA002,
# which refuses the edit that would remove it.
#
# This is defence in depth behind --text, not a substitute for it: --text is what
# keeps every rule reading hunks, and it also covers the `-diff` attribute, which
# has no NUL to find. DA004 covers what --text does not — a NUL breaks the rest
# of the toolchain (grep, the site renderer, an editor) long before it reaches
# this gate, and refusing it here is how it never gets committed.
check_nul() {
	local commit="$1" subject="$2" parents="$3" p
	has_nul "$commit" || return 0
	for p in $parents; do
		# A parent already carrying a NUL makes this commit an inheritor, not the
		# author of the fault.
		has_nul "$p" && return 0
	done
	fail "DA004 commit ${commit:0:12} (${subject}) introduces a NUL byte into $LEDGER. The ledger is prose; a NUL is never honest content, and it makes git treat the file as binary — every hunk-reading rule here would see nothing at all."
}

check_commits() {
	local base="$1" head="$2"
	local rc=0 range
	# Merges are NOT excluded. A hand-forged conflict resolution is authored
	# content that appears in no single-parent commit, so skipping merges would
	# leave an unchecked write path into the ledger.
	range="$(git rev-list "$base".."$head" 2>&1)" || rc=$?
	if [ "$rc" -ne 0 ]; then
		echo "check-decisions-append: git rev-list failed for $base..$head (exit $rc) — refusing rather than reporting a vacuous pass:" >&2
		echo "$range" >&2
		exit 2
	fi
	if [ -z "$range" ]; then
		echo "check-decisions-append: no commits in $base..$head — nothing to check"
		return 0
	fi
	local checked=0
	while IFS= read -r sha; do
		[ -n "$sha" ] || continue
		check_commit "$sha"
		checked=$((checked + 1))
	done <<<"$range"
	echo "check-decisions-append: DA001-DA004 checked $checked commit(s) in $base..$head"
}

case "${1:-}" in
commits)
	[ $# -eq 3 ] || usage
	check_commits "$2" "$3"
	;;
*)
	usage
	;;
esac

if [ "$violations" -gt 0 ]; then
	printf 'check-decisions-append: FAILED — %d violation(s)\n' "$violations" >&2
	exit 1
fi
echo "check-decisions-append: OK"
