package lint

import (
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/intent"
)

// intentParityCfg is the record-lint config the parity tests run: the intent
// tree at its real repo-relative home, so ONE fixture is read by both the lint
// and the loader.
func intentParityCfg() Config {
	return Config{
		Roots: []string{".abcd/development"},
		Rules: map[string]RuleConfig{
			"intent_lifecycle": {Enabled: true, Severity: severityBlocker, IntentsDir: "intents"},
		},
	}
}

// TestIntentLintRefusesEveryIDTheLoaderRefuses pins the lint↔loader contract for
// the intent store: any intent file intent.Load rejects must be a record-lint
// BLOCKER, so it can never merge green.
//
// The asymmetry this closes: intent.Load fail-closes the whole CORPUS on one
// malformed id (Validate → ^itd-[0-9]+$), while intent_lifecycle validated
// status/kind/spec_id/superseded_by and never the id itself — so a hand-edited
// record whose `id:` line was absent, empty, or a YAML null passed every merge
// gate and then bricked every intent verb, `abcd spec close` included, for
// everyone who pulled it (iss-2608270500198764).
func TestIntentLintRefusesEveryIDTheLoaderRefuses(t *testing.T) {
	cases := []struct {
		name        string
		frontmatter string
	}{
		{"absent", "---\nkind: standalone\nspec_id: spc-10-thing\n---\n# x\n"},
		{"empty", "---\nid:\nkind: standalone\nspec_id: spc-10-thing\n---\n# x\n"},
		{"null", "---\nid: null\nkind: standalone\nspec_id: spc-10-thing\n---\n# x\n"},
		{"tilde", "---\nid: ~\nkind: standalone\nspec_id: spc-10-thing\n---\n# x\n"},
		{"not-an-id", "---\nid: TBD\nkind: standalone\nspec_id: spc-10-thing\n---\n# x\n"},
	}
	const rel = ".abcd/development/intents/shipped/itd-999-thing.md"
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, rel, tc.frontmatter)

			// Half one: the loader refuses the file (and with it the whole corpus).
			if _, err := intent.Load(root); err == nil {
				t.Fatalf("intent.Load accepted %s: the fixture no longer pins the loader's rule", tc.name)
			}

			// Half two: the lint must therefore refuse it too.
			fs, err := Lint(intentParityCfg(), root)
			if err != nil {
				t.Fatal(err)
			}
			var blocked bool
			for _, f := range fs {
				if f.File == rel && f.RuleID == "intent_lifecycle" &&
					f.Severity == severityBlocker && strings.Contains(f.Message, "id") {
					blocked = true
				}
			}
			if !blocked {
				t.Fatalf("lint passed an intent id the loader refuses (%s); findings: %+v", tc.name, fs)
			}
		})
	}
}

// TestIntentLintAcceptsWhatTheLoaderAccepts is the other direction of the same
// contract: a well-formed record the loader reads must be lint-clean, so arming
// the id check cannot turn a healthy corpus red.
func TestIntentLintAcceptsWhatTheLoaderAccepts(t *testing.T) {
	root := t.TempDir()
	const rel = ".abcd/development/intents/shipped/itd-999-thing.md"
	writeFile(t, root, rel, "---\nid: itd-999\nslug: thing\nkind: standalone\nspec_id: spc-10-thing\n---\n# x\n")

	corpus, err := intent.Load(root)
	if err != nil {
		t.Fatalf("intent.Load refused a well-formed record: %v", err)
	}
	if len(corpus.Intents) != 1 {
		t.Fatalf("expected 1 intent, got %d", len(corpus.Intents))
	}
	fs, err := Lint(intentParityCfg(), root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "intent_lifecycle"); n != 0 {
		t.Fatalf("well-formed intent must be lint-clean; got %+v", fs)
	}
}
