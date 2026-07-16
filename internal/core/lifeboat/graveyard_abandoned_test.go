package lifeboat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// abandonedWriter builds an empty temp repo and returns its root plus a closure
// that writes a repo-relative file (creating parent dirs). It is the layer-2
// analogue of nativeTierFixture: each test writes exactly the record material it
// asserts on, so a fixture is never coupled to another test's expectations.
func abandonedWriter(t *testing.T) (string, func(rel, content string)) {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, content string) {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir, write
}

func abandonedCtx(t *testing.T, dir string) *SourceContext {
	t.Helper()
	ctx, err := newSourceContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ctx.Close() })
	return ctx
}

func gvFindingByID(fs []Finding, id string) (Finding, bool) {
	for _, f := range fs {
		if f.ID == id {
			return f, true
		}
	}
	return Finding{}, false
}

func gvCountSignal(fs []Finding, sig Signal) int {
	n := 0
	for _, f := range fs {
		if f.Signal == sig {
			n++
		}
	}
	return n
}

func gvEvidenceContains(f Finding, sub string) bool {
	for _, e := range f.Evidence {
		if strings.Contains(e, sub) {
			return true
		}
	}
	return false
}

// --- 1. superseded intents ---------------------------------------------------

func TestAbandonedSupersededIntent(t *testing.T) {
	dir, write := abandonedWriter(t)
	// A superseded intent (its bucket IS the lifecycle state).
	write(".abcd/development/intents/superseded/itd-31-cross-document-fidelity-reviewer.md",
		"---\nid: itd-31\nslug: cross-document-fidelity-reviewer\nsuperseded_by: itd-48\n---\n\n# Superseded intent\n")
	// A live intent in drafts/ must NOT be reported as superseded.
	write(".abcd/development/intents/drafts/itd-9-live.md",
		"---\nid: itd-9\nslug: live\n---\n\n# A live intent\n")

	fs := gvSupersededIntents(abandonedCtx(t, dir))
	if gvCountSignal(fs, SignalSupersededIntent) != 1 {
		t.Fatalf("want exactly one superseded-intent finding, got %d (%v)", len(fs), fs)
	}
	f, ok := gvFindingByID(fs, "itd-31")
	if !ok {
		t.Fatalf("want a finding keyed itd-31, got %v", fs)
	}
	if f.Signal != SignalSupersededIntent {
		t.Errorf("signal = %s, want %s", f.Signal, SignalSupersededIntent)
	}
	if !gvEvidenceContains(f, ".abcd/development/intents/superseded/itd-31-cross-document-fidelity-reviewer.md") {
		t.Errorf("evidence = %v, want the superseded path cited", f.Evidence)
	}
	if _, ok := gvFindingByID(fs, "itd-9"); ok {
		t.Errorf("a live intent (itd-9) must not appear as superseded")
	}
}

func TestAbandonedSupersededIntentsSortedNumerically(t *testing.T) {
	dir, write := abandonedWriter(t)
	write(".abcd/development/intents/superseded/itd-10-ten.md", "---\nid: itd-10\n---\n")
	write(".abcd/development/intents/superseded/itd-2-two.md", "---\nid: itd-2\n---\n")
	fs := gvSupersededIntents(abandonedCtx(t, dir))
	if len(fs) != 2 {
		t.Fatalf("want 2 findings, got %d", len(fs))
	}
	if fs[0].ID != "itd-2" || fs[1].ID != "itd-10" {
		t.Errorf("order = [%s %s], want numeric [itd-2 itd-10]", fs[0].ID, fs[1].ID)
	}
}

// --- 2. superseded ADRs ------------------------------------------------------

func TestAbandonedSupersededADRByStatus(t *testing.T) {
	dir, write := abandonedWriter(t)
	write(".abcd/development/decisions/adrs/0007-old.md",
		"---\nid: adr-7\nstatus: superseded\nsuperseded_by: null\n---\n\n# Old decision\n")
	fs := gvSupersededADRs(abandonedCtx(t, dir))
	f, ok := gvFindingByID(fs, "adr-7")
	if !ok {
		t.Fatalf("status: superseded ADR should yield adr-7, got %v", fs)
	}
	if f.Signal != SignalSupersededADR {
		t.Errorf("signal = %s, want %s", f.Signal, SignalSupersededADR)
	}
}

func TestAbandonedSupersededADRBySupersededBy(t *testing.T) {
	dir, write := abandonedWriter(t)
	write(".abcd/development/decisions/adrs/0012-thing.md",
		"---\nid: adr-12\nstatus: accepted\nsuperseded_by: adr-31\n---\n\n# Thing\n")
	fs := gvSupersededADRs(abandonedCtx(t, dir))
	f, ok := gvFindingByID(fs, "adr-12")
	if !ok {
		t.Fatalf("superseded_by ADR should yield adr-12, got %v", fs)
	}
	if !gvEvidenceContains(f, "adr-31") {
		t.Errorf("evidence = %v, want the superseding target adr-31 named", f.Evidence)
	}
}

func TestAbandonedAcceptedADRIsNotReported(t *testing.T) {
	dir, write := abandonedWriter(t)
	// Mirrors the real adr-35 frontmatter: accepted, superseded_by null.
	write(".abcd/development/decisions/adrs/0035-live.md",
		"---\nid: adr-35\nstatus: accepted\nsuperseded_by: null\n---\n\n# Live decision\n")
	fs := gvSupersededADRs(abandonedCtx(t, dir))
	if len(fs) != 0 {
		t.Fatalf("an accepted ADR (superseded_by null) must not be reported, got %v", fs)
	}
}

func TestAbandonedSupersededADRAcrossBothHomesDedupes(t *testing.T) {
	dir, write := abandonedWriter(t)
	// Same ADR id present in the native home AND a conventional home.
	write(".abcd/development/decisions/adrs/0012-thing.md",
		"---\nid: adr-12\nstatus: superseded\nsuperseded_by: adr-31\n---\n\n# Native copy\n")
	write("docs/adr/0012-thing.md",
		"---\nid: adr-12\nstatus: superseded\nsuperseded_by: adr-31\n---\n\n# Conventional copy\n")
	fs := gvSupersededADRs(abandonedCtx(t, dir))
	if gvCountSignal(fs, SignalSupersededADR) != 1 {
		t.Fatalf("same ADR in both homes must dedupe to one finding, got %d (%v)", len(fs), fs)
	}
	f, _ := gvFindingByID(fs, "adr-12")
	if !gvEvidenceContains(f, ".abcd/development/decisions/adrs/0012-thing.md") {
		t.Errorf("first-wins should cite the native home, evidence = %v", f.Evidence)
	}
}

func TestAbandonedSupersededADRIDFromFilenameFallback(t *testing.T) {
	dir, write := abandonedWriter(t)
	// No usable frontmatter id — must derive adr-12 from the NNNN- filename.
	write("docs/adrs/0012-thing.md",
		"---\nstatus: superseded\n---\n\n# No id in frontmatter\n")
	fs := gvSupersededADRs(abandonedCtx(t, dir))
	if _, ok := gvFindingByID(fs, "adr-12"); !ok {
		t.Fatalf("want adr-12 derived from filename, got %v", fs)
	}
}

// --- 3. alternatives considered ---------------------------------------------

func TestAbandonedAlternativesConsidered(t *testing.T) {
	dir, write := abandonedWriter(t)
	write(".abcd/development/decisions/adrs/0004-voyage.md",
		"---\nid: adr-4\nstatus: accepted\n---\n\n# 4. Voyage\n\n## Context\n\nWe need to pack.\n\n"+
			"## Alternatives Considered\n\n- Voyage inside the source repo — rejected because it mutates the tree.\n"+
			"- A second clone — rejected because it doubles disk.\n\n## Decision\n\nOut-of-tree.\n")
	// An ADR with no such section must not produce an alternatives finding.
	write(".abcd/development/decisions/adrs/0005-plain.md",
		"---\nid: adr-5\nstatus: accepted\n---\n\n# 5. Plain\n\n## Context\n\nNo alternatives here.\n")

	fs := gvAlternativesConsidered(abandonedCtx(t, dir))
	if gvCountSignal(fs, SignalAlternativesConsidered) != 1 {
		t.Fatalf("want exactly one alternatives finding, got %d (%v)", len(fs), fs)
	}
	f, ok := gvFindingByID(fs, "adr-4-alt")
	if !ok {
		t.Fatalf("want adr-4-alt, got %v", fs)
	}
	if !gvEvidenceContains(f, "Voyage inside the source repo") {
		t.Errorf("evidence = %v, want the first bullet quoted", f.Evidence)
	}
	if len(f.Evidence) != 2 {
		t.Errorf("want the two top-level bullets, got %d (%v)", len(f.Evidence), f.Evidence)
	}
	if _, ok := gvFindingByID(fs, "adr-5-alt"); ok {
		t.Errorf("adr-5 has no Alternatives section and must not appear")
	}
}

// --- 4. wontfix issues -------------------------------------------------------

func TestAbandonedWontfixIssue(t *testing.T) {
	dir, write := abandonedWriter(t)
	// Mirrors the real issue frontmatter shape: quoted id/slug.
	write(".abcd/work/issues/wontfix/iss-30-atomic-write.md",
		"---\nschema_version: 1\nid: \"iss-30\"\nslug: \"atomic-write\"\nseverity: \"minor\"\n"+
			"wontfix_reason: \"superseded by the atomic-write consolidation\"\n---\n\nbody.\n")
	fs := gvWontfixIssues(abandonedCtx(t, dir))
	f, ok := gvFindingByID(fs, "iss-30")
	if !ok {
		t.Fatalf("want iss-30 finding, got %v", fs)
	}
	if f.Signal != SignalWontfixIssue {
		t.Errorf("signal = %s, want %s", f.Signal, SignalWontfixIssue)
	}
	if !gvEvidenceContains(f, "superseded by the atomic-write consolidation") {
		t.Errorf("evidence = %v, want the wontfix reason quoted (unquoted)", f.Evidence)
	}
}

func TestAbandonedEmptyWontfixDirIsNone(t *testing.T) {
	dir, write := abandonedWriter(t)
	// A wontfix dir that exists but holds no iss-*.md.
	write(".abcd/work/issues/wontfix/README.md", "# wontfix ledger\n")
	fs := gvWontfixIssues(abandonedCtx(t, dir))
	if len(fs) != 0 {
		t.Fatalf("an empty wontfix ledger must yield no findings, got %v", fs)
	}
}

// --- 5. rejected options in DECISIONS.md ------------------------------------

func TestAbandonedRejectedOptions(t *testing.T) {
	dir, write := abandonedWriter(t)
	write(".abcd/work/DECISIONS.md",
		"# DECISIONS\n\n"+ // line 1, 2
			"- 2026-07-06 — Adopt Cobra as the CLI framework.\n"+ // line 3 (neutral)
			"- 2026-07-08 — RAG rejected at this scale; grep corpus instead.\n"+ // line 4 (rejected)
			"- 2026-07-09 — flow-next dropped in favour of native.\n"+ // line 5 (dropped)
			"- 2026-07-10 — Private companion repo deferred (trigger: shared transcripts).\n") // line 6 (deferred)

	fs := gvRejectedOptions(abandonedCtx(t, dir))
	if gvCountSignal(fs, SignalRejectedOption) != 3 {
		t.Fatalf("want 3 rejected-option findings (rejected/dropped/deferred), got %d (%v)", len(fs), fs)
	}
	// dec-L<line> keyed by 1-based line number, file order preserved.
	if fs[0].ID != "dec-L4" || fs[1].ID != "dec-L5" || fs[2].ID != "dec-L6" {
		t.Errorf("ids = [%s %s %s], want [dec-L4 dec-L5 dec-L6]", fs[0].ID, fs[1].ID, fs[2].ID)
	}
	if !gvEvidenceContains(fs[0], "RAG rejected at this scale") {
		t.Errorf("evidence = %v, want the line quoted verbatim", fs[0].Evidence)
	}
	if _, ok := gvFindingByID(fs, "dec-L3"); ok {
		t.Errorf("the neutral bullet on line 3 must not be reported")
	}
}

func TestAbandonedRejectedOptionsConservativeMatcher(t *testing.T) {
	dir, write := abandonedWriter(t)
	// "rejection" is a different word from the verb "rejected": the substring
	// matcher must NOT fire on it (documented conservative-verbs-only edge).
	write(".abcd/work/DECISIONS.md",
		"- 2026-07-06 — Handling of the rejection path is documented in the spec.\n")
	fs := gvRejectedOptions(abandonedCtx(t, dir))
	if len(fs) != 0 {
		t.Fatalf("a bullet containing \"rejection\" (not the verb) must not fire, got %v", fs)
	}
}

// --- 6. empty record ---------------------------------------------------------

func TestAbandonedEmptyRecordIsEmptySlice(t *testing.T) {
	dir := t.TempDir() // no .abcd at all
	ab := buildAbandoned(abandonedCtx(t, dir))
	if ab.SchemaVersion != GraveyardSchemaVersion {
		t.Errorf("schema_version = %d, want %d", ab.SchemaVersion, GraveyardSchemaVersion)
	}
	if ab.Findings == nil {
		t.Fatal("Findings must be a non-nil empty slice, not nil")
	}
	if len(ab.Findings) != 0 {
		t.Fatalf("want no findings for an empty record, got %v", ab.Findings)
	}
	j, err := json.MarshalIndent(ab, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(j), "\"findings\": []") {
		t.Errorf("empty abandoned must marshal findings as [], got:\n%s", j)
	}
}

// --- 7. integration + id stability ------------------------------------------

func TestAbandonedBuildGroupsAllSignalsInRankOrder(t *testing.T) {
	dir := abandonedFullFixture(t)
	ab := buildAbandoned(abandonedCtx(t, dir))
	// Signals must appear grouped in signalRank order, never re-sorted globally.
	lastRank := -1
	for _, f := range ab.Findings {
		r, ok := signalRank[f.Signal]
		if !ok {
			t.Fatalf("finding %s carries an unranked signal %s", f.ID, f.Signal)
		}
		if r < lastRank {
			t.Fatalf("signal %s (rank %d) appears after rank %d — not grouped by signalRank", f.Signal, r, lastRank)
		}
		lastRank = r
	}
	// Sanity: at least one finding from each of the five layer-2 signals.
	for _, sig := range []Signal{
		SignalSupersededIntent, SignalSupersededADR, SignalAlternativesConsidered,
		SignalWontfixIssue, SignalRejectedOption,
	} {
		if gvCountSignal(ab.Findings, sig) == 0 {
			t.Errorf("full fixture produced no %s finding", sig)
		}
	}
}

func TestAbandonedIDStabilityAcrossCalls(t *testing.T) {
	dir := abandonedFullFixture(t)
	a := buildAbandoned(abandonedCtx(t, dir))
	b := buildAbandoned(abandonedCtx(t, dir))
	ja, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	jb, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if string(ja) != string(jb) {
		t.Errorf("buildAbandoned is not byte-stable across calls:\n--- a ---\n%s\n--- b ---\n%s", ja, jb)
	}
}

// --- 8. sanitisation ---------------------------------------------------------

func TestAbandonedSanitisesControlChars(t *testing.T) {
	dir, write := abandonedWriter(t)
	// A DECISIONS line and a wontfix reason each carrying an ANSI escape (0x1b)
	// and a NUL (0x00) plus an HTML-comment marker. sanitize() maps C0/DEL to '?'
	// and tab to space; markers are inert in JSON so they pass through unchanged.
	write(".abcd/work/DECISIONS.md",
		"- 2026-07-06 — option \x1bdropped\x00 here <!-- hide -->\n")
	write(".abcd/work/issues/wontfix/iss-9-x.md",
		"---\nid: \"iss-9\"\nwontfix_reason: \"bad\x1breason\x00 <!-- x -->\"\n---\n")

	dec := gvRejectedOptions(abandonedCtx(t, dir))
	if len(dec) != 1 {
		t.Fatalf("want one decision finding, got %v", dec)
	}
	for _, e := range dec[0].Evidence {
		if strings.ContainsRune(e, 0x1b) || strings.ContainsRune(e, 0x00) {
			t.Errorf("evidence retains a control char: %q", e)
		}
		if !strings.Contains(e, "?dropped?") {
			t.Errorf("control chars should map to '?', got %q", e)
		}
	}

	won := gvWontfixIssues(abandonedCtx(t, dir))
	f, ok := gvFindingByID(won, "iss-9")
	if !ok {
		t.Fatalf("want iss-9, got %v", won)
	}
	for _, e := range f.Evidence {
		if strings.ContainsRune(e, 0x1b) || strings.ContainsRune(e, 0x00) {
			t.Errorf("wontfix evidence retains a control char: %q", e)
		}
	}
}

// abandonedFullFixture writes a record exercising all five layer-2 signals at
// once. It is the integration/stability fixture; the per-signal tests keep their
// own minimal fixtures.
func abandonedFullFixture(t *testing.T) string {
	t.Helper()
	dir, write := abandonedWriter(t)
	write(".abcd/development/intents/superseded/itd-47-oracle-gates.md",
		"---\nid: itd-47\nslug: oracle-gates\nsuperseded_by: itd-48\n---\n\n# Superseded\n")
	write(".abcd/development/decisions/adrs/0012-old.md",
		"---\nid: adr-12\nstatus: superseded\nsuperseded_by: adr-31\n---\n\n# Old\n")
	write(".abcd/development/decisions/adrs/0004-voyage.md",
		"---\nid: adr-4\nstatus: accepted\n---\n\n# Voyage\n\n"+
			"## Alternatives Considered\n\n- Voyage inside the source repo — rejected.\n- A second clone — rejected.\n")
	write(".abcd/work/issues/wontfix/iss-30-atomic.md",
		"---\nid: \"iss-30\"\nslug: \"atomic\"\nwontfix_reason: \"superseded by consolidation\"\n---\n\nbody.\n")
	write(".abcd/work/DECISIONS.md",
		"# DECISIONS\n\n- 2026-07-08 — RAG rejected at this scale.\n")
	return dir
}
