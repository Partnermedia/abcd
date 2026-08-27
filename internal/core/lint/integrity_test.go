package lint

import (
	"path/filepath"
	"strings"
	"testing"
)

// GitHub #333: the markdown-discovery filters tested the `.md` extension
// case-sensitively, so renaming a record to `.MD` made it INVISIBLE to every
// blocking rule (exit 0) while the lifeboat still packed it as a live record. A
// record whose only defect is the extension casing must still be discovered and
// gated, not skipped.
func TestRecordSchemaGatesUppercaseExtension(t *testing.T) {
	root := t.TempDir()
	// A record in a declared bucket whose ONLY defect is the extension casing.
	// Case-sensitive discovery dropped it silently; it must be discovered and
	// reported as a malformed filename (the `.md` contract is lowercase).
	writeFile(t, root, "rec/intents/planned/itd-9999-probe.MD", "---\nid: itd-9999\nkind: standalone\nspec_id: null\n---\n# probe\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	rel := filepath.Join("rec", "intents", "planned", "itd-9999-probe.MD")
	if !findingWith(fs, rel, ruleRecordSchema, "not a well-formed") {
		t.Fatalf("a `.MD` record must be discovered and gated, not skipped: %+v", fs)
	}
}

// GitHub #333: stray_root_docs also failed open on a case-flipped extension —
// NOTES.MD at the repo root slipped the allowlist gate that NOTES.md is caught
// by.
func TestStrayRootDocsGatesUppercaseExtension(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "NOTES.MD", "# stray\n")
	cfg := Config{
		Rules: map[string]RuleConfig{
			"stray_root_docs": {Enabled: true, Severity: "blocker",
				Allowlist: []string{"README", "AGENTS"}},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(fs, "NOTES.MD", "stray_root_docs", 0) {
		t.Fatalf("a `.MD` stray root doc must be gated, not skipped: %+v", fs)
	}
}

// GitHub #357: record-lint read frontmatter with the lenient canonical scanner
// (first-occurrence-wins) while capture's strict parser hard-rejects a
// duplicated top-level key. A duplicate `impact:` therefore shipped lint-green
// while silencing the armed issue_impact_valid blocker and dropping the record
// out of capture's corpus. The gate must reject a duplicated key the way its
// consumer does.
func TestRecordSchemaFlagsDuplicateFrontmatterKey(t *testing.T) {
	root := t.TempDir()
	// A duplicated top-level key on a record with no other defect. Lenient
	// first-wins hid it; capture refuses it.
	writeFile(t, root, "rec/decisions/adrs/0012-issue-ledger.md",
		"---\nid: adr-12\nsupersedes: null\nsupersedes: null\nsuperseded_by: null\n---\n# ADR-12\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	rel := filepath.Join("rec", "decisions", "adrs", "0012-issue-ledger.md")
	if !findingWith(fs, rel, ruleRecordSchema, "duplicate") {
		t.Fatalf("a duplicated top-level frontmatter key must be gated: %+v", fs)
	}
}

// GitHub #357: `id : value` (one space before the colon) was read by the lenient
// scanner's keyRe as NO id at all, so checkRecordFilename hit its absent-id early
// return and the filename↔id blocker never fired — while capture parses the id
// fine and reports the mismatch. The lint reader must see the key its consumer
// sees, so the armed blocker cannot be disarmed by a stray space.
func TestRecordSchemaFilenameIDSeesSpaceBeforeColon(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/intents/planned/itd-4-capture.md",
		"---\nid : itd-999\nkind: standalone\nspec_id: null\n---\n# x\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	rel := filepath.Join("rec", "intents", "planned", "itd-4-capture.md")
	if !findingWith(fs, rel, ruleRecordSchema, "filename claims id 'itd-4'") {
		t.Fatalf("a space before the colon must not disarm the filename↔id blocker: %+v", fs)
	}
}

// GitHub #360: a `roots` entry pointing at a directory that does not exist made
// the per-file lint families walk nothing and report zero findings at exit 0 —
// silently disarming every blocker for that tree. A missing configured root is
// misconfiguration and must fail loudly, not pass vacuously.
func TestLintMissingRootIsLoud(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "docs/ok.md", "# ok\n")
	cfg := Config{
		Roots: []string{"documentation"}, // the tree lives in docs/, not documentation/
		BannedTokens: []BannedToken{{
			ID: "t1", Pattern: "SEKRET", Message: "no", Severity: "blocker",
			Successor: "x", AllowContext: []string{"never"},
		}},
	}
	_, err := Lint(cfg, root)
	if err == nil {
		t.Fatal("a configured root that does not exist must fail loudly, not pass with zero findings")
	}
	if !strings.Contains(err.Error(), "documentation") {
		t.Errorf("the loud error should name the unresolved root: %v", err)
	}
}
