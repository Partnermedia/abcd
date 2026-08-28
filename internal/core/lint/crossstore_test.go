package lint

import (
	"path/filepath"
	"strings"
	"testing"
)

// crossStoreCfg builds a config with the cross-store rule armed and the ADR store
// declared, so the rule has a corpus to weigh an id claim against.
func crossStoreCfg() Config {
	return Config{
		Rules: map[string]RuleConfig{
			ruleCrossStoreIDClaim: {Enabled: true, Severity: severityBlocker},
			ruleRecordSchema: {Enabled: false, RecordStores: map[string]string{
				"adr": ".abcd/development/decisions/adrs",
				"itd": ".abcd/development/intents",
				"spc": ".abcd/development/specs",
				"iss": ".abcd/work/issues",
			}},
		},
		ExemptPaths: []string{".abcd/development/research/"},
	}
}

// writeTakenADR puts a real ADR in the ADR store, so its id is TAKEN. It is the
// sibling of writeADR (contextcurrency_test.go), which knobs the lifecycle
// frontmatter this rule never reads and carries no H1 to claim an id with.
func writeTakenADR(t *testing.T, root, file, id, title string) {
	t.Helper()
	writeFile(t, root, filepath.Join(".abcd", "development", "decisions", "adrs", file),
		"---\nid: "+id+"\nstatus: accepted\n---\n\n# "+title+"\n")
}

// probeNote is the iss-2608230752354926 probe, verbatim in shape: an H1 claiming a
// taken ADR id, a decision-shaped Status block, and no frontmatter at all.
const probeNote = "# ADR-23: Transport Agnostic Core (probe)\n" +
	"\n" +
	"## Status\n" +
	"\n" +
	"Accepted (locked)\n"

// TestCrossStoreIDClaimFlagsTheProbe is the criterion the issue's own probe
// states: the file passed record-lint clean, exit 0, zero findings. It must now
// produce a finding naming the outside-store id claim.
func TestCrossStoreIDClaimFlagsTheProbe(t *testing.T) {
	root := t.TempDir()
	writeTakenADR(t, root, "0023-transport-agnostic-core.md", "adr-23", "Transport-agnostic core")
	writeFile(t, root, filepath.Join("research", "notes", "zz-recurrence-probe.md"), probeNote)

	fs, err := Lint(crossStoreCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleCrossStoreIDClaim); n != 1 {
		t.Fatalf("expected exactly the probe to fire, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join("research", "notes", "zz-recurrence-probe.md"), ruleCrossStoreIDClaim, 1) {
		t.Errorf("expected the finding on the claiming heading; got %+v", fs)
	}
	if !messageContains(fs, "adr-23") {
		t.Errorf("expected the message to name the claimed id; got %+v", fs)
	}
	// The severity is blocking, which is what makes record-lint exit non-zero.
	for _, f := range fs {
		if f.RuleID == ruleCrossStoreIDClaim && f.Severity != severityBlocker {
			t.Errorf("cross-store finding severity = %q, want blocker", f.Severity)
		}
	}
}

// TestCrossStoreIDClaimLeavesRealRecordsAlone: a genuine ADR sitting in its own
// store claims its own id with an accepted status, and must never fire.
func TestCrossStoreIDClaimLeavesRealRecordsAlone(t *testing.T) {
	root := t.TempDir()
	writeTakenADR(t, root, "0023-transport-agnostic-core.md", "adr-23", "ADR-23: Transport-agnostic core")
	writeFile(t, root, filepath.Join(".abcd", "development", "decisions", "adrs", "0024-second.md"),
		"---\nid: adr-24\n---\n\n# ADR-24: Second\n\n## Status\n\nAccepted\n")

	fs, err := Lint(crossStoreCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleCrossStoreIDClaim); n != 0 {
		t.Fatalf("expected records inside their own store to be left alone; got %d: %+v", n, fs)
	}
}

// TestCrossStoreIDClaimNeedsBothSignals is the grandfathering criterion, stated
// as the fire condition rather than as a path list: an undated Phase 0 note
// whose FILENAME reads like a record id, and a note that names a taken id in
// prose, are each missing one half of the pair and neither fires.
func TestCrossStoreIDClaimNeedsBothSignals(t *testing.T) {
	root := t.TempDir()
	writeTakenADR(t, root, "0023-transport-agnostic-core.md", "adr-23", "Transport-agnostic core")

	// Phase 0 note: the filename looks like an ordinal, the heading claims no id,
	// and there is no decision shape.
	writeFile(t, root, filepath.Join("notes", "01-harness-interface.md"),
		"# Harness interface\n\nA Phase 0 note. It mentions adr-23 in prose.\n")
	// A decision-shaped document that claims no id at all.
	writeFile(t, root, filepath.Join("notes", "meeting.md"),
		"# Wednesday review\n\n## Status\n\nAccepted\n")
	// An id-claiming heading with no decision shape anywhere in the body.
	writeFile(t, root, filepath.Join("notes", "adr-23-summary.md"),
		"# adr-23 in one paragraph\n\nA reading note, not a decision.\n")
	// The corpus's own commonest shape, and the one this rule must never fire on:
	// a design plan headed with the record it plans FOR, carrying a status of its
	// own that is not a record lifecycle state.
	writeFile(t, root, filepath.Join(".abcd", "development", "plans", "2026-07-11-adr-23-options.md"),
		"# adr-23 transport core — design options (STOP for sign-off)\n\n"+
			"**Status:** SIGNED OFF 2026-07-11. The maintainer chose Option C.\n")

	fs, err := Lint(crossStoreCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleCrossStoreIDClaim); n != 0 {
		t.Fatalf("expected neither half alone to fire; got %d: %+v", n, fs)
	}
}

// TestCrossStoreIDClaimUntakenIDDoesNotFire pins the baseline half: the claim is
// weighed against what the corpus HOLDS, so a decision-shaped note claiming an id
// no record has taken is not this rule's finding.
func TestCrossStoreIDClaimUntakenIDDoesNotFire(t *testing.T) {
	root := t.TempDir()
	writeTakenADR(t, root, "0023-transport-agnostic-core.md", "adr-23", "Transport-agnostic core")
	writeFile(t, root, filepath.Join("research", "notes", "zz-probe.md"),
		strings.Replace(probeNote, "ADR-23", "ADR-99", 1))

	fs, err := Lint(crossStoreCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleCrossStoreIDClaim); n != 0 {
		t.Fatalf("expected an untaken id to be out of scope; got %d: %+v", n, fs)
	}
}

// TestCrossStoreIDClaimReachesEveryStore pins that the detector is not
// ADR-shaped: an intent id claimed outside the intents store fires the same way.
func TestCrossStoreIDClaimReachesEveryStore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, filepath.Join(".abcd", "development", "intents", "planned", "itd-7-real.md"),
		"---\nid: itd-7\n---\n\n# The real intent\n")
	writeFile(t, root, filepath.Join("docs", "notes", "duplicate.md"),
		"# itd-7: a second intent with the same handle\n\nStatus: Accepted\n")

	fs, err := Lint(crossStoreCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleCrossStoreIDClaim); n != 1 {
		t.Fatalf("expected the outside-store intent claim to fire, got %d: %+v", n, fs)
	}
	if !messageContains(fs, "itd-7") {
		t.Errorf("expected the message to name the claimed id; got %+v", fs)
	}
}

// TestCrossStoreIDClaimIgnoresFencedStatus pins that a fenced example — a
// document QUOTING the shape, which is how this rule gets documented — is not
// read as the document's own decision shape.
func TestCrossStoreIDClaimIgnoresFencedStatus(t *testing.T) {
	root := t.TempDir()
	writeTakenADR(t, root, "0023-transport-agnostic-core.md", "adr-23", "Transport-agnostic core")
	writeFile(t, root, filepath.Join("docs", "explainer.md"),
		"# ADR-23 is the example this page explains\n\n```\n## Status\n\nAccepted\n```\n")

	fs, err := Lint(crossStoreCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleCrossStoreIDClaim); n != 0 {
		t.Fatalf("expected a fenced example not to count as a decision shape; got %d: %+v", n, fs)
	}
}
