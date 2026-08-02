package lint

import (
	"path/filepath"
	"strings"
	"testing"
)

// schemaStores is the store map every record_schema fixture below shares.
func schemaStores() map[string]string {
	return map[string]string{
		"adr": "rec/decisions/adrs",
		"itd": "rec/intents",
		"spc": "rec/specs",
		"iss": "work/issues",
	}
}

func schemaConfig() Config {
	return Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			ruleRecordSchema: {Enabled: true, Severity: severityBlocker, RecordStores: schemaStores()},
		},
	}
}

// findingWith reports whether a finding on file:line quotes the given substring.
func findingWith(fs []Finding, file, ruleID string, substr string) bool {
	for _, f := range fs {
		if f.File == file && f.RuleID == ruleID && strings.Contains(f.Message, substr) {
			return true
		}
	}
	return false
}

// Corpus fixture (a): the shape of the real adr-6 defect — two ADRs name adr-6 as
// a live design input in related_adrs, no 0006-*.md is in the store, and no record
// declares it superseded. The control in the same fixture is adr-8, which is
// equally absent but IS declared retired by its successor's `supersedes` (the ADR
// store's documented pruning convention), and must stay silent.
func TestRecordSchemaPhantomCrossReference(t *testing.T) {
	root := t.TempDir()
	adrs := "rec/decisions/adrs"
	writeFile(t, root, adrs+"/0022-adapters.md", "---\nid: adr-22\nsupersedes: [adr-14]\nsuperseded_by: null\nrelated_adrs: []\n---\n# ADR-22\n")
	writeFile(t, root, adrs+"/0025-host-delegated.md", "---\nid: adr-25\nsupersedes: [adr-8]\nsuperseded_by: null\nrelated_adrs: [adr-6, adr-22]\n---\n# ADR-25\n")
	writeFile(t, root, adrs+"/0029-transcripts.md", "---\nid: adr-29\nsupersedes: null\nsuperseded_by: null\nrelated_adrs: [adr-6, adr-22]\n---\n# ADR-29\n")
	// A retired handle cited from an intent must resolve through the successor's
	// declaration, not a file.
	writeFile(t, root, "rec/intents/drafts/itd-17-tracking.md", "---\nid: itd-17\nkind: null\nspec_id: null\nrelated_adrs: [adr-8, adr-25]\n---\n# draft\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleRecordSchema); n != 2 {
		t.Fatalf("expected exactly 2 record_schema findings (the two adr-6 citations), got %d: %+v", n, fs)
	}
	for _, f := range []string{"0025-host-delegated.md", "0029-transcripts.md"} {
		if !findingWith(fs, filepath.Join(adrs, f), ruleRecordSchema, "'adr-6'") {
			t.Errorf("expected an unresolved adr-6 finding on %s: %+v", f, fs)
		}
	}
}

// Corpus fixture (b): the shape of the real adr-12 / adr-32 defect — a supersession
// declared from one side only. Both directions are checked, because either half
// alone leaves the record contradicting itself about which decision is in force.
func TestRecordSchemaSupersessionIsBidirectional(t *testing.T) {
	adrs := "rec/decisions/adrs"

	t.Run("forward half only", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-12\nsupersedes: null\nsuperseded_by: adr-32\n---\n# ADR-12\n")
		writeFile(t, root, adrs+"/0032-working-tier.md", "---\nid: adr-32\nsupersedes: null\nsuperseded_by: null\n---\n# ADR-32\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRule(fs, ruleRecordSchema); n != 1 {
			t.Fatalf("expected 1 one-way supersession finding, got %d: %+v", n, fs)
		}
		if !findingWith(fs, filepath.Join(adrs, "0012-issue-ledger.md"), ruleRecordSchema, "one-way supersession") {
			t.Errorf("expected the finding on the declaring record: %+v", fs)
		}
	})

	t.Run("reverse half only", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-12\nsupersedes: null\nsuperseded_by: null\n---\n# ADR-12\n")
		writeFile(t, root, adrs+"/0032-working-tier.md", "---\nid: adr-32\nsupersedes: adr-12\nsuperseded_by: null\n---\n# ADR-32\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if !findingWith(fs, filepath.Join(adrs, "0032-working-tier.md"), ruleRecordSchema, "one-way supersession") {
			t.Fatalf("expected a one-way finding on the supersedes side: %+v", fs)
		}
	})

	t.Run("both halves present", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-12\nsupersedes: null\nsuperseded_by: adr-32\n---\n# ADR-12\n")
		writeFile(t, root, adrs+"/0032-working-tier.md", "---\nid: adr-32\nsupersedes: adr-12\nsuperseded_by: null\n---\n# ADR-32\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRule(fs, ruleRecordSchema); n != 0 {
			t.Fatalf("a fully-linked pair must be clean, got %d: %+v", n, fs)
		}
	})
}

// A supersession may cross stores: an ADR that redecides the question an intent
// rested on retires that intent (the real itd-47 → adr-22, itd-49 → adr-26). The
// bidirectional rule then binds the ADR to name the intent back.
func TestRecordSchemaSupersessionCrossesStores(t *testing.T) {
	root := t.TempDir()
	adrs := "rec/decisions/adrs"
	sup := "rec/intents/superseded"
	writeFile(t, root, adrs+"/0022-adapters.md", "---\nid: adr-22\nsupersedes: [itd-47]\nsuperseded_by: null\n---\n# ADR-22\n")
	writeFile(t, root, adrs+"/0026-spec-layer.md", "---\nid: adr-26\nsupersedes: null\nsuperseded_by: null\n---\n# ADR-26\n")
	writeFile(t, root, sup+"/itd-47-oracle-gates.md", "---\nid: itd-47\nkind: standalone\nsuperseded_by: adr-22\n---\n# ok\n")
	writeFile(t, root, sup+"/itd-49-flow-drift.md", "---\nid: itd-49\nkind: standalone\nsuperseded_by: adr-26\n---\n# bad\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleRecordSchema); n != 1 {
		t.Fatalf("expected 1 finding (adr-26 does not name itd-49 back), got %d: %+v", n, fs)
	}
	if !findingWith(fs, filepath.Join(sup, "itd-49-flow-drift.md"), ruleRecordSchema, "one-way supersession") {
		t.Errorf("expected the one-way finding on itd-49: %+v", fs)
	}
}

// A record answers to exactly one handle: the id its filename claims and the id
// its frontmatter declares must agree, because half the rules key on one and half
// on the other.
func TestRecordSchemaFilenameMatchesID(t *testing.T) {
	root := t.TempDir()
	adrs := "rec/decisions/adrs"
	writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-21\n---\n# ADR-12\n")
	writeFile(t, root, adrs+"/0013-memory.md", "---\nid: adr-13\n---\n# ADR-13\n")
	// A record with no id at all is a different (larger) schema question, not this
	// rule's: it must stay silent rather than invent an agreement to check.
	writeFile(t, root, adrs+"/0007-grill.md", "# ADR-7\n")
	writeFile(t, root, "rec/intents/planned/itd-4-capture.md", "---\nid: itd-5\nkind: standalone\nspec_id: null\n---\n# x\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleRecordSchema); n != 2 {
		t.Fatalf("expected 2 filename/id findings, got %d: %+v", n, fs)
	}
	if !findingWith(fs, filepath.Join(adrs, "0012-issue-ledger.md"), ruleRecordSchema, "filename claims id 'adr-12'") {
		t.Errorf("expected the ADR mismatch: %+v", fs)
	}
	if !findingWith(fs, filepath.Join("rec", "intents", "planned", "itd-4-capture.md"), ruleRecordSchema, "filename claims id 'itd-4'") {
		t.Errorf("expected the intent mismatch: %+v", fs)
	}
}

// Every lifecycle directory is enumerated, so a bucket nobody declared cannot hold
// records that no rule ever reads — and a record cannot escape its lifecycle by
// sitting in the store root.
func TestRecordSchemaLifecycleDirectoriesAreDeclared(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/intents/shipped/itd-1-good.md", "---\nid: itd-1\nkind: standalone\nspec_id: spc-1-x\n---\n# ok\n")
	writeFile(t, root, "rec/intents/archived/itd-2-hidden.md", "---\nid: itd-2\nkind: standalone\n---\n# hidden\n")
	writeFile(t, root, "rec/intents/itd-3-loose.md", "---\nid: itd-3\nkind: standalone\n---\n# loose\n")
	writeFile(t, root, "rec/intents/README.md", "# intents\n")
	writeFile(t, root, "rec/specs/open/notes.md", "# not a spec filename\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !findingWith(fs, filepath.Join("rec", "intents", "archived"), ruleRecordSchema, "is not a declared intent bucket") {
		t.Errorf("expected an undeclared-bucket finding: %+v", fs)
	}
	if !findingWith(fs, filepath.Join("rec", "intents", "itd-3-loose.md"), ruleRecordSchema, "rather than a lifecycle bucket") {
		t.Errorf("expected a store-root record finding: %+v", fs)
	}
	if !findingWith(fs, filepath.Join("rec", "specs", "open", "notes.md"), ruleRecordSchema, "not a well-formed spec filename") {
		t.Errorf("expected a malformed-filename finding: %+v", fs)
	}
	// The README and the well-formed record are clean.
	if n := countRule(fs, ruleRecordSchema); n != 3 {
		t.Fatalf("expected exactly 3 walk findings, got %d: %+v", n, fs)
	}
}

// A FLAT store declares no buckets, so any subdirectory of it is undeclared. Left
// unsaid, a directory one level down hides every record inside it from every check
// in the rule — the same escape the bucketed branch closes, in the branch that
// looked like it had nothing to close.
func TestRecordSchemaFlatStoreHasNoSubdirectories(t *testing.T) {
	root := t.TempDir()
	adrs := "rec/decisions/adrs"
	writeFile(t, root, adrs+"/0013-memory.md", "---\nid: adr-13\n---\n# ADR-13\n")
	// Three defects, each of a class the rule checks, all one directory down.
	writeFile(t, root, adrs+"/archive/0012-issue-ledger.md", "---\nid: adr-21\nsuperseded_by: adr-9999\n---\n# ADR-12\n")
	writeFile(t, root, adrs+"/archive/loose-notes.md", "# notes\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !findingWith(fs, filepath.Join(adrs, "archive"), ruleRecordSchema, "store is flat, so subdirectory 'archive' is undeclared") {
		t.Fatalf("expected an undeclared-subdirectory finding on the flat store: %+v", fs)
	}
}

// A DECLARED bucket holds its records directly, so a directory inside one is the
// same escape as a directory inside a flat store — every check in the rule stops
// at it. Closing it only for the store roots left three of the four stores open.
func TestRecordSchemaBucketHasNoSubdirectories(t *testing.T) {
	root := t.TempDir()
	shipped := "rec/intents/shipped"
	writeFile(t, root, shipped+"/itd-1-good.md", "---\nid: itd-1\nkind: standalone\nspec_id: spc-1-x\n---\n# ok\n")
	// Two defects, each of a class the rule checks, one directory down.
	writeFile(t, root, shipped+"/archive/itd-2-bad.md", "---\nid: itd-7\nkind: standalone\nsuperseded_by: itd-9999\n---\n# hidden\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !findingWith(fs, filepath.Join(shipped, "archive"), ruleRecordSchema, "holds records directly, so subdirectory 'archive' is undeclared") {
		t.Fatalf("expected an undeclared-subdirectory finding inside the bucket: %+v", fs)
	}
	if n := countRule(fs, ruleRecordSchema); n != 1 {
		t.Fatalf("expected exactly 1 finding, got %d: %+v", n, fs)
	}
}

// The spec store is indexed by this rule, so a spec handle must PARSE as one. A
// prefix the rule indexes but the pattern omits reads as "no handle at all": the
// forward direction becomes a false blocker and the reverse is never checked.
func TestRecordSchemaResolvesSpecHandles(t *testing.T) {
	specs := "rec/specs"

	t.Run("a bidirectional spec supersession is clean", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, specs+"/closed/spc-5-old.md", "---\nid: spc-5\nintent: itd-1\nslug: old\nsuperseded_by: spc-6\n---\n# x\n")
		writeFile(t, root, specs+"/open/spc-6-new.md", "---\nid: spc-6\nintent: itd-1\nslug: new\nsupersedes: [spc-5]\n---\n# x\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRule(fs, ruleRecordSchema); n != 0 {
			t.Fatalf("a fully-linked spec pair must be clean, got %d: %+v", n, fs)
		}
	})

	t.Run("a one-way spec supersession is caught", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, specs+"/closed/spc-5-old.md", "---\nid: spc-5\nintent: itd-1\nslug: old\nsuperseded_by: spc-6\n---\n# x\n")
		writeFile(t, root, specs+"/open/spc-6-new.md", "---\nid: spc-6\nintent: itd-1\nslug: new\nsupersedes: null\n---\n# x\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if !findingWith(fs, filepath.Join(specs, "closed", "spc-5-old.md"), ruleRecordSchema, "one-way supersession") {
			t.Fatalf("expected the reverse direction to be checked for specs: %+v", fs)
		}
	})
}

// An empty flow sequence is this record's house spelling for an empty list, so it
// says "nothing here" — not "a value that is not a handle".
func TestRecordSchemaEmptyListIsAbsent(t *testing.T) {
	root := t.TempDir()
	adrs := "rec/decisions/adrs"
	writeFile(t, root, adrs+"/0001-model.md", "---\nid: adr-1\nsupersedes: []\nsuperseded_by: []\nrelated_adrs: []\n---\n# ADR-1\n")
	writeFile(t, root, adrs+"/0002-kinds.md", "---\nid: adr-2\nsuperseded_by: [ ]\n---\n# ADR-2\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleRecordSchema); n != 0 {
		t.Fatalf("an empty list is an absence, got %d: %+v", n, fs)
	}
}

// A zero-padded id and its bare spelling are one handle — the rule says so about
// its cross-references, and must not contradict itself about its filenames.
func TestRecordSchemaFilenameIDComparesNumerically(t *testing.T) {
	root := t.TempDir()
	adrs := "rec/decisions/adrs"
	writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-0012\n---\n# ADR-12\n")
	writeFile(t, root, adrs+"/0013-memory.md", "---\nid: ADR-13\n---\n# ADR-13\n")
	// Still a mismatch: a different number, and a value that is no handle at all.
	writeFile(t, root, adrs+"/0021-go.md", "---\nid: adr-0022\n---\n# ADR-21\n")
	writeFile(t, root, adrs+"/0023-core.md", "---\nid: core\n---\n# ADR-23\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleRecordSchema); n != 2 {
		t.Fatalf("expected 2 filename/id findings, got %d: %+v", n, fs)
	}
	for _, f := range []string{"0021-go.md", "0023-core.md"} {
		if !findingWith(fs, filepath.Join(adrs, f), ruleRecordSchema, "filename claims id") {
			t.Errorf("expected a mismatch on %s: %+v", f, fs)
		}
	}
}

// The retirement escape hatch is bounded by what the store has actually issued.
// Unbounded it is self-attesting: one record naming a phantom in `supersedes`
// would make that phantom resolvable in every other record's cross-references.
func TestRecordSchemaRetirementIsBoundedByAllocation(t *testing.T) {
	adrs := "rec/decisions/adrs"

	t.Run("an id the store never issued", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, adrs+"/0025-host-delegated.md", "---\nid: adr-25\nsupersedes: [adr-9999]\nrelated_adrs: [adr-9999]\n---\n# ADR-25\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRule(fs, ruleRecordSchema); n != 2 {
			t.Fatalf("expected 2 findings (the declaration and the reference it must not legitimise), got %d: %+v", n, fs)
		}
		if !findingWith(fs, filepath.Join(adrs, "0025-host-delegated.md"), ruleRecordSchema, "nor an id the adr store has issued") {
			t.Errorf("expected the supersedes declaration itself to be refused: %+v", fs)
		}
		if !findingWith(fs, filepath.Join(adrs, "0025-host-delegated.md"), ruleRecordSchema, "related_adrs names 'adr-9999'") {
			t.Errorf("expected the cross-reference to stay unresolved: %+v", fs)
		}
	})

	t.Run("a genuinely pruned id", func(t *testing.T) {
		root := t.TempDir()
		// adr-8 has no file: it was pruned when adr-25 landed. adr-25 says so, and
		// 8 is below the store's high-water mark, so citing it elsewhere resolves.
		writeFile(t, root, adrs+"/0025-host-delegated.md", "---\nid: adr-25\nsupersedes: [adr-8]\n---\n# ADR-25\n")
		writeFile(t, root, "rec/intents/drafts/itd-17-tracking.md", "---\nid: itd-17\nkind: null\nspec_id: null\nrelated_adrs: [adr-8]\n---\n# draft\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRule(fs, ruleRecordSchema); n != 0 {
			t.Fatalf("a pruned id below the high-water mark must resolve, got %d: %+v", n, fs)
		}
	})
}

// YAML spells a list two ways and the record uses both. Reading only the same-line
// form makes a block sequence look empty — and an empty read here does not merely
// under-report, it makes the bidirectional check assert that ANOTHER file omits a
// link that file plainly carries.
func TestRecordSchemaReadsBlockSequences(t *testing.T) {
	adrs := "rec/decisions/adrs"

	t.Run("a linked pair written in block form is clean", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-12\nsuperseded_by: adr-32\n---\n# ADR-12\n")
		writeFile(t, root, adrs+"/0032-working-tier.md", "---\nid: adr-32\nsupersedes:\n  - adr-12\nrelated_adrs:\n  - adr-12\n---\n# ADR-32\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRule(fs, ruleRecordSchema); n != 0 {
			t.Fatalf("block-sequence links must read as links, got %d: %+v", n, fs)
		}
	})

	t.Run("a one-way pair written in block form is still caught", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-12\nsuperseded_by:\n  - adr-32\n---\n# ADR-12\n")
		writeFile(t, root, adrs+"/0032-working-tier.md", "---\nid: adr-32\nsupersedes: null\n---\n# ADR-32\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if !findingWith(fs, filepath.Join(adrs, "0012-issue-ledger.md"), ruleRecordSchema, "one-way supersession") {
			t.Fatalf("expected the one-way finding to survive the block spelling: %+v", fs)
		}
	})

	// YAML nests a block sequence under a mapping key with NO extra indentation:
	// `supersedes:\n- adr-12` is the same list as `supersedes:\n  - adr-12`, and the
	// record writes both. Reading only the indented spelling makes the column-0 one
	// look empty — the same false assertion about ANOTHER file, reached by a purely
	// cosmetic difference in how the list was typed.
	t.Run("the column-0 spelling reads the same as the indented one", func(t *testing.T) {
		lint := func(t *testing.T, supersedes, related string) []Finding {
			t.Helper()
			root := t.TempDir()
			writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-12\nsuperseded_by: adr-32\n---\n# ADR-12\n")
			writeFile(t, root, adrs+"/0032-working-tier.md",
				"---\nid: adr-32\nsupersedes:\n"+supersedes+"related_adrs:\n"+related+"---\n# ADR-32\n")
			fs, err := Lint(schemaConfig(), root)
			if err != nil {
				t.Fatal(err)
			}
			return fs
		}

		indented := lint(t, "  - adr-12\n", "  - adr-12\n")
		column0 := lint(t, "- adr-12\n", "- adr-12\n")
		if n := countRule(indented, ruleRecordSchema); n != 0 {
			t.Fatalf("the indented spelling must be clean, got %d: %+v", n, indented)
		}
		if n := countRule(column0, ruleRecordSchema); n != 0 {
			t.Fatalf("the column-0 spelling is the same list, got %d: %+v", n, column0)
		}
	})

	t.Run("a one-way pair written at column 0 is still caught", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, adrs+"/0012-issue-ledger.md", "---\nid: adr-12\nsuperseded_by:\n- adr-32\n---\n# ADR-12\n")
		writeFile(t, root, adrs+"/0032-working-tier.md", "---\nid: adr-32\nsupersedes: null\n---\n# ADR-32\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if !findingWith(fs, filepath.Join(adrs, "0012-issue-ledger.md"), ruleRecordSchema, "one-way supersession") {
			t.Fatalf("expected the one-way finding to survive the column-0 spelling: %+v", fs)
		}
	})

	t.Run("an unrelated block key does not leak into the next", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, root, "rec/intents/superseded/itd-31-cross-doc.md",
			"---\nid: itd-31\nkind: standalone\nreclassification_history:\n  - { date: 2026-05-27, reason: \"absorbed by itd-48\" }\nsupersedes: null\nsuperseded_by: itd-48\n---\n# x\n")
		writeFile(t, root, "rec/intents/shipped/itd-48-successor.md", "---\nid: itd-48\nkind: standalone\nspec_id: spc-2-x\nsupersedes:\n  - itd-31\n---\n# ok\n")
		fs, err := Lint(schemaConfig(), root)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRule(fs, ruleRecordSchema); n != 0 {
			t.Fatalf("expected a clean pair, got %d: %+v", n, fs)
		}
	})
}

// A dot-directory is tooling state, not a lifecycle the record authored, so it is
// not an undeclared bucket.
func TestRecordSchemaIgnoresDotDirectories(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/intents/shipped/itd-1-good.md", "---\nid: itd-1\nkind: standalone\nspec_id: spc-1-x\n---\n# ok\n")
	writeFile(t, root, "rec/intents/.obsidian/workspace.json", "{}\n")
	writeFile(t, root, "rec/decisions/adrs/.vscode/settings.json", "{}\n")

	fs, err := Lint(schemaConfig(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleRecordSchema); n != 0 {
		t.Fatalf("dot-directories are not undeclared buckets, got %d: %+v", n, fs)
	}
}

// Corpus fixture (c)+(d): the superseded/ exemption is a CONTENT exemption. A
// record in an exempt bucket still answers the schema rules — which is the only
// reason the missing supersession on the two real superseded intents is visible at
// all — while the content-authoring ban above it stays suppressed.
func TestSupersededBucketExemptsContentNotSchema(t *testing.T) {
	root := t.TempDir()
	sup := "rec/intents/superseded"
	writeFile(t, root, "rec/intents/shipped/itd-48-successor.md", "---\nid: itd-48\nkind: standalone\nspec_id: spc-2-x\nsupersedes: itd-31\n---\n# ok\n")
	// Blockquote prose only, exactly as the two real records carried it.
	writeFile(t, root, sup+"/itd-47-oracle-gates.md",
		"---\nid: itd-47\nkind: standalone\n---\n\n> **Superseded by ADR-22** — intent_lint.py is gone.\n")
	writeFile(t, root, sup+"/itd-31-cross-doc.md",
		"---\nid: itd-31\nkind: standalone\nsuperseded_by: itd-48\n---\n\n> intent_lint.py is gone.\n")

	cfg := Config{
		Roots: []string{"rec"},
		BannedTokens: []BannedToken{
			{ID: "py", Pattern: `intent_lint\.py`, Message: "no python name", Severity: severityBlocker,
				Successor: "internal/core/lint", AllowContext: []string{"historical"}},
		},
		Rules: map[string]RuleConfig{
			"intent_lifecycle": {Enabled: true, Severity: severityBlocker, IntentsDir: "intents"},
			ruleRecordSchema:   {Enabled: true, Severity: severityBlocker, RecordStores: schemaStores()},
		},
		ExemptPaths: []string{filepath.Join("rec", "intents", "superseded")},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	// Content stays exempt: neither superseded record trips the banned token.
	if n := countRule(fs, "py"); n != 0 {
		t.Errorf("the superseded bucket must stay exempt from content bans, got %d: %+v", n, fs)
	}
	// Schema does not: the record with prose but no superseded_by field is caught.
	if !hasFinding(fs, filepath.Join(sup, "itd-47-oracle-gates.md"), "intent_lifecycle", 1) {
		t.Errorf("expected a lifecycle finding on the superseded record with no supersession field: %+v", fs)
	}
	// And the well-formed superseded record stays clean.
	for _, f := range fs {
		if filepath.Base(f.File) == "itd-31-cross-doc.md" {
			t.Errorf("unexpected finding on the well-formed superseded record: %+v", f)
		}
	}
}

// A record whose lifecycle bucket is superseded/ must name its successor; an ADR
// is a legal successor, so the lifecycle rule accepts an adr-N target and only
// resolves an intent one against the intent tree.
func TestIntentLifecycleAcceptsAnADRSuccessor(t *testing.T) {
	root := t.TempDir()
	base := "rec/intents"
	writeFile(t, root, base+"/superseded/itd-47-adr-successor.md", "---\nid: itd-47\nkind: standalone\nsuperseded_by: adr-22\n---\n# ok\n")
	writeFile(t, root, base+"/superseded/itd-49-no-successor.md", "---\nid: itd-49\nkind: standalone\n---\n# bad\n")
	writeFile(t, root, base+"/superseded/itd-50-ghost.md", "---\nid: itd-50\nkind: standalone\nsuperseded_by: itd-999\n---\n# bad\n")

	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"intent_lifecycle": {Enabled: true, Severity: severityBlocker, IntentsDir: "intents"},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "intent_lifecycle"); n != 2 {
		t.Fatalf("expected 2 lifecycle findings, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join(base, "superseded", "itd-49-no-successor.md"), "intent_lifecycle", 1) {
		t.Errorf("expected a finding for the missing superseded_by: %+v", fs)
	}
	if !findingWith(fs, filepath.Join(base, "superseded", "itd-50-ghost.md"), "intent_lifecycle", "does not exist in any bucket") {
		t.Errorf("expected a finding for the unresolvable intent successor: %+v", fs)
	}
}

// TestRecordSchemaGuardsTheRealRecord loads the committed record-lint config and
// asserts the rule is armed as a blocker over the real stores. Deleting the rule,
// dropping a store, or downgrading its severity fails here rather than silently
// re-opening the drift iss-39 closed.
func TestRecordSchemaGuardsTheRealRecord(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join("..", "..", "..", ".abcd", "record-lint.json"))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	rc, ok := cfg.Rules[ruleRecordSchema]
	if !ok || !rc.Enabled {
		t.Fatalf("record_schema must be enabled in the committed config: %+v", rc)
	}
	if rc.Severity != severityBlocker {
		t.Errorf("record_schema severity = %q, want blocker", rc.Severity)
	}
	for _, prefix := range []string{"adr", "itd", "spc", "iss"} {
		if rc.RecordStores[prefix] == "" {
			t.Errorf("record_schema declares no store for %q", prefix)
		}
	}
}
