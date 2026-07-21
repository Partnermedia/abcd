// Package lint is abcd's record-drift gate: it reads a JSON config and lints
// the markdown design record, returning findings. It performs no I/O beyond
// reading files under a caller-supplied repo root — no printing, no os.Exit —
// so it is fully testable and reusable across surfaces.
package lint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/REPPL/abcd-cli/internal/core/frontmatter"
	"github.com/REPPL/abcd-cli/internal/fsutil"
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
	// Issue id embedded in a ledger filename (issue_id_unique).
	issueIDRe   = regexp.MustCompile(`iss-\d+`)
	issueFileRe = regexp.MustCompile(`^iss-\d+.*\.md$`)
	// Surface registry Command cell: the bare "/abcd" top-level, or "/abcd:<name>".
	surfaceCmdRe = regexp.MustCompile(`^/abcd(?::([a-z0-9-]+))?$`)
	// receipt_gate arming inputs are release-time and become externally supplied
	// (release.yml) — validated as safe path components before use.
	receiptShaRe  = regexp.MustCompile(`^[0-9a-f]{7,64}$`)
	receiptGateRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	// gate_lockstep hand-parsers (no YAML library — the repo has none and adds no
	// dependency): a markdown numbered-list item; the `jobs:` line; a 2-space job
	// header (identifier, optional quotes/trailing comment, never a comment line);
	// a job's `steps:` key; a step list-item marker; and a step `name:` key line.
	// A non-empty floor (RuleConfig.MinGates) is the fail-closed backstop that
	// makes any parser under-count block rather than silently pass.
	numberedItemRe = regexp.MustCompile(`^\s*\d+\.\s+(.+?)\s*$`)
	jobsSectionRe  = regexp.MustCompile(`^jobs:\s*(#.*)?$`)
	jobHeaderRe    = regexp.MustCompile(`^  ["']?([A-Za-z0-9_-]+)["']?:\s*(#.*)?$`)
	stepsKeyRe     = regexp.MustCompile(`^\s+steps:\s*(#.*)?$`)
	stepMarkerRe   = regexp.MustCompile(`^\s+-\s+(.*\S)\s*$`)
	stepNameKeyRe  = regexp.MustCompile(`^\s+name:\s*(.+?)\s*$`)
	specIDRe       = regexp.MustCompile(`^spc-`)
	supersededRe   = regexp.MustCompile(`^itd-\d+`)
	// spec_lifecycle: anchored id/link/filename matchers for the spec store.
	specFileRe     = regexp.MustCompile(`^spc-\d+.*\.md$`)
	specIDFullRe   = regexp.MustCompile(`^spc-\d+$`)
	intentIDFullRe = regexp.MustCompile(`^itd-\d+$`)
	intentBuckets  = map[string]bool{
		"drafts": true, "planned": true, "shipped": true,
		"disciplines": true, "superseded": true,
	}
	// issueStatusDirs are the issue ledger's status directories (issue_id_unique
	// scans all three for a duplicated iss-N id).
	issueStatusDirs = []string{"open", "resolved", "wontfix"}
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
				findings = append(findings, checkBrittleRefs(rel, lines, mask, brittleCfg)...)
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

		if specCfg, ok := cfg.Rules["spec_lifecycle"]; ok && specCfg.Enabled {
			sl, err := checkSpecLifecycle(repoRoot, rootAbs, specCfg, cfg)
			if err != nil {
				return nil, err
			}
			findings = append(findings, sl...)
		}

		if fsCfg, ok := cfg.Rules["forbidden_synonyms"]; ok && fsCfg.Enabled {
			fs, err := checkForbiddenSynonyms(repoRoot, rootAbs, fsCfg)
			if err != nil {
				return nil, err
			}
			findings = append(findings, fs...)
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

	// receipt_gate is the release-time verification of the semantic gates. It is
	// disabled for ordinary development (a commit under review has no receipt yet)
	// and armed only at release time with a target commit; it reads sha-keyed
	// receipts outside cfg.Roots, so it also runs once here.
	if rgCfg, ok := cfg.Rules["receipt_gate"]; ok && rgCfg.Enabled {
		rg, err := checkReceiptGate(repoRoot, rgCfg)
		if err != nil {
			return nil, err
		}
		findings = append(findings, rg...)
	}

	// gate_lockstep keeps the release-gate runbook's deterministic-gate list in
	// lockstep with the CI workflow (both outside cfg.Roots), so it runs once here.
	if glCfg, ok := cfg.Rules["gate_lockstep"]; ok && glCfg.Enabled {
		gl, err := checkGateLockstep(repoRoot, glCfg)
		if err != nil {
			return nil, err
		}
		findings = append(findings, gl...)
	}

	// issue_id_unique scans the issue ledger (.abcd/work/issues/{open,resolved,
	// wontfix} — outside cfg.Roots) for an iss-N id claimed by two or more files,
	// so it too runs once here.
	if iiCfg, ok := cfg.Rules["issue_id_unique"]; ok && iiCfg.Enabled {
		ii, err := checkIssueIDUnique(repoRoot, iiCfg)
		if err != nil {
			return nil, err
		}
		findings = append(findings, ii...)
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

// receipt is the parsed shape of a semantic-pass receipt — a Verification
// Summary Attestation (VSA): only the fields the release gate checks.
type receipt struct {
	Subject struct {
		Digest struct {
			GitCommit string `json:"gitCommit"`
		} `json:"digest"`
	} `json:"subject"`
	VerificationResult string `json:"verificationResult"`
	JudgeModel         string `json:"judgeModel"`
	// Policy.Detector names the gate this receipt attests. The gate binding is the
	// receipt's content, not its filename — so a genuine receipt for one detector
	// cannot be copied to another gate's path and satisfy it.
	Policy struct {
		Detector string `json:"detector"`
	} `json:"policy"`
}

// checkReceiptGate is the fail-closed, release-time verification of the semantic
// gates (the LLM passes CI cannot run). For the target commit it asserts that
// every required gate has a receipt whose subject digest is that commit, whose
// verdict is PROMOTE, and which pins a judge model. A missing, mismatched,
// malformed, HOLD, or model-less receipt BLOCKS — an un-run semantic pass is
// never a silent pass. The rule is release-time only: it stays disabled for
// ordinary development (a commit under review has no receipt yet), and
// release.yml supplies the target commit when it arms the rule. An enabled rule
// with no configured commit fails closed rather than passing vacuously.
func checkReceiptGate(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	dir := cfg.ReceiptsDir
	if dir == "" {
		dir = filepath.Join(".abcd", "work", "reviews")
	}
	// Every arming-input defect fails closed with a single finding — an armed gate
	// that has nothing valid to check is never a pass. Guards are symmetric so no
	// missing/blank/malformed arming input can slip through to a vacuous success.
	failClosed := func(msg string) []Finding {
		return []Finding{{File: dir, Line: 0, RuleID: "receipt_gate", Severity: cfg.Severity, Message: msg}}
	}
	if cfg.Commit == "" {
		return failClosed("receipt_gate is enabled but no target commit is configured; the release gate fails closed"), nil
	}
	if !receiptShaRe.MatchString(cfg.Commit) {
		return failClosed("receipt_gate target commit '" + cfg.Commit + "' is not a valid commit sha; the release gate fails closed"), nil
	}
	if len(cfg.RequiredGates) == 0 {
		return failClosed("receipt_gate is enabled but lists no required gates; the release gate fails closed"), nil
	}

	var out []Finding
	add := func(rel, msg string) {
		out = append(out, Finding{
			File: rel, Line: 0, RuleID: "receipt_gate", Severity: cfg.Severity, Message: msg,
		})
	}
	for _, gate := range cfg.RequiredGates {
		if !receiptGateRe.MatchString(gate) {
			add(dir, "receipt_gate required gate name '"+gate+"' is not a safe path component; the release gate fails closed")
			continue
		}
		rel := filepath.Join(dir, cfg.Commit, gate+".json")
		data, err := os.ReadFile(filepath.Join(repoRoot, rel))
		if err != nil {
			if os.IsNotExist(err) {
				add(rel, "no '"+gate+"' receipt for commit "+cfg.Commit+"; the semantic gate has not run (fail-closed)")
				continue
			}
			return nil, err
		}
		var r receipt
		if err := json.Unmarshal(data, &r); err != nil {
			add(rel, "'"+gate+"' receipt is malformed JSON: "+err.Error())
			continue
		}
		if r.Subject.Digest.GitCommit != cfg.Commit {
			add(rel, "'"+gate+"' receipt subject '"+r.Subject.Digest.GitCommit+"' does not match the target commit "+cfg.Commit)
			continue
		}
		if r.VerificationResult != "PROMOTE" {
			add(rel, "'"+gate+"' receipt verdict is '"+r.VerificationResult+"', not PROMOTE")
			continue
		}
		// Bind the receipt to the gate it attests: a receipt whose policy.detector
		// names a different gate (or none) does not satisfy this one, even if it is a
		// genuine PROMOTE for the target commit. Without this, one receipt copied
		// across every gate's path would satisfy them all.
		if strings.TrimSpace(r.Policy.Detector) != gate {
			add(rel, "'"+gate+"' receipt attests detector '"+r.Policy.Detector+"', not this gate; a receipt is bound to its detector, not its filename")
			continue
		}
		if strings.TrimSpace(r.JudgeModel) == "" {
			add(rel, "'"+gate+"' receipt pins no judge model; a floating judge is not auditable")
		}
	}
	return out, nil
}

// checkGateLockstep asserts the release-gate runbook's deterministic-gate list
// is in lockstep with the CI workflow's gate steps — neither is a projection of
// the other (the runbook carries prose the workflow can't, the workflow carries
// setup the runbook shouldn't), so the anti-drift shape is a consistency check,
// not generate-from-source. A gate in one but not the other BLOCKS. Both are
// hand-parsed (no YAML dependency): the runbook's numbered "Deterministic gates"
// list, and the named steps of the workflow's target job minus the configured
// setup steps.
func checkGateLockstep(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	if cfg.Runbook == "" || cfg.Workflow == "" {
		return nil, nil
	}
	minGates := cfg.MinGates
	if minGates < 1 {
		minGates = 1 // enabled ⇒ at least a non-empty floor, never a vacuous pass
	}

	var out []Finding
	block := func(file, msg string) {
		out = append(out, Finding{File: file, Line: 0, RuleID: "gate_lockstep", Severity: cfg.Severity, Message: msg})
	}

	// A configured path that does not resolve is drift/misconfig, not "clean" —
	// fail loud rather than parsing an empty list that would pass vacuously.
	runbookExists, err := fsutil.Exists(filepath.Join(repoRoot, cfg.Runbook))
	if err != nil {
		return nil, err
	}
	workflowExists, err := fsutil.Exists(filepath.Join(repoRoot, cfg.Workflow))
	if err != nil {
		return nil, err
	}
	if !runbookExists {
		block(cfg.Runbook, "gate_lockstep runbook '"+cfg.Runbook+"' does not exist; the release gate fails closed")
	}
	if !workflowExists {
		block(cfg.Workflow, "gate_lockstep workflow '"+cfg.Workflow+"' does not exist; the release gate fails closed")
	}
	if !runbookExists || !workflowExists {
		return out, nil
	}

	runbookGates, err := runbookGateList(repoRoot, cfg.Runbook)
	if err != nil {
		return nil, err
	}
	workflowGates, err := workflowStepNames(repoRoot, cfg.Workflow, cfg.Job, cfg.IgnoreSteps)
	if err != nil {
		return nil, err
	}

	// Non-empty floor: an under-count means a heading/job rename, a missed step
	// form, or a mis-scoped job silently dropped gates. This is the fail-closed
	// backstop that turns every such under-count loud instead of a silent pass.
	if len(runbookGates) < minGates {
		block(cfg.Runbook, "gate_lockstep parsed "+strconv.Itoa(len(runbookGates))+" runbook gate(s), below the expected minimum "+strconv.Itoa(minGates)+"; the release gate fails closed")
	}
	if len(workflowGates) < minGates {
		block(cfg.Workflow, "gate_lockstep parsed "+strconv.Itoa(len(workflowGates))+" '"+cfg.Job+"' gate(s) from "+cfg.Workflow+", below the expected minimum "+strconv.Itoa(minGates)+"; the release gate fails closed")
	}

	inWorkflow := make(map[string]bool, len(workflowGates))
	for _, g := range workflowGates {
		inWorkflow[g] = true
	}
	inRunbook := make(map[string]bool, len(runbookGates))
	for _, g := range runbookGates {
		inRunbook[g] = true
	}
	for _, g := range runbookGates {
		if !inWorkflow[g] {
			block(cfg.Runbook, "runbook lists deterministic gate '"+g+"' that is not a '"+cfg.Job+"' step in "+cfg.Workflow)
		}
	}
	for _, g := range workflowGates {
		if !inRunbook[g] {
			block(cfg.Workflow, cfg.Workflow+" '"+cfg.Job+"' step '"+g+"' is missing from the runbook's deterministic gates ("+cfg.Runbook+")")
		}
	}
	return out, nil
}

// runbookGateList extracts the numbered items under the runbook's "Deterministic
// gates" heading (case-insensitive), stopping at the next heading, with the same
// surrounding-quote normalization the workflow side uses so identical gates never
// look different. A missing file yields nil — the caller has already failed it
// closed.
func runbookGateList(repoRoot, rel string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var gates []string
	inSection := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			inSection = strings.Contains(strings.ToLower(trimmed), "deterministic gate")
			continue
		}
		if !inSection {
			continue
		}
		if m := numberedItemRe.FindStringSubmatch(line); m != nil {
			gates = append(gates, strings.Trim(strings.TrimSpace(m[1]), `"'`))
		}
	}
	return gates, nil
}

// workflowStepNames extracts the step names of the named job, dropping configured
// setup steps. It scopes to the `jobs:` block (so 2-space keys under on:/
// permissions:/concurrency: never masquerade as jobs), identifies a job only by a
// strict 2-space header regex (so a comment such as `  # NOTE:` cannot close a
// job early), and captures each step's name wherever it appears in the step block
// — inline `- name:` OR a following `name:` line after `- uses:` — so the
// alternate step form is not invisible. A missing file yields nil; the caller has
// already failed it closed.
func workflowStepNames(repoRoot, rel, job string, ignore []string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	ignored := make(map[string]bool, len(ignore))
	for _, s := range ignore {
		ignored[s] = true
	}

	var names []string
	seenJobs, inJob, inSteps, inStep := false, false, false, false
	var cur string
	// Column where the current step's own mapping keys sit (the marker's content
	// column). A `name:` deeper than this belongs to a nested mapping (e.g.
	// `with: name:`) and is NOT the step's name.
	stepKeyCol := -1
	flush := func() {
		if inStep {
			name := strings.Trim(strings.TrimSpace(cur), `"'`)
			if name != "" && !ignored[name] {
				names = append(names, name)
			}
		}
		cur, inStep = "", false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !seenJobs {
			if jobsSectionRe.MatchString(line) {
				seenJobs = true
			}
			continue
		}
		if m := jobHeaderRe.FindStringSubmatch(line); m != nil {
			flush()
			inJob, inSteps = m[1] == job, false
			continue
		}
		if !inJob {
			continue
		}
		if stepsKeyRe.MatchString(line) {
			inSteps = true
			continue
		}
		if !inSteps {
			continue
		}
		if m := stepMarkerRe.FindStringSubmatch(line); m != nil {
			flush()
			inStep = true
			// The step's keys align at the marker's content column, so a later
			// `name:` at that exact column is the step's name; anything deeper is nested.
			stepKeyCol = strings.Index(line, m[1])
			if content := m[1]; strings.HasPrefix(content, "name:") {
				cur = strings.TrimSpace(strings.TrimPrefix(content, "name:"))
			}
			continue
		}
		if inStep && cur == "" {
			if nm := stepNameKeyRe.FindStringSubmatch(line); nm != nil {
				if indent := len(line) - len(strings.TrimLeft(line, " ")); indent == stepKeyCol {
					cur = strings.TrimSpace(nm[1])
				}
			}
		}
	}
	flush()
	return names, nil
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
func checkBrittleRefs(rel string, lines []string, mask []bool, cfg RuleConfig) []Finding {
	var out []Finding
	for i, line := range lines {
		if mask[i] {
			continue
		}
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
	// superseded_by target can be checked for existence. Track the files each id
	// claims: ids are the intent's identity across the record, and parallel
	// branches each allocating "the next free id" collide silently otherwise.
	known := map[string]bool{}
	idFiles := map[string][]string{}
	_ = filepath.WalkDir(intentsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if intentFileRe.MatchString(d.Name()) {
			id := intentIDRe.FindString(d.Name())
			known[id] = true
			idFiles[id] = append(idFiles[id], path)
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
			out = append(out, validateIntentIDUnique(repoRoot, rel, e.Name(), fields, idFiles, cfg.Severity)...)
		}
	}
	return out, nil
}

// validateIntentIDUnique flags a duplicated intent id, delegating to the shared
// validateIDUnique primitive (the issue-id rule uses the same logic).
func validateIntentIDUnique(repoRoot, rel, name string, fields map[string]fmField, idFiles map[string][]string, severity string) []Finding {
	id := intentIDRe.FindString(name)
	return validateIDUnique(repoRoot, rel, id, "intent", "intent_lifecycle", severity, fields, idFiles)
}

// validateIDUnique flags every file in a colliding id set, not just one: the
// linter cannot know which claimant is authoritative, and flagging a single file
// would imply the others are fine. It is the one primitive behind both the
// intent-id (intent_lifecycle) and issue-id (issue_id_unique) uniqueness rules —
// an id is a record's identity across its register and must be unique within it.
// noun names the record kind in the message; ruleID and severity tag the emitted
// Finding. idFiles maps each id to the repo-absolute paths that claim it.
func validateIDUnique(repoRoot, rel, id, noun, ruleID, severity string, fields map[string]fmField, idFiles map[string][]string) []Finding {
	claimants := idFiles[id]
	if len(claimants) < 2 {
		return nil
	}
	others := make([]string, 0, len(claimants)-1)
	for _, p := range claimants {
		if r := repoRel(repoRoot, p); r != rel {
			others = append(others, r)
		}
	}
	sort.Strings(others)

	line := 1
	if f, ok := fields["id"]; ok && f.line > 0 {
		line = f.line
	}
	return []Finding{{
		File: rel, Line: line, RuleID: ruleID, Severity: severity,
		Message: "duplicate " + noun + " id " + id + "; also claimed by " + strings.Join(others, ", ") +
			" — an id is the " + noun + "'s identity across the record and must be unique",
	}}
}

// checkIssueIDUnique flags any iss-N id claimed by two or more files across the
// issue ledger's status directories (open/, resolved/, wontfix/). The capture
// allocator rejects a duplicate on the reservation path, but a hand-added issue
// file that bypassed it — how a past iss-56 collision arose — slips straight
// through; this is the record-lint backstop that catches it. It is the issue-side
// mirror of the intent-id uniqueness rule and shares the same validateIDUnique
// primitive. The ledger lives outside cfg.Roots, so the rule runs once, not
// per-root. A missing ledger is not an error — it yields no findings.
func checkIssueIDUnique(repoRoot string, cfg RuleConfig) ([]Finding, error) {
	issuesDir := cfg.IssuesDir
	if issuesDir == "" {
		issuesDir = ".abcd/work/issues"
	}
	issuesRoot := filepath.Join(repoRoot, issuesDir)
	if _, err := os.Stat(issuesRoot); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Track every file each iss-N id claims across the three status dirs: an id is
	// the issue's identity across the ledger, and a bypassed-allocator file (or two
	// parallel branches) can land the same id silently otherwise.
	idFiles := map[string][]string{}
	var files []string
	for _, sub := range issueStatusDirs {
		entries, err := os.ReadDir(filepath.Join(issuesRoot, sub))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !issueFileRe.MatchString(e.Name()) {
				continue
			}
			fileAbs := filepath.Join(issuesRoot, sub, e.Name())
			id := issueIDRe.FindString(e.Name())
			idFiles[id] = append(idFiles[id], fileAbs)
			files = append(files, fileAbs)
		}
	}

	var out []Finding
	for _, fileAbs := range files {
		id := issueIDRe.FindString(filepath.Base(fileAbs))
		content, err := os.ReadFile(fileAbs)
		if err != nil {
			return nil, err
		}
		rel := repoRel(repoRoot, fileAbs)
		fields := frontmatterFields(strings.Split(string(content), "\n"))
		out = append(out, validateIDUnique(repoRoot, rel, id, "issue", "issue_id_unique", cfg.Severity, fields, idFiles)...)
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

// checkSpecLifecycle is the spec-side mirror of checkIntentLifecycle (check
// family G). It discovers spec files under specs/{open,closed}/ and validates
// each has a well-formed id/slug/intent link, that the named intent EXISTS in the
// corpus, and — the load-bearing cross-check — that the intent points back at
// this spec (bidirectional agreement). A missing specs/ directory is soft.
func checkSpecLifecycle(repoRoot, rootAbs string, cfg RuleConfig, top Config) ([]Finding, error) {
	specsDir := cfg.SpecsDir
	if specsDir == "" {
		specsDir = "specs"
	}
	specsRoot := filepath.Join(rootAbs, specsDir)
	if _, err := os.Stat(specsRoot); err != nil {
		return nil, nil // missing specs/ is soft, mirroring intent_lifecycle
	}

	// Index the intent corpus: which ids exist, and each intent's spec_id value.
	// This is what lets a spec's link be checked for existence and back-agreement.
	intentsDir := cfg.IntentsDir
	if intentsDir == "" {
		intentsDir = "intents"
	}
	knownIntent := map[string]bool{}
	intentSpecID := map[string]string{}
	_ = filepath.WalkDir(filepath.Join(rootAbs, intentsDir), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !intentFileRe.MatchString(d.Name()) {
			return nil
		}
		content, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		fields := frontmatterFields(strings.Split(string(content), "\n"))
		id := fields["id"].value
		if !intentIDFullRe.MatchString(id) {
			id = intentIDRe.FindString(d.Name())
		}
		if id != "" {
			knownIntent[id] = true
			intentSpecID[id] = fields["spec_id"].value
		}
		return nil
	})

	var out []Finding
	for _, bucket := range []string{"open", "closed"} {
		bucketDir := filepath.Join(specsRoot, bucket)
		entries, err := os.ReadDir(bucketDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !specFileRe.MatchString(e.Name()) {
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
			out = append(out, validateSpec(rel, fields, knownIntent, intentSpecID, cfg.Severity)...)
		}
	}
	return out, nil
}

func validateSpec(rel string, fields map[string]fmField, knownIntent map[string]bool, intentSpecID map[string]string, severity string) []Finding {
	var out []Finding
	add := func(line int, msg string) {
		if line == 0 {
			line = 1
		}
		out = append(out, Finding{
			File: rel, Line: line, RuleID: "spec_lifecycle",
			Severity: severity, Message: msg,
		})
	}

	// Status is the bucket directory (open/closed), never a field.
	if st, ok := fields["status"]; ok {
		add(st.line, "status: key forbidden; spec status is the bucket directory (open/closed), not a field")
	}

	id := fields["id"]
	idValid := specIDFullRe.MatchString(id.value)
	if !idValid {
		add(id.line, "spec id must be present and match ^spc-\\d+$ (got '"+id.value+"')")
	}
	if slug, ok := fields["slug"]; !ok || isNull(slug.value) {
		add(slug.line, "spec slug must be present")
	}

	intent := fields["intent"]
	if !intentIDFullRe.MatchString(intent.value) {
		add(intent.line, "spec intent link must be present and match ^itd-\\d+$ (got '"+intent.value+"')")
		return out // no existence/agreement check possible without a well-formed link
	}
	if !knownIntent[intent.value] {
		add(intent.line, "spec intent '"+intent.value+"' does not exist in any bucket")
		return out
	}
	// Bidirectional agreement: the named intent must carry spec_id == this spec's
	// id. Drift either way (the intent points elsewhere, or at null) is flagged.
	if idValid {
		back := intentSpecID[intent.value]
		if specNum(back) != specNum(id.value) {
			add(intent.line, "bidirectional drift: spec '"+id.value+"' names intent '"+intent.value+"' but that intent's spec_id is '"+back+"'")
		}
	}
	return out
}

// specNum extracts the numeric N from a spec id or spec_id value (which may carry
// a trailing slug, e.g. spc-1-thing), or -1 when there is no spc- id at all. It
// lets the bidirectional check compare spc-1 against a reserved spec_id: spc-1
// written with or without a slug suffix.
func specNum(v string) int {
	if !strings.HasPrefix(v, "spc-") {
		return -1
	}
	rest := v[len("spc-"):]
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return -1
	}
	n, err := strconv.Atoi(rest[:end])
	if err != nil {
		return -1
	}
	return n
}

// fmField is a frontmatter key's value and 1-based source line.
// checkForbiddenSynonyms implements the GL002 forbidden-synonym family. It reads
// the glossary term files under cfg.GlossaryDir (the single source of truth for
// what a forbidden synonym is), then flags live prose that uses an *enforced*
// synonym as a standalone word. Enforcement is scoped to cfg.Enforce because most
// forbidden synonyms are common English words; each enforced entry must be a
// declared forbidden_synonym or the config is rejected (an error), so the config
// can never gate a word the glossary does not forbid.
//
// Matching is case-insensitive with explicit Unicode word boundaries — Go's
// regexp \b is ASCII-only, so a run of unicode letters adjacent to the synonym
// (é, а, …) would otherwise read as a boundary and leak a false hit. Code spans
// (fenced and inline single-backtick), YAML frontmatter, exempt path prefixes, the
// glossary directory itself, and any line matching an allow_context regexp are all
// out of scope: a term file names its own forbidden synonyms legitimately, and a
// code span or external token (`epic-review`) is a mention, not a substitution.
func checkForbiddenSynonyms(repoRoot, rootAbs string, cfg RuleConfig) ([]Finding, error) {
	glossaryDir := cfg.GlossaryDir
	if glossaryDir == "" {
		glossaryDir = ".abcd/development/brief/glossary"
	}
	forbidden, canonical, err := loadForbiddenSynonyms(repoRoot, glossaryDir)
	if err != nil {
		return nil, err
	}

	// Compile one boundary-aware matcher per enforced synonym, rejecting any the
	// glossary does not actually forbid (the glossary is the source of truth).
	type synMatcher struct {
		word string
		re   *regexp.Regexp
	}
	var matchers []synMatcher
	for _, s := range cfg.Enforce {
		key := strings.ToLower(strings.TrimSpace(s))
		if key == "" {
			continue
		}
		if !forbidden[key] {
			return nil, &configError{"forbidden_synonyms: enforced word " + strconv.Quote(s) +
				" is not declared as a forbidden_synonym by any glossary term under " + glossaryDir}
		}
		matchers = append(matchers, synMatcher{word: key, re: regexp.MustCompile("(?i)" + regexp.QuoteMeta(key))})
	}
	if len(matchers) == 0 {
		return nil, nil
	}

	allow := make([]*regexp.Regexp, 0, len(cfg.AllowContext))
	for _, a := range cfg.AllowContext {
		re, err := regexp.Compile(a)
		if err != nil {
			return nil, err
		}
		allow = append(allow, re)
	}

	glossaryPrefix := filepath.ToSlash(glossaryDir)
	files, err := markdownFiles(rootAbs)
	if err != nil {
		return nil, err
	}

	var out []Finding
	for _, fileAbs := range files {
		rel := repoRel(repoRoot, fileAbs)
		relSlash := filepath.ToSlash(rel)
		if strings.HasPrefix(relSlash, glossaryPrefix) || hasAnyPrefix(relSlash, cfg.ExemptPrefixes) {
			continue
		}
		content, err := os.ReadFile(fileAbs)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(content), "\n")
		mask := fenceMask(lines)
		bodyStart := frontmatterBodyStart(lines)
		for i, line := range lines {
			if i < bodyStart || mask[i] {
				continue // YAML frontmatter and fenced code are not prose
			}
			if matchesAny(allow, line) {
				continue
			}
			stripped := stripInlineCode(line)
			for _, m := range matchers {
				for _, loc := range m.re.FindAllStringIndex(stripped, -1) {
					if !wordBoundaryAt(stripped, loc[0], loc[1]) {
						continue
					}
					term := canonical[m.word]
					out = append(out, Finding{
						File: rel, Line: i + 1, RuleID: "GL002", Severity: cfg.Severity,
						Message: "forbidden synonym '" + m.word + "' for glossary term '" + term +
							"' in live prose; use '" + term + "' (itd-43). If this is a mention (not a substitution), quote it in a code span.",
					})
					break // one finding per enforced word per line is enough signal
				}
			}
		}
	}
	return out, nil
}

// configError is a malformed-configuration error (an enforced synonym the glossary
// does not forbid), surfaced through Lint's error return like an uncompilable regexp.
type configError struct{ msg string }

func (e *configError) Error() string { return e.msg }

// loadForbiddenSynonyms walks the glossary directory and returns the set of
// forbidden synonyms (lower-cased) and a map from each synonym to its canonical
// term. A missing glossary directory yields empty maps, not an error.
func loadForbiddenSynonyms(repoRoot, glossaryDir string) (map[string]bool, map[string]string, error) {
	forbidden := map[string]bool{}
	canonical := map[string]string{}
	dirAbs := filepath.Join(repoRoot, glossaryDir)
	files, err := markdownFiles(dirAbs)
	if err != nil {
		return nil, nil, err
	}
	for _, fileAbs := range files {
		content, err := os.ReadFile(fileAbs)
		if err != nil {
			return nil, nil, err
		}
		lines := strings.Split(string(content), "\n")
		// Glossary term files carry a leading attribution comment before the `---`
		// block (the mattpocock template), so slice from the opening delimiter — the
		// shared frontmatter scanner requires `---` on line 0.
		if start := frontmatterOpen(lines); start > 0 {
			lines = lines[start:]
		}
		fields := frontmatterFields(lines)
		term, ok := fields["term"]
		if !ok {
			continue
		}
		syns, ok := fields["forbidden_synonyms"]
		if !ok {
			continue
		}
		for _, s := range parseYAMLStringList(syns.value) {
			key := strings.ToLower(strings.TrimSpace(s))
			if key == "" {
				continue
			}
			forbidden[key] = true
			if _, seen := canonical[key]; !seen {
				canonical[key] = term.value
			}
		}
	}
	return forbidden, canonical, nil
}

// parseYAMLStringList parses an inline YAML flow sequence of strings, e.g.
// `["sprint", "milestone", "epic"]`, into its members. It is deliberately small —
// the glossary frontmatter only ever uses the inline `[...]` form — and tolerates
// quotes and surrounding whitespace.
func parseYAMLStringList(v string) []string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "[")
	v = strings.TrimSuffix(v, "]")
	if strings.TrimSpace(v) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"'`)
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// frontmatterOpen returns the index of the opening `---` frontmatter delimiter,
// skipping leading blank lines and HTML comments; -1 when the leading block is not
// frontmatter. It lets a term file carry an attribution comment above its `---`.
func frontmatterOpen(lines []string) int {
	i := 0
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if t == "" || (strings.HasPrefix(t, "<!--") && strings.HasSuffix(t, "-->")) {
			i++
			continue
		}
		break
	}
	if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
		return i
	}
	return -1
}

// frontmatterBodyStart returns the index of the first body line after a leading
// YAML frontmatter block (the line after the closing `---`); 0 when there is none.
// It shares frontmatterOpen's tolerance of a leading attribution comment, so a
// file whose frontmatter carries a `core/epic` term reference is never scanned as
// prose just because a comment precedes its `---`.
func frontmatterBodyStart(lines []string) int {
	open := frontmatterOpen(lines)
	if open < 0 {
		return 0
	}
	for j := open + 1; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) == "---" {
			return j + 1
		}
	}
	return 0 // unterminated frontmatter: treat all as body rather than swallow the file
}

// stripInlineCode blanks the contents of single-backtick inline code spans (and
// their delimiters) so a forbidden synonym named inside a code span is a mention,
// not a match. Unpaired backticks leave the remainder untouched.
func stripInlineCode(line string) string {
	b := []rune(line)
	out := make([]rune, len(b))
	copy(out, b)
	inSpan := false
	for i, r := range b {
		if r == '`' {
			out[i] = ' '
			inSpan = !inSpan
			continue
		}
		if inSpan {
			out[i] = ' '
		}
	}
	if inSpan {
		// Unpaired backtick: nothing was a real span — restore the tail.
		return line
	}
	return string(out)
}

// wordBoundaryAt reports whether [start,end) in s is bounded by non-word runes on
// both sides. Word runes are Unicode letters, digits, and underscore — explicit
// because Go's regexp \b is ASCII-only and would treat a unicode-letter neighbour
// as a boundary, leaking "épic"/"epicа" style false hits past an ASCII \b.
func wordBoundaryAt(s string, start, end int) bool {
	if start > 0 {
		r, _ := utf8.DecodeLastRuneInString(s[:start])
		if isWordRune(r) {
			return false
		}
	}
	if end < len(s) {
		r, _ := utf8.DecodeRuneInString(s[end:])
		if isWordRune(r) {
			return false
		}
	}
	return true
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(s, filepath.ToSlash(p)) {
			return true
		}
	}
	return false
}

type fmField struct {
	value string
	line  int
}

// frontmatterFields returns the top-level keys of the leading YAML frontmatter
// (the block between the first two `---` lines) in lint's local fmField shape. It
// adapts the canonical scanner in internal/core/frontmatter, so the frontmatter
// line-scanner (and its delimiter handling) lives in ONE place, not a divergent
// copy here.
func frontmatterFields(lines []string) map[string]fmField {
	fields := map[string]fmField{}
	for key, f := range frontmatter.Fields(lines) {
		fields[key] = fmField{value: f.Value, line: f.Line}
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
