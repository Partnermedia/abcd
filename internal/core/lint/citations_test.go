package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDaysBetweenIsTimezoneStable pins that the staleness arithmetic depends
// only on the instants, not on the caller's timezone. The gate reads a local
// clock (time.Now()) while the baseline dates are stamped UTC, so a bare
// Year()/Month()/Day() put the 180-day boundary — and the release gate's
// citation-overdue blocker — one day apart for a maintainer east of UTC.
func TestDaysBetweenIsTimezoneStable(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	// 2026-06-29T20:00Z is already 2026-06-30 in UTC+13, so a local-date read
	// would return 180 there and 179 at UTC — the boundary crossing.
	inst := time.Date(2026, 6, 29, 20, 0, 0, 0, time.UTC)
	want := -1
	for _, offset := range []int{0, -7 * 3600, 13 * 3600} {
		got := DaysBetween(from, inst.In(time.FixedZone("z", offset)))
		if want < 0 {
			want = got
		}
		if got != want {
			t.Errorf("DaysBetween at offset %ds = %d, want %d (timezone must not shift the boundary)", offset, got, want)
		}
	}
	if want != 179 {
		t.Errorf("DaysBetween(2026-01-01, 2026-06-29T20:00Z) = %d, want 179", want)
	}
}

// citeRepo plants a fixture tree and returns its root. Keys are repo-relative
// slash paths.
func citeRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// citeConfig builds a docs-lint config with exactly one citation rule armed, so
// each test isolates the rule under examination.
func citeConfig(ruleID string, rc RuleConfig) Config {
	rc.Enabled = true
	if rc.Severity == "" {
		rc.Severity = severityBlocker
	}
	return Config{Roots: []string{"docs"}, Rules: map[string]RuleConfig{ruleID: rc}}
}

// findingsFor filters findings down to one rule id.
func findingsFor(findings []Finding, ruleID string) []Finding {
	var out []Finding
	for _, f := range findings {
		if f.RuleID == ruleID {
			out = append(out, f)
		}
	}
	return out
}

// fixedNow is the clock every staleness test runs against, so the arithmetic is
// deterministic rather than a function of the day the suite happens to run.
var fixedNow = time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)

// ---------------------------------------------------------------------------
// citation_footnotes — marker/definition bijection
// ---------------------------------------------------------------------------

// TestCitationFootnotesBijection pins the structural half of AC 1: on a page
// that cites, every marker resolves to a definition and every definition is
// referenced. Both directions matter — an orphan definition is a citation the
// page lost, which is exactly the silent rot the gate exists to catch.
func TestCitationFootnotesBijection(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/good.md": "# Good\n\nA claim.[^a] Another.[^b]\n\n" +
			"[^a]: Source A, https://example.org/a.\n\n[^b]: Source B, https://example.org/b.\n",
		"docs/orphan-marker.md": "# Orphan marker\n\nA claim.[^missing]\n",
		"docs/orphan-def.md": "# Orphan definition\n\nA claim.[^a]\n\n" +
			"[^a]: Source A, https://example.org/a.\n\n[^unused]: Source U, https://example.org/u.\n",
		"docs/duplicate-def.md": "# Duplicate\n\nA claim.[^a]\n\n" +
			"[^a]: Source A, https://example.org/a.\n\n[^a]: Source A again, https://example.org/a2.\n",
		"docs/no-footnotes.md": "# Plain page\n\nNo citations here at all.\n",
	})
	findings, err := LintAt(citeConfig("citation_footnotes", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	got := findingsFor(findings, "citation_footnotes")

	byFile := map[string][]string{}
	for _, f := range got {
		byFile[f.File] = append(byFile[f.File], f.Message)
	}
	if n := len(byFile[filepath.FromSlash("docs/good.md")]); n != 0 {
		t.Errorf("a page in bijection produced %d finding(s): %v", n, byFile[filepath.FromSlash("docs/good.md")])
	}
	if n := len(byFile[filepath.FromSlash("docs/no-footnotes.md")]); n != 0 {
		t.Errorf("a page with no footnotes produced %d finding(s)", n)
	}
	mustContain(t, byFile[filepath.FromSlash("docs/orphan-marker.md")], "[^missing]", "no definition")
	mustContain(t, byFile[filepath.FromSlash("docs/orphan-def.md")], "[^unused]", "never referenced")
	mustContain(t, byFile[filepath.FromSlash("docs/duplicate-def.md")], "[^a]", "more than once")
}

// mustContain asserts some message in the set mentions every needle.
func mustContain(t *testing.T, msgs []string, needles ...string) {
	t.Helper()
	for _, m := range msgs {
		hit := true
		for _, n := range needles {
			if !strings.Contains(m, n) {
				hit = false
				break
			}
		}
		if hit {
			return
		}
	}
	t.Errorf("no finding mentions all of %v; got %v", needles, msgs)
}

// TestCitationFootnotesIgnoresCode proves the parser reads prose, not code. A
// regexp character class `[^abc]` inside a fence or inline code is not a
// citation, and treating it as one would make the rule unusable on any page that
// documents a pattern.
func TestCitationFootnotesIgnoresCode(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/code.md": "# Code\n\nA real claim.[^a]\n\n" +
			"Inline `re.match(r'[^A-Za-z]')` is not a citation.\n\n" +
			"```bash\ngrep '[^0-9]' file\n```\n\n" +
			"[^a]: Source A, https://example.org/a.\n",
	})
	findings, err := LintAt(citeConfig("citation_footnotes", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if got := findingsFor(findings, "citation_footnotes"); len(got) != 0 {
		t.Errorf("code-fenced and inline-code bracket classes were read as footnotes: %+v", got)
	}
}

// TestCitationFootnotesIgnoresReferenceLinks guards the neighbouring syntax: a
// reference-style link definition `[label]: url` is not a footnote definition,
// and must not be counted as an orphan.
func TestCitationFootnotesIgnoresReferenceLinks(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/refs.md": "# Refs\n\nSee [the principle][srp].\n\n" +
			"[srp]: https://example.org/srp \"Single responsibility\"\n",
	})
	findings, err := LintAt(citeConfig("citation_footnotes", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if got := findingsFor(findings, "citation_footnotes"); len(got) != 0 {
		t.Errorf("reference-style link definitions were read as footnotes: %+v", got)
	}
}

// ---------------------------------------------------------------------------
// citation_crosswalk_rows — the spc-15 deferral
// ---------------------------------------------------------------------------

// TestCitationCrosswalkRows pins the rule deferred from spc-15: every body row
// of a crosswalk table carries at least one footnote. The identification
// heuristic is deliberately the narrowest that matches the repo's real crosswalk
// — the table's nearest preceding heading names it — so ordinary tables on the
// same page are untouched.
func TestCitationCrosswalkRows(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/crosswalk.md": "# Terminology crosswalk\n\n" +
			"## Plain table\n\n" +
			"| Directory | For |\n|---|---|\n| tutorials/ | learning |\n| how-to/ | tasks |\n\n" +
			"## The crosswalk\n\n" +
			"| Term | Established meaning | Position |\n" +
			"|------|---------------------|----------|\n" +
			"| **A** | Means a thing.[^a] | USES |\n" +
			"| **B** | Means another thing. | ADAPTS |\n" +
			"| **C** | Means a third thing.[^c] | REJECTS |\n\n" +
			"[^a]: Source A, https://example.org/a.\n\n[^c]: Source C, https://example.org/c.\n",
	})
	findings, err := LintAt(citeConfig("citation_crosswalk_rows", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	got := findingsFor(findings, "citation_crosswalk_rows")
	if len(got) != 1 {
		t.Fatalf("want exactly 1 finding (the uncited row **B**), got %d: %+v", len(got), got)
	}
	if !strings.Contains(got[0].Message, "footnote") {
		t.Errorf("message should say the row carries no footnote; got %q", got[0].Message)
	}
	// The uncited row **B** is the 15th line of the fixture (1-based).
	if got[0].Line != 15 {
		t.Errorf("finding anchored at line %d, want 15 (the **B** row)", got[0].Line)
	}
}

// TestCitationCrosswalkRowsLeavesOrdinaryTablesAlone is the false-positive
// budget for the heuristic. The repo's docs are full of ordinary tables with no
// citations at all; a rule that swept them up would be uninstallable.
func TestCitationCrosswalkRowsLeavesOrdinaryTablesAlone(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/README.md": "# Docs\n\n## Map\n\n" +
			"| Directory | Diataxis type | For |\n|---|---|---|\n" +
			"| tutorials/ | tutorial | learning |\n| how-to/ | how-to | tasks |\n",
	})
	findings, err := LintAt(citeConfig("citation_crosswalk_rows", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if got := findingsFor(findings, "citation_crosswalk_rows"); len(got) != 0 {
		t.Errorf("ordinary tables under a non-crosswalk heading were swept up: %+v", got)
	}
}

// ---------------------------------------------------------------------------
// citation_url_syntax
// ---------------------------------------------------------------------------

// TestCitationURLSyntax pins the syntax half of AC 1 against the punctuation the
// repo's real footnotes actually contain: several URLs on one line, sentence
// periods abutting them, and URLs closed inside parentheses.
func TestCitationURLSyntax(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/ok.md": "# OK\n\nClaim.[^a] Claim.[^b] Claim.[^c]\n\n" +
			"[^a]: Spec, https://example.org/spec/; repo, https://github.com/e/x.\n\n" +
			"[^b]: Page, https://example.org/page (mirror: https://example.net/page).\n\n" +
			"[^c]: Paper, NeurIPS 2020, DOI 10.48550/arXiv.2005.11401.\n",
		"docs/bad.md": "# Bad\n\nClaim.[^u] Claim.[^d]\n\n" +
			"[^u]: Source, https://exam ple.org/x and http://nohost/.\n\n" +
			"[^d]: Paper, DOI 10.notanumber/xyz.\n",
	})
	findings, err := LintAt(citeConfig("citation_url_syntax", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	got := findingsFor(findings, "citation_url_syntax")
	byFile := map[string][]string{}
	for _, f := range got {
		byFile[f.File] = append(byFile[f.File], f.Message)
	}
	if n := len(byFile[filepath.FromSlash("docs/ok.md")]); n != 0 {
		t.Errorf("well-formed citations flagged %d time(s): %v", n, byFile[filepath.FromSlash("docs/ok.md")])
	}
	mustContain(t, byFile[filepath.FromSlash("docs/bad.md")], "nohost")
	mustContain(t, byFile[filepath.FromSlash("docs/bad.md")], "10.notanumber/xyz")
}

// TestCitationURLSyntaxScopedToCitations proves the rule reads citation lines
// only. The word "DOI" in ordinary prose, and a URL in a body paragraph, are not
// the citation corpus and must not be judged as one.
func TestCitationURLSyntaxScopedToCitations(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/prose.md": "# Prose\n\nAdmission is strict: a peer-reviewed paper with a DOI, or a\n" +
			"specification. See http://nohost/ mentioned in passing.\n",
	})
	findings, err := LintAt(citeConfig("citation_url_syntax", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if got := findingsFor(findings, "citation_url_syntax"); len(got) != 0 {
		t.Errorf("prose outside a footnote definition was judged as a citation: %+v", got)
	}
}

// TestCitationDOIMentionIsAnchored is the false-positive budget for the DOI
// check. "DOI" is an ordinary English word in citation prose, and the resolvable
// https://doi.org/... form is a URL, not a bare DOI. Claiming either as a
// malformed DOI would make a blocker-severity rule fire on correct citations.
func TestCitationDOIMentionIsAnchored(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/doi.md": "# DOI\n\nA.[^a] B.[^b] C.[^c] D.[^d]\n\n" +
			"[^a]: Lewis, P., NeurIPS 2020, DOI https://doi.org/10.48550/arXiv.2005.11401.\n\n" +
			"[^b]: A standard with a DOI and a stable URL, https://example.org/a.\n\n" +
			"[^c]: Report, 2024. No DOI is assigned; see https://example.org/c.\n\n" +
			"[^d]: Paper, DOI 10.1145/3586183.3606763.\n",
	})
	findings, err := LintAt(citeConfig("citation_url_syntax", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if got := findingsFor(findings, "citation_url_syntax"); len(got) != 0 {
		t.Errorf("correct citations were flagged as malformed DOIs: %+v", got)
	}
}

// TestCitationDefinitionContinuationLines closes a silent gate hole: a footnote
// definition wrapped across lines still cites everything below its first line.
// Dropping the continuation would let a malformed address — or one the baseline
// records as broken — through the gate unexamined, which is worse than a noisy
// rule because it fails OPEN.
func TestCitationDefinitionContinuationLines(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/wrapped.md": "# Wrapped\n\nA.[^a] B.[^b]\n\n" +
			"[^a]: Author, Title, a long citation that wraps,\n" +
			"  see http://nohost/ for the record.\n\n" +
			"[^b]: Short, https://example.org/b.\n\n" +
			"Ordinary prose resumes here, mentioning http://alsonohost/ harmlessly.\n",
	})
	findings, err := LintAt(citeConfig("citation_url_syntax", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	got := findingsFor(findings, "citation_url_syntax")
	if len(got) != 1 {
		t.Fatalf("want exactly 1 finding (the wrapped nohost URL), got %d: %+v", len(got), got)
	}
	if !strings.Contains(got[0].Message, "http://nohost/") {
		t.Errorf("finding should name the continuation-line URL; got %q", got[0].Message)
	}
	// The continuation line, not the definition's opening line.
	if got[0].Line != 6 {
		t.Errorf("finding anchored at line %d, want 6 (the continuation line)", got[0].Line)
	}
}

// TestCitationBaselineDedupesIdenticalRefs: one line citing the same URL twice
// is one problem, and must not be reported twice.
func TestCitationBaselineDedupesIdenticalRefs(t *testing.T) {
	root := baselineRepo(t,
		"# Cites\n\nA.[^a]\n\n[^a]: S, https://example.org/gone and mirror https://example.org/gone.\n",
		`{"schema_version":1,"entries":{}}`)
	findings, err := LintAt(citeConfig("citation_baseline", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if got := findingsFor(findings, "citation_baseline"); len(got) != 1 {
		t.Fatalf("want 1 finding for one URL cited twice on one line, got %d: %+v", len(got), got)
	}
}

// ---------------------------------------------------------------------------
// citation_source_policy
// ---------------------------------------------------------------------------

// TestCitationSourcePolicyRefusesConfiguredDomains pins that the refused list
// comes from committed config, matches subdomains, and ignores a www. prefix.
func TestCitationSourcePolicyRefusesConfiguredDomains(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/sources.md": "# Sources\n\nA.[^a] B.[^b] C.[^c] D.[^d]\n\n" +
			"[^a]: Write-up, https://aggregator.example/post/1.\n\n" +
			"[^b]: Write-up, https://www.aggregator.example/post/2.\n\n" +
			"[^c]: Write-up, https://blog.aggregator.example/post/3.\n\n" +
			"[^d]: Spec, https://standards.example/spec.\n",
	})
	cfg := citeConfig("citation_source_policy", RuleConfig{RefusedDomains: []string{"aggregator.example"}})
	findings, err := LintAt(cfg, root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	got := findingsFor(findings, "citation_source_policy")
	if len(got) != 3 {
		t.Fatalf("want 3 refusals (bare, www., and subdomain), got %d: %+v", len(got), got)
	}
	for _, f := range got {
		if !strings.Contains(f.Message, "aggregator.example") {
			t.Errorf("finding should name the refused domain; got %q", f.Message)
		}
	}
}

// TestCitationSourcePolicyEmptyListRefusesNothing pins the shipped default. The
// repo documents no refused-domain list anywhere, so the config seeds empty and
// the rule is inert until a maintainer authors one — the gate never invents a
// blacklist.
func TestCitationSourcePolicyEmptyListRefusesNothing(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/sources.md": "# Sources\n\nA.[^a]\n\n[^a]: Write-up, https://aggregator.example/post/1.\n",
	})
	findings, err := LintAt(citeConfig("citation_source_policy", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if got := findingsFor(findings, "citation_source_policy"); len(got) != 0 {
		t.Errorf("an empty refused list produced findings: %+v", got)
	}
}

// ---------------------------------------------------------------------------
// citation_baseline
// ---------------------------------------------------------------------------

// baselineRepo plants a citing page plus a baseline document.
func baselineRepo(t *testing.T, page, baseline string) string {
	t.Helper()
	files := map[string]string{"docs/cites.md": page}
	if baseline != "" {
		files[DefaultBaselinePath] = baseline
	}
	return citeRepo(t, files)
}

// TestCitationBaselineEnforcement pins every offline baseline refusal in AC 1
// and the warn half of AC 3, against one page citing five URLs.
func TestCitationBaselineEnforcement(t *testing.T) {
	page := "# Cites\n\nA.[^a] B.[^b] C.[^c] D.[^d] E.[^e]\n\n" +
		"[^a]: OK, https://example.org/alive.\n\n" +
		"[^b]: Broken, https://example.org/broken.\n\n" +
		"[^c]: Drifted, https://example.org/drifted.\n\n" +
		"[^d]: Stale, https://example.org/stale.\n\n" +
		"[^e]: Absent, https://example.org/absent.\n"
	// fixedNow is 2026-07-27. 2026-07-01 is 26 days old (fresh); 2026-01-01 is
	// 207 days old (past the 180-day warn, short of the 365-day blocker).
	baseline := `{"schema_version":1,"entries":{
		"https://example.org/alive":{"final_url":"https://example.org/alive","last_checked":"2026-07-01","outcome":"alive","verification":"automatic","verified_on":"2026-07-01"},
		"https://example.org/broken":{"final_url":"https://example.org/broken","last_checked":"2026-07-01","outcome":"broken","verification":"automatic","verified_on":"2026-07-01"},
		"https://example.org/drifted":{"final_url":"https://example.org/moved-here","last_checked":"2026-07-01","outcome":"alive","verification":"automatic","verified_on":"2026-07-01"},
		"https://example.org/stale":{"final_url":"https://example.org/stale","last_checked":"2026-01-01","outcome":"alive","verification":"manual","verified_on":"2026-01-01"}
	}}`
	root := baselineRepo(t, page, baseline)
	findings, err := LintAt(citeConfig("citation_baseline", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	msgs := []string{}
	for _, f := range findingsFor(findings, "citation_baseline") {
		msgs = append(msgs, f.Severity+" "+f.Message)
	}
	mustContain(t, msgs, "https://example.org/absent", "not in the citation baseline")
	mustContain(t, msgs, "https://example.org/broken", "recorded broken")
	mustContain(t, msgs, "https://example.org/drifted", "https://example.org/moved-here")
	mustContain(t, msgs, "https://example.org/stale", "180", severityWarn)
	for _, m := range msgs {
		if strings.Contains(m, "https://example.org/alive") {
			t.Errorf("a fresh, alive, undrifted citation was flagged: %q", m)
		}
	}
}

// TestCitationBaselineStalenessBoundaries pins the exact 180-day edge and the
// distinct overdue finding the RELEASE gate promotes. The commit gate never
// blocks on the calendar (spc-17), so 365 days is reported at warn here under
// its own rule id — part (b) arms it as a blocker at release time.
func TestCitationBaselineStalenessBoundaries(t *testing.T) {
	cases := []struct {
		name        string
		lastChecked string
		wantRule    string
		wantSev     string
	}{
		{"179 days is fresh", "2026-01-29", "", ""},
		{"180 days warns", "2026-01-28", "citation_baseline", severityWarn},
		{"364 days warns", "2025-07-28", "citation_baseline", severityWarn},
		{"365 days is overdue", "2025-07-27", "citation_baseline_overdue", severityWarn},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page := "# Cites\n\nA.[^a]\n\n[^a]: S, https://example.org/a.\n"
			baseline := `{"schema_version":1,"entries":{"https://example.org/a":{` +
				`"final_url":"https://example.org/a","last_checked":"` + tc.lastChecked + `",` +
				`"outcome":"alive","verification":"automatic","verified_on":"` + tc.lastChecked + `"}}}`
			root := baselineRepo(t, page, baseline)
			findings, err := LintAt(citeConfig("citation_baseline", RuleConfig{}), root, fixedNow)
			if err != nil {
				t.Fatalf("LintAt: %v", err)
			}
			var aged []Finding
			for _, f := range findings {
				if strings.HasPrefix(f.RuleID, "citation_baseline") {
					aged = append(aged, f)
				}
			}
			if tc.wantRule == "" {
				if len(aged) != 0 {
					t.Fatalf("a %s entry produced findings: %+v", tc.name, aged)
				}
				return
			}
			if len(aged) != 1 {
				t.Fatalf("want 1 finding, got %d: %+v", len(aged), aged)
			}
			if aged[0].RuleID != tc.wantRule {
				t.Errorf("rule id = %q, want %q", aged[0].RuleID, tc.wantRule)
			}
			if aged[0].Severity != tc.wantSev {
				t.Errorf("severity = %q, want %q (the commit gate never calendar-blocks)", aged[0].Severity, tc.wantSev)
			}
		})
	}
}

// TestCitationBaselineOverdueSeverityIsConfigurable is the seam part (b) needs:
// the release gate promotes the overdue finding to a blocker without the commit
// gate's semantics changing.
func TestCitationBaselineOverdueSeverityIsConfigurable(t *testing.T) {
	page := "# Cites\n\nA.[^a]\n\n[^a]: S, https://example.org/a.\n"
	baseline := `{"schema_version":1,"entries":{"https://example.org/a":{` +
		`"final_url":"https://example.org/a","last_checked":"2024-01-01",` +
		`"outcome":"alive","verification":"automatic","verified_on":"2024-01-01"}}}`
	root := baselineRepo(t, page, baseline)
	cfg := citeConfig("citation_baseline", RuleConfig{OverdueSeverity: severityBlocker})
	findings, err := LintAt(cfg, root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	got := findingsFor(findings, "citation_baseline_overdue")
	if len(got) != 1 || got[0].Severity != severityBlocker {
		t.Fatalf("overdue_severity did not promote the finding: %+v", got)
	}
}

// TestCitationBaselineMissingFileFailsClosed: a repo that cites but has no
// baseline gets ONE actionable finding naming the refresh verb, not one per URL.
func TestCitationBaselineMissingFileFailsClosed(t *testing.T) {
	root := baselineRepo(t, "# Cites\n\nA.[^a] B.[^b]\n\n"+
		"[^a]: S, https://example.org/a.\n\n[^b]: S, https://example.org/b.\n", "")
	findings, err := LintAt(citeConfig("citation_baseline", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	got := findingsFor(findings, "citation_baseline")
	if len(got) != 1 {
		t.Fatalf("want exactly 1 finding for a missing baseline, got %d: %+v", len(got), got)
	}
	if got[0].Severity != severityBlocker || !strings.Contains(got[0].Message, "docs cite refresh") {
		t.Errorf("finding should block and name the refresh verb; got %+v", got[0])
	}
	if got[0].File != DefaultBaselinePath {
		t.Errorf("finding should anchor on the baseline path, got %q", got[0].File)
	}
}

// TestCitationBaselineAbsentWithNoCitations: a repo that cites nothing needs no
// baseline. The rule stays silent rather than nagging every repo that enables it.
func TestCitationBaselineAbsentWithNoCitations(t *testing.T) {
	root := citeRepo(t, map[string]string{"docs/plain.md": "# Plain\n\nNo citations.\n"})
	findings, err := LintAt(citeConfig("citation_baseline", RuleConfig{}), root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if got := findingsFor(findings, "citation_baseline"); len(got) != 0 {
		t.Errorf("a repo with no citations was nagged for a baseline: %+v", got)
	}
}

// TestCitationBaselineCorruptIsAnError separates a corrupt record from a missing
// one: the gate refuses to run rather than reporting misleading findings.
func TestCitationBaselineCorruptIsAnError(t *testing.T) {
	root := baselineRepo(t, "# Cites\n\nA.[^a]\n\n[^a]: S, https://example.org/a.\n",
		`{"schema_version":1,"entries":{"https://example.org/a":{"final_url":"x","last_checked":"2026-07-01","outcome":"alive","verification":"manual","verified_on":"2026-07-01","method":"browser"}}}`)
	if _, err := LintAt(citeConfig("citation_baseline", RuleConfig{}), root, fixedNow); err == nil {
		t.Fatal("a baseline smuggling a method field was accepted by the lint")
	}
}

// TestCitationRulesOffByDefault: an unconfigured repo sees no citation findings,
// so adding the family cannot retroactively fail a repo that never opted in.
func TestCitationRulesOffByDefault(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/x.md": "# X\n\nA.[^missing]\n\n[^unused]: S, https://exam ple.org/a.\n",
	})
	findings, err := LintAt(Config{Roots: []string{"docs"}}, root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	for _, f := range findings {
		if strings.HasPrefix(f.RuleID, "citation_") {
			t.Errorf("citation rule fired with no config: %+v", f)
		}
	}
}

// TestLintDelegatesToLintAt pins that the wall-clock entry point still exists
// and agrees with the injected-clock one on everything the clock does not touch.
func TestLintDelegatesToLintAt(t *testing.T) {
	root := citeRepo(t, map[string]string{
		"docs/x.md": "# X\n\nA.[^missing]\n",
	})
	cfg := citeConfig("citation_footnotes", RuleConfig{})
	a, err := Lint(cfg, root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	b, err := LintAt(cfg, root, fixedNow)
	if err != nil {
		t.Fatalf("LintAt: %v", err)
	}
	if len(a) != len(b) || len(a) != 1 {
		t.Fatalf("Lint and LintAt disagree: %+v vs %+v", a, b)
	}
}
