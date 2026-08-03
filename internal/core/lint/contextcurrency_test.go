package lint

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// repoRoot locates this repository's root from the package directory, so a test
// can run the shipped configuration against the real record.
const repoRootFromPackage = "../../.."

// contextCurrencyCfg builds a one-rule config pointed at a fixture tree. The
// stores are named exactly as the live config names them, so a fixture that
// passes here is a fixture the shipped configuration would also judge.
func contextCurrencyCfg() Config {
	return Config{Rules: map[string]RuleConfig{
		ruleContextCitationCurrency: {
			Enabled: true, Severity: severityBlocker,
			RecordStores: map[string]string{
				"adr": ".abcd/development/decisions/adrs",
				"itd": ".abcd/development/intents",
				"spc": ".abcd/development/specs",
				"iss": ".abcd/work/issues",
			},
		},
	}}
}

// writeIssue puts an issue in a ledger status directory; the directory IS the
// lifecycle state.
func writeIssue(t *testing.T, root, status, name string) {
	t.Helper()
	writeFile(t, root, filepath.Join(".abcd", "work", "issues", status, name),
		"---\nid: \""+issueIDOf(name)+"\"\n---\n\nfixture issue.\n")
}

// writeSpec puts a spec in a lifecycle bucket.
func writeSpec(t *testing.T, root, bucket, name string) {
	t.Helper()
	writeFile(t, root, filepath.Join(".abcd", "development", "specs", bucket, name),
		"---\nid: \""+specIDOf(name)+"\"\n---\n\nfixture spec.\n")
}

// writeADR puts an ADR in the flat decision store. The ADR store carries its
// lifecycle in frontmatter rather than in a directory, so status and
// superseded_by are the fixture's lifecycle knobs.
func writeADR(t *testing.T, root, name, id, status, supersededBy string) {
	t.Helper()
	writeFile(t, root, filepath.Join(".abcd", "development", "decisions", "adrs", name),
		"---\nid: "+id+"\nstatus: "+status+"\nsuperseded_by: "+supersededBy+"\n---\n\nfixture ADR.\n")
}

var specIDOfRe = regexp.MustCompile(`spc-\d+`)

func issueIDOf(name string) string { return issueIDRe.FindString(name) }
func specIDOf(name string) string  { return specIDOfRe.FindString(name) }

// writeContext writes a fixture CONTEXT.md whose sharp-edges section holds the
// given bullet lines.
func writeContext(t *testing.T, root string, bullets ...string) {
	t.Helper()
	body := []string{
		"# CONTEXT",
		"",
		"Shared orientation. Durable design truth lives in ../development/.",
		"",
		"## What this repo is",
		"",
		"A fixture.",
		"",
		"## Live constraints / sharp edges",
		"",
	}
	body = append(body, bullets...)
	writeFile(t, root, filepath.Join(".abcd", "work", "CONTEXT.md"), strings.Join(body, "\n")+"\n")
}

// liveCorpus populates one live record in each store, so a fixture always has a
// non-empty corpus and the fail-closed guard is not what a test is measuring.
func liveCorpus(t *testing.T, root string) {
	t.Helper()
	writeIssue(t, root, "open", "iss-42-record-orientation-currency.md")
	writeIntent(t, root, "planned", "itd-73-derived-versioning.md")
	writeSpec(t, root, "open", "spc-21-plugin-provisions-its-binary.md")
	writeADR(t, root, "0001-three-layer-mental-model.md", "adr-1", "accepted", "null")
}

// TestContextCitationCurrencyCatchesTerminalCitations reproduces the iss-42
// corpus shape in every store: the sharp-edges list grounds a live caveat in a
// record that has since reached the end of its lifecycle. The iss-35 bullet is
// the instance that motivated the rule — the trust caveat still called a shipped
// reconciliation "the open cross-check".
func TestContextCitationCurrencyCatchesTerminalCitations(t *testing.T) {
	root := t.TempDir()
	liveCorpus(t, root)
	writeIssue(t, root, "resolved", "iss-35-brief-surface-reconciliation.md")
	writeIssue(t, root, "wontfix", "iss-60-agent-git-revert-hazard.md")
	writeIntent(t, root, "shipped", "itd-88-lifeboat-coverage-experiment.md")
	writeIntent(t, root, "superseded", "itd-4-issue-ledger.md")
	writeSpec(t, root, "closed", "spc-3-lifeboat-coverage-experiment.md")
	writeADR(t, root, "0012-issue-ledger-live-vs-structured.md", "adr-12", "superseded", "adr-32")
	writeADR(t, root, "0032-issue-ledger-is-working-tier-data.md", "adr-32", "accepted", "null")
	writeContext(t, root,
		"- The brief-vs-surface reconciliation (iss-35 in the ledger) is the open cross-check.",
		"- The coverage experiment (itd-88) is still running.",
		"- The ledger shape is under review (adr-12).",
		"- The lifeboat spec (spc-3) is not settled.",
		"- Two live pointers that must stay quiet: iss-42 and itd-73.",
	)

	fs, err := Lint(contextCurrencyCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleContextCitationCurrency); n != 4 {
		t.Fatalf("expected the four terminal citations to fire, got %d: %+v", n, fs)
	}
	for _, want := range []string{"iss-35", "itd-88", "adr-12", "spc-3"} {
		if !messageContains(fs, want) {
			t.Errorf("expected a finding naming %s; got %+v", want, fs)
		}
	}
	// Line 11 is the first bullet of the section — the rule must anchor the
	// finding on the citing line so the reader is sent to the sentence to rewrite.
	if !hasFinding(fs, filepath.Join(".abcd", "work", "CONTEXT.md"), ruleContextCitationCurrency, 11) {
		t.Errorf("expected the iss-35 finding anchored on its own line; got %+v", fs)
	}
}

// TestContextCitationCurrencyStaysQuietOnLiveRecords is the precision half: the
// normal state of the file is bullets grounded in records that are still open,
// planned, or in force, and none of them may fire. A rule that cannot stay quiet
// on the healthy corpus gets disabled, and then it gates nothing.
func TestContextCitationCurrencyStaysQuietOnLiveRecords(t *testing.T) {
	root := t.TempDir()
	liveCorpus(t, root)
	writeIntent(t, root, "drafts", "itd-85-audit-verb.md")
	writeIntent(t, root, "disciplines", "itd-81-judge-calibration.md")
	writeContext(t, root,
		"- The orientation-currency work (iss-42) is open.",
		"- Derived versioning (itd-73) is planned; the audit verb (itd-85) is captured only.",
		"- Judge calibration (itd-81) is a standing discipline.",
		"- The plugin-provisioning spec (spc-21) is open; adr-1 is in force.",
	)

	fs, err := Lint(contextCurrencyCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleContextCitationCurrency); n != 0 {
		t.Fatalf("expected live citations to be silent, got %d: %+v", n, fs)
	}
}

// TestContextCitationCurrencyReadsOnlyTheSharpEdges pins the rule's scope. The
// file's own preamble explains what the document is and may legitimately name a
// closed record; so may a fenced example. Only the live-constraints bullets claim
// to describe what is true right now.
func TestContextCitationCurrencyReadsOnlyTheSharpEdges(t *testing.T) {
	root := t.TempDir()
	liveCorpus(t, root)
	writeIssue(t, root, "resolved", "iss-35-brief-surface-reconciliation.md")
	writeFile(t, root, filepath.Join(".abcd", "work", "CONTEXT.md"), strings.Join([]string{
		"# CONTEXT",
		"",
		"This file went status-free after iss-35 shipped its reconciliation.",
		"",
		"## Live constraints / sharp edges",
		"",
		"```md",
		"- an example bullet citing iss-35",
		"```",
		"",
		"- Nothing terminal is cited here: iss-42 is open.",
		"",
		"## Appendix",
		"",
		"- A later section may cite iss-35 freely.",
	}, "\n")+"\n")

	fs, err := Lint(contextCurrencyCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleContextCitationCurrency); n != 0 {
		t.Fatalf("expected the preamble, the fenced example, and the later section to be out of scope, got %d: %+v", n, fs)
	}
}

// TestContextCitationCurrencyFailsClosed covers the three ways an armed gate
// could check nothing at all: no target to read, no sharp-edges section in it,
// and a record corpus that resolves nothing (every citation would then read as
// "not a record", which is indistinguishable from a clean section).
func TestContextCitationCurrencyFailsClosed(t *testing.T) {
	t.Run("target missing", func(t *testing.T) {
		root := t.TempDir()
		liveCorpus(t, root)
		if _, err := Lint(contextCurrencyCfg(), root); err == nil {
			t.Fatal("expected an error when the configured target is absent")
		}
	})

	t.Run("section missing", func(t *testing.T) {
		root := t.TempDir()
		liveCorpus(t, root)
		writeIssue(t, root, "resolved", "iss-35-brief-surface-reconciliation.md")
		writeFile(t, root, filepath.Join(".abcd", "work", "CONTEXT.md"),
			"# CONTEXT\n\n## Sharp bits\n\n- The reconciliation (iss-35) is the open cross-check.\n")

		fs, err := Lint(contextCurrencyCfg(), root)
		if err != nil {
			t.Fatal(err)
		}
		if n := countRule(fs, ruleContextCitationCurrency); n != 1 {
			t.Fatalf("expected a renamed section to be reported rather than to disarm the gate, got %d: %+v", n, fs)
		}
	})

	t.Run("record stores hold no records", func(t *testing.T) {
		root := t.TempDir()
		writeContext(t, root, "- The reconciliation (iss-35) is the open cross-check.")
		if _, err := Lint(contextCurrencyCfg(), root); err == nil {
			t.Fatal("expected an error when the configured record stores resolve no records")
		}
	})
}

// TestContextCitationCurrencyGatesTheLiveRecord runs the rule with the shipped
// configuration against the real repository, so the gate that is armed in
// .abcd/record-lint.json is the gate this package proves. Without it the rule
// could be correct on fixtures and mis-wired in the file that actually runs it.
func TestContextCitationCurrencyGatesTheLiveRecord(t *testing.T) {
	root := repoRootFromPackage
	cfg, err := LoadConfig(filepath.Join(root, ".abcd", "record-lint.json"))
	if err != nil {
		t.Fatal(err)
	}
	rc, ok := cfg.Rules[ruleContextCitationCurrency]
	if !ok || !rc.Enabled {
		t.Fatalf("%s is not enabled in the shipped record-lint config", ruleContextCitationCurrency)
	}
	fs, err := checkContextCitationCurrency(root, rc)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 0 {
		t.Fatalf("the live CONTEXT.md cites a terminal record: %+v", fs)
	}
}
