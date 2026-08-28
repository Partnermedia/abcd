package lint

import (
	"path/filepath"
	"strings"
	"testing"
)

func harnessLeakCfg() Config {
	return Config{
		Roots: []string{"docs"},
		Rules: map[string]RuleConfig{
			ruleHarnessLeak: {Enabled: true, Severity: severityBlocker},
		},
	}
}

// TestHarnessLeakInLintedProse is the "any committed or posted text, not only a
// freshly created pull-request body" half of the criterion: the same class the
// scanner defines reaches the documents the lint walks.
func TestHarnessLeakInLintedProse(t *testing.T) {
	root := t.TempDir()
	// Assembled at runtime so this repository's own tree carries no specimen.
	footer := strings.Join([]string{"Generated", "with"}, " ") + " [Some Tool](https://tool.dev)"
	sessionURL := "https://agent-host.dev/code/" + strings.Join([]string{"session", "01Gy4Zo93PdMmggA8sfGyb"}, "_")
	writeFile(t, root, filepath.Join("docs", "release.md"),
		"# Release\n\nShipped.\n\n"+footer+"\n")
	writeFile(t, root, filepath.Join("docs", "run.md"),
		"# Run\n\nRecorded at "+sessionURL+"\n")

	fs, err := Lint(harnessLeakCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleHarnessLeak); n != 2 {
		t.Fatalf("expected both leak shapes to be flagged, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join("docs", "release.md"), ruleHarnessLeak, 5) {
		t.Errorf("expected the footer finding on its own line; got %+v", fs)
	}
	if !messageContains(fs, "Assisted-by") {
		t.Errorf("expected the message to name the sanctioned alternative; got %+v", fs)
	}
}

// TestHarnessLeakSparesFencedExamples: a fenced example is quoted material, and
// the whole point of a fence in this corpus is to show a literal. A gate that
// cannot be written about is one people route around.
func TestHarnessLeakSparesFencedExamples(t *testing.T) {
	root := t.TempDir()
	footer := strings.Join([]string{"Generated", "with"}, " ") + " [Some Tool](https://tool.dev)"
	writeFile(t, root, filepath.Join("docs", "policy.md"),
		"# Policy\n\nThe banned shape:\n\n```\n"+footer+"\n```\n\nUse the trailer instead.\n")

	fs, err := Lint(harnessLeakCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleHarnessLeak); n != 0 {
		t.Fatalf("expected a fenced example to be spared, got %d: %+v", n, fs)
	}
}

// TestHarnessLeakHonoursTheWaiver: the line-scoped escape the privacy rule
// teaches works here too, so one deliberately illustrative line does not force a
// whole page into a fence.
func TestHarnessLeakHonoursTheWaiver(t *testing.T) {
	root := t.TempDir()
	footer := strings.Join([]string{"Generated", "with"}, " ") + " [Some Tool](https://tool.dev)"
	writeFile(t, root, filepath.Join("docs", "policy.md"),
		"# Policy\n\n"+footer+" <!-- abcd-lint:allow -->\n")

	fs, err := Lint(harnessLeakCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleHarnessLeak); n != 0 {
		t.Fatalf("expected the waiver to be honoured, got %d: %+v", n, fs)
	}
}

// TestHarnessLeakDisabledByDefault: the rule is opt-in like every other, so a
// config that does not name it changes nothing.
func TestHarnessLeakDisabledByDefault(t *testing.T) {
	root := t.TempDir()
	footer := strings.Join([]string{"Generated", "with"}, " ") + " [Some Tool](https://tool.dev)"
	writeFile(t, root, filepath.Join("docs", "release.md"), "# Release\n\n"+footer+"\n")

	fs, err := Lint(Config{Roots: []string{"docs"}}, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleHarnessLeak); n != 0 {
		t.Fatalf("expected an unconfigured rule to be inert, got %d: %+v", n, fs)
	}
}

// TestHarnessLeakSecondMatchOnALine: a benign leftmost candidate must not disarm
// the pattern for the rest of the line. Stopping at the first skipped match made
// this blocker strictly weaker than the scanner it shares a definition with —
// the exact drift the class exists to prevent.
func TestHarnessLeakSecondMatchOnALine(t *testing.T) {
	root := t.TempDir()
	sessionURL := "https://agent-host.dev/code/" + strings.Join([]string{"session", "01Gy4Zo93PdMmggA8sfGyb"}, "_")
	writeFile(t, root, filepath.Join("docs", "run.md"),
		"# Run\n\nBackground https://agent-host.dev/blog/using-agent-session-management-and-1m and the run at "+sessionURL+"\n")

	fs, err := Lint(harnessLeakCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleHarnessLeak); n != 1 {
		t.Fatalf("a skipped leftmost candidate hid a real session URL; got %d: %+v", n, fs)
	}
}
