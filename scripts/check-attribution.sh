#!/usr/bin/env bash
# Deterministic gate for abcd's AI-attribution convention (AGENTS.md § Attribution,
# CONTRIBUTING.md): an AI-assisted commit message and an AI-assisted pull-request
# body each carry the kernel-format trailer
#
#   Assisted-by: Claude:<model-version>
#
# and never `Co-Authored-By:` for an AI, and never a tool's own "Generated with
# <tool>" footer.
#
# Stopgap: the convention has lived as prose in AGENTS.md and CONTRIBUTING.md
# since the beginning and drifted anyway — itd-91 records a reconciliation sweep
# across 78 pull requests after PR bodies picked up a tool's default footer. Prose
# is not the missing piece; a check that fails closed is. itd-91 owns the general
# version (a per-repo declared preference every surface reads); this enforces
# abcd's own answer until that lands.
#
# Usage:
#   check-attribution.sh commits <base-ref> <head-ref>   # every non-merge commit in the range
#   check-attribution.sh body <file>                     # a pull-request body on stdin-file
#
# Exit 0 clean, 1 a violation, 2 a usage/environment fault.
set -euo pipefail

# The sanctioned trailer. Kernel format: a `Key: value` line, value naming the
# assisting model as <Vendor>:<model-version>. Anchored to line start so a mention
# inside prose is not a pass.
#
# The optional bracketed suffix admits a context-window variant such as
# claude-opus-5[1m]. That is an exact model identifier, not a typo — this repo's
# own history carries 33 commits using it and 5 using claude-opus-4-8[1m] — and
# the first cut of this gate rejected every one of them, so an agent reporting its
# precise model followed the prose and failed the check (iss-214). A bare vendor
# with no version stays refused: the convention asks for the version.
#
# The vendor half is deliberately NOT pinned to one name. The convention is
# <Vendor>:<model-version>, with Claude an example rather than the literal
# (iss-215).
TRAILER_RE='^Assisted-by: [A-Za-z][A-Za-z0-9._-]*:[A-Za-z0-9._-]+(\[[A-Za-z0-9._-]+\])?$'

# Footers a tool emits by default. These name a tool outside the two credit
# surfaces AGENTS.md sanctions (the README badge and ACKNOWLEDGEMENTS.md), which
# is why they are refused here rather than merely discouraged. Matched on the
# SHAPE — "generated with/by" followed by a markdown link, or the robot-emoji
# form — so a different tool's footer is caught as readily as the one that
# prompted this (iss-215).
#
# ANCHORED TO LINE START, and that is load-bearing: a real footer occupies its own
# line, while prose ABOUT the rule quotes it mid-sentence. Unanchored, this rejected
# the very commit that introduced it, and would reject the CHANGELOG entry, the
# issue records and the docs that describe what is banned. A gate that cannot be
# written about is a gate people route around.
GENERATED_RE='^[[:space:]]*(🤖[[:space:]]*)?[Gg]enerated (with|by) \['

# AI co-authorship, matched on the TRAILER KEY rather than on a vendor name: the
# objection AGENTS.md states — an authorship the tool does not hold, and an
# inflated contributor graph — is not specific to one provider, and a ban naming
# only Claude let `Co-authored-by: ChatGPT` through (iss-215).
#
# This refuses EVERY co-authorship trailer, including a genuinely human one. That
# over-reach is deliberate and worth stating: abcd defers DCO/Signed-off-by until
# the repo is public or takes its first outside contribution, so there is no human
# co-authorship case in flight today, and refusing is the fail-safe direction. If
# a real human co-author arrives, this is the line to revisit.
COAUTHOR_RE='^[[:space:]]*[Cc]o-[Aa]uthored-[Bb]y:'

fail=0
note() { echo "  $1" >&2; }

usage() {
	echo "usage: check-attribution.sh commits <base-ref> <head-ref> | body <file>" >&2
	exit 2
}

# check_text applies both halves to one artefact: the banned footer must be
# absent, and the trailer must be present.
check_text() {
	local label="$1" text="$2"
	if printf '%s' "$text" | grep -Eq "$GENERATED_RE"; then
		echo "check-attribution: $label carries a tool's default 'generated with' footer" >&2
		note "A 'Generated with <tool>' footer names a tool outside the two credit surfaces"
		note "AGENTS.md sanctions (the README badge and ACKNOWLEDGEMENTS.md). Replace it with"
		note "the trailer: Assisted-by: <Vendor>:<model-version>"
		fail=1
		return
	fi
	if printf '%s' "$text" | grep -Eq "$COAUTHOR_RE"; then
		echo "check-attribution: $label carries a 'Co-authored-by:' trailer" >&2
		note "abcd never uses Co-Authored-By: for AI — it asserts an authorship the tool does"
		note "not hold and inflates the contributor graph. Disclosure goes in the kernel"
		note "trailer instead: Assisted-by: <Vendor>:<model-version>"
		note "(This refuses human co-authorship too. abcd defers DCO until the repo is public"
		note "or takes its first outside contribution, so there is no such case today; if one"
		note "arrives, COAUTHOR_RE in this script is the line to revisit.)"
		fail=1
		return
	fi
	if ! printf '%s' "$text" | grep -Eq "$TRAILER_RE"; then
		echo "check-attribution: $label has no 'Assisted-by:' trailer" >&2
		note "Add a final line of the form: Assisted-by: <Vendor>:<model-version>"
		note "for example  Assisted-by: Claude:claude-opus-5  or  Assisted-by: Claude:claude-opus-5[1m]"
		note "(If this artefact had no AI assistance, say so in the PR and this gate can be"
		note "revisited — the convention is disclosure, and a human-only change has nothing to"
		note "disclose. Today the gate asks for the trailer on every non-bot contribution.)"
		fail=1
	fi
}

case "${1:-}" in
commits)
	[ $# -eq 3 ] || usage
	base="$2"
	head="$3"
	# --no-merges: a merge commit is generated by the forge, not authored here, and
	# carries no trailer of its own.
	range="$(git rev-list --no-merges "$base".."$head")"
	if [ -z "$range" ]; then
		echo "check-attribution: no non-merge commits in $base..$head — nothing to check"
		exit 0
	fi
	while IFS= read -r sha; do
		[ -n "$sha" ] || continue
		# A bot's commit is exempt: dependabot cannot write a trailer, and failing its
		# PRs would train the maintainer to discount a red gate on exactly the PRs where
		# a red gate most needs to be trusted.
		author_email="$(git show -s --format='%ae' "$sha")"
		case "$author_email" in
		*"[bot]@users.noreply.github.com" | *"@dependabot.com")
			continue
			;;
		esac
		check_text "commit ${sha:0:12} ($(git show -s --format='%s' "$sha" | cut -c1-50))" \
			"$(git show -s --format='%B' "$sha")"
	done <<<"$range"
	;;
body)
	[ $# -eq 2 ] || usage
	[ -f "$2" ] || {
		echo "check-attribution: no such file: $2" >&2
		exit 2
	}
	check_text "the pull-request body" "$(cat "$2")"
	;;
*)
	usage
	;;
esac

if [ "$fail" -ne 0 ]; then
	echo "check-attribution: FAILED — see AGENTS.md § Attribution and acknowledgements" >&2
	exit 1
fi
echo "check-attribution: clean"
