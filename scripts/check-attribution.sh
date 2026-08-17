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

# The human-only declaration. The convention is DISCLOSURE, and a change no AI
# touched has nothing to disclose — but silence cannot say that: an absent
# trailer and a forgotten trailer are the same bytes, which is why the gate
# refuses a bare omission. `Assisted-by: None` is the positive form, so a
# human-only artefact states its provenance as explicitly as an assisted one
# and stays auditable. It is deliberately the ONLY accepted non-vendor value —
# a free-text escape ("n/a", "human") would reopen the omission it closes.
#
# It is not a bypass to reach for when a trailer is merely inconvenient:
# claiming None for assisted work is a false disclosure, the exact failure this
# gate exists to prevent, and the reviewer reading the diff is the check on it.
NONE_RE='^Assisted-by: [Nn]one$'

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

# The git IDENTITY itself, not only the message. A commit authored AND committed
# as `Claude <noreply@anthropic.com>` carried a fully compliant message and
# sailed through the message-only gate — but the contributor graph is built from
# commit authorship plus Co-authored-by trailers, so it put an AI at #2 in the
# graph. Worse, a squash merge auto-appends a Co-authored-by for any branch
# author who is not the PR author, so one mis-identified branch commit inflates
# the graph again on every squash. Refusing the identity here closes both routes.
#
# Names are matched whole (a human named Claudette is not an AI identity) and
# case-insensitively; the address rule covers the vendor domains AI tools stamp
# by default. As with COAUTHOR_RE the intent is vendor-agnostic, but an identity
# ban can only enumerate — extend both lists as new defaults are met in the wild.
AI_IDENT_NAME_RE='^[[:space:]]*(claude|chatgpt|copilot|github copilot|gemini|codex|devin)[[:space:]]*$'
AI_IDENT_MAIL_RE='@anthropic\.com$|@openai\.com$'

fail=0
note() { echo "  $1" >&2; }

usage() {
	echo "usage: check-attribution.sh commits <base-ref> <head-ref> | body <file>" >&2
	exit 2
}

# check_ident refuses an AI git identity in one role (author or committer) of
# one commit; a human is the author of record, and the tool's disclosure lives
# in the trailer, never in the identity fields the contributor graph reads.
check_ident() {
	local label="$1" role="$2" name="$3" mail="$4"
	if printf '%s' "$name" | grep -Eiq "$AI_IDENT_NAME_RE" ||
		printf '%s' "$mail" | grep -Eiq "$AI_IDENT_MAIL_RE"; then
		echo "check-attribution: $label has an AI $role identity: $name <$mail>" >&2
		note "The human is the author of record (AGENTS.md); the contributor graph is built"
		note "from these identity fields, so an AI here asserts an authorship the tool does"
		note "not hold. Fix the commit identity (git commit --amend --reset-author with"
		note "user.name/user.email set to the human) and disclose the tool in the trailer:"
		note "Assisted-by: <Vendor>:<model-version>"
		fail=1
	fi
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
	if ! printf '%s' "$text" | grep -Eq "$TRAILER_RE" && ! printf '%s' "$text" | grep -Eq "$NONE_RE"; then
		echo "check-attribution: $label has no 'Assisted-by:' trailer" >&2
		note "Add a final line of the form: Assisted-by: <Vendor>:<model-version>"
		note "for example  Assisted-by: Claude:claude-opus-5  or  Assisted-by: Claude:claude-opus-5[1m]"
		note "If no AI assisted this artefact, disclose that positively instead:"
		note "  Assisted-by: None"
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
		label="commit ${sha:0:12} ($(git show -s --format='%s' "$sha" | cut -c1-50))"
		{
			IFS= read -r author_name
			IFS= read -r committer_name
			IFS= read -r committer_email
		} <<<"$(git show -s --format='%an%n%cn%n%ce' "$sha")"
		check_ident "$label" author "$author_name" "$author_email"
		check_ident "$label" committer "$committer_name" "$committer_email"
		check_text "$label" "$(git show -s --format='%B' "$sha")"
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
