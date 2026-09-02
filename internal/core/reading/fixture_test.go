package reading

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/gitutil"
)

// The fixture repository the assembler tests read. Every warm class the record
// names is planted in it with a distinctive sentinel token, so a leak shows up
// as a token rather than as a judgement call.
const (
	sentinelEvidence    = "SENTINEL-EVIDENCE-CHAPTER"
	sentinelDecision    = "SENTINEL-DECISION-RECORD"
	sentinelIssue       = "SENTINEL-ISSUE-LEDGER"
	sentinelAuditNotes  = "SENTINEL-AUDIT-NOTES"
	sentinelWhyItMatter = "SENTINEL-WHY-THIS-MATTERS"
	sentinelOrigin      = "SENTINEL-ORIGIN-KEY"
	sentinelProdMode    = "SENTINEL-PRODUCTION-MODE"
	sentinelSuperseded  = "SENTINEL-SUPERSEDED-INTENT"
	sentinelPlan        = "SENTINEL-PLAN-RECORD"
	sentinelPriorRun    = "SENTINEL-PRIOR-MANIFEST"
	sentinelDraftBody   = "SENTINEL-DRAFT-BODY"
	sentinelDefinition  = "SENTINEL-READING-DEFINITION"
	sentinelLapse       = "SENTINEL-LAPSE-LOG"
)

// writeFile writes one fixture file, creating its parents.
func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// fixtureRepo materialises a small repository carrying one file of every class
// the include table admits and one of every class the exclusion floor refuses,
// commits it, and returns its root.
func fixtureRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// fixtureScope selects every material kind at every assembling position, so
	// a test written before scopes existed sees the item set it was written
	// against. A test that cares about NARROWING names its own preset.
	writeFile(t, root, ".abcd/config/reading-presets.json", fixturePresets())

	writeFile(t, root, ".abcd/record-lint.json", `{
  "schema_version": 1,
  "rules": {
    "record_schema": {
      "enabled": true,
      "severity": "blocker",
      "record_stores": {
        "adr": ".abcd/development/decisions/adrs",
        "itd": ".abcd/development/intents",
        "spc": ".abcd/development/specs",
        "iss": ".abcd/work/issues"
      }
    }
  }
}
`)

	// The admitted floor.
	writeFile(t, root, ".abcd/development/brief/01-product/06-framing.md",
		"# Framing\n\n## Construal\n\nWe are treating this as a gap in what the repository records.\n")
	writeFile(t, root, ".abcd/development/brief/01-product/01-press-release.md",
		"---\nproduction_mode: hand-written "+sentinelProdMode+"\n---\n\n# Press release\n\nThe product, stated.\n")
	writeFile(t, root, ".abcd/development/brief/02-constraints/03-invariants.md",
		"# Invariants\n\n1. The core is transport agnostic.\n")
	// The rest of brief current text (itd-194): the meta chapter at the brief's
	// root and the three chapters below the evidence chapter. A walk row's
	// source directory must exist or the run refuses, so these are not optional
	// decoration — the table names six chapters and the fixture carries six.
	writeFile(t, root, ".abcd/development/brief/00-meta.md",
		"# Meta\n\nHow this brief is organised.\n")
	writeFile(t, root, ".abcd/development/brief/04-surfaces/23-reading.md",
		"# The reading surface\n\nWhat the verb does.\n")
	writeFile(t, root, ".abcd/development/brief/05-internals/03-configuration.md",
		"# Configuration\n\nWhere the settings live.\n")
	writeFile(t, root, ".abcd/development/brief/06-delivery/01-shipping.md",
		"# Shipping\n\nHow a release is cut.\n")
	writeFile(t, root, ".abcd/development/brief/glossary/core/construal.md",
		"# Construal\n\nWhat the situation is treated as.\n")
	writeFile(t, root, ".abcd/development/intents/disciplines/itd-4-selection-criteria.md",
		"---\nid: itd-4\n---\n\n# Selection criteria\n\nSix criteria, recorded.\n")
	writeFile(t, root, ".abcd/development/specs/open/spc-1-a-design-record.md",
		"---\nid: spc-1\nintent: itd-1\norigin: "+sentinelOrigin+"\n---\n\n# A design record\n\nThe mechanics.\n")
	writeFile(t, root, "README.md", "# The repository\n\nWhat it is.\n")
	writeFile(t, root, "docs/reference/thing.md", "# Thing\n\nReference prose.\n")
	writeFile(t, root, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, root, "go.mod", "module example\n\ngo 1.24\n")
	writeFile(t, root, "Makefile", "build:\n\t@echo build\n")

	// A shipped intent carrying both the fields that travel and the fields that
	// must not.
	writeFile(t, root, ".abcd/development/intents/shipped/itd-1-a-shipped-intent.md",
		"---\nid: itd-1\nspec_id: spc-1\norigin: "+sentinelOrigin+"\n---\n\n"+
			"# A shipped intent\n\n"+
			"## Press Release\n\nThe promise, as made.\n\n"+
			"## Why This Matters\n\n"+sentinelWhyItMatter+"\n\n"+
			"## Acceptance Criteria\n\n- Given a state, when it runs, then it holds.\n\n"+
			"## Scope Conditions\n\n- Holds while the record is one repository.\n\n"+
			"## Mechanism\n\nA positive include table.\n\n"+
			"## Audit Notes\n\n"+sentinelAuditNotes+"\n")

	// The candidate set: cold at entailment, warm everywhere else.
	writeFile(t, root, ".abcd/development/intents/drafts/itd-2-a-draft.md",
		"---\nid: itd-2\norigin: "+sentinelOrigin+"\n---\n\n# A draft\n\n## Press Release\n\n"+sentinelDraftBody+"\n")
	writeFile(t, root, ".abcd/development/intents/planned/itd-3-a-planned.md",
		"---\nid: itd-3\nspec_id: spc-1\n---\n\n# A planned intent\n\n## Press Release\n\nA planned promise.\n")

	// Every refused class.
	writeFile(t, root, ".abcd/development/brief/03-evidence/01-open-questions.md",
		"# Open questions\n\n"+sentinelEvidence+"\n")
	writeFile(t, root, ".abcd/development/decisions/adrs/0001-a-decision.md",
		"---\nid: adr-1\n---\n\n# ADR-1\n\n"+sentinelDecision+"\n")
	writeFile(t, root, ".abcd/development/intents/superseded/itd-5-a-retired.md",
		"---\nid: itd-5\n---\n\n# A retired intent\n\n"+sentinelSuperseded+"\n")
	writeFile(t, root, ".abcd/development/plans/2026-08-30-a-plan.md", "# A plan\n\n"+sentinelPlan+"\n")
	writeFile(t, root, ".abcd/development/research/notes/a-note.md", "# A note\n\nresearch prose.\n")
	writeFile(t, root, ".abcd/development/roadmap/rfcs/an-rfc.md", "# An RFC\n\nrfc prose.\n")
	writeFile(t, root, ".abcd/work/issues/open/iss-1-a-defect.md",
		"---\nid: iss-1\n---\n\n"+sentinelIssue+"\n")
	writeFile(t, root, ".abcd/work/issues/open/iss-2-a-lapse.md",
		"---\nid: iss-2\ncategory: lapse-log\n---\n\n"+sentinelLapse+"\n")
	writeFile(t, root, ".abcd/work/DECISIONS.md", "# Decisions\n\n"+sentinelDecision+"\n")
	writeFile(t, root, ".abcd/development/readings/rdg-2608301200000001/manifest.json",
		"{\"note\": \""+sentinelPriorRun+"\"}\n")
	writeFile(t, root, "agents/cold-reading-widening.md", "# Widening\n\n"+sentinelDefinition+"\n")
	writeFile(t, root, "evals/read_block_test.go", "package evals\n\n// "+sentinelDefinition+"\n")

	gitInit(t, root)
	return root
}

// gitInit turns a fixture directory into a repository with one commit, so HEAD
// resolves and no included path is dirty.
func gitInit(t *testing.T, root string) {
	t.Helper()
	steps := [][]string{
		{"init", "-q", "-b", "main"},
		{"-c", "user.name=abcd test", "-c", "user.email=test@example.invalid", "add", "-A"},
		{"-c", "user.name=abcd test", "-c", "user.email=test@example.invalid",
			"commit", "-q", "-m", "fixture"},
	}
	for _, args := range steps {
		if out, err := gitutil.Run(root, args...); err != nil {
			t.Fatalf("git %s: %v (%s)", strings.Join(args, " "), err, out)
		}
	}
}

// headOf returns the fixture repository's HEAD sha.
func headOf(t *testing.T, root string) string {
	t.Helper()
	sha, err := gitutil.Run(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return sha
}

// assembleFixture runs one assembly over the fixture repository at p.
func assembleFixture(t *testing.T, root string, p Position) AssembleResult {
	t.Helper()
	res, err := Assemble(AssembleRequest{
		RepoRoot: root, Position: p, Target: "HEAD", DryRun: true,
	})
	if err != nil {
		t.Fatalf("assemble at %s: %v", p, err)
	}
	return res
}

// bundleText joins every item's text, for a sentinel scan.
func bundleText(b Bundle) string {
	var sb strings.Builder
	for _, it := range b.Items {
		sb.WriteString(it.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// itemPaths returns the manifest's item paths in order.
func itemPaths(m Manifest) []string {
	out := make([]string, 0, len(m.Items))
	for _, it := range m.Items {
		out = append(out, it.Path)
	}
	return out
}

// gitCommitAll commits everything a test added after the fixture was built, so
// the tree stays clean for the dirty-path gate.
func gitCommitAll(t *testing.T, root string) {
	t.Helper()
	ident := []string{"-c", "user.name=abcd test", "-c", "user.email=test@example.invalid"}
	for _, args := range [][]string{
		append(append([]string{}, ident...), "add", "-A"),
		append(append([]string{}, ident...), "commit", "-q", "-m", "fixture update"),
	} {
		if out, err := gitutil.Run(root, args...); err != nil {
			t.Fatalf("git %s: %v (%s)", strings.Join(args, " "), err, out)
		}
	}
}

// gitRun runs one git command against a fixture repository under the test
// identity, failing the test on error.
func gitRun(t *testing.T, root string, args ...string) {
	t.Helper()
	full := append([]string{"-c", "user.name=abcd test", "-c", "user.email=test@example.invalid"}, args...)
	if out, err := gitutil.Run(root, full...); err != nil {
		t.Fatalf("git %s: %v (%s)", strings.Join(args, " "), err, out)
	}
}

// fixtureScopeName is the all-selecting preset every fixture carries.
const fixtureScopeName = "everything"

// fixturePresets renders a preset file selecting every kind at every position
// that assembles. It is generated from Kinds() and AssemblingPositions() rather
// than written out, so a new kind or position cannot leave the fixture quietly
// narrower than the table it is meant to mirror.
func fixturePresets() string {
	kinds := make([]string, 0, len(Kinds()))
	for _, k := range Kinds() {
		kinds = append(kinds, strconv.Quote(string(k)))
	}
	paths := make([]string, 0, len(fixtureTreePaths))
	for _, p := range fixtureTreePaths {
		paths = append(paths, strconv.Quote(p))
	}
	positions := make([]string, 0, len(AssemblingPositions()))
	for _, p := range AssemblingPositions() {
		positions = append(positions, fmt.Sprintf(
			`    %q: {"object": {"records": [], "paths": [%s]}, "kinds": [%s], `+
				`"window": {"tokens_est": 1000000, "measured_tokens_est": 0, `+
				`"measured_bytes": 0, "measured_at": "0000000"}}`,
			string(p), strings.Join(paths, ", "), strings.Join(kinds, ", ")))
	}
	return fmt.Sprintf(`{
  "schema_version": %d,
  "positions": {
%s
  }
}
`, PresetSchemaVersion, strings.Join(positions, ",\n"))
}

// fixtureTreePaths is every root-level path the fixture carries tree material
// under. The object set NARROWS the tree rows, and an entry naming no path hands
// nothing from the tree whatever kinds it lists (spc-2609020626048722), so the
// all-selecting fixture entry has to name where the fixture's tree material
// sits. `main_test.go` is listed although the base fixture does not carry it: a
// test that adds one is naming a path the entry already reaches, rather than
// having to rewrite the entry to see it.
var fixtureTreePaths = []string{
	"Gadget_TEST.go", "Makefile", "README.md", "build", "docs", "fence.go",
	"go.mod", "ignored.json", "main.go", "main_test.go", "run-dir", "runs",
	"widget.go", "widget_test.go",
}
