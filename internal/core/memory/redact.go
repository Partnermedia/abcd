package memory

import (
	"fmt"
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
// acquired-text write passes through here before it is written: page BODIES
// inside WritePages, the one primitive every verb (ingest, ask --file-back)
// writes through, so no body lands unscanned whichever verb built it; the
// --keep-original copy in Ingest; and, transitively, since they are derived
// from the redacted bodies, index.md and log.md. Page FRONTMATTER and the
// registry leaves a write introduces go through the same detector by way of
// redactLeaves (GHSA-x46m-mw9h-5jwj, iss-2608291941064448): a host-supplied
// citation title or recall entry, the licence the core lifts from an SPDX line
// or a License: header, and a redirect-controlled origin are acquired text as
// much as a body is, and the one place every verb writes through is where they
// are judged. contradictions.md is derived from the redacted frontmatter.

// storeRedactor holds a per-repo scanner plus the caller's resolved $HOME for
// the deterministic literal backstop. Construct it with newStoreRedactor, which
// fails closed on a degraded scanner exactly as history.Capture does.
type storeRedactor struct {
	sc   *scanner.Scanner
	home string
}

// openStoreRedactor builds the shared scanner for repoRoot and says, as a plain
// error, why it cannot be trusted: init failed, or the per-repo pii.json left
// it degraded. ScanText/Redact cannot signal the unavailable state in-band
// (only ScanBundle does), so a caller that skipped this check would sanitise
// with a silently weakened pattern set — the exact fail-open this closes. The
// write side (newStoreRedactor) turns that error into a refusal; the lint
// (GHSA-xj89-cc2c-wgwr) turns it into a blocker finding, because its contract
// is to always crawl and write its report.
func openStoreRedactor(repoRoot string) (*storeRedactor, error) {
	sc, err := scanner.New(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("scanner init failed: %v", err)
	}
	if unavail, reason := sc.Unavailable(); unavail {
		return nil, fmt.Errorf("degraded scanner: %s", reason)
	}
	return &storeRedactor{sc: sc, home: scanner.CallerHome()}, nil
}

// newStoreRedactor is the write side of openStoreRedactor: it refuses the whole
// ingest on a broken per-repo pii.json, matching history's posture — a
// committed store must never be written with a detector known to be weakened.
func newStoreRedactor(repoRoot string) (*storeRedactor, error) {
	r, err := openStoreRedactor(repoRoot)
	if err != nil {
		return nil, newIngestError("refusing to ingest: %v", err)
	}
	return r, nil
}

// residueHomeKind labels the literal-home backstop's finding. It mirrors the
// scanner's own kind for the caller's home, so a report reads the same
// whichever detector saw the path.
const residueHomeKind = "home_path_self"

// residue is the read side of redactText, for text ALREADY in the store: the
// spans the write side would refuse or rewrite — every blocking finding of the
// canonical scanner (a hard_fail secret, any identity or network span whatever
// its severity) plus the deterministic literal-home backstop, reported per
// line and deduplicated against a scanner finding of the same kind on that
// line. It rewrites nothing: the lint reports and the operator repairs
// (GHSA-xj89-cc2c-wgwr). A finding carries the kind and the line; the caller
// must never put its Matched span into free text.
func (r *storeRedactor) residue(text, label string) []scanner.Finding {
	out := scanner.BlockingResidual(r.sc.ScanText(text, label))
	if r.home == "" {
		return out
	}
	seen := map[int]bool{}
	for _, f := range out {
		if f.Kind == residueHomeKind {
			seen[f.Line] = true
		}
	}
	for i, line := range strings.Split(text, "\n") {
		if seen[i+1] || scanner.SweepCallerHome(line, r.home) == line {
			continue
		}
		out = append(out, scanner.Finding{File: label, Line: i + 1, Kind: residueHomeKind, Matched: "~"})
	}
	return out
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

// redactLeaves walks target — the nested map[string]any / []any / string shape
// the frontmatter dumper and the JSON registry share — IN PLACE, and sanitises
// through redactText every string leaf the write is INTRODUCING. A leaf is
// introduced when current holds no string at the same path, or a different
// one; a leaf present and equal in current is the store's already, not this
// write's to judge, and stays byte-identical — so a legacy registry carrying
// a dirty cached citation is neither refused on re-ingest (the one verb that
// repairs it) nor rewritten behind the operator's back; reporting it is the
// lint's job. A nil current introduces every leaf. An introduced KEY is judged
// too, by judgeKey — keys are NOT schema-fixed, and a key carrying a secret is
// refused rather than rewritten; non-string scalars pass through untouched.
//
// Mutating in place is the contract, not a shortcut: the map a RegistryMerge
// returned IS what the store then writes, so a caller that kept hold of it —
// the registry-only fast path reads its result licence and citation off the
// merged map — sees the written bytes rather than a pre-redaction copy.
func (r *storeRedactor) redactLeaves(current, target any, label string) error {
	switch v := target.(type) {
	case map[string]any:
		cm, _ := current.(map[string]any)
		for k, item := range v {
			var cv any
			known := false
			if cm != nil {
				cv, known = cm[k]
			}
			// A key the store does not already hold is this write's, exactly as
			// a leaf is, and is judged before its value: the key is the one part
			// of the shape a host controls that no value-side pass can see.
			if !known {
				if err := r.judgeKey(k, label); err != nil {
					return err
				}
			}
			if s, ok := item.(string); ok {
				red, err := r.judgeLeaf(cv, s, label)
				if err != nil {
					return err
				}
				v[k] = red
				continue
			}
			if err := r.redactLeaves(cv, item, label); err != nil {
				return err
			}
		}
	case []any:
		cl, _ := current.([]any)
		for i, item := range v {
			var cv any
			if i < len(cl) {
				cv = cl[i]
			}
			if s, ok := item.(string); ok {
				red, err := r.judgeLeaf(cv, s, label)
				if err != nil {
					return err
				}
				v[i] = red
				continue
			}
			if err := r.redactLeaves(cv, item, label); err != nil {
				return err
			}
		}
	}
	return nil
}

// judgeKey is redactLeaves' one KEY rule. Keys are not schema-fixed:
// validateSourceBlock rejects no unknown key in a page's source: block and the
// frontmatter dumper admits any identifier-shaped key, a `ghp_`-prefixed token
// among them, so a host distiller's page JSON can carry a credential as a YAML
// key. It is refused, never rewritten — renaming a key renames the field the
// reader looks up, and dropping it discards the value it names, so neither is a
// redaction. The refusal names the label and the kinds and NEVER the key
// itself: here the key IS the secret, and an error is the one artefact that
// reaches the terminal and the run log unredacted.
func (r *storeRedactor) judgeKey(key, label string) error {
	resid := r.residue(key, label)
	if len(resid) == 0 {
		return nil
	}
	kinds := make([]string, 0, len(resid))
	for _, f := range resid {
		kinds = append(kinds, f.Kind)
	}
	return newIngestError("refusing to write: a map key in %s carries %d blocking span(s) [%s]; a key cannot be redacted without renaming the field it names, so repair the source", label, len(resid), strings.Join(kinds, ", "))
}

// judgeLeaf is redactLeaves' one leaf rule: unchanged from current, keep it;
// otherwise it is this write's and goes through redactText.
func (r *storeRedactor) judgeLeaf(current any, leaf, label string) (string, error) {
	if c, ok := current.(string); ok && c == leaf {
		return leaf, nil
	}
	red, _, err := r.redactText(leaf, label)
	return red, err
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
