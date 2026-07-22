package lint

import (
	"path/filepath"
	"strings"
	"testing"
)

// impactCase is one record fixture plus the single finding it must (or must not)
// produce. wantMsg empty means the record is clean; otherwise the rule must emit
// exactly one finding, on wantLine, whose message contains wantMsg.
type impactCase struct {
	name     string
	rel      string
	body     string
	wantLine int
	wantMsg  string
}

// runImpactCases writes each fixture into its own temp repo and lints it with
// only the rule under test enabled, so a finding can come from nowhere else.
func runImpactCases(t *testing.T, ruleID string, rule RuleConfig, cases []impactCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, c.rel, c.body)
			cfg := Config{
				Roots: []string{"rec"},
				Rules: map[string]RuleConfig{ruleID: rule},
			}
			fs, err := Lint(cfg, root)
			if err != nil {
				t.Fatal(err)
			}
			if c.wantMsg == "" {
				if n := countRule(fs, ruleID); n != 0 {
					t.Fatalf("expected no %s finding, got %d: %+v", ruleID, n, fs)
				}
				return
			}
			if n := countRule(fs, ruleID); n != 1 {
				t.Fatalf("expected exactly 1 %s finding, got %d: %+v", ruleID, n, fs)
			}
			f := fs[0]
			if f.File != filepath.FromSlash(c.rel) || f.Line != c.wantLine {
				t.Errorf("finding at %s:%d, want %s:%d", f.File, f.Line, c.rel, c.wantLine)
			}
			if f.Severity != severityBlocker {
				t.Errorf("severity = %q, want %q", f.Severity, severityBlocker)
			}
			if !strings.Contains(f.Message, c.wantMsg) {
				t.Errorf("message = %q, want it to contain %q", f.Message, c.wantMsg)
			}
		})
	}
}

// An intent that has reached shipped/ is in the release set, so its impact is
// what the derived version is computed from: absent, misspelled, and internal
// must each block, with a message that says which of the three failures it is.
// A drafts/ or planned/ intent is NOT required to carry the judgement yet — the
// gate is the move into shipped/ — but a value it DOES carry must be a legal
// one, because a misspelling caught in the bucket where it was written is
// actionable and the same misspelling caught at the shipped/ gate is archaeology.
func TestIntentImpactValid(t *testing.T) {
	const base = "rec/intents"
	rule := RuleConfig{Enabled: true, Severity: severityBlocker, IntentsDir: "intents"}

	runImpactCases(t, "intent_impact_valid", rule, []impactCase{
		{
			name:     "shipped without impact",
			rel:      base + "/shipped/itd-48-thing.md",
			body:     "---\nid: itd-48\nkind: standalone\nspec_id: spc-10-x\n---\n# x\n",
			wantLine: 1,
			wantMsg:  "impact must be set explicitly",
		},
		{
			name:     "shipped with an explicit null impact",
			rel:      base + "/shipped/itd-49-thing.md",
			body:     "---\nid: itd-49\nkind: standalone\nspec_id: spc-10-x\nimpact: null\n---\n# x\n",
			wantLine: 5,
			wantMsg:  "impact must be set explicitly",
		},
		{
			name:     "shipped with a misspelled impact",
			rel:      base + "/shipped/itd-50-thing.md",
			body:     "---\nid: itd-50\nkind: standalone\nspec_id: spc-10-x\nimpact: additiv\n---\n# x\n",
			wantLine: 5,
			wantMsg:  "invalid impact 'additiv'",
		},
		{
			name:     "shipped with a capitalised impact",
			rel:      base + "/shipped/itd-51-thing.md",
			body:     "---\nid: itd-51\nkind: standalone\nspec_id: spc-10-x\nimpact: Additive\n---\n# x\n",
			wantLine: 5,
			wantMsg:  "invalid impact 'Additive'",
		},
		{
			name:     "shipped with impact internal",
			rel:      base + "/shipped/itd-52-thing.md",
			body:     "---\nid: itd-52\nkind: standalone\nspec_id: spc-10-x\nimpact: internal\n---\n# x\n",
			wantLine: 5,
			wantMsg:  "must not be internal on an intent",
		},
		{
			name: "shipped with impact additive",
			rel:  base + "/shipped/itd-53-thing.md",
			body: "---\nid: itd-53\nkind: standalone\nspec_id: spc-10-x\nimpact: additive\n---\n# x\n",
		},
		{
			name: "shipped with impact breaking",
			rel:  base + "/shipped/itd-54-thing.md",
			body: "---\nid: itd-54\nkind: standalone\nspec_id: spc-10-x\nimpact: breaking\n---\n# x\n",
		},
		{
			name: "shipped with impact fix",
			rel:  base + "/shipped/itd-55-thing.md",
			body: "---\nid: itd-55\nkind: standalone\nspec_id: spc-10-x\nimpact: fix\n---\n# x\n",
		},
		{
			name: "drafts without impact is not gated",
			rel:  base + "/drafts/itd-10-thing.md",
			body: "---\nid: itd-10\nkind: null\nspec_id: null\n---\n# x\n",
		},
		{
			name: "planned without impact is not gated",
			rel:  base + "/planned/itd-20-thing.md",
			body: "---\nid: itd-20\nkind: standalone\nspec_id: null\n---\n# x\n",
		},
		{
			name: "disciplines without impact is not gated",
			rel:  base + "/disciplines/itd-1-thing.md",
			body: "---\nid: itd-1\nkind: discipline\nspec_id: null\n---\n# x\n",
		},
		{
			name:     "drafts with a misspelled impact is still flagged",
			rel:      base + "/drafts/itd-11-thing.md",
			body:     "---\nid: itd-11\nkind: null\nspec_id: null\nimpact: braking\n---\n# x\n",
			wantLine: 5,
			wantMsg:  "invalid impact 'braking'",
		},
		{
			name:     "planned with impact internal is still flagged",
			rel:      base + "/planned/itd-21-thing.md",
			body:     "---\nid: itd-21\nkind: standalone\nspec_id: null\nimpact: internal\n---\n# x\n",
			wantLine: 5,
			wantMsg:  "must not be internal on an intent",
		},
	})
}

// An issue in resolved/ is in the release set exactly like a shipped intent, so
// the same gate applies — with one difference: internal is a LEGAL judgement on
// an issue, because most resolved issues are plumbing that no user can be told
// about. An open/ or wontfix/ issue is not required to carry the judgement.
func TestIssueImpactValid(t *testing.T) {
	const base = ".abcd/work/issues"
	rule := RuleConfig{Enabled: true, Severity: severityBlocker, IssuesDir: base}

	runImpactCases(t, "issue_impact_valid", rule, []impactCase{
		{
			name:     "resolved without impact",
			rel:      base + "/resolved/iss-56-thing.md",
			body:     "---\nid: \"iss-56\"\n---\n# x\n",
			wantLine: 1,
			wantMsg:  "impact must be set explicitly",
		},
		{
			name:     "resolved with an explicit null impact",
			rel:      base + "/resolved/iss-57-thing.md",
			body:     "---\nid: \"iss-57\"\nimpact: null\n---\n# x\n",
			wantLine: 3,
			wantMsg:  "impact must be set explicitly",
		},
		{
			name:     "resolved with a misspelled impact",
			rel:      base + "/resolved/iss-58-thing.md",
			body:     "---\nid: \"iss-58\"\nimpact: fixx\n---\n# x\n",
			wantLine: 3,
			wantMsg:  "invalid impact \"fixx\"",
		},
		{
			// The ledger quotes ids (id: "iss-56"), so quoting the impact is the
			// plausible authoring slip. It is rejected, not absorbed: the shared
			// parser matches the enum exactly, and the frontmatter scanner is a line
			// scanner rather than a YAML parser, so a tolerated `"fix"` here would
			// read as a different string in every other consumer of the field.
			name:     "resolved with a YAML-quoted impact",
			rel:      base + "/resolved/iss-62-thing.md",
			body:     "---\nid: \"iss-62\"\nimpact: \"fix\"\n---\n# x\n",
			wantLine: 3,
			wantMsg:  "invalid impact",
		},
		{
			name: "resolved with impact internal",
			rel:  base + "/resolved/iss-59-thing.md",
			body: "---\nid: \"iss-59\"\nimpact: internal\n---\n# x\n",
		},
		{
			name: "resolved with impact fix",
			rel:  base + "/resolved/iss-60-thing.md",
			body: "---\nid: \"iss-60\"\nimpact: fix\n---\n# x\n",
		},
		{
			name: "resolved with impact additive",
			rel:  base + "/resolved/iss-61-thing.md",
			body: "---\nid: \"iss-61\"\nimpact: additive\n---\n# x\n",
		},
		{
			name: "open without impact is not gated",
			rel:  base + "/open/iss-70-thing.md",
			body: "---\nid: \"iss-70\"\n---\n# x\n",
		},
		{
			name: "wontfix without impact is not gated",
			rel:  base + "/wontfix/iss-71-thing.md",
			body: "---\nid: \"iss-71\"\n---\n# x\n",
		},
		{
			name:     "open with a misspelled impact is still flagged",
			rel:      base + "/open/iss-72-thing.md",
			body:     "---\nid: \"iss-72\"\nimpact: breaking!\n---\n# x\n",
			wantLine: 3,
			wantMsg:  "invalid impact \"breaking!\"",
		},
	})
}

// The two rules only gate anything if the shipped config arms them. This is the
// wiring test: it reads the real .abcd/record-lint.json, so disabling either
// rule, or downgrading it out of blocker severity, fails here. It deliberately
// does NOT assert the live record is clean — the records are back-filled
// separately, and a rule that gates nothing until then is still armed.
func TestImpactRulesArmedInRealConfig(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..")
	cfg, err := LoadConfig(filepath.Join(repoRoot, ".abcd", "record-lint.json"))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	for _, id := range []string{"intent_impact_valid", "issue_impact_valid"} {
		rc, ok := cfg.Rules[id]
		if !ok || !rc.Enabled {
			t.Errorf("record-lint.json must enable the %s rule (spc-10)", id)
			continue
		}
		if rc.Severity != severityBlocker {
			t.Errorf("%s severity = %q, want %q", id, rc.Severity, severityBlocker)
		}
	}
	if rc := cfg.Rules["intent_impact_valid"]; rc.IntentsDir == "" {
		t.Error("intent_impact_valid must declare intents_dir, like intent_lifecycle")
	}
	if rc := cfg.Rules["issue_impact_valid"]; rc.IssuesDir == "" {
		t.Error("issue_impact_valid must declare issues_dir, like issue_id_unique")
	}
}
