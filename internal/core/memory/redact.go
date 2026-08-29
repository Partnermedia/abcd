package memory

import (
	"strings"
	"unicode/utf8"

	"github.com/intentdriven/abcd/internal/adapter/scanner"
)

// redact.go — the write-time secret/PII sanitiser for the committed
// .abcd/memory/ store (GHSA-j5f5-phgm-9m73).
//
// capture, history and launch all route free text through the shared
// internal/adapter/scanner before it lands in a committed artefact; memory
// ingest did not. Ingest acquires a local file or URL, distils it, and writes
// page bodies (and, under --keep-original, the raw source bytes) under
// .abcd/memory/ — a tree .gitignore does NOT exclude, so an operator commits
// whatever it holds. A PAT or an absolute home path in the source therefore
// reached the repository with every secret-pattern lint gate green, because
// those gates are not this package.
//
// The fix reuses the ONE canonical detector (no second scanner) and mirrors
// history.Capture's fail-closed, two-stage discipline: refuse on a degraded
// scanner, redact-and-report through scanner.Redact, apply a literal $HOME
// backstop, then re-scan and refuse if a blocking span survived. Every
// acquired-text write passes through here before it is written: page bodies
// inside WritePages, the one primitive every verb (ingest, ask --file-back)
// writes through, so no PageWrite lands unscanned whichever verb built it; the
// --keep-original copy in Ingest; and, transitively, since they are derived
// from the redacted bodies, index.md and log.md.

// storeRedactor holds a per-repo scanner plus the caller's resolved $HOME for
// the deterministic literal backstop. Construct it with newStoreRedactor, which
// fails closed on a degraded scanner exactly as history.Capture does.
type storeRedactor struct {
	sc   *scanner.Scanner
	home string
}

// newStoreRedactor builds the shared scanner for repoRoot and refuses when it is
// degraded. ScanText/Redact cannot signal the unavailable state in-band (only
// ScanBundle does), so a caller that skipped this check would sanitise with a
// silently weakened pattern set — the exact fail-open this closes. Refusing the
// whole ingest on a broken per-repo pii.json matches history's posture: a
// committed store must never be written with a detector known to be weakened.
func newStoreRedactor(repoRoot string) (*storeRedactor, error) {
	sc, err := scanner.New(repoRoot)
	if err != nil {
		return nil, newIngestError("refusing to ingest: scanner init failed: %v", err)
	}
	if unavail, reason := sc.Unavailable(); unavail {
		return nil, newIngestError("refusing to ingest with a degraded scanner: %s", reason)
	}
	return &storeRedactor{sc: sc, home: scanner.CallerHome()}, nil
}

// redactText sanitises one free-text blob bound for the store. label is the
// logical name stamped into each finding. It returns the redacted text and the
// number of spans rewritten, or an IngestError when a blocking span (any
// identity/network span whatever its severity, plus every hard_fail one)
// survives redaction — the same fail-closed gate history.Capture applies.
func (r *storeRedactor) redactText(text, label string) (string, int, error) {
	findings := r.sc.ScanText(text, label)
	redacted, _ := scanner.Redact(text, findings)

	// Deterministic literal $HOME backstop, independent of the scanner
	// heuristic (defence in depth on this trust boundary): collapse every
	// remaining occurrence of the resolved home that stands as a path to "~",
	// rewrite any /Users/<user> or /home/<user> segment for the caller's own
	// username that the literal sweep cannot see, then fail closed only if the
	// literal home still shows. The same gate history.Capture holds.
	if r.home != "" {
		redacted = scanner.SweepCallerHome(redacted, r.home)
		var resid []scanner.Finding
		redacted, resid = scanner.SurvivingCallerHome(redacted, r.home)
		if len(resid) > 0 {
			return "", 0, newIngestError("refusing to write: the caller's home path survived redaction")
		}
	}
	if resid := scanner.BlockingResidual(r.sc.ScanText(redacted, label)); len(resid) > 0 {
		kinds := make([]string, 0, len(resid))
		for _, f := range resid {
			kinds = append(kinds, f.Kind)
		}
		return "", 0, newIngestError("refusing to write: redaction left %d blocking span(s) unresolved [%s]", len(resid), strings.Join(kinds, ", "))
	}
	return redacted, len(findings), nil
}

// redactOriginalBytes sanitises the --keep-original source copy. A text source's
// raw bytes ARE its text, so they are redacted in place. A binary source (a PDF)
// cannot be redacted byte-wise without corrupting it, so its distilled text is
// scanned instead and the keep-original copy is REFUSED when that text carries a
// blocking span — a refused best-effort copy is recorded, never fatal, and never
// leaks. A binary source with no blocking span is stored verbatim.
func (r *storeRedactor) redactOriginalBytes(material sourceMaterial) ([]byte, error) {
	if isRedactableText(material.rawBytes) {
		red, _, err := r.redactText(string(material.rawBytes), "source")
		if err != nil {
			return nil, err
		}
		return []byte(red), nil
	}
	if _, _, err := r.redactText(material.text, "source"); err != nil {
		return nil, newIngestError("refusing to keep original: source carries a secret/PII span that cannot be redacted in a binary file")
	}
	return material.rawBytes, nil
}

// isRedactableText reports whether data is UTF-8 text with no NUL byte — the
// same shape decodeText accepts — so it can be safely redacted as a string. A
// binary blob (a PDF) is not, and must never be rewritten span-wise.
func isRedactableText(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return false
		}
	}
	return utf8.Valid(data)
}
