package lint

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile writes content to root/rel, creating parent directories.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// countRule returns how many findings carry the given rule id.
func countRule(fs []Finding, ruleID string) int {
	n := 0
	for _, f := range fs {
		if f.RuleID == ruleID {
			n++
		}
	}
	return n
}

func hasFinding(fs []Finding, file, ruleID string, line int) bool {
	for _, f := range fs {
		if f.File == file && f.RuleID == ruleID && f.Line == line {
			return true
		}
	}
	return false
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "record-lint.json")
	body := `{
      "roots": ["rec"],
      "banned_tokens": [{"id":"t1","pattern":"foo","message":"no foo","severity":"blocker","allow_context":["ok"]}],
      "rules": {"no_git_metadata": {"enabled": true, "severity": "blocker", "fields": ["created"]}}
    }`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Roots) != 1 || cfg.Roots[0] != "rec" {
		t.Errorf("roots = %v", cfg.Roots)
	}
	if len(cfg.BannedTokens) != 1 || cfg.BannedTokens[0].ID != "t1" {
		t.Errorf("banned_tokens = %v", cfg.BannedTokens)
	}
	if r := cfg.Rules["no_git_metadata"]; !r.Enabled || r.Fields[0] != "created" {
		t.Errorf("no_git_metadata rule = %+v", r)
	}
}

func TestBannedTokens(t *testing.T) {
	root := t.TempDir()
	// bad: matches with no allow context
	writeFile(t, root, "rec/bad.md", "the tool intent_lint.py is gone\nplain line\n")
	// allow context suppresses
	writeFile(t, root, "rec/allowed.md", "historical note: intent_lint.py existed\n")
	// inside a code fence -> skipped by default
	writeFile(t, root, "rec/fenced.md", "text\n```\nintent_lint.py\n```\nmore\n")
	// clean file
	writeFile(t, root, "rec/clean.md", "nothing to see here\n")

	cfg := Config{
		Roots: []string{"rec"},
		BannedTokens: []BannedToken{
			{ID: "py", Pattern: `intent_lint\.py`, Message: "no python name", Severity: "blocker", AllowContext: []string{"historical"}},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "py"); n != 1 {
		t.Fatalf("expected 1 banned-token finding, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join("rec", "bad.md"), "py", 1) {
		t.Errorf("missing expected finding on bad.md:1: %+v", fs)
	}
}

func TestNoGitMetadata(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/bad.md", "---\nid: x\nupdated: 2026-01-01\nauthor: someone\n---\n# Title\n")
	writeFile(t, root, "rec/clean.md", "---\nid: x\nkind: standalone\n---\n# Title\n")

	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"no_git_metadata": {Enabled: true, Severity: "blocker", Fields: []string{"created", "updated", "author"}},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "no_git_metadata"); n != 2 {
		t.Fatalf("expected 2 metadata findings, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join("rec", "bad.md"), "no_git_metadata", 3) {
		t.Errorf("expected finding on updated (line 3): %+v", fs)
	}
}

func TestLinksResolve(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/target.md", "# Target\n")
	writeFile(t, root, "rec/doc.md",
		"good: [t](target.md)\n"+
			"anchor ok: [t](target.md#section)\n"+
			"external: [x](https://example.com)\n"+
			"broken: [t](missing.md)\n"+
			"escape: [t](../../../etc/passwd)\n"+
			"fenced skip:\n```\n[t](also-missing.md)\n```\n")

	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"links_resolve": {Enabled: true, Severity: "blocker"},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	// missing.md and the escape -> 2 findings; the fenced link and externals ignored.
	if n := countRule(fs, "links_resolve"); n != 2 {
		t.Fatalf("expected 2 link findings, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join("rec", "doc.md"), "links_resolve", 4) {
		t.Errorf("expected broken-link finding on line 4: %+v", fs)
	}
	if !hasFinding(fs, filepath.Join("rec", "doc.md"), "links_resolve", 5) {
		t.Errorf("expected escape finding on line 5: %+v", fs)
	}
}

func TestBrittleLineRefs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/doc.md", "see configuration.md:171 for detail\nno ref here\nalso other.md:9 inline\n")

	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"no_brittle_line_refs": {Enabled: true, Severity: "warn"},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "no_brittle_line_refs"); n != 2 {
		t.Fatalf("expected 2 brittle-ref findings, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join("rec", "doc.md"), "no_brittle_line_refs", 1) {
		t.Errorf("expected finding on line 1: %+v", fs)
	}
}

func TestDirectoryCoverage(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/README.md", "# rec\n")
	writeFile(t, root, "rec/covered/README.md", "# covered\n")
	writeFile(t, root, "rec/covered/note.md", "hi\n")
	writeFile(t, root, "rec/bare/note.md", "no readme here\n")

	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"directory_coverage": {Enabled: true, Severity: "warn", Exempt: []string{}},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "directory_coverage"); n != 1 {
		t.Fatalf("expected 1 coverage finding, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join("rec", "bare"), "directory_coverage", 0) {
		t.Errorf("expected coverage finding on rec/bare: %+v", fs)
	}

	// exempting the bare dir clears the finding.
	cfg.Rules["directory_coverage"] = RuleConfig{Enabled: true, Severity: "warn", Exempt: []string{filepath.Join("rec", "bare")}}
	fs, err = Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "directory_coverage"); n != 0 {
		t.Fatalf("exempt glob should clear finding, got %d: %+v", n, fs)
	}
}

func TestIntentLifecycle(t *testing.T) {
	root := t.TempDir()
	base := "rec/intents"

	// Well-formed intents across buckets (the reference targets for superseded).
	writeFile(t, root, base+"/shipped/itd-48-good.md", "---\nid: itd-48\nkind: standalone\nspec_id: spc-10-thing\n---\n# ok\n")
	writeFile(t, root, base+"/planned/itd-20-good.md", "---\nid: itd-20\nkind: bundle-member\nspec_id: spc-83-thing\n---\n# ok\n")
	writeFile(t, root, base+"/drafts/itd-10-good.md", "---\nid: itd-10\nkind: null\nspec_id: null\n---\n# ok\n")
	writeFile(t, root, base+"/disciplines/itd-1-good.md", "---\nid: itd-1\nkind: discipline\nspec_id: null\n---\n# ok\n")
	writeFile(t, root, base+"/superseded/itd-31-good.md", "---\nid: itd-31\nkind: standalone\nsuperseded_by: itd-48\n---\n# ok\n")

	// Violations, one family per file.
	writeFile(t, root, base+"/drafts/itd-11-bad.md", "---\nid: itd-11\nkind: null\nspec_id: spc-99-nope\n---\n# bad\n")                        // drafts spec_id must be null
	writeFile(t, root, base+"/planned/itd-21-bad.md", "---\nid: itd-21\nkind: null\nspec_id: spc-1-x\n---\n# bad\n")                           // planned kind must be non-null
	writeFile(t, root, base+"/planned/itd-22-bad.md", "---\nid: itd-22\nkind: standalone\nspec_id: itd-5-wrong\n---\n# bad\n")                 // planned spec_id must be ^spc-
	writeFile(t, root, base+"/shipped/itd-49-bad.md", "---\nid: itd-49\nkind: standalone\nspec_id: null\n---\n# bad\n")                        // shipped spec_id must be non-null
	writeFile(t, root, base+"/disciplines/itd-2-bad.md", "---\nid: itd-2\nkind: standalone\nspec_id: null\n---\n# bad\n")                      // disciplines kind must be discipline
	writeFile(t, root, base+"/superseded/itd-32-bad.md", "---\nid: itd-32\nkind: standalone\nsuperseded_by: itd-999\n---\n# bad\n")            // target missing
	writeFile(t, root, base+"/shipped/itd-50-status.md", "---\nid: itd-50\nkind: standalone\nspec_id: spc-2-x\nstatus: shipped\n---\n# bad\n") // status: forbidden

	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"intent_lifecycle": {Enabled: true, Severity: "blocker", IntentsDir: "intents"},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		file string
		line int
	}{
		{filepath.Join(base, "drafts", "itd-11-bad.md"), 4},  // spec_id line
		{filepath.Join(base, "planned", "itd-21-bad.md"), 3}, // kind line
		{filepath.Join(base, "planned", "itd-22-bad.md"), 4}, // spec_id line
		{filepath.Join(base, "shipped", "itd-49-bad.md"), 4}, // spec_id line
		{filepath.Join(base, "disciplines", "itd-2-bad.md"), 3},
		{filepath.Join(base, "superseded", "itd-32-bad.md"), 4},
		{filepath.Join(base, "shipped", "itd-50-status.md"), 5}, // status line
	}
	for _, c := range checks {
		if !hasFinding(fs, c.file, "intent_lifecycle", c.line) {
			t.Errorf("expected intent_lifecycle finding on %s:%d; got %+v", c.file, c.line, fs)
		}
	}
	// The five good intents must produce no lifecycle findings.
	good := []string{"itd-48-good.md", "itd-20-good.md", "itd-10-good.md", "itd-1-good.md", "itd-31-good.md"}
	for _, f := range fs {
		for _, g := range good {
			if filepath.Base(f.File) == g {
				t.Errorf("unexpected finding on clean intent %s: %+v", g, f)
			}
		}
	}
}

func TestExemptions(t *testing.T) {
	root := t.TempDir()
	// A banned token in an exempt_paths file → no finding.
	writeFile(t, root, "rec/research/note.md", "the old intent_lint.py ran here\n")
	// The same token in a non-exempt file → finding.
	writeFile(t, root, "rec/live/doc.md", "the tool intent_lint.py is referenced\n")
	// A status:-exempt file: banned token is excused, but a broken link in it
	// STILL fires — structural checks stay universal.
	writeFile(t, root, "rec/live/old.md",
		"---\nid: x\nstatus: superseded\n---\nintent_lint.py mentioned\n[dead](missing.md)\n")

	cfg := Config{
		Roots: []string{"rec"},
		BannedTokens: []BannedToken{
			{ID: "py", Pattern: `intent_lint\.py`, Message: "no python name", Severity: "blocker"},
		},
		Rules: map[string]RuleConfig{
			"links_resolve": {Enabled: true, Severity: "blocker"},
		},
		ExemptPaths:    []string{filepath.Join("rec", "research") + string(filepath.Separator)},
		ExemptIfStatus: []string{"superseded"},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	// research note: exempt from banned tokens.
	if hasFinding(fs, filepath.Join("rec", "research", "note.md"), "py", 1) {
		t.Errorf("exempt_paths file should not fire banned token: %+v", fs)
	}
	// live doc: not exempt.
	if !hasFinding(fs, filepath.Join("rec", "live", "doc.md"), "py", 1) {
		t.Errorf("non-exempt file should fire banned token: %+v", fs)
	}
	// status:superseded file: banned token exempt...
	if hasFinding(fs, filepath.Join("rec", "live", "old.md"), "py", 5) {
		t.Errorf("status-exempt file should not fire banned token: %+v", fs)
	}
	// ...but its broken link STILL fires (structural, universal).
	if !hasFinding(fs, filepath.Join("rec", "live", "old.md"), "links_resolve", 6) {
		t.Errorf("structural link check must stay universal in exempt file: %+v", fs)
	}
	if n := countRule(fs, "py"); n != 1 {
		t.Fatalf("expected exactly 1 banned-token finding, got %d: %+v", n, fs)
	}
}

func TestCleanRecordProducesNoFindings(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/README.md", "# rec\n")
	writeFile(t, root, "rec/doc.md", "---\nid: x\nkind: standalone\n---\n# Clean\n\nA [link](README.md) and prose. No banned words.\n")

	cfg := Config{
		Roots: []string{"rec"},
		BannedTokens: []BannedToken{
			{ID: "py", Pattern: `intent_lint\.py`, Message: "no", Severity: "blocker"},
		},
		Rules: map[string]RuleConfig{
			"no_git_metadata":      {Enabled: true, Severity: "blocker", Fields: []string{"created", "updated"}},
			"links_resolve":        {Enabled: true, Severity: "blocker"},
			"no_brittle_line_refs": {Enabled: true, Severity: "warn"},
			"directory_coverage":   {Enabled: true, Severity: "warn"},
			"intent_lifecycle":     {Enabled: true, Severity: "blocker", IntentsDir: "intents"},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 0 {
		t.Fatalf("expected no findings on clean record, got %d: %+v", len(fs), fs)
	}
}
