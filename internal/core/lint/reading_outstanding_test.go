package lint

import (
	"strings"
	"testing"
)

// readingLedger writes one reading record and returns the repo root, so each
// test starts from a run that returned something.
func readingLedger(t *testing.T, run, item, position string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "rec/.keep", "")
	writeFile(t, root, ".abcd/work/issues/readings/"+run+"/"+item+".md",
		"---\n"+
			"schema_version: 1\n"+
			"id: \""+item+"\"\n"+
			"run: \""+run+"\"\n"+
			"manifest: \"sha256:beef\"\n"+
			"position: \""+position+"\"\n"+
			"regime: \"supplied\"\n"+
			"pattern: \"a stated constraint\"\n"+
			"---\n\n")
	return root
}

// readingOutstandingConfig arms only the rule under test, at the severity the
// caller names — which is the point of the severity test below: the rule must
// ignore it.
func readingOutstandingConfig(severity string) Config {
	return Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			ruleReadingOutstanding: {
				Enabled: true, Severity: severity, IssuesDir: ".abcd/work/issues",
			},
		},
	}
}

// Every reading record either carries a disposition or is REPORTED as
// outstanding. Nothing meaning "already covered" exists at any position, so an
// unanswered item has no state to sit in — the report is the only thing that
// says it is unanswered, and silence would let it disappear.
func TestOutstandingReportNamesUndispositionedItems(t *testing.T) {
	run, item := "rdg-2608300000000001", "rdi-2608300000000002"
	root := readingLedger(t, run, item, "detection")

	fs, err := Lint(readingOutstandingConfig(severityBlocker), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleReadingOutstanding); n != 1 {
		t.Fatalf("expected exactly 1 %s finding, got %d: %+v", ruleReadingOutstanding, n, fs)
	}
	if !strings.Contains(fs[0].Message, item) {
		t.Fatalf("the report must name the item; got %q", fs[0].Message)
	}

	// An answered item drops off the report — the status signal is the presence
	// of the keyed disposition, and one probe is all it takes.
	writeFile(t, root, ".abcd/work/issues/dispositions/"+item+"/dsp-2608300000000003.md",
		"---\nschema_version: 1\nid: \"dsp-2608300000000003\"\nitem: \""+item+"\"\n"+
			"state: \"accepted\"\ndisposition_grounds: \"worth acting on\"\n---\n\n")
	fs, err = Lint(readingOutstandingConfig(severityBlocker), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleReadingOutstanding); n != 0 {
		t.Fatalf("a dispositioned item must not be outstanding, got %d finding(s): %+v", n, fs)
	}
}

// The severity is pinned in CODE, not read from the config. A rule whose
// severity a config could raise to blocker is a gate waiting to happen, and a
// reading must never block an unrelated push — the config above asks for
// `blocker` precisely so the refusal to honour it is what this test observes.
func TestOutstandingReportSeverityIsInfoNotBlocker(t *testing.T) {
	root := readingLedger(t, "rdg-2608300000000001", "rdi-2608300000000002", "detection")

	fs, err := Lint(readingOutstandingConfig(severityBlocker), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) == 0 {
		t.Fatal("expected the outstanding report to produce a finding")
	}
	for _, f := range fs {
		if f.RuleID != ruleReadingOutstanding {
			continue
		}
		if f.Severity != severityInfo {
			t.Fatalf("severity = %q, want %q — a config that could raise this to blocker is a gate waiting to happen",
				f.Severity, severityInfo)
		}
	}
}

// A hold is directional and exits only through a superseding disposition that
// cites it — never by expiry, and never silently. So an open hold renders WITH
// its exit condition: a hold whose exit condition is not in front of the reader
// is indistinguishable from a parking space.
func TestOpenHoldRendersItsExitCondition(t *testing.T) {
	item := "rdi-2608300000000002"
	root := readingLedger(t, "rdg-2608300000000001", item, "detection")
	writeFile(t, root, ".abcd/work/issues/dispositions/"+item+"/dsp-2608300000000003.md",
		"---\nschema_version: 1\nid: \"dsp-2608300000000003\"\nitem: \""+item+"\"\n"+
			"state: \"held\"\nexit_condition: \"the closing run returns it again\"\n---\n\n")

	fs, err := Lint(readingOutstandingConfig(severityBlocker), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleReadingOutstanding); n != 1 {
		t.Fatalf("expected exactly 1 %s finding for the open hold, got %d: %+v", ruleReadingOutstanding, n, fs)
	}
	msg := fs[0].Message
	if !strings.Contains(msg, "held") || !strings.Contains(msg, "the closing run returns it again") {
		t.Fatalf("an open hold must render with its exit condition; got %q", msg)
	}

	// A superseding disposition closes the hold, and the superseded record stays
	// in place: the report follows the STANDING answer, not the file count.
	writeFile(t, root, ".abcd/work/issues/dispositions/"+item+"/dsp-2608300000000004.md",
		"---\nschema_version: 1\nid: \"dsp-2608300000000004\"\nitem: \""+item+"\"\n"+
			"state: \"accepted\"\ndisposition_grounds: \"the closing run returned it\"\n"+
			"supersedes_disposition: \"dsp-2608300000000003\"\n---\n\n")
	fs, err = Lint(readingOutstandingConfig(severityBlocker), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, ruleReadingOutstanding); n != 0 {
		t.Fatalf("a superseded hold must leave the report, got %d finding(s): %+v", n, fs)
	}
}
