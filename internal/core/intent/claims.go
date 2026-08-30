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
	malformedMarkerRe = regexp.MustCompile(`<!--\s*cond\s*:`)
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
	ConditionsDuplicated bool `json:"conditions_duplicated,omitempty"`
}

// ParseClaims reads an intent record's `## Mechanism` and `## Scope Conditions`
// sections. It is the single reader for both: every downstream consumer asks it
// rather than re-deciding what a section, a bullet, or the nullity token is.
func ParseClaims(content string) Claims {
	lines := strings.Split(content, "\n")
	fenced := fenceMask(lines)
	c := Claims{
		Mechanism:        claimState(lines, mechanismHeadingRe),
		MechanismPrompt:  sectionIsPrompt(lines, mechanismHeadingRe),
		ConditionsState:  claimState(lines, scopeHeadingRe),
		ConditionsPrompt: sectionIsPrompt(lines, scopeHeadingRe),
		Conditions:       []ScopeCondition{},
	}
	if start, end, ok := sectionLineRangeIn(lines, fenced, scopeHeadingRe); ok {
		c.ConditionsFenced = anyFenced(fenced, start, end)
		c.ConditionsDuplicated = countHeadings(lines, fenced, scopeHeadingRe) > 1
	}
	if c.ConditionsState == ClaimStated {
		c.Conditions = parseConditions(lines)
	}
	return c
}

// fenceMask reports, per line, whether that line lies inside a fenced code
// block — the delimiter lines included, since nothing on them is markdown
// either. A markdown parser that cannot see a fence reads an EXAMPLE as an
// instruction: a fenced `## Scope Conditions` shadows the record's real one, and
// a fenced bullet is counted as a condition and written into by the stamp
// (iss-2608300235388164). An unclosed fence runs to end of file, per CommonMark.
func fenceMask(lines []string) []bool {
	mask := make([]bool, len(lines))
	open := ""
	for i, raw := range lines {
		ln := strings.TrimRight(raw, "\r")
		m := fenceRe.FindStringSubmatch(ln)
		if open == "" {
			// A backtick opener's info string may not itself contain a backtick.
			if m != nil && !(m[1][0] == '`' && strings.Contains(m[2], "`")) {
				open = m[1]
				mask[i] = true
			}
			continue
		}
		mask[i] = true
		if m != nil && m[1][0] == open[0] && len(m[1]) >= len(open) && strings.TrimSpace(m[2]) == "" {
			open = ""
		}
	}
	return mask
}

// anyFenced reports whether any line of [start, end) lies inside a fence.
func anyFenced(fenced []bool, start, end int) bool {
	for i := start; i < end && i < len(fenced); i++ {
		if fenced[i] {
			return true
		}
	}
	return false
}

// countHeadings counts the unfenced headings matching headRe.
func countHeadings(lines []string, fenced []bool, headRe *regexp.Regexp) int {
	n := 0
	for i, ln := range lines {
		if !fenced[i] && headRe.MatchString(strings.TrimRight(ln, "\r")) {
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
	fenced := fenceMask(lines)
	start, end, ok := sectionLineRangeIn(lines, fenced, scopeHeadingRe)
	if !ok {
		return []ScopeCondition{}
	}
	conds := []ScopeCondition{}
	for _, b := range conditionBlocks(lines, fenced, start, end) {
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
// fenced line neither opens a bullet nor continues one.
func conditionBlocks(lines []string, fenced []bool, start, end int) []conditionBlock {
	var blocks []conditionBlock
	for i := start; i < end; i++ {
		if fenced[i] || !bulletRe.MatchString(strings.TrimRight(lines[i], "\r")) {
			continue
		}
		j := i + 1
		for ; j < end; j++ {
			cont := strings.TrimRight(lines[j], "\r")
			if fenced[j] || strings.TrimSpace(cont) == "" || anyBulletRe.MatchString(cont) {
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
		for _, m := range condMarkerRe.FindAllStringSubmatch(ln, -1) {
			ids = append(ids, m[1])
		}
		ln = condMarkerRe.ReplaceAllString(ln, " ")
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
	fenced := fenceMask(lines)
	start, end, ok := sectionLineRangeIn(lines, fenced, scopeHeadingRe)
	if !ok {
		return content, 0, nil
	}
	// Two structural faults make the section unsafe to WRITE, as opposed to
	// merely ambiguous to read, so the stamp refuses rather than guessing. Both
	// are reported by the readiness gate too, so the remedy it names is never a
	// command that cannot run (iss-2608300235388164).
	if countHeadings(lines, fenced, scopeHeadingRe) > 1 {
		return "", 0, fmt.Errorf("intent: more than one '## Scope Conditions' heading; refusing to stamp an ambiguous section")
	}
	if anyFenced(fenced, start, end) {
		return "", 0, fmt.Errorf("intent: '## Scope Conditions' contains a fenced block; refusing to stamp a section whose bullets may be an example")
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
	for _, b := range conditionBlocks(lines, fenced, start, end) {
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
		lines[i] = strings.TrimRight(trimmed, " \t") + " <!-- cond: " + id + " -->" + cr
		stamped++
	}
	if stamped == 0 {
		return content, 0, nil
	}
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
	return sectionLineRangeIn(lines, fenceMask(lines), headRe)
}

// sectionLineRangeIn is sectionLineRange over a fence mask the caller already
// computed. A fenced line is neither the heading that opens a section nor the
// heading that closes one.
func sectionLineRangeIn(lines []string, fenced []bool, headRe *regexp.Regexp) (start, end int, ok bool) {
	for i, ln := range lines {
		if fenced[i] || !headRe.MatchString(strings.TrimRight(ln, "\r")) {
			continue
		}
		end = len(lines)
		for j := i + 1; j < len(lines); j++ {
			if !fenced[j] && headingRe.MatchString(strings.TrimRight(lines[j], "\r")) {
				end = j
				break
			}
		}
		return i + 1, end, true
	}
	return 0, 0, false
}
