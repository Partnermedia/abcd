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
	intentIDRe    = regexp.MustCompile(`itd-\d+`)
	intentFileRe  = regexp.MustCompile(`^itd-\d+.*\.md$`)
	specIDRe      = regexp.MustCompile(`^spc-`)
	supersededRe  = regexp.MustCompile(`^itd-\d+`)
	intentBuckets = map[string]bool{
		"drafts": true, "planned": true, "shipped": true,
		"disciplines": true, "superseded": true,
	}
)

// Lint runs every enabled check family against the record under repoRoot and
// returns the findings sorted deterministically. An error is returned only for
// malformed configuration (e.g. an uncompilable regexp); a walkable-but-missing
// root is skipped, not an error.
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
