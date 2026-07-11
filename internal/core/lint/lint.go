// Package lint is abcd's record-drift gate: it reads a JSON config and lints
// the markdown design record, returning findings. It performs no I/O beyond
// reading files under a caller-supplied repo root — no printing, no os.Exit —
// so it is fully testable and reusable across surfaces.
package lint

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Finding is one lint violation. File is repo-relative; Line is 1-based (0 when
// the finding is not tied to a specific line, e.g. a missing directory README).
type Finding struct {
	File     string
	Line     int
	RuleID   string
	Severity string
	Message  string
}

const (
	severityBlocker = "blocker"
	severityWarn    = "warn"
)

var (
	// Top-level YAML frontmatter key (column 0, no indentation).
	fmKeyRe = regexp.MustCompile(`^([A-Za-z0-9_]+):(.*)$`)
	// Inline markdown link: [text](target). Also matches the link part of an
	// image (![alt](src)), which resolves the same way.
	linkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	// URL scheme prefix (http:, mailto:, ...); such targets are not local paths.
	schemeRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.\-]*:`)
	// Brittle line reference: some-file.md:171 (check D).
	brittleRefRe = regexp.MustCompile(`[A-Za-z0-9_./-]+\.md:\d+`)
	// Intent id embedded in a filename or a superseded_by value.
	intentIDRe   = regexp.MustCompile(`itd-\d+`)
	intentFileRe = regexp.MustCompile(`^itd-\d+.*\.md$`)
	// Surface registry Command cell: the bare "/abcd" top-level, or "/abcd:<name>".
	surfaceCmdRe  = regexp.MustCompile(`^/abcd(?::([a-z0-9-]+))?$`)
	specIDRe      = regexp.MustCompile(`^spc-`)
	supersededRe  = regexp.MustCompile(`^itd-\d+`)
	intentBuckets = map[string]bool{
		"drafts": true, "planned": true, "shipped": true,
		"disciplines": true, "superseded": true,
	}
)

// Lint runs every enabled check family against the record under repoRoot and
// returns the findings sorted deterministically. An error is returned only for
// malformed configuration (e.g. an uncompilable regexp, or a persona_registry
// roster that is missing, unparsable, or empty); a walkable-but-missing root
// is skipped, not an error.
func Lint(cfg Config, repoRoot string) ([]Finding, error) {
	var findings []Finding

	tokenChecks, err := compileTokens(cfg.BannedTokens)
	if err != nil {
		return nil, err
	}

	linksCfg, linksOn := cfg.Rules["links_resolve"]
	gitMetaCfg, gitMetaOn := cfg.Rules["no_git_metadata"]
	brittleCfg, brittleOn := cfg.Rules["no_brittle_line_refs"]
	linksOn = linksOn && linksCfg.Enabled
	gitMetaOn = gitMetaOn && gitMetaCfg.Enabled
	brittleOn = brittleOn && brittleCfg.Enabled

	personaCfg, personaOn := cfg.Rules["persona_registry"]
	personaOn = personaOn && personaCfg.Enabled
	var personaRoster map[string]bool
	if personaOn {
		personaRoster, err = loadPersonaRoster(repoRoot, personaCfg.Registry)
		if err != nil {
			return nil, err
		}
	}

	for _, root := range cfg.Roots {
		rootAbs := filepath.Join(repoRoot, root)
		mdFiles, err := markdownFiles(rootAbs)
		if err != nil {
			return nil, err
		}
		for _, fileAbs := range mdFiles {
			content, err := os.ReadFile(fileAbs)
			if err != nil {
				return nil, err
			}
			rel := repoRel(repoRoot, fileAbs)
			lines := strings.Split(string(content), "\n")
			mask := fenceMask(lines)
			// Content-drift checks (banned_tokens, intent_lifecycle) cover only
			// the forward-looking record; structural checks below stay universal.
			exempt := contentExempt(rel, frontmatterFields(lines), cfg)

			if len(tokenChecks) > 0 && !exempt {
				findings = append(findings, checkBannedTokens(rel, lines, mask, tokenChecks)...)
			}
			if personaOn && !exempt {
				findings = append(findings, checkPersonaRegistry(rel, lines, mask, personaRoster, personaCfg)...)
			}
			if gitMetaOn {
				findings = append(findings, checkGitMetadata(rel, lines, gitMetaCfg)...)
			}
			if linksOn {
				findings = append(findings, checkLinks(rel, fileAbs, repoRoot, lines, mask, linksCfg)...)
			}
			if brittleOn {
				findings = append(findings, checkBrittleRefs(rel, lines, brittleCfg)...)
			}
		}

		if dirCfg, ok := cfg.Rules["directory_coverage"]; ok && dirCfg.Enabled {
			dc, err := checkDirectoryCoverage(repoRoot, rootAbs, dirCfg)
			if err != nil {
				return nil, err
			}
			findings = append(findings, dc...)
		}

		if intentCfg, ok := cfg.Rules["intent_lifecycle"]; ok && intentCfg.Enabled {
			il, err := checkIntentLifecycle(repoRoot, rootAbs, intentCfg, cfg)
			if err != nil {
				return nil, err
			}
			findings = append(findings, il...)
		}
	}

	// stray_root_docs is repo-root scoped and non-recursive — independent of
	// cfg.Roots, so it runs once, outside the per-root loop.
	if strayCfg, ok := cfg.Rules["stray_root_docs"]; ok && strayCfg.Enabled {
		sr, err := checkStrayRootDocs(repoRoot, strayCfg)
		if err != nil {
			return nil, err
		}
		findings = append(findings, sr...)
	}

	// context_status_free targets one work-tier file (CONTEXT.md) that lives
	// outside cfg.Roots, so it too runs once, outside the per-root loop.
	if ctxCfg, ok := cfg.Rules["context_status_free"]; ok && ctxCfg.Enabled {
		cs, err := checkContextStatusFree(repoRoot, ctxCfg)
		if err != nil {
			return nil, err
		}
		findings = append(findings, cs...)
	}

	// surface_coverage cross-checks the plugin surface (commands/, skills/ —
	// outside cfg.Roots) against the brief's surface registry (inside a root), so
	// it too runs once, outside the per-root loop.
	if scCfg, ok := cfg.Rules["surface_coverage"]; ok && scCfg.Enabled {
		sc, err := checkSurfaceCoverage(repoRoot, scCfg)
		if err != nil {
			return nil, err
		}
		findings = append(findings, sc...)
	}

	sortFindings(findings)
	return findings, nil
}

// checkStrayRootDocs flags top-level markdown files whose basename stem is not
// in the configured allowlist. It is non-recursive (os.ReadDir on the repo root
// only) — subdirectories such as docs/ and .abcd/ are never touched. A root
// markdown file that is a symlink is judged by its resolved target's stem, so
// bridge links like CLAUDE.md -> AGENTS.md and GEMINI.md -> AGENTS.md are exempt
// without allowlisting the link name; a symlink with an unresolvable target is a
// finding.
func checkStrayRootDocs(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	allow := make(map[string]bool, len(cfg.Allowlist))
	for _, a := range cfg.Allowlist {
		allow[strings.ToUpper(a)] = true
	}

	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var out []Finding
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}

		// A root markdown symlink is judged by its resolved target's stem, not
		// its own name — this is what lets the CLAUDE.md / GEMINI.md bridges
		// (symlinks to the allowlisted AGENTS.md) pass. Known tradeoff: a stray
		// name pointed at an allowlisted target (e.g. NOTES.md -> AGENTS.md) is
		// also exempt. Acceptable here — creating such a symlink is a deliberate
		// act with the same trust level as adding the allowlisted file itself.
		stemName := name
		info, err := os.Lstat(filepath.Join(repoRoot, name))
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(filepath.Join(repoRoot, name))
			if err != nil {
				out = append(out, Finding{
					File: name, Line: 0, RuleID: "stray_root_docs",
					Severity: cfg.Severity,
					Message:  "top-level markdown symlink '" + name + "' has an unresolvable target",
				})
				continue
			}
			stemName = filepath.Base(resolved)
		}

		stem := strings.ToUpper(strings.TrimSuffix(stemName, filepath.Ext(stemName)))
		if allow[stem] {
			continue
		}
		out = append(out, Finding{
			File: name, Line: 0, RuleID: "stray_root_docs",
			Severity: cfg.Severity,
			Message:  "top-level markdown '" + name + "' is not in the root allowlist; documentation belongs under docs/",
		})
	}
	return out, nil
}

// contextStatusDefaultPatterns is the fallback line-match set for
// context_status_free when the config supplies no patterns. It targets the
// phase/status idioms an orientation doc drifts into (headings, **Current
// phase**, **Next:**, "Phase N — ...", and a status: frontmatter key).
var contextStatusDefaultPatterns = []string{
	`(?i)^#+\s*current (phase|status)`,
	`(?i)\*\*current phase`,
	`(?i)^\*\*next:`,
	`(?i)\bphase [0-9]+(\.[0-9]+)? — `,
	`(?i)^status:`,
}

const contextStatusMessage = "CONTEXT.md is status-free (DECISIONS.md 2026-07-10): status lives in the live surfaces (CLI, ledger), not in orientation docs"

// checkContextStatusFree scans a single target file (the work-tier CONTEXT.md,
// which sits outside cfg.Roots) line by line, flagging every line that matches
// any configured pattern. Fenced code blocks are masked via the shared
// fenceMask. A missing target is not an error — it yields no findings. This is a
// file-scoped rule: banning "Phase N" corpus-wide would explode, so the ban is
// confined to the one orientation doc that must stay status-free.
func checkContextStatusFree(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	target := cfg.Target
	if target == "" {
		target = filepath.Join(".abcd", "work", "CONTEXT.md")
	}
	rawPatterns := cfg.Patterns
	if len(rawPatterns) == 0 {
		rawPatterns = contextStatusDefaultPatterns
	}
	patterns := make([]*regexp.Regexp, 0, len(rawPatterns))
	for _, p := range rawPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, re)
	}

	fileAbs := filepath.Join(repoRoot, target)
	content, err := os.ReadFile(fileAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	rel := repoRel(repoRoot, fileAbs)
	lines := strings.Split(string(content), "\n")
	mask := fenceMask(lines)

	var out []Finding
	for i, line := range lines {
		if mask[i] {
			continue
		}
		if matchesAny(patterns, line) {
			out = append(out, Finding{
				File: rel, Line: i + 1, RuleID: "context_status_free",
				Severity: cfg.Severity, Message: contextStatusMessage,
			})
		}
	}
	return out, nil
}

// surfaceRow is one parsed row of the brief's surface registry table.
type surfaceRow struct {
	name   string // sub-verb after "/abcd:"; empty for the bare "/abcd" top-level
	status string // lower-cased "shipped" | "staged"; any other value is flagged
	line   int    // 1-based line of the row in the registry file
}

// checkSurfaceCoverage is the deterministic (Direction-B) half of the iss-35
// brief↔surface cross-check. It reads the plugin surface (commands/ + skills/,
// which live outside cfg.Roots) and the brief's surface registry table, then
// asserts three invariants:
//   - coverage: every real surface (a commands/abcd/*.md file or a skills/*/
//     directory) has a registry row;
//   - status fidelity: a row marked "shipped" has a backing surface, and a row
//     marked "staged" does not — the bare "/abcd" top-level is binary-backed,
//     has no command file, and is exempt from the file check;
//   - registry integrity: every row's status is "shipped" or "staged".
//
// The semantic half (a brief claim vs. binary behaviour — flags, exit codes,
// schema fields) stays an agent/release-gate check, not a structural lint. A
// missing registry file is not an error.
func checkSurfaceCoverage(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	if cfg.Registry == "" {
		return nil, nil
	}
	rows, err := parseSurfaceRegistry(repoRoot, cfg.Registry)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, nil // registry file absent — nothing to cross-check
	}
	real, realPaths, err := realSurfaces(repoRoot, cfg)
	if err != nil {
		return nil, err
	}

	var out []Finding
	rowNames := make(map[string]bool, len(rows))
	for _, r := range rows {
		if r.name != "" {
			rowNames[r.name] = true
		}
		switch r.status {
		case "shipped":
			if r.name != "" && !real[r.name] {
				out = append(out, Finding{
					File: cfg.Registry, Line: r.line, RuleID: "surface_coverage", Severity: cfg.Severity,
					Message: "surface row '" + surfaceLabel(r.name) + "' is marked shipped but no commands/ or skills/ surface backs it",
				})
			}
		case "staged":
			if r.name != "" && real[r.name] {
				out = append(out, Finding{
					File: cfg.Registry, Line: r.line, RuleID: "surface_coverage", Severity: cfg.Severity,
					Message: "surface row '" + surfaceLabel(r.name) + "' is marked staged but a backing surface exists; mark it shipped",
				})
			}
		default:
			out = append(out, Finding{
				File: cfg.Registry, Line: r.line, RuleID: "surface_coverage", Severity: cfg.Severity,
				Message: "surface row '" + surfaceLabel(r.name) + "' has unknown status '" + r.status + "' (want shipped|staged)",
			})
		}
	}

	// Coverage: every real surface must resolve to a registry row. Iterated in
	// sorted order so findings are deterministic before the final sort.
	names := make([]string, 0, len(realPaths))
	for name := range realPaths {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !rowNames[name] {
			out = append(out, Finding{
				File: realPaths[name], Line: 0, RuleID: "surface_coverage", Severity: cfg.Severity,
				Message: "surface '" + name + "' has no row in the brief surface registry (" + cfg.Registry + ")",
			})
		}
	}
	return out, nil
}

// surfaceLabel renders a surface name as its slash-command spelling.
func surfaceLabel(name string) string {
	if name == "" {
		return "/abcd"
	}
	return "/abcd:" + name
}

// realSurfaces enumerates the shipped plugin surface as a name set plus a
// name→repo-relative-path map. Command surfaces are the *.md files directly under
// CommandsDir (README excepted); skill surfaces are the immediate subdirectories
// of SkillsDir. A missing directory contributes nothing and is not an error.
func realSurfaces(repoRoot string, cfg RuleConfig) (map[string]bool, map[string]string, error) {
	set := map[string]bool{}
	paths := map[string]string{}

	if cfg.CommandsDir != "" {
		entries, err := os.ReadDir(filepath.Join(repoRoot, cfg.CommandsDir))
		if err != nil && !os.IsNotExist(err) {
			return nil, nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".md")
			if strings.EqualFold(name, "README") {
				continue
			}
			set[name] = true
			paths[name] = filepath.Join(cfg.CommandsDir, e.Name())
		}
	}

	if cfg.SkillsDir != "" {
		entries, err := os.ReadDir(filepath.Join(repoRoot, cfg.SkillsDir))
		if err != nil && !os.IsNotExist(err) {
			return nil, nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			set[e.Name()] = true
			paths[e.Name()] = filepath.Join(cfg.SkillsDir, e.Name())
		}
	}
	return set, paths, nil
}

// parseSurfaceRegistry reads the surface registry markdown and returns its table
// rows. It locates the one pipe-table whose header names both a Command and a
// Status column, then reads each data row's Command and Status cells (keyed by
// header position, so column order is free). Fenced code blocks are masked (the
// house convention, per fenceMask) so a table shown as a markdown *example* — a
// real risk in a doc whose subject is the surface table itself — is never
// mistaken for the registry. A missing file yields (nil, nil); a present file
// with no such table yields an empty, non-nil slice.
func parseSurfaceRegistry(repoRoot, registry string) ([]surfaceRow, error) {
	content, err := os.ReadFile(filepath.Join(repoRoot, registry))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	mask := fenceMask(lines)

	cmdCol, statusCol, headerIdx := -1, -1, -1
	for i, line := range lines {
		if mask[i] || !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		c, s := -1, -1
		for j, cell := range tableCells(line) {
			switch strings.ToLower(cell) {
			case "command":
				c = j
			case "status":
				s = j
			}
		}
		if c >= 0 && s >= 0 {
			cmdCol, statusCol, headerIdx = c, s, i
			break
		}
	}
	rows := []surfaceRow{}
	if headerIdx < 0 {
		return rows, nil
	}

	for i := headerIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if mask[i] || !strings.HasPrefix(trimmed, "|") {
			break // table ended (a fence closes it too)
		}
		if isTableSeparator(trimmed) {
			continue
		}
		// A row whose Command/Status cell is missing or malformed is skipped
		// here (status-fidelity unchecked) but still caught by the coverage
		// direction if it names a real backing surface — fail-loud, not silent.
		cells := tableCells(lines[i])
		if cmdCol >= len(cells) || statusCol >= len(cells) {
			continue
		}
		name, ok := parseSurfaceCommand(cells[cmdCol])
		if !ok {
			continue // not a /abcd surface row
		}
		rows = append(rows, surfaceRow{
			name:   name,
			status: strings.ToLower(strings.TrimSpace(cells[statusCol])),
			line:   i + 1,
		})
	}
	return rows, nil
}

// tableCells splits a markdown table row into trimmed cell strings, dropping the
// empty cells the leading and trailing border pipes produce.
func tableCells(line string) []string {
	parts := strings.Split(strings.TrimSpace(line), "|")
	cells := make([]string, 0, len(parts))
	for _, p := range parts {
		cells = append(cells, strings.TrimSpace(p))
	}
	if len(cells) > 0 && cells[0] == "" {
		cells = cells[1:]
	}
	if len(cells) > 0 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	return cells
}

// isTableSeparator reports whether a table row is the header separator (only
// pipes, dashes, colons, and whitespace).
func isTableSeparator(line string) bool {
	for _, r := range line {
		switch r {
		case '|', '-', ':', ' ', '\t':
		default:
			return false
		}
	}
	return true
}

// parseSurfaceCommand extracts the surface name from a Command cell. The bare
// "/abcd" yields ("", true); "/abcd:<name>" yields (name, true); anything else
// yields ("", false) so non-command rows are skipped rather than misread.
func parseSurfaceCommand(cell string) (string, bool) {
	c := strings.TrimSpace(strings.Trim(strings.TrimSpace(cell), "`"))
	m := surfaceCmdRe.FindStringSubmatch(c)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// tokenCheck is a compiled BannedToken ready for line matching.
type tokenCheck struct {
	token   BannedToken
	pattern *regexp.Regexp
	allow   []*regexp.Regexp
}

func compileTokens(tokens []BannedToken) ([]tokenCheck, error) {
	checks := make([]tokenCheck, 0, len(tokens))
	for _, t := range tokens {
		pat, err := regexp.Compile(t.Pattern)
		if err != nil {
			return nil, err
		}
		allow := make([]*regexp.Regexp, 0, len(t.AllowContext))
		for _, a := range t.AllowContext {
			re, err := regexp.Compile(a)
			if err != nil {
				return nil, err
			}
			allow = append(allow, re)
		}
		checks = append(checks, tokenCheck{token: t, pattern: pat, allow: allow})
	}
	return checks, nil
}

// checkBannedTokens implements check family A.
func checkBannedTokens(rel string, lines []string, mask []bool, checks []tokenCheck) []Finding {
	var out []Finding
	for i, line := range lines {
		for _, c := range checks {
			if c.token.skipFences() && mask[i] {
				continue
			}
			if !c.pattern.MatchString(line) {
				continue
			}
			if matchesAny(c.allow, line) {
				continue
			}
			out = append(out, Finding{
				File: rel, Line: i + 1, RuleID: c.token.ID,
				Severity: c.token.Severity, Message: c.token.Message,
			})
		}
	}
	return out
}

// checkGitMetadata implements check family B.
func checkGitMetadata(rel string, lines []string, cfg RuleConfig) []Finding {
	banned := make(map[string]bool, len(cfg.Fields))
	for _, f := range cfg.Fields {
		banned[f] = true
	}
	var out []Finding
	for key, field := range frontmatterFields(lines) {
		if banned[key] {
			out = append(out, Finding{
				File: rel, Line: field.line, RuleID: "no_git_metadata",
				Severity: cfg.Severity,
				Message:  "git-inferable metadata key '" + key + "' in frontmatter; git log/blame is canonical",
			})
		}
	}
	return out
}

// checkLinks implements check family C.
func checkLinks(rel, fileAbs, repoRoot string, lines []string, mask []bool, cfg RuleConfig) []Finding {
	fileDir := filepath.Dir(fileAbs)
	var out []Finding
	for i, line := range lines {
		if mask[i] {
			continue
		}
		for _, m := range linkRe.FindAllStringSubmatch(line, -1) {
			target := strings.TrimSpace(m[1])
			if target == "" || strings.HasPrefix(target, "#") ||
				strings.HasPrefix(target, "//") || schemeRe.MatchString(target) {
				continue
			}
			// Strip anchor/query before resolving to a filesystem path.
			if idx := strings.IndexAny(target, "#?"); idx >= 0 {
				target = target[:idx]
			}
			if target == "" {
				continue
			}
			resolved := filepath.Join(fileDir, target)
			relTo, err := filepath.Rel(repoRoot, resolved)
			if err != nil || relTo == ".." || strings.HasPrefix(relTo, ".."+string(filepath.Separator)) {
				out = append(out, Finding{
					File: rel, Line: i + 1, RuleID: "links_resolve",
					Severity: cfg.Severity,
					Message:  "link target escapes repo root: " + m[1],
				})
				continue
			}
			if _, err := os.Stat(resolved); err != nil {
				out = append(out, Finding{
					File: rel, Line: i + 1, RuleID: "links_resolve",
					Severity: cfg.Severity,
					Message:  "link target does not resolve: " + m[1],
				})
			}
		}
	}
	return out
}

// checkBrittleRefs implements check family D.
func checkBrittleRefs(rel string, lines []string, cfg RuleConfig) []Finding {
	var out []Finding
	for i, line := range lines {
		for _, m := range brittleRefRe.FindAllString(line, -1) {
			out = append(out, Finding{
				File: rel, Line: i + 1, RuleID: "no_brittle_line_refs",
				Severity: cfg.Severity,
				Message:  "brittle line reference '" + m + "'; use a section anchor instead",
			})
		}
	}
	return out
}

// checkDirectoryCoverage implements check family E.
func checkDirectoryCoverage(repoRoot, rootAbs string, cfg RuleConfig) ([]Finding, error) {
	var out []Finding
	err := filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		rel := repoRel(repoRoot, path)
		if matchesGlob(cfg.Exempt, rel) {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, "README.md")); err != nil {
			out = append(out, Finding{
				File: rel, Line: 0, RuleID: "directory_coverage",
				Severity: cfg.Severity,
				Message:  "directory has no README.md",
			})
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}

// checkIntentLifecycle implements check family F.
func checkIntentLifecycle(repoRoot, rootAbs string, cfg RuleConfig, top Config) ([]Finding, error) {
	intentsDir := cfg.IntentsDir
	if intentsDir == "" {
		intentsDir = "intents"
	}
	intentsRoot := filepath.Join(rootAbs, intentsDir)
	if _, err := os.Stat(intentsRoot); err != nil {
		return nil, nil
	}

	// Collect every intent id that exists as a file in any bucket, so the
	// superseded_by target can be checked for existence.
	known := map[string]bool{}
	_ = filepath.WalkDir(intentsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if intentFileRe.MatchString(d.Name()) {
			known[intentIDRe.FindString(d.Name())] = true
		}
		return nil
	})

	buckets, err := os.ReadDir(intentsRoot)
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, b := range buckets {
		if !b.IsDir() || !intentBuckets[b.Name()] {
			continue
		}
		bucket := b.Name()
		bucketDir := filepath.Join(intentsRoot, bucket)
		entries, err := os.ReadDir(bucketDir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !intentFileRe.MatchString(e.Name()) {
				continue
			}
			fileAbs := filepath.Join(bucketDir, e.Name())
			content, err := os.ReadFile(fileAbs)
			if err != nil {
				return nil, err
			}
			rel := repoRel(repoRoot, fileAbs)
			fields := frontmatterFields(strings.Split(string(content), "\n"))
			if contentExempt(rel, fields, top) {
				continue
			}
			out = append(out, validateIntent(rel, bucket, fields, known, cfg.Severity)...)
		}
	}
	return out, nil
}

func validateIntent(rel, bucket string, fields map[string]fmField, known map[string]bool, severity string) []Finding {
	var out []Finding
	add := func(line int, msg string) {
		if line == 0 {
			line = 1
		}
		out = append(out, Finding{
			File: rel, Line: line, RuleID: "intent_lifecycle",
			Severity: severity, Message: msg,
		})
	}

	// ANY bucket: lifecycle state is the directory, never a status: field.
	if st, ok := fields["status"]; ok {
		add(st.line, "status: key forbidden; lifecycle state is the bucket directory, not a field")
	}

	kind, kindOK := fields["kind"]
	spec, specOK := fields["spec_id"]
	kindNull := !kindOK || isNull(kind.value)
	specNull := !specOK || isNull(spec.value)
	kindVal := kind.value

	switch bucket {
	case "drafts":
		if !(kindNull || kindVal == "standalone" || kindVal == "bundle-member") {
			add(kind.line, "drafts: kind must be null, standalone, or bundle-member (got '"+kindVal+"')")
		}
		if !specNull {
			add(spec.line, "drafts: spec_id must be null")
		}
	case "planned":
		if kindNull || !(kindVal == "standalone" || kindVal == "bundle-member") {
			add(kind.line, "planned: kind must be standalone or bundle-member (non-null)")
		}
		// Re-baselined for the Go build: a planned intent is committed to build, but
		// its native spec_id is assigned only when the spec layer schedules it
		// (Phase 4). So spec_id may be null (unscheduled) or a spc- id — never other.
		if !specNull && !specIDRe.MatchString(spec.value) {
			add(spec.line, "planned: spec_id must be null (unscheduled) or match ^spc- (got '"+spec.value+"')")
		}
	case "shipped":
		if kindNull || !(kindVal == "standalone" || kindVal == "bundle-member") {
			add(kind.line, "shipped: kind must be standalone or bundle-member (non-null)")
		}
		if specNull {
			add(spec.line, "shipped: spec_id must be non-null")
		}
	case "disciplines":
		if kindVal != "discipline" {
			add(kind.line, "disciplines: kind must be discipline (got '"+kindVal+"')")
		}
		if !specNull {
			add(spec.line, "disciplines: spec_id must be null")
		}
	case "superseded":
		sup, supOK := fields["superseded_by"]
		if !supOK || !supersededRe.MatchString(sup.value) {
			add(sup.line, "superseded: superseded_by must be present and match ^itd-\\d+")
		} else if id := intentIDRe.FindString(sup.value); !known[id] {
			add(sup.line, "superseded: superseded_by target '"+id+"' does not exist in any bucket")
		}
		if kindNull {
			add(kind.line, "superseded: kind must be non-null")
		}
	}
	return out
}

// fmField is a frontmatter key's value and 1-based source line.
type fmField struct {
	value string
	line  int
}

// frontmatterFields returns the top-level keys of the leading YAML frontmatter
// (the block between the first two `---` lines). It is a line scanner, not a
// YAML parser: nested keys and list items are ignored, which is sufficient for
// the top-level checks here.
func frontmatterFields(lines []string) map[string]fmField {
	fields := map[string]fmField{}
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return fields
	}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if line == "---" {
			break
		}
		m := fmKeyRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := m[1]
		if _, exists := fields[key]; !exists {
			fields[key] = fmField{value: strings.TrimSpace(m[2]), line: i + 1}
		}
	}
	return fields
}

// contentExempt reports whether a file is excused from the content-drift checks
// (banned_tokens, intent_lifecycle) because it is part of the historical record:
// its repo-relative path begins with a configured prefix, or its leading
// frontmatter status: value is listed. Structural checks never consult this.
func contentExempt(rel string, fields map[string]fmField, cfg Config) bool {
	for _, p := range cfg.ExemptPaths {
		if strings.HasPrefix(rel, p) {
			return true
		}
	}
	if st, ok := fields["status"]; ok {
		for _, s := range cfg.ExemptIfStatus {
			if st.value == s {
				return true
			}
		}
	}
	return false
}

func isNull(v string) bool {
	return v == "" || v == "null" || v == "~"
}

// fenceMask marks lines that are inside (or are a marker for) a triple-backtick
// fenced code block.
func fenceMask(lines []string) []bool {
	mask := make([]bool, len(lines))
	inFence := false
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			mask[i] = true
			inFence = !inFence
			continue
		}
		mask[i] = inFence
	}
	return mask
}

func markdownFiles(rootAbs string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func matchesAny(res []*regexp.Regexp, s string) bool {
	for _, re := range res {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func matchesGlob(globs []string, rel string) bool {
	for _, g := range globs {
		if ok, _ := filepath.Match(g, rel); ok {
			return true
		}
	}
	return false
}

func repoRel(repoRoot, abs string) string {
	if rel, err := filepath.Rel(repoRoot, abs); err == nil {
		return rel
	}
	return abs
}

func sortFindings(f []Finding) {
	sort.SliceStable(f, func(i, j int) bool {
		if f[i].File != f[j].File {
			return f[i].File < f[j].File
		}
		if f[i].Line != f[j].Line {
			return f[i].Line < f[j].Line
		}
		if f[i].RuleID != f[j].RuleID {
			return f[i].RuleID < f[j].RuleID
		}
		return f[i].Message < f[j].Message
	})
}
