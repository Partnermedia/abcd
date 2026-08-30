package intent

// grounds.go is the intent record's grounds writer and reader (spc-57). Grounds
// were recorded only for deliberate non-action — a wontfix carries a note, an
// ADR carries its alternatives — while the reasoning behind what went FORWARD
// had no home and evaporated at the gate. It lives on the intent record because
// the conjecture is the intent's, and the gate that reads it takes an itd-N.
//
// The vocabulary and the grammar are core/grounds's, held once for the three
// writers; what this file owns is the record form — one top-level bullet under
// `## Grounds` — and the append that puts an entry there. It borrows the section
// and bullet machinery claims.go already owns (sectionLineRangeIn, maskLines,
// conditionBlocks), so a fenced example or a bullet parked in an HTML comment is
// no more a recorded ground than it is a scope condition.

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/intentdriven/abcd/internal/core/grounds"
	"github.com/intentdriven/abcd/internal/core/recordid"
)

// GroundsHeading is the section an intent record carries its grounds under. It
// is spelled once: the writer creates it, the reader locates it, and the
// readiness gate names it in a remedy.
const GroundsHeading = "Grounds"

// groundsHeadingRe matches the `## Grounds` heading at any heading depth, on the
// mechanismHeadingRe pattern.
var groundsHeadingRe = regexp.MustCompile(`^#{1,6}\s+` + GroundsHeading + `\s*$`)

// GroundsResult is the outcome of one RecordGrounds call. Redacted counts the
// spans the redactor rewrote before the write, so a surface can SAY the text was
// altered: rewriting somebody's reasoning in silence is worse than not recording
// it.
type GroundsResult struct {
	IntentID string `json:"intent_id"`
	Path     string `json:"path"`  // repo-relative intent path
	Token    string `json:"token"` // pursued | deferred | declined
	Text     string `json:"text"`  // the text as WRITTEN (post-redaction)
	Entries  int    `json:"entries"`
	Redacted int    `json:"redacted,omitempty"`
}

// RecordGrounds appends one grounds entry to an intent record, creating the
// `## Grounds` section when it is absent.
//
// Recording is APPEND-ONLY: a second gate decision adds an entry beside the
// first rather than replacing it, because the earlier conjecture is precisely
// what a later reader checks the outcome against. Rewriting it would leave the
// record saying only what was believed last.
//
// The text is redacted BEFORE it is validated, never after, so no rewritten span
// can reach a field the validator has already passed; the redactor is the same
// fail-closed one the quoted-text create path uses, because a grounds text is
// durable committed prose. The write goes through writeIntentFile, the package's
// one intent-record writer.
func RecordGrounds(repoRoot, intentID string, g grounds.Grounds) (GroundsResult, error) {
	if !recordid.ValidIntentID(intentID) {
		return GroundsResult{}, fmt.Errorf("intent: id %q must match ^itd-[0-9]+$", intentID)
	}
	corpus, err := Load(repoRoot)
	if err != nil {
		return GroundsResult{}, err
	}
	it, ok := corpus.Lookup(intentID)
	if !ok {
		return GroundsResult{}, fmt.Errorf("intent: %s not found in any bucket", intentID)
	}
	redText, redacted, err := redactIntentText(repoRoot, g.Text)
	if err != nil {
		return GroundsResult{}, err
	}
	// Re-validated on the redacted text: the token against the closed set, the
	// text against the substance floor. A redaction that emptied the text, or a
	// token no caller checked, is refused here with nothing written.
	validated, err := grounds.New(g.Token, redText)
	if err != nil {
		return GroundsResult{}, fmt.Errorf("intent: %w", err)
	}

	abs := filepath.Join(repoRoot, it.Path)
	data, err := readRepoFile(abs, it.Path)
	if err != nil {
		return GroundsResult{}, err
	}
	updated := appendGroundsBullet(string(data), validated)
	if err := writeIntentFile(abs, it.Path, updated); err != nil {
		return GroundsResult{}, err
	}
	return GroundsResult{
		IntentID: it.ID,
		Path:     it.Path,
		Token:    string(validated.Token),
		Text:     validated.Text,
		Entries:  len(ParseGrounds(updated)),
		Redacted: redacted,
	}, nil
}

// ParseGrounds reads a record's recorded grounds, in the order they were
// written. It is the single reader every consumer asks — the readiness gate and
// the result count alike — so no two of them can disagree about what an entry is.
//
// A bullet that does not parse is not an entry: it is prose under the heading,
// and reporting it as a malformed ground would put a gate verdict on a sentence
// somebody wrote for a human. What the gate asks is whether at least one
// WELL-FORMED entry is there.
func ParseGrounds(content string) []grounds.Grounds {
	lines := strings.Split(content, "\n")
	mask := maskLines(lines)
	start, end, ok := sectionLineRangeIn(lines, mask, groundsHeadingRe)
	if !ok {
		return nil
	}
	var out []grounds.Grounds
	for _, b := range conditionBlocks(lines, mask, start, end) {
		g, err := grounds.Parse(groundsBlockText(lines, b))
		if err != nil {
			continue
		}
		out = append(out, g)
	}
	return out
}

// groundsBlockText folds one top-level bullet — and the continuation lines
// wrapped into it — back into the single line the grammar is written on.
func groundsBlockText(lines []string, b conditionBlock) string {
	parts := make([]string, 0, b.end-b.start)
	for i := b.start; i < b.end; i++ {
		ln := strings.TrimRight(lines[i], "\r")
		if i == b.start {
			ln = bulletPrefixRe.ReplaceAllString(ln, "")
		}
		parts = append(parts, ln)
	}
	return grounds.Fold(strings.Join(parts, " "))
}

// appendGroundsBullet puts one entry at the end of the record's `## Grounds`
// section, creating the section at end of file when it is absent. It mirrors
// appendToAuditNotes, including the trailing link-reference peel: a `[ref]: url`
// definition parked at the end of a section belongs below its prose, and
// appending under it would detach the entry from the section it belongs to.
func appendGroundsBullet(content string, g grounds.Grounds) string {
	lines := strings.Split(content, "\n")
	mask := maskLines(lines)
	start, end, ok := sectionLineRangeIn(lines, mask, groundsHeadingRe)
	if !ok {
		body := strings.TrimRight(content, "\n")
		return body + "\n\n## " + GroundsHeading + "\n\n" + g.Bullet() + "\n"
	}
	section := append([]string{}, lines[start:end]...)
	for len(section) > 0 && strings.TrimSpace(section[len(section)-1]) == "" {
		section = section[:len(section)-1]
	}
	trailingRefs := peelTrailingLinkRefs(&section)
	rebuilt := make([]string, 0, len(lines)+6)
	rebuilt = append(rebuilt, lines[:start]...)
	rebuilt = append(rebuilt, section...)
	if len(section) > 0 && strings.TrimSpace(section[len(section)-1]) != "" &&
		!bulletRe.MatchString(strings.TrimRight(section[len(section)-1], "\r")) {
		// Prose immediately above the first entry needs its blank line; a run of
		// bullets is one list and must not be split by one.
		rebuilt = append(rebuilt, "")
	}
	rebuilt = append(rebuilt, g.Bullet())
	if len(trailingRefs) > 0 {
		rebuilt = append(rebuilt, "")
		rebuilt = append(rebuilt, trailingRefs...)
	}
	rebuilt = append(rebuilt, "")
	rebuilt = append(rebuilt, lines[end:]...)
	return strings.Join(rebuilt, "\n")
}
