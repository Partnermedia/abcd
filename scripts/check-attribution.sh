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
#
# The optional [_*]{1,3} admits the MARKDOWN-EMPHASIS run the footer is actually
# appended inside — italic (1), bold (2) or bold-italic (3). The first cut matched
# leading whitespace only, so the precise form that motivated this gate went
# straight through the body check; it was found live on two pull requests whose
# failing leg happened to be something else (iss-262).
#
# The emoji group is repeated on BOTH sides of the emphasis run because a footer
# may put the robot inside the italics or before them, and neither order is more
# natural than the other. At most one of the two can match a given line; a single
# group in one position is exactly the gap that let `🤖 _Generated with [` through.
#
# The marker must be ATTACHED to the word, with no space between, because that is
# how markdown itself reads it: `*text*` is emphasis, `* text` is a list item. A
# bullet describing the banned footer is documentation — the same writable-about
# case the anchor already protects for `- Generated with …` — so requiring
# attachment is what keeps the widening from swallowing list items.
#
# The class admits mixed runs (`_*`, `*_`, `__*`) that nobody writes. That is
# over-REJECTION of text no one produces, which costs nothing; pinning the class
# to well-formed markdown would buy nothing back.
GENERATED_RE='^[[:space:]]*(🤖[[:space:]]*)?([_*]{1,3})?(🤖[[:space:]]*)?[Gg]enerated (with|by) \['

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

# strip_fenced_blocks removes markdown fenced code blocks from stdin.
#
# A fenced block is QUOTED MATERIAL, not the document's own voice, and every rule
# below is about what the document itself says. So the rule reads in both
# directions: a banned form inside a fence is an example rather than a violation,
# and the required trailer inside a fence is an example too — it cannot serve as
# the disclosure. Stripping before any check gets both at once.
#
# This exists because mid-sentence prose was the only way to document the banned
# shape, and a fenced example — the natural way to show a literal — was refused
# exactly as a real footer would be. The change that tightened the rule tripped
# over this in its own pull-request body (iss-268). The line anchor already grants
# this to `- Generated with …` bullets; a fence is the same concession, made
# deliberately.
#
# APPLIED TO THE PULL-REQUEST BODY ONLY. The whole justification is a RENDERING
# argument — a fence reads as an example because GitHub draws it as a code block.
# A commit message is never rendered as markdown: `git log` and the forge's commit
# view show it verbatim, so a fenced footer there is displayed exactly as a real
# one and lands in permanent history. There is no reading under which it is
# quotation, so the commits arm keeps the unstripped text and the mid-sentence
# convention (which every commit message in this repo's history already uses —
# not one contains a fence marker at all).
#
# Permitting a fenced footer in a body costs little against the actual threat.
# This gate defends against a footer a TOOL APPENDS BY DEFAULT, and no tool we
# have met wraps its own footer in a code fence — a judgement about observed
# tools, not a law about all of them, and the line to revisit if one appears. It
# was never a defence against a determined author, who would simply delete the
# footer rather than fence it.
#
# THE STRIPPED REGION MUST BE A SUBSET OF WHAT MARKDOWN RENDERS AS CODE, or the
# concession stops being safe: a line the gate skips but the forge renders as an
# ordinary paragraph is a real footer, silently passed. A presence test on the
# first three characters is NOT that subset — it is blind to marker type, run
# length and info string, and each blindness is a working bypass:
#
#   ~~~ / ``` / ~~~ / footer / ```   — CommonMark closes a tilde block only with
#                                  tildes, so the footer is a paragraph; a
#                                  type-blind toggle pairs 1-2 and 3-4 and strips it.
#   ```` / ``` / ```` / footer / ```  — a closer must be at least as long as its opener.
#   ```make preflight``` is the gate.
#   footer
#   ```make build``` builds it. — no exotic markdown at all: a backtick fence's info
#                                  string may not contain a backtick, so these are
#                                  inline code spans in ONE paragraph and the footer
#                                  renders in full.
#
# So the matcher is pairing-aware. An opener records its character and run length;
# a line closes it only with the SAME character, AT LEAST as many of them, and
# nothing but whitespace after the run; a backtick opener whose info string
# carries a backtick is not an opener at all.
#
# UNTERMINATED BLOCKS FAIL CLOSED. Markdown renders an unclosed fence as code to
# the end of the document, so honouring one would let a single stray line suppress
# every check beneath it — the cheapest bypass imaginable. Ending the scan still
# inside a block strips nothing. Tracking the opener rather than counting markers
# also makes NESTED documentation work: a ```` block quoting a ``` example is one
# block, where a parity count saw three markers and failed closed on the very
# shape needed to document this rule.
#
# A fence may be indented up to three spaces (markdown's own rule); at four it is
# an indented code block, and a tab-led line is not a fence either. Both stay
# unstripped, which is the conservative direction. A fence nested inside a list
# item is indented past three and so is NOT stripped — iss-270 carries that.
strip_fenced_blocks() {
	awk '
		# fenceparse fills f[] with the marker character, its run length and the
		# text after the run, or returns 0 when the line cannot open or close a
		# fence at all.
		function fenceparse(l, f,   t, indent, c, n) {
			t = l
			sub(/^ +/, "", t)
			indent = length(l) - length(t)
			if (indent > 3) return 0
			c = substr(t, 1, 1)
			if (c != "`" && c != "~") return 0
			n = 0
			while (substr(t, n + 1, 1) == c) n++
			if (n < 3) return 0
			f["char"] = c
			f["len"] = n
			f["rest"] = substr(t, n + 1)
			return 1
		}
		function blank(s) { gsub(/[ \t\r]/, "", s); return s == "" }
		{ line[NR] = $0 }
		END {
			inside = 0
			for (i = 1; i <= NR; i++) {
				if (!inside) {
					if (!fenceparse(line[i], f)) continue
					# A backtick opener may not carry a backtick in its info
					# string; such a line is an inline code span, not a fence.
					if (f["char"] == "`" && index(f["rest"], "`") > 0) continue
					inside = 1
					openchar = f["char"]
					openlen = f["len"]
					drop[i] = 1
					continue
				}
				drop[i] = 1
				if (fenceparse(line[i], f) && f["char"] == openchar &&
					f["len"] >= openlen && blank(f["rest"])) inside = 0
			}
			# Unterminated: strip nothing.
			if (inside) {
				for (i = 1; i <= NR; i++) print line[i]
				exit
			}
			for (i = 1; i <= NR; i++) if (!drop[i]) print line[i]
		}
	'
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
# absent, and the trailer must be present. The caller decides whether fenced
# quotation was removed first: the body arm strips, the commits arm does not (see
# strip_fenced_blocks).
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
	# The body is markdown a forge renders, so a fenced block reads as an example.
	check_text "the pull-request body" "$(strip_fenced_blocks <"$2")"
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
