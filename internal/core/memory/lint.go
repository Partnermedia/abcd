package memory

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// lint.go — the `abcd memory lint` verb (fn-39): a full-store curator
// health-check. Page-local checks (MS001/MS002/ML001/MQ001/MQ003) per typed
// page, plus a corpus pass (MQ002 + per-source MQ003) that rebuilds the
// regenerable .coverage_index.json. Writes ONE logbook report; mutates no
// memory-store state. Exit contract: blocker -> 1; warn/info/clean -> 0.

// Finding is a single memory-lint finding.
type Finding struct {
	Code       string `json:"code"`
	Severity   string `json:"severity"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

// LintSummary tallies findings by severity.
type LintSummary struct {
	Blockers int `json:"blockers"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
}

// LintRequest is the input to Lint.
type LintRequest struct {
	RepoRoot string
	Now      time.Time
}

// LintResult is the structured result of Lint.
type LintResult struct {
	Findings      []Finding      `json:"findings"`
	Summary       LintSummary    `json:"summary"`
	CoverageIndex map[string]any `json:"coverage_index"`
	ReportDir     string         `json:"report_dir"`
	GeneratedAt   string         `json:"generated_at"`
	StorePath     string         `json:"store_path"`
	ExitCode      int            `json:"exit_code"`
}

var defaultSeverities = map[string]string{
	"MQ001": "warn",
	"MQ002": "warn",
	"MQ003": "info",
	"MS001": "info",
	"MS002": "blocker",
	"ML001": "blocker",
}

func severityFor(code string) string {
	if s, ok := defaultSeverities[code]; ok {
		return s
	}
	return "warn"
}

// ---------------------------------------------------------------------------
// Typed-page gate
// ---------------------------------------------------------------------------

func isTypedMemoryPagePath(mem, path string) bool {
	base := filepath.Base(path)
	if siblingFiles[base] {
		return false
	}
	if _, _, _, ok := ParsePageFilename(base); !ok {
		return false
	}
	rel, err := filepath.Rel(mem, path)
	if err != nil {
		return false
	}
	segs := strings.Split(filepath.ToSlash(rel), "/")
	for i := 0; i < len(segs)-1; i++ {
		if segs[i] == "sources" {
			return false
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	fm, err := parseFrontmatter(string(raw))
	if err != nil {
		return false
	}
	_, ok := fm["source"]
	return ok
}

// ---------------------------------------------------------------------------
// Page-local linter
// ---------------------------------------------------------------------------

type memoryLinter struct {
	pagePath    string
	repoRoot    string
	content     string
	frontmatter map[string]any
	sourceLine  int
	findings    []Finding
}

func newMemoryLinter(pagePath, repoRoot, content string) *memoryLinter {
	fm, err := parseFrontmatter(content)
	if err != nil {
		fm = map[string]any{}
	}
	return &memoryLinter{
		pagePath:    pagePath,
		repoRoot:    repoRoot,
		content:     content,
		frontmatter: fm,
		sourceLine:  frontmatterKeyLine(content, "source"),
	}
}

func (l *memoryLinter) emit(code, message string, line int, suggestion string) {
	l.findings = append(l.findings, Finding{
		Code:       code,
		Severity:   severityFor(code),
		File:       l.pagePath,
		Line:       line,
		Message:    message,
		Suggestion: suggestion,
	})
}

func (l *memoryLinter) sourceBlock() map[string]any {
	if src, ok := l.frontmatter["source"].(map[string]any); ok {
		return src
	}
	return map[string]any{}
}

func (l *memoryLinter) derivedClasses() []string {
	src := l.sourceBlock()
	var raw []string
	if classes, ok := src["classes"].([]any); ok {
		for _, c := range classes {
			if s, ok := c.(string); ok {
				raw = append(raw, s)
			}
		}
	} else if single, ok := src["class"].(string); ok {
		raw = []string{single}
	}
	var out []string
	for _, c := range raw {
		if !contains(out, c) {
			out = append(out, c)
		}
	}
	return out
}

func (l *memoryLinter) run() []Finding {
	l.checkSourceClasses()
	l.checkLicence()
	l.checkQuotation()
	return l.findings
}

func (l *memoryLinter) checkSourceClasses() {
	classes := l.derivedClasses()
	switch {
	case len(classes) == 1:
		l.emit("MS001",
			fmt.Sprintf("memory page synthesised from a single source class (%s) — low cross-validation (advisory).", classes[0]),
			l.sourceLine,
			"Cross-validate the page against a second source class when one becomes available; advisory only.")
	case len(classes) >= 2:
		note, ok := l.sourceBlock()["weighting_note"].(string)
		if !ok || strings.TrimSpace(note) == "" {
			l.emit("MS002",
				fmt.Sprintf("memory page mixes %d source classes (%s) without a `source.weighting_note` acknowledging the asymmetric trust gradient between class types.", len(classes), strings.Join(classes, ", ")),
				l.sourceLine,
				"Add `weighting_note: \"<text>\"` under `source:` explaining how the classes are weighted against each other.")
		}
	}
}

func hasLicence(block map[string]any) bool {
	lic, ok := block["licence"].(string)
	return ok && strings.TrimSpace(lic) != ""
}

func (l *memoryLinter) checkLicence() {
	src := l.sourceBlock()
	if sources, ok := src["sources"].([]any); ok {
		for idx, e := range sources {
			em, ok := e.(map[string]any)
			if !ok {
				continue
			}
			if isExternalClass(em["class"]) && !hasLicence(em) {
				l.emit("ML001",
					fmt.Sprintf("source.sources[%d] (class %v) has no per-source `licence` field; explicit `licence: unknown` is acceptable — missing is the violation.", idx, em["class"]),
					l.sourceLine,
					fmt.Sprintf("Add `licence: <spdx-id|declared-by-user|unknown>` to source.sources[%d].", idx))
			}
		}
		return
	}
	if isExternalClass(src["class"]) && !hasLicence(src) {
		l.emit("ML001",
			fmt.Sprintf("memory page with `source.class: %v` has no `licence` field; explicit `licence: unknown` is acceptable — missing is the violation.", src["class"]),
			l.sourceLine,
			"Add `licence: <spdx-id|declared-by-user|unknown>` under `source:`.")
	}
}

func (l *memoryLinter) checkQuotation() {
	spans := extractQuotedSpans(l.content)
	if len(spans) == 0 {
		return
	}
	src := l.sourceBlock()
	externals := externalSourceHashes(src)
	if len(externals) == 0 {
		l.emit("MQ003",
			"[reason: no_external_source] page has quoted spans but no `external_*` source to budget — quotation budget skipped (coverage unavailable).",
			l.sourceLine, "")
		return
	}
	budget := loadQuotationBudget(l.repoRoot)
	for _, span := range spans {
		if span.tokenCount > budget.MaxContiguousQuoteWords {
			l.emit("MQ001",
				fmt.Sprintf("contiguous quoted span of %d words exceeds the %d-word ceiling.", span.tokenCount, budget.MaxContiguousQuoteWords),
				span.line,
				"Summarise the passage in your own words; quote only the load-bearing fragment.")
		}
	}
	pageTokens := pageQuotedTokenTotal(spans)
	registry, err := LoadRegistry(SourcesIndexPath(l.repoRoot))
	if err != nil {
		registry = nil // corrupt index -> every lookup degrades to malformed
	}
	for _, sh := range externals {
		tokenCount, reason := lookupSourceTokenCount(registry, sh)
		if reason != "" {
			fix := "Re-ingest the source (or backfill `source_token_count` in `.sources_index.json`)."
			if reason == reasonMalformedEntry {
				fix = "Repair the corrupt registry entry in `.abcd/memory/.sources_index.json`."
			}
			l.emit("MQ003",
				fmt.Sprintf("[reason: %s] coverage unavailable for source %s — per-page quotation budget skipped for this source.", reason, short12(sh)),
				l.sourceLine, fix)
			continue
		}
		pct := float64(pageTokens) / float64(tokenCount)
		if pct > budget.PerPagePct {
			l.emit("MQ001",
				fmt.Sprintf("page quotes %s of source %s (%d/%d tokens) — over the %.0f%% per-page budget.", fmtPct(pct), short12(sh), pageTokens, tokenCount, budget.PerPagePct*100),
				l.sourceLine,
				"Summarise quoted passages in your own words until the page is back under the per-page budget.")
		}
	}
}

// ---------------------------------------------------------------------------
// Full-corpus coverage lint (MQ002 + per-source MQ003)
// ---------------------------------------------------------------------------

func runMemoryCoverageLint(repoRoot string) ([]Finding, map[string]any, error) {
	mem := Dir(repoRoot)
	indexPath := CoverageIndexPath(repoRoot)
	report := map[string]any{
		"path":            indexPath,
		"stale":           false,
		"old_fingerprint": nil,
		"new_fingerprint": nil,
		"written":         false,
	}
	if fi, err := os.Stat(mem); err != nil || !fi.IsDir() {
		return nil, report, nil
	}

	var pages []crawledPage
	err := filepath.WalkDir(mem, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if !isTypedMemoryPagePath(mem, path) {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(mem, path)
		pages = append(pages, crawledPage{rel: filepath.ToSlash(rel), text: string(raw)})
		return nil
	})
	if err != nil {
		return nil, report, err
	}
	sort.Slice(pages, func(i, j int) bool { return pages[i].rel < pages[j].rel })

	budget := loadQuotationBudget(repoRoot)
	registry, regErr := LoadRegistry(SourcesIndexPath(repoRoot))
	if regErr != nil {
		registry = nil
	}
	result := buildCoverage(pages, registry, budget)

	oldFP := readStoredFingerprint(indexPath)
	if _, err := writeCoverageIndex(indexPath, result, budget); err != nil {
		return nil, report, err
	}
	report["stale"] = oldFP != "" && oldFP != result.fingerprint
	if oldFP != "" {
		report["old_fingerprint"] = oldFP
	}
	report["new_fingerprint"] = result.fingerprint
	report["written"] = true

	var findings []Finding
	unavailKeys := make([]string, 0, len(result.unavailable))
	for sh := range result.unavailable {
		unavailKeys = append(unavailKeys, sh)
	}
	sort.Strings(unavailKeys)
	for _, sh := range unavailKeys {
		reason := result.unavailable[sh]
		fix := "Re-ingest the source (or backfill `source_token_count` in `.sources_index.json`)."
		if reason == reasonMalformedEntry {
			fix = "Repair the corrupt registry entry in `.abcd/memory/.sources_index.json`."
		}
		findings = append(findings, Finding{
			Code: "MQ003", Severity: "info", File: indexPath,
			Message:    fmt.Sprintf("[reason: %s] cumulative coverage unavailable for source %s — MQ002 skipped for this source.", reason, short12(sh)),
			Suggestion: fix,
		})
	}

	srcKeys := make([]string, 0, len(result.sources))
	for sh := range result.sources {
		srcKeys = append(srcKeys, sh)
	}
	sort.Strings(srcKeys)
	warnPct := budget.CumulativeWarnPct
	blockPct := budget.CumulativeBlockPct
	for _, sh := range srcKeys {
		cov := result.sources[sh]
		pagesStr := strings.Join(cov.Pages, ", ")
		switch {
		case cov.CoverageUnambiguous >= blockPct:
			findings = append(findings, Finding{
				Code: "MQ002", Severity: "blocker", File: indexPath,
				Message: fmt.Sprintf("cumulative quoted coverage of source %s is %s — unambiguous single-source attribution alone is %s, at or over the %.0f%% block threshold. Pages: %s.",
					short12(sh), fmtPct(cov.CoverageTotal), fmtPct(cov.CoverageUnambiguous), blockPct*100, pagesStr),
				Suggestion: "Rewrite quoted passages as summaries until coverage drops below the threshold.",
			})
		case cov.CoverageTotal >= blockPct:
			findings = append(findings, Finding{
				Code: "MQ002", Severity: "warn", File: indexPath,
				Message: fmt.Sprintf("cumulative quoted coverage of source %s is %s (>= the %.0f%% block threshold) but is driven by AMBIGUOUS multi-source attribution (unambiguous: %s) — capped at warn; only unambiguous single-source coverage can block. Pages: %s.",
					short12(sh), fmtPct(cov.CoverageTotal), blockPct*100, fmtPct(cov.CoverageUnambiguous), pagesStr),
				Suggestion: "Review the multi-source pages quoting this source; rewrite quotes as summaries to reduce coverage.",
			})
		case cov.CoverageTotal >= warnPct:
			findings = append(findings, Finding{
				Code: "MQ002", Severity: "warn", File: indexPath,
				Message: fmt.Sprintf("cumulative quoted coverage of source %s is %s — in the warn band [%.0f%%, %.0f%%). Pages: %s.",
					short12(sh), fmtPct(cov.CoverageTotal), warnPct*100, blockPct*100, pagesStr),
				Suggestion: "Rewrite quoted passages as summaries before coverage reaches the block threshold.",
			})
		}
	}
	return findings, report, nil
}

// ---------------------------------------------------------------------------
// Lint orchestration
// ---------------------------------------------------------------------------

// Lint runs the full-store curator health-check and writes one logbook report.
// Mutates no memory-store state (only the regenerable coverage index + report).
func Lint(req LintRequest) (LintResult, error) {
	root := req.RepoRoot
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	mem := Dir(root)

	var findings []Finding
	if fi, err := os.Stat(mem); err == nil && fi.IsDir() {
		var pagePaths []string
		err := filepath.WalkDir(mem, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			if isTypedMemoryPagePath(mem, path) {
				pagePaths = append(pagePaths, path)
			}
			return nil
		})
		if err != nil {
			return LintResult{}, err
		}
		sort.Strings(pagePaths)
		for _, path := range pagePaths {
			raw, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			findings = append(findings, newMemoryLinter(path, root, string(raw)).run()...)
		}
	}

	corpusFindings, coverageReport, err := runMemoryCoverageLint(root)
	if err != nil {
		return LintResult{}, err
	}
	findings = append(findings, corpusFindings...)

	summary := LintSummary{}
	for _, f := range findings {
		switch f.Severity {
		case "blocker":
			summary.Blockers++
		case "warn":
			summary.Warnings++
		case "info":
			summary.Infos++
		}
	}
	exitCode := 0
	if summary.Blockers > 0 {
		exitCode = 1
	}
	generatedAt := now.Format("2006-01-02T15:04:05Z")
	coverageIndex := map[string]any{}
	if coverageReport != nil {
		coverageIndex = coverageReport
	}

	reportDir, err := lintReportDir(root, now)
	if err != nil {
		return LintResult{}, err
	}
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		return LintResult{}, err
	}
	reportFields := map[string]any{
		"findings":       findingsToMaps(findings),
		"summary":        map[string]any{"blockers": summary.Blockers, "warnings": summary.Warnings, "infos": summary.Infos},
		"coverage_index": coverageIndex,
		"generated_at":   generatedAt,
		"store_path":     mem,
	}
	if err := durableWrite(filepath.Join(reportDir, "report.json"), marshalIndentNoEscape(reportFields)); err != nil {
		return LintResult{}, err
	}
	if err := durableWrite(filepath.Join(reportDir, "report.md"), renderLintReportMD(reportFields)); err != nil {
		return LintResult{}, err
	}

	return LintResult{
		Findings:      findings,
		Summary:       summary,
		CoverageIndex: coverageIndex,
		ReportDir:     reportDir,
		GeneratedAt:   generatedAt,
		StorePath:     mem,
		ExitCode:      exitCode,
	}, nil
}

func lintReportDir(repoRoot string, now time.Time) (string, error) {
	ts := now.Format("20060102T150405.000000Z")
	logbook := filepath.Join(repoRoot, ".abcd", "logbook", "memory")
	base := filepath.Join(logbook, "lint-"+ts)
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return base, nil
	}
	for n := 1; n < 1000; n++ {
		candidate := filepath.Join(logbook, fmt.Sprintf("lint-%s-%03d", ts, n))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique lint logbook dir for %s", ts)
}

func findingsToMaps(findings []Finding) []any {
	out := make([]any, len(findings))
	for i, f := range findings {
		out[i] = map[string]any{
			"code": f.Code, "severity": f.Severity, "file": f.File,
			"line": f.Line, "message": f.Message, "suggestion": f.Suggestion,
		}
	}
	return out
}

func renderLintReportMD(fields map[string]any) string {
	summary, _ := fields["summary"].(map[string]any)
	cov, _ := fields["coverage_index"].(map[string]any)
	lines := []string{
		"# /abcd:memory lint — curator health-check",
		"",
		fmt.Sprintf("Generated: %v", fields["generated_at"]),
		fmt.Sprintf("Store: %v", fields["store_path"]),
		fmt.Sprintf("Summary: %d blocker(s), %d warning(s), %d info(s)",
			toInt(summary["blockers"]), toInt(summary["warnings"]), toInt(summary["infos"])),
	}
	if written, _ := cov["written"].(bool); written {
		newFP, _ := cov["new_fingerprint"].(string)
		if stale, _ := cov["stale"].(bool); stale {
			oldFP, _ := cov["old_fingerprint"].(string)
			lines = append(lines, fmt.Sprintf("Coverage index: rebuilt (fingerprint %s); drift detected against the stored index (was %s) — the fresh crawl is authoritative.", short12(newFP), short12(oldFP)))
		} else {
			lines = append(lines, fmt.Sprintf("Coverage index: rebuilt (fingerprint %s); no drift against the stored index.", short12(newFP)))
		}
	} else {
		lines = append(lines, "Coverage index: not written (no memory store present).")
	}
	lines = append(lines, "Exit contract: blockers exit nonzero; warnings are curator-advisory (exit 0, non-blocking); infos never affect exit.", "")
	findings, _ := fields["findings"].([]any)
	if len(findings) == 0 {
		lines = append(lines, "No findings — store is clean.")
	} else {
		lines = append(lines, "## Findings", "")
		for _, sev := range []string{"blocker", "warn", "info"} {
			for _, fa := range findings {
				f, _ := fa.(map[string]any)
				if f["severity"] != sev {
					continue
				}
				loc := fmt.Sprintf("%v", f["file"])
				if line := toInt(f["line"]); line != 0 {
					loc += fmt.Sprintf(":%d", line)
				}
				lines = append(lines, fmt.Sprintf("- [%s] %v %s — %v", sev, f["code"], loc, f["message"]))
				if sug, _ := f["suggestion"].(string); sug != "" {
					lines = append(lines, "  fix: "+sug)
				}
			}
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func short12(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
