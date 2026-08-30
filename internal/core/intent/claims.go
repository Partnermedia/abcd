package intent

// claims.go is the intent record's claim reader (spc-55). An intent carries up
// to three kinds of claim — criterion, mechanism, context — and the
// claim-recording-gradient discipline (itd-190) gives each its own recording
// requirement at the readiness gate. The criterion claim already has a parser
// (countAcceptanceCriteria); this file adds the other two, as ONE reader every
// consumer shares — the readiness gate, `Plan`'s identity stamp, and the
// `--json` render — on the countAcceptanceCriteria precedent, so no two gates
// can disagree about what a claim section says.
//
// It owns no notion of a section or a bullet of its own: sectionLineRange is the
// single bound both this file and audit.go's sectionBody read a section through,
// and headingRe/bulletRe are audit.go's.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/intentdriven/abcd/internal/core/recordid"
)

// NullityToken is the one exact spelling that records a claim as considered and
// declined: alone on its line under the section heading, byte for byte. Prose
// that merely contains the word "none" is a stated claim — the whole point of an
// exact token is that a gate can tell a declined claim from a discussed one.
const NullityToken = "None stated."

// condFamily is the record-id family the per-condition identity is minted in.
// It is deliberately NOT a record store: the marker identifies a claim INSIDE a
// record, so `abcd <id>` dispatch is untouched.
const condFamily = "cond"

// condMintAttempts bounds the redraw loop that separates two conditions minted
// in the same second. The mint reads no ledger (adr-45), so a suffix
// coincidence within one stamping pass is this caller's to resolve.
const condMintAttempts = 100

// ClaimState is the byte state of a claim section — the three states the
// gradient refuses to collapse, plus the ordinary one. An ABSENT section is a
// claim not carried; an EMPTY section is a gate fault; the NULLITY token is a
// claim considered and declined.
type ClaimState string

const (
	ClaimAbsent  ClaimState = "absent"
	ClaimEmpty   ClaimState = "empty"
	ClaimNullity ClaimState = "nullity"
	ClaimStated  ClaimState = "stated"
)

var (
	// mechanismHeadingRe matches the `## Mechanism` heading (any heading depth).
	mechanismHeadingRe = regexp.MustCompile(`^#{1,6}\s+Mechanism\s*$`)
	// scopeHeadingRe matches the `## Scope Conditions` heading (any depth).
	scopeHeadingRe = regexp.MustCompile(`^#{1,6}\s+Scope Conditions\s*$`)
	// condMarkerRe matches the identity marker ANYWHERE inside a condition bullet
	// — first physical line or a folded continuation line. The HTML-comment form
	// is this repository's machine-marker idiom (see audit.go's
	// `<!-- abcd-review: … -->`): invisible once rendered, so the bullet's prose
	// stays exactly what a human wrote. It is deliberately not anchored to the end
	// of a line: an editor that rewraps an 80-column bullet moves the marker, and
	// a positional read would orphan every disposition keyed on it and then mint a
	// second identity for the same condition (iss-2608300235377731).
	condMarkerRe = regexp.MustCompile(`<!-- cond: (cond-[0-9]{16}) -->`)
	// spaceRunRe collapses the whitespace left behind when a marker is excised
	// from the middle of a line, and by folding a wrapped bullet into one string.
	spaceRunRe = regexp.MustCompile(`\s+`)
	// bulletPrefixRe strips a top-level list marker to leave the bullet's prose.
	bulletPrefixRe = regexp.MustCompile(`^[-*]\s+`)
	// fenceRe matches a fenced-code-block delimiter: a run of three or more
	// backticks or tildes at up to three spaces of indent (CommonMark).
	fenceRe = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})(.*)$")
	// malformedMarkerRe matches the opening of something SHAPED like an identity
	// marker. Anything it matches that condMarkerRe did not is a hand-typed
	// near-miss: prose to a parser, an identity to the human who wrote it.
	malformedMarkerRe = regexp.MustCompile(`(?i)<!--\s*cond\s*:`)
	// anyBulletRe matches a list item at ANY indent. A line that is a bullet ends
	// the preceding bullet, so an indented sub-bullet is never folded into its
	// parent's text.
	anyBulletRe = regexp.MustCompile(`^\s*[-*]\s+\S`)
)

// ScopeCondition is one recorded context claim: the condition's prose and the
// minted identity that outlives a rewrite of it. ID is empty for a bullet no
// `Plan` run has stamped yet — an unstamped condition is a gate finding, never
// silently repaired at read time.
type ScopeCondition struct {
	Ordinal int    `json:"ordinal"` // 1-based position in the section
	ID      string `json:"id"`      // cond-<16 digits>, empty when unmarked
	Text    string `json:"text"`    // the bullet's prose, marker removed
	// ExtraIDs holds every well-formed marker after the first. A bullet carrying
	// two identities is an ambiguity the gate names: a disposition keyed on one of
	// them would attach to whichever the reader reached first.
	ExtraIDs []string `json:"extra_ids,omitempty"`
	// MalformedMarker reports a `<!-- cond:` in the bullet that is not a
	// well-formed identity. Left unread it is silent prose, so the stamp glues a
	// real marker beside it and the bullet ends up carrying two things that look
	// like identities to a human and one to the machine.
	MalformedMarker bool `json:"malformed_marker,omitempty"`
}

// Claims is what one intent record says about its mechanism and context claims.
//
// The two Prompt flags mark a section still holding the create-path scaffold:
// bytes that ASK for a claim. They read as ClaimStated because that is what
// they are on disk, and the flag is what stops a gate reporting an unanswered
// question as an answer (iss-2608300210588414).
type Claims struct {
	Mechanism        ClaimState       `json:"mechanism"`
	MechanismPrompt  bool             `json:"mechanism_prompt"`
	ConditionsState  ClaimState       `json:"conditions_state"`
	ConditionsPrompt bool             `json:"conditions_prompt"`
	Conditions       []ScopeCondition `json:"conditions"` // populated only when stated
	// ConditionsFenced and ConditionsDuplicated are structural faults in the
	// `## Scope Conditions` section that make its bullets unsafe to WRITE: a
	// fenced example a stamp could land inside, and a second heading that makes
	// "the section" ambiguous. Both are reported by the gate and refused by the
	// stamp, so the remedy the gate names is never a command that cannot run.
	ConditionsFenced     bool `json:"conditions_fenced,omitempty"`
	ConditionsCommented  bool `json:"conditions_commented,omitempty"`
	ConditionsDuplicated bool `json:"conditions_duplicated,omitempty"`
}

// ParseClaims reads an intent record's `## Mechanism` and `## Scope Conditions`
// sections. It is the single reader for both: every downstream consumer asks it
// rather than re-deciding what a section, a bullet, or the nullity token is.
func ParseClaims(content string) Claims {
	lines := strings.Split(content, "\n")
	mask := maskLines(lines)
	c := Claims{
		Mechanism:        claimState(lines, mechanismHeadingRe),
		MechanismPrompt:  sectionIsPrompt(lines, mechanismHeadingRe),
		ConditionsState:  claimState(lines, scopeHeadingRe),
		ConditionsPrompt: sectionIsPrompt(lines, scopeHeadingRe),
		Conditions:       []ScopeCondition{},
	}
	if start, end, ok := sectionLineRangeIn(lines, mask, scopeHeadingRe); ok {
		c.ConditionsFenced = anyMasked(mask, start, end, maskFence)
		c.ConditionsCommented = anyMasked(mask, start, end, maskComment)
		c.ConditionsDuplicated = countHeadings(lines, mask, scopeHeadingRe) > 1
	}
	if c.ConditionsState == ClaimStated {
		c.Conditions = parseConditions(lines)
	}
	return c
}

// Mask flags: why a line is not live markdown. Both are read the same way — the
// line is neither matched nor written — but they are reported separately, so a
// refusal names the thing the reader has to go and look at.
const (
	maskFence uint8 = 1 << iota
	maskComment
)

// maskLines reports, per line, whether that line lies inside a fenced code block
// or an HTML comment span — the delimiter lines included, since nothing on them
// is live markdown either. A parser that cannot see either reads an EXAMPLE, or
// a deliberately PARKED bullet, as an instruction: a fenced or commented
// `## Scope Conditions` shadows the record's real one, and a bullet inside
// either is counted as a condition and written into by the stamp. Writing a
// marker inside a comment is the worst of the three — the marker's own `-->`
// closes the comment early, so the rest of the parked block starts rendering
// (iss-2608300235388164, iss-2608300259316871).
//
// Neither construct nests in the other: inside a fence `<!--` is literal text,
// and inside a comment a fence delimiter is. An unclosed opener of either kind
// runs to end of file — CommonMark's rule for a fence, and the only safe reading
// of a comment nobody closed.
func maskLines(lines []string) []uint8 {
	mask := make([]uint8, len(lines))
	fenceOpen := ""
	inComment := false
	for i, raw := range lines {
		ln := strings.TrimRight(raw, "\r")
		switch {
		case inComment:
			mask[i] |= maskComment
			// Inside a comment every byte is literal, so the raw first `-->` is the
			// closer; the remainder of the line is live markdown again and may open
			// a fresh span.
			if k := strings.Index(ln, "-->"); k >= 0 {
				inComment = opensCommentFrom(ln, k+len("-->"))
			}
		case fenceOpen != "":
			mask[i] |= maskFence
			if m := fenceRe.FindStringSubmatch(ln); m != nil && m[1][0] == fenceOpen[0] && len(m[1]) >= len(fenceOpen) && strings.TrimSpace(m[2]) == "" {
				fenceOpen = ""
			}
		default:
			// A backtick opener's info string may not itself contain a backtick.
			if m := fenceRe.FindStringSubmatch(ln); m != nil && !(m[1][0] == '`' && strings.Contains(m[2], "`")) {
				fenceOpen = m[1]
				mask[i] |= maskFence
				continue
			}
			if opensComment(ln) {
				inComment = true
				mask[i] |= maskComment
			}
		}
	}
	return mask
}

// opensComment reports whether a line leaves an HTML comment open. It walks the
// line left to right and lets the FIRST construct win, which is CommonMark's own
// precedence: at a backtick run it skips a matched code span whole (or treats an
// unmatched run as literal backticks), and at `<!--` it skips to the `-->` that
// closes it — consuming any backticks in between, because inside a comment they
// are literal text.
//
// Resolving code spans first, as this did, diverged in both directions: a
// backtick inside a LIVE comment re-paired the rest of the line, and a comment
// opener quoted in backticks read as live. One cursor closes both
// (iss-2608300320418618, iss-2608300335369473). A well-formed identity marker
// closes on its own line, so it never opens a span; a genuinely unclosed opener
// masks to end of file.
func opensComment(ln string) bool { return opensCommentFrom(ln, 0) }

// opensCommentFrom is opensComment beginning at an offset — the form maskLines
// needs when a line closes the span it was already inside and the remainder is
// live markdown again.
func opensCommentFrom(ln string, start int) bool {
	for i := start; i < len(ln); {
		switch {
		case ln[i] == '`':
			j := backtickRunEnd(ln, i)
			if _, end, ok := findBacktickRun(ln, j, j-i); ok {
				i = end
			} else {
				i = j
			}
		case strings.HasPrefix(ln[i:], "<!--"):
			rest := ln[i+len("<!--"):]
			k := strings.Index(rest, "-->")
			if k < 0 {
				return true
			}
			i += len("<!--") + k + len("-->")
		default:
			i++
		}
	}
	return false
}

// masked reports whether a line is not live markdown, for any reason.
func masked(mask []uint8, i int) bool { return i < len(mask) && mask[i] != 0 }

// anyMasked reports whether any line of [start, end) carries the given flag.
func anyMasked(mask []uint8, start, end int, flag uint8) bool {
	for i := start; i < end && i < len(mask); i++ {
		if mask[i]&flag != 0 {
			return true
		}
	}
	return false
}

// countHeadings counts the live headings matching headRe.
func countHeadings(lines []string, mask []uint8, headRe *regexp.Regexp) int {
	n := 0
	for i, ln := range lines {
		if !masked(mask, i) && headRe.MatchString(strings.TrimRight(ln, "\r")) {
			n++
		}
	}
	return n
}

// claimState classifies one claim section into its byte state. A heading with
// nothing but blank lines under it is EMPTY; a heading whose body is exactly the
// nullity token is NULLITY; anything else written there is STATED.
func claimState(lines []string, headRe *regexp.Regexp) ClaimState {
	start, end, ok := sectionLineRange(lines, headRe)
	if !ok {
		return ClaimAbsent
	}
	var nonBlank []string
	for _, ln := range lines[start:end] {
		if strings.TrimSpace(ln) != "" {
			nonBlank = append(nonBlank, strings.TrimRight(ln, "\r"))
		}
	}
	switch {
	case len(nonBlank) == 0:
		return ClaimEmpty
	case len(nonBlank) == 1 && nonBlank[0] == NullityToken:
		return ClaimNullity
	default:
		return ClaimStated
	}
}

// sectionIsPrompt reports whether a claim section holds nothing but the
// create-path scaffold, asking create.go — which owns the templates and is the
// only place the answer can stay true when they are reworded.
func sectionIsPrompt(lines []string, headRe *regexp.Regexp) bool {
	start, end, ok := sectionLineRange(lines, headRe)
	if !ok {
		return false
	}
	return IsClaimPrompt(strings.Join(lines[start:end], "\n"))
}

// parseConditions enumerates the top-level bullets of `## Scope Conditions` in
// order, each with its identity marker (wherever in the bullet it sits) and its
// prose. Continuation lines — the wrap of a long bullet — are folded into the
// prose; an indented sub-bullet is detail of its parent and ends it, exactly as
// bulletRe already rules for acceptance criteria.
func parseConditions(lines []string) []ScopeCondition {
	mask := maskLines(lines)
	start, end, ok := sectionLineRangeIn(lines, mask, scopeHeadingRe)
	if !ok {
		return []ScopeCondition{}
	}
	conds := []ScopeCondition{}
	for _, b := range conditionBlocks(lines, mask, start, end) {
		ids, malformed, text := readConditionBlock(lines, b)
		c := ScopeCondition{Ordinal: len(conds) + 1, Text: text, MalformedMarker: malformed}
		if len(ids) > 0 {
			c.ID, c.ExtraIDs = ids[0], ids[1:]
		}
		conds = append(conds, c)
	}
	return conds
}

// conditionBlock is one top-level bullet and the continuation lines folded into
// it, as the half-open line range [start, end).
type conditionBlock struct{ start, end int }

// conditionBlocks splits a section body into its top-level bullets. It is the
// single notion of where a condition starts and stops, so the reader and the
// stamper can never disagree about which lines belong to which condition. A
// masked line — fenced or commented — neither opens a bullet nor continues one.
func conditionBlocks(lines []string, mask []uint8, start, end int) []conditionBlock {
	var blocks []conditionBlock
	for i := start; i < end; i++ {
		if masked(mask, i) || !bulletRe.MatchString(strings.TrimRight(lines[i], "\r")) {
			continue
		}
		j := i + 1
		for ; j < end; j++ {
			cont := strings.TrimRight(lines[j], "\r")
			if masked(mask, j) || strings.TrimSpace(cont) == "" || anyBulletRe.MatchString(cont) {
				break
			}
		}
		blocks = append(blocks, conditionBlock{start: i, end: j})
	}
	return blocks
}

// readConditionBlock returns every well-formed identity marker in the bullet, in
// the order they appear, whether the bullet also carries a near-miss of one, and
// the bullet's prose with the well-formed markers excised. A malformed marker is
// left in the prose deliberately: the gate names it and the remedy is to delete
// it, so the reader has to be able to see what to delete.
func readConditionBlock(lines []string, b conditionBlock) (ids []string, malformed bool, text string) {
	var parts []string
	for i := b.start; i < b.end; i++ {
		ln := strings.TrimRight(lines[i], "\r")
		spans := codeSpanRanges(ln)
		var live []string
		for _, m := range condMarkerRe.FindAllStringSubmatchIndex(ln, -1) {
			// A marker quoted in backticks is DOCUMENTATION of the grammar, not an
			// identity — the gradient's own prose writes it that way. Reading it as
			// one hands a condition an id nobody minted.
			if inAnyRange(spans, m[0]) {
				malformed = true
				continue
			}
			ids = append(ids, ln[m[2]:m[3]])
			live = append(live, ln[m[0]:m[1]])
		}
		for _, marker := range live {
			ln = strings.Replace(ln, marker, " ", 1)
		}
		if malformedMarkerRe.MatchString(ln) {
			malformed = true
		}
		if i == b.start {
			ln = bulletPrefixRe.ReplaceAllString(ln, "")
		}
		parts = append(parts, ln)
	}
	return ids, malformed, strings.TrimSpace(spaceRunRe.ReplaceAllString(strings.Join(parts, " "), " "))
}

// codeSpanRanges returns the byte ranges of the inline code spans on one line:
// a run of backticks closed by a run of the same length (CommonMark).
func codeSpanRanges(ln string) [][2]int {
	var out [][2]int
	for i := 0; i < len(ln); {
		if ln[i] != '`' {
			i++
			continue
		}
		j := backtickRunEnd(ln, i)
		closeStart, closeEnd, ok := findBacktickRun(ln, j, j-i)
		if !ok {
			i = j
			continue
		}
		_ = closeStart
		out = append(out, [2]int{i, closeEnd})
		i = closeEnd
	}
	return out
}

// backtickRunEnd returns the index just past the run of backticks at i.
func backtickRunEnd(ln string, i int) int {
	for i < len(ln) && ln[i] == '`' {
		i++
	}
	return i
}

// findBacktickRun finds the next run of exactly n backticks at or after from.
func findBacktickRun(ln string, from, n int) (start, end int, ok bool) {
	for k := from; k < len(ln); {
		if ln[k] != '`' {
			k++
			continue
		}
		e := backtickRunEnd(ln, k)
		if e-k == n {
			return k, e, true
		}
		k = e
	}
	return 0, 0, false
}

// inAnyRange reports whether pos falls inside one of the ranges.
func inAnyRange(ranges [][2]int, pos int) bool {
	for _, r := range ranges {
		if pos >= r[0] && pos < r[1] {
			return true
		}
	}
	return false
}

// DuplicateConditionIDs returns the identity markers carried by more than one
// condition, in first-seen order. Two conditions sharing an identity is a fault
// the gate names rather than repairs: a disposition keyed on a duplicated
// marker would attach to whichever bullet the reader happened to reach first.
func DuplicateConditionIDs(conds []ScopeCondition) []string {
	seen := map[string]int{}
	var dupes []string
	for _, c := range conds {
		if c.ID == "" {
			continue
		}
		seen[c.ID]++
		if seen[c.ID] == 2 {
			dupes = append(dupes, c.ID)
		}
	}
	return dupes
}

// MultiplyMarkedConditions returns the 1-based positions of bullets carrying
// more than one identity.
func MultiplyMarkedConditions(conds []ScopeCondition) []int {
	var out []int
	for _, c := range conds {
		if len(c.ExtraIDs) > 0 {
			out = append(out, c.Ordinal)
		}
	}
	return out
}

// MalformedMarkerOrdinals returns the 1-based positions of bullets carrying a
// near-miss of an identity marker.
func MalformedMarkerOrdinals(conds []ScopeCondition) []int {
	var out []int
	for _, c := range conds {
		if c.MalformedMarker {
			out = append(out, c.Ordinal)
		}
	}
	return out
}

// UnmarkedConditionOrdinals returns the 1-based positions of the conditions no
// `Plan` run has stamped yet.
func UnmarkedConditionOrdinals(conds []ScopeCondition) []int {
	var out []int
	for _, c := range conds {
		if c.ID == "" {
			out = append(out, c.Ordinal)
		}
	}
	return out
}

// stampScopeConditions returns content with an identity marker appended to every
// unmarked top-level `## Scope Conditions` bullet, and the number it stamped. It
// is the ONLY writer of a condition identity — markers are never hand-typed —
// and it is idempotent: an already-stamped bullet is left byte-identical, so a
// re-run after an edit mints only for what is new. An absent section is a no-op.
//
// Identity lifecycle falls out of that rule with no diffing engine anywhere: an
// edit keeps its marker because the marker is bytes in the bullet; a split keeps
// the marker on the first part and mints for the second; a merge keeps the
// surviving marker and the retired one simply stops appearing.
func stampScopeConditions(content string, minter recordid.Minter) (string, int, error) {
	lines := strings.Split(content, "\n")
	mask := maskLines(lines)
	start, end, ok := sectionLineRangeIn(lines, mask, scopeHeadingRe)
	if !ok {
		return content, 0, nil
	}
	// Two structural faults make the section unsafe to WRITE, as opposed to
	// merely ambiguous to read, so the stamp refuses rather than guessing. Both
	// are reported by the readiness gate too, so the remedy it names is never a
	// command that cannot run (iss-2608300235388164).
	if countHeadings(lines, mask, scopeHeadingRe) > 1 {
		return "", 0, fmt.Errorf("intent: more than one '## Scope Conditions' heading; refusing to stamp an ambiguous section")
	}
	if anyMasked(mask, start, end, maskFence) {
		return "", 0, fmt.Errorf("intent: '## Scope Conditions' contains a fenced block; refusing to stamp a section whose bullets may be an example")
	}
	if anyMasked(mask, start, end, maskComment) {
		return "", 0, fmt.Errorf("intent: '## Scope Conditions' contains an HTML comment; refusing to stamp a section where a marker's own `-->` would close the comment early")
	}
	used := map[string]bool{}
	for _, c := range parseConditions(lines) {
		if c.ID != "" {
			used[c.ID] = true
		}
		for _, extra := range c.ExtraIDs {
			used[extra] = true
		}
	}
	stamped := 0
	for _, b := range conditionBlocks(lines, mask, start, end) {
		// Marker-free bullets only. A bullet that already carries an identity —
		// wherever in its wrapped text that identity sits — is never re-minted, and
		// one carrying a near-miss of a marker is left for the human to correct
		// rather than given a second thing that looks like an identity.
		ids, malformed, _ := readConditionBlock(lines, b)
		if len(ids) > 0 || malformed {
			continue
		}
		id, err := mintConditionID(minter, used)
		if err != nil {
			return "", 0, err
		}
		used[id] = true
		i := b.start
		trimmed := strings.TrimRight(lines[i], "\r")
		cr := ""
		if strings.HasSuffix(lines[i], "\r") {
			cr = "\r"
		}
		// Two trailing spaces are a markdown hard line break — content, not slack —
		// so the marker goes in FRONT of the trailing run rather than replacing it.
		body := strings.TrimRight(trimmed, " \t")
		lines[i] = body + " <!-- cond: " + id + " -->" + trimmed[len(body):] + cr
		stamped++
	}
	if stamped == 0 {
		return content, 0, nil
	}
	// The byte cap is enforced by writeIntentFile, at the write itself: the stamp
	// is only one of three steps that grow a record on the way to disk, and a
	// guard on one producer left the others free to cross the line.
	return strings.Join(lines, "\n"), stamped, nil
}

// mintConditionID draws a fresh identity, redrawing past one already used in
// this pass. Two bullets stamped in the same second differ only by the mint's
// random suffix, and a duplicate is exactly what the gate refuses — so the
// writer resolves the coincidence instead of emitting a record it would reject.
func mintConditionID(minter recordid.Minter, used map[string]bool) (string, error) {
	for attempt := 0; attempt < condMintAttempts; attempt++ {
		id, err := minter.Mint(condFamily)
		if err != nil {
			return "", fmt.Errorf("intent: minting a scope-condition identity: %w", err)
		}
		if !used[id] {
			return id, nil
		}
	}
	return "", fmt.Errorf("intent: minting a scope-condition identity: %d draws all collided", condMintAttempts)
}

// sectionLineRange returns the [start, end) line bounds of the body of the
// section introduced by the FIRST heading matching headRe — the body runs to the
// next heading of any depth, or to end of file. ok is false when no such heading
// exists, which is what separates an absent section from an empty one (a
// distinction sectionBody alone cannot make, since both read as "").
//
// This is the single notion of where a section starts and stops in this package:
// sectionBody is expressed on top of it, so the claim parse, the identity stamp
// and the acceptance-criteria count can never drift apart.
func sectionLineRange(lines []string, headRe *regexp.Regexp) (start, end int, ok bool) {
	return sectionLineRangeIn(lines, maskLines(lines), headRe)
}

// sectionLineRangeIn is sectionLineRange over a mask the caller already
// computed. A masked line — fenced or commented — is neither the heading that
// opens a section nor the heading that closes one.
func sectionLineRangeIn(lines []string, mask []uint8, headRe *regexp.Regexp) (start, end int, ok bool) {
	for i, ln := range lines {
		if masked(mask, i) || !headRe.MatchString(strings.TrimRight(ln, "\r")) {
			continue
		}
		end = len(lines)
		for j := i + 1; j < len(lines); j++ {
			if !masked(mask, j) && headingRe.MatchString(strings.TrimRight(lines[j], "\r")) {
				end = j
				break
			}
		}
		return i + 1, end, true
	}
	return 0, 0, false
}
