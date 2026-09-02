package scanner

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// pem.go — the block consumer for a PEM private key (GHSA-gmp7-9rvm-qcr3,
// GHSA-5qr6-f78x-g2cx, GHSA-29jw-3jg9-qmhx).
//
// The scanner is line-oriented and the bundled pem_private_key pattern keys on
// the BEGIN line, which is what makes a key block DETECTABLE: the header is
// single-line and self-identifying, while the base64 body after it has no
// prefix, no fixed width and — with no entropy rule in the bundled set —
// matches nothing. Detection is not redaction, though. Redact masks matched
// spans, so masking the header alone wrote every body line and the END line
// verbatim into every store, and the stage-two rescan, running the same
// header rule over a fingerprinted header, was clean by construction.
//
// So the pattern reaches over whatever body shares the header's OWN line (a
// resolve note, a JSON or K8s secret dump with literal \n escapes — see
// patterns.go), and Redact consumes the block's FOLLOWING lines through the
// END line here. The consumer is bounded three ways. It consumes NOTHING until
// a body demonstrably opened — a base64 run long enough to be key material, or
// an armour header, within a line or two of the header (pemBodyEvidence) —
// because a header that is only NAMED in a rotation note or a runbook opens no
// block, and "body-shaped" on its own accepts a blank line, a code fence, a
// setext underline, a bare number and a single-token list item. Once a body
// has opened it takes only body-shaped lines — base64 runs, the armour headers
// a legacy encrypted PEM or a PGP block carries, blank lines, each behind an
// optional gutter (indentation, a diff or quote marker, a line number, a
// quote) — so a truncated block with no END line never swallows the prose
// after it. And it stops at maxPEMBodyLines, so no block consumes a record
// without limit. The consumed lines collapse
// into one placeholder that says how many lines went; the header span itself
// is masked whole (maskedWhole), since a head/tail fingerprint of a span that
// ends in body bytes would keep two bytes of key.
//
// Residual, stated rather than hidden: a body line the shape rule declines —
// a gutter this list does not know — ends the block early and survives, and
// the stage-two rescan cannot see a headerless base64 line any more than
// stage one could (iss-96 tracks the entropy residue). The rule is wide
// enough for a pasted, indented, quoted, diffed or line-numbered block.

// maxPEMBodyLines bounds the lines one block consumer may take after the
// header. A PGP private-key block with several subkeys runs to a few hundred
// lines; the bound clears that by an order of magnitude and still caps a
// pathological block at a fraction of any transcript.
const maxPEMBodyLines = 4096

var (
	// pemEndRe is the END line of any PEM/PGP private-key block; a line that
	// carries it closes the block and is consumed with it.
	pemEndRe = regexp.MustCompile(`-----END (?:[A-Z0-9]+ )*PRIVATE KEY(?: BLOCK)?-----`)
	// pemBodyLineRe is one body-shaped line: an optional gutter, then a base64
	// run or an armour header or nothing at all, then optional quoting.
	pemBodyLineRe = regexp.MustCompile(`^[\s\d+\-|>:"'` + "`" + `│]*(?:[A-Za-z0-9+/=]+|(?:Proc-Type|DEK-Info|Version|Comment|Charset|Hash|MessageID):.*)?[\s"',;\\]*$`)
	// pemBase64RunRe is a run from the base64 alphabet long enough to be key
	// material — the SAME rule the same-line pattern applies before it reaches
	// past a header (patterns.go).
	pemBase64RunRe = regexp.MustCompile(`[A-Za-z0-9+/=]{16,}`)
	// pemArmourRe is an armour header at the head of a line, behind the same
	// optional gutter pemBodyLineRe allows.
	pemArmourRe = regexp.MustCompile(`^[\s\d+\-|>:"'` + "`" + `│]*(?:Proc-Type|DEK-Info|Version|Comment|Charset|Hash|MessageID):`)
)

// pemEvidenceWindow is how far past the header the consumer looks for evidence
// that a body opened. A block's first body line follows the header directly, or
// after one blank line; nothing further out is a block this consumer opened.
const pemEvidenceWindow = 2

// pemBodyOpened reports whether a line carries positive evidence that a key
// body opened on it: a base64 run long enough to be key material, or a PEM/PGP
// armour header. A run of pure padding ("====…", a setext underline) is not
// evidence — base64 padding is at most two bytes and never stands alone — so
// the run must carry at least one alphanumeric byte.
func pemBodyOpened(line string) bool {
	if pemArmourRe.MatchString(line) {
		return true
	}
	for _, run := range pemBase64RunRe.FindAllString(line, -1) {
		if strings.ContainsFunc(run, isBase64Alnum) {
			return true
		}
	}
	return false
}

func isBase64Alnum(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

// pemBodyEvidence reports whether a body opened after the header at lines[h].
// The shape rule alone cannot answer this: "body-shaped" accepts a blank line,
// a code fence, a setext underline, a bare number and a single-token list item,
// so a sentence that merely NAMES a BEGIN marker — a rotation note, a runbook,
// an issue record — used to consume the lines after it and could leave an
// unbalanced fence where a fenced block had been. Evidence is required before
// anything is taken; a line that is not even body-shaped ends the search.
func pemBodyEvidence(lines []string, h int) bool {
	last := h + pemEvidenceWindow
	if last >= len(lines) {
		last = len(lines) - 1
	}
	for j := h + 1; j <= last; j++ {
		if pemBodyOpened(lines[j]) {
			return true
		}
		if !pemBodyLineRe.MatchString(lines[j]) {
			return false
		}
	}
	return false
}

// pemBodyPlaceholder is the one line a consumed block collapses to.
func pemBodyPlaceholder(n int) string {
	return fmt.Sprintf("[redacted-pem-body: %d lines]", n)
}

// consumePEMBodies collapses, for every pem_private_key finding whose span did
// not reach an END marker on the header's own line, the body-shaped lines that
// follow the header through the END line into one placeholder line. It returns
// the rewritten lines and the number of blocks collapsed. Headers are handled
// from the last line upward so an earlier header's indices stay valid.
//
// original is the text as it was BEFORE the per-line redaction; lines is the
// sealed slice the placeholder is written into. Where a block ends is a
// question about the source text — a seal writes '*' bytes, which are not
// body-shaped, so a second finding on a body line would otherwise close the
// block at that line and leave the rest of the key behind. The two slices are
// the same length by construction (per-line redaction never adds or removes a
// line); if they ever diverge, the sealed slice is the safe fallback.
func consumePEMBodies(original, lines []string, findings []Finding) ([]string, int) {
	if len(original) != len(lines) {
		original = lines
	}
	var headers []int
	seen := map[int]bool{}
	for _, f := range findings {
		if f.Kind != kindPEMPrivateKey || pemEndRe.MatchString(f.Matched) {
			continue
		}
		idx := f.Line - 1
		if idx < 0 || idx >= len(lines) || seen[idx] {
			continue
		}
		seen[idx] = true
		headers = append(headers, idx)
	}
	if len(headers) == 0 {
		return lines, 0
	}
	sort.Sort(sort.Reverse(sort.IntSlice(headers)))
	blocks := 0
	for _, h := range headers {
		end := pemBlockEnd(original, h)
		if end <= h+1 {
			continue
		}
		rest := append([]string{pemBodyPlaceholder(end - (h + 1))}, lines[end:]...)
		lines = append(lines[:h+1:h+1], rest...)
		blocks++
	}
	return lines, blocks
}

// pemBlockEnd returns the exclusive index of the last line the block whose
// header is lines[h] consumes: through the END line when one is reached within
// the bound, else through the last body-shaped line — with trailing blank
// lines given back, since they belong to the prose after a truncated block.
func pemBlockEnd(lines []string, h int) int {
	if !pemBodyEvidence(lines, h) {
		return h + 1 // the header was named, not opened: take nothing
	}
	limit := h + 1 + maxPEMBodyLines
	if limit > len(lines) {
		limit = len(lines)
	}
	j := h + 1
	for ; j < limit; j++ {
		if pemEndRe.MatchString(lines[j]) {
			return j + 1
		}
		if !pemBodyLineRe.MatchString(lines[j]) {
			break
		}
	}
	for j > h+1 && strings.TrimSpace(lines[j-1]) == "" {
		j--
	}
	return j
}
