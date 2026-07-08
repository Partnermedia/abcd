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

// TestDocsLintHarnessNameGate guards the real .abcd/docs-lint.json harness-name
// family (the prevention gate): a specific agent-harness name in user-facing
// content is a blocker, and the docs-lint:allow comment on the same line
// suppresses it. Loading the actual config means deleting the family (or dropping
// its blocker severity) fails this test.
func TestDocsLintHarnessNameGate(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join("..", "..", "..", ".abcd", "docs-lint.json"))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	root := t.TempDir()
	writeFile(t, root, "docs/named.md", "# t\n\nRun this in Claude Code.\n")
	writeFile(t, root, "docs/allowed.md", "# t\n\n<!-- docs-lint: allow --> Claude Code is named deliberately.\n")
	writeFile(t, root, "docs/clean.md", "# t\n\nUse the agent harness.\n")

	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(fs, filepath.Join("docs", "named.md"), "harness/claude-code", 3) {
		t.Fatalf("expected harness/claude-code blocker on docs/named.md:3: %+v", fs)
	}
	for _, f := range fs {
		if f.RuleID == "harness/claude-code" {
			if f.Severity != "blocker" {
				t.Errorf("harness/claude-code severity = %q, want blocker", f.Severity)
			}
			if f.File != filepath.Join("docs", "named.md") {
				t.Errorf("harness gate fired outside named.md (allow-context/clean leaked): %+v", f)
			}
		}
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
	writeFile(t, root, base+"/planned/itd-23-good.md", "---\nid: itd-23\nkind: standalone\nspec_id: null\n---\n# ok\n") // re-baseline: planned may be unscheduled (null spec_id)
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
	// The good intents must produce no lifecycle findings.
	good := []string{"itd-48-good.md", "itd-20-good.md", "itd-23-good.md", "itd-10-good.md", "itd-1-good.md", "itd-31-good.md"}
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

// presentTenseTokens mirrors the change-narration phrase list shipped in
// .abcd/docs-lint.json (word-boundary, case-insensitive, inline-escape).
func presentTenseTokens() []BannedToken {
	allow := []string{`(?i)<!--\s*docs-lint:\s*allow\b`}
	msg := "change-narration in a doc body; docs state present reality only."
	mk := func(id, pat, sev string) BannedToken {
		return BannedToken{ID: id, Pattern: pat, Severity: sev, AllowContext: allow, Message: msg}
	}
	return []BannedToken{
		mk("present_tense/previously", `(?i)\bpreviously\b`, "blocker"),
		mk("present_tense/formerly", `(?i)\bformerly\b`, "blocker"),
		mk("present_tense/to-be-implemented", `(?i)\bto be implemented\b`, "blocker"),
		mk("present_tense/has-been-replaced", `(?i)\b(has|have|had)\s+been\s+replaced\b`, "blocker"),
		mk("present_tense/we-switched", `(?i)\bwe switched\b`, "blocker"),
		mk("present_tense/renamed-from", `(?i)\brenamed from\b`, "blocker"),
		// Ambiguous with legitimate present-state prose -> advisory (warn), never block.
		mk("present_tense/no-longer", `(?i)\bno longer\b`, "warn"),
		mk("present_tense/deprecated", `(?i)\bdeprecated\b`, "warn"),
		mk("present_tense/migrated-from", `(?i)\bmigrated from\b`, "warn"),
		// NB: "used to" is intentionally absent — RE2 has no lookbehind to tell
		// change-narration ("used to be X") from passive present ("is used to X").
	}
}

func TestPresentTense(t *testing.T) {
	root := t.TempDir()
	// Unambiguous change-narration -> each phrase is a BLOCKER finding.
	writeFile(t, root, "docs/history.md",
		"The flag was previously named --old.\n"+ // previously
			"It was formerly a shell script.\n"+ // formerly
			"The engine has been replaced by a Go port.\n"+ // has been replaced
			"We switched to TOML for config.\n"+ // we switched
			"The type was renamed from Foo.\n"+ // renamed from
			"The MCP surface is to be implemented.\n") // to be implemented
	// Ambiguous phrasing that also states present reality -> WARN, never blocks.
	writeFile(t, root, "docs/ambiguous.md",
		"The upstream token is deprecated.\n"+ // deprecated (present-state)
			"Records migrated from the legacy store are validated on read.\n") // migrated from (provenance)
	// Legitimate present tense, incl. passive "is used to" -> NO finding at all
	// (this is the regression that broke a blocking gate before "used to" was dropped).
	writeFile(t, root, "docs/clean.md",
		"The --config flag is used to override defaults.\n"+
			"This directory is used to store build output.\n"+
			"Run the command now.\nFirst do X, then do Y.\nThe engine currently reads config.json.\n")
	// The inline escape suppresses an otherwise-flagged line.
	writeFile(t, root, "docs/escaped.md",
		"The API was previously named --old. <!-- docs-lint: allow -->\n")
	// Inside a fenced code block -> skipped by default.
	writeFile(t, root, "docs/fenced.md",
		"prose\n```\nthis was previously broken\n```\nmore\n")

	cfg := Config{Roots: []string{"docs"}, BannedTokens: presentTenseTokens()}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	sevOf := func(base, rule string, line int) (string, bool) {
		for _, f := range fs {
			if filepath.Base(f.File) == base && f.RuleID == rule && f.Line == line {
				return f.Severity, true
			}
		}
		return "", false
	}

	// history.md: the six unambiguous phrases each fire as a BLOCKER, one per line.
	for _, w := range []struct {
		rule string
		line int
	}{
		{"present_tense/previously", 1},
		{"present_tense/formerly", 2},
		{"present_tense/has-been-replaced", 3},
		{"present_tense/we-switched", 4},
		{"present_tense/renamed-from", 5},
		{"present_tense/to-be-implemented", 6},
	} {
		if sev, ok := sevOf("history.md", w.rule, w.line); !ok || sev != "blocker" {
			t.Errorf("expected blocker %s on history.md:%d; got sev=%q ok=%v; all=%+v", w.rule, w.line, sev, ok, fs)
		}
	}

	// ambiguous.md: deprecated + migrated-from fire, but only as WARN (gate stays satisfiable).
	for _, w := range []struct {
		rule string
		line int
	}{
		{"present_tense/deprecated", 1},
		{"present_tense/migrated-from", 2},
	} {
		if sev, ok := sevOf("ambiguous.md", w.rule, w.line); !ok || sev != "warn" {
			t.Errorf("expected warn %s on ambiguous.md:%d; got sev=%q ok=%v", w.rule, w.line, sev, ok)
		}
	}

	// clean/escaped/fenced docs produce NO finding; nothing in ambiguous.md may block.
	for _, f := range fs {
		switch filepath.Base(f.File) {
		case "clean.md", "escaped.md", "fenced.md":
			t.Errorf("unexpected present-tense finding on %s: %+v", f.File, f)
		case "ambiguous.md":
			if f.Severity == "blocker" {
				t.Errorf("ambiguous present-state prose must not block the gate: %+v", f)
			}
		}
	}
}

func TestStrayRootDocs(t *testing.T) {
	root := t.TempDir()
	// Allowlisted regular files at root -> exempt.
	writeFile(t, root, "README.md", "# readme\n")
	writeFile(t, root, "AGENTS.md", "# agents\n")
	// A stray, non-allowlisted root markdown -> finding.
	writeFile(t, root, "NOTES.md", "# stray\n")
	// Markdown under a subdirectory is never touched (non-recursive).
	writeFile(t, root, "docs/guide.md", "# guide\n")
	// A symlink whose target stem is allowlisted -> exempt (the CLAUDE.md ->
	// AGENTS.md case). Judged by the resolved target, not the link name.
	if err := os.Symlink("AGENTS.md", filepath.Join(root, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	// A symlink with a missing target -> finding.
	if err := os.Symlink("GONE.md", filepath.Join(root, "BROKEN.md")); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		Rules: map[string]RuleConfig{
			"stray_root_docs": {Enabled: true, Severity: "blocker",
				Allowlist: []string{"README", "AGENTS", "CHANGELOG", "CONTRIBUTING", "SECURITY", "LICENSE"}},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}

	if !hasFinding(fs, "NOTES.md", "stray_root_docs", 0) {
		t.Errorf("expected stray finding on NOTES.md: %+v", fs)
	}
	if !hasFinding(fs, "BROKEN.md", "stray_root_docs", 0) {
		t.Errorf("expected broken-symlink finding on BROKEN.md: %+v", fs)
	}
	// README, AGENTS, the CLAUDE.md symlink, and docs/guide.md must not fire.
	for _, f := range fs {
		switch f.File {
		case "README.md", "AGENTS.md", "CLAUDE.md", filepath.Join("docs", "guide.md"):
			t.Errorf("unexpected stray finding on %s: %+v", f.File, f)
		}
	}
	if n := countRule(fs, "stray_root_docs"); n != 2 {
		t.Fatalf("expected exactly 2 stray_root_docs findings, got %d: %+v", n, fs)
	}
}

func TestDocsLinkResolveInDocs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "docs/target.md", "# target\n")
	writeFile(t, root, "docs/doc.md",
		"good: [t](target.md)\n"+
			"dir ok: [d](../docs/)\n"+
			"broken: [t](missing.md)\n")

	cfg := Config{
		Roots: []string{"docs"},
		Rules: map[string]RuleConfig{
			"links_resolve": {Enabled: true, Severity: "blocker"},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "links_resolve"); n != 1 {
		t.Fatalf("expected 1 link finding, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, filepath.Join("docs", "doc.md"), "links_resolve", 3) {
		t.Errorf("expected broken-link finding on docs/doc.md:3: %+v", fs)
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

func TestPersonaRegistry(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".abcd/development/personas.json",
		`{"schema_version":2,"personas":[{"name":"Alice","role_hints":["solo founder"]},{"name":"Bob","role_hints":["staff engineer"]}]}`)
	writeFile(t, root, "rec/ok.md", "# ok\n\n> \"Fine,\" said Alice, solo founder.\n")
	writeFile(t, root, "rec/bad.md", "# bad\n\n> \"Nope,\" said Zorro, pirate captain.\n")
	writeFile(t, root, "rec/fenced.md", "# fenced\n\n```\nsaid Zorro, pirate captain.\n```\n")
	writeFile(t, root, "rec/hist/old.md", "# old\n\n> \"Old,\" said Zorro, pirate captain.\n")

	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"persona_registry": {Enabled: true, Severity: "blocker", Registry: ".abcd/development/personas.json"},
		},
		ExemptPaths: []string{"rec/hist/"},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "persona_registry"); n != 1 {
		t.Fatalf("expected exactly 1 persona finding, got %d: %+v", n, fs)
	}
	if !hasFinding(fs, "rec/bad.md", "persona_registry", 3) {
		t.Errorf("expected finding on rec/bad.md:3: %+v", fs)
	}
}

func TestPersonaRegistryMissingFileErrors(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/ok.md", "# ok\n")
	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"persona_registry": {Enabled: true, Severity: "blocker", Registry: "nope/personas.json"},
		},
	}
	if _, err := Lint(cfg, root); err == nil {
		t.Fatal("expected error for enabled rule with missing registry")
	}
}

func TestPersonaRegistryEdgeCases(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "reg.json",
		`{"personas":[{"name":"Alice"},{"name":"Zoë"}]}`)
	writeFile(t, root, "rec/multi.md",
		"# m\n\n> \"A,\" said Alice, founder. \"B,\" said Zorro, pirate. And said Anne-Marie, sailor.\n")
	writeFile(t, root, "rec/unicode.md", "# u\n\n> \"Fine,\" said Zoë, designer.\n")
	writeFile(t, root, "rec/nocomma.md", "# n\n\n> as Zorro said before, nothing here attributes a quote.\n")

	cfg := Config{
		Roots: []string{"rec"},
		Rules: map[string]RuleConfig{
			"persona_registry": {Enabled: true, Severity: "blocker", Registry: "reg.json"},
		},
	}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if n := countRule(fs, "persona_registry"); n != 2 {
		t.Fatalf("expected 2 findings (Zorro + Anne-Marie on one line), got %d: %+v", n, fs)
	}
	if !hasFinding(fs, "rec/multi.md", "persona_registry", 3) {
		t.Errorf("expected findings on rec/multi.md:3: %+v", fs)
	}
}

func TestPersonaRegistryConfigGuards(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "rec/ok.md", "# ok\n")
	writeFile(t, root, "empty.json", `{"personas":[]}`)

	// Enabled with empty registry path: loud error, not a directory read.
	cfg := Config{Roots: []string{"rec"}, Rules: map[string]RuleConfig{
		"persona_registry": {Enabled: true, Severity: "blocker"},
	}}
	if _, err := Lint(cfg, root); err == nil {
		t.Fatal("expected error for enabled rule with empty registry path")
	}

	// Enabled with zero-persona roster: loud error, not whole-record flagging.
	cfg.Rules["persona_registry"] = RuleConfig{Enabled: true, Severity: "blocker", Registry: "empty.json"}
	if _, err := Lint(cfg, root); err == nil {
		t.Fatal("expected error for zero-persona roster")
	}

	// Disabled rule with missing registry: no error, no findings.
	cfg.Rules["persona_registry"] = RuleConfig{Enabled: false, Registry: "nope.json"}
	fs, err := Lint(cfg, root)
	if err != nil {
		t.Fatalf("disabled rule must not touch the registry: %v", err)
	}
	if n := countRule(fs, "persona_registry"); n != 0 {
		t.Fatalf("disabled rule produced findings: %+v", fs)
	}
}
